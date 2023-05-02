package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"strings"
	"time"
)

// declare login variables
var loggedIn bool

type GraphQLRequest struct {
	Query string `json:"query"`
}

// Create a new struct to hold the login request data
type loginRequest struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}
type Response struct {
	Data Data `json:"data"`
}

type Data struct {
	Users []User `json:"user"`
}

type User struct {
	ID           int           `json:"id"`
	Login        string        `json:"login"`
	AuditRatio   float64       `json:"auditRatio"`
	Campus       string        `json:"campus"`
	Transactions []Transaction `json:"transactions"`
}

type Attribute struct {
	AuditId int `json:"auditId"`
}

type Transaction struct {
	Path      string    `json:"path"`
	CreatedAt time.Time `json:"createdAt"`
	Amount    float64   `json:"amount"`
	Type      string    `json:"type"`
	Attrs     Attribute `json:"attrs"`
}

// create a struct for transactions which contain "/gritlab/school-curriculum/" in the path
type SchoolTransaction struct {
	Path      string    `json:"path"`
	CreatedAt time.Time `json:"createdAt"`
	Amount    float64   `json:"amount"`
	Type      string    `json:"type"`
	Attrs     Attribute `json:"attrs"`
}

// declare a global variable to store the highestAmounts which is a map
var highestAmounts = make(map[string]float64)

// declare a global variable to store user id, login, auditRatio, campus
var userdata User

// declare a global varible to store schoolTransactions
var xpTransactions []SchoolTransaction

var res1 Response
var res2 Response

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	//handle /login
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/api/execute-query", DataReceiver)
	//handle /dashboard
	http.HandleFunc("/dashboard", dashboardHandler)
	//handle /logout
	http.HandleFunc("/logout", logoutHandler)
	port := "8080"
	fmt.Printf("Starting server at http://localhost:%s\n", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
	}
}

func DataReceiver(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Decode the request body
	var requestBody struct {
		Token       string   `json:"token"`
		Response    Response `json:"response"`
		ResponseTwo Response `json:"responsetwo"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Store the received data in global variables res1 and res2
	res1 = requestBody.Response
	res2 = requestBody.ResponseTwo

	loggedIn = true
	manipulateData()

	fmt.Fprint(w, "OK")
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		http.ServeFile(w, r, "login.html")
		return
	}
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	loggedIn = false
	http.SetCookie(w, &http.Cookie{
		Name:   "jwt_token",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	//remove the data
	xpTransactions = []SchoolTransaction{}
	highestAmounts = map[string]float64{}
	userdata = User{
		ID:           0,
		Login:        "",
		AuditRatio:   0,
		Campus:       "",
		Transactions: []Transaction{},
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	// Check if the user is logged in
	if !loggedIn {
		// Redirect to the login page or display an error message
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if r.Method == "GET" {

		XpTransactionsJSON, err := json.Marshal(xpTransactions)
		if err != nil {
			fmt.Println(err)
		}

		HighestAmountsJSON, err := json.Marshal(highestAmounts)
		if err != nil {
			fmt.Println(err)
		}

		UserJSON, err := json.Marshal(userdata)
		if err != nil {
			fmt.Println(err)
		}

		data := struct {
			XPTransactionsJSON template.JS
			HighestAmountsJSON template.JS
			UserJSON           template.JS
		}{
			XPTransactionsJSON: template.JS(strings.TrimSpace(string(XpTransactionsJSON))),
			HighestAmountsJSON: template.JS(strings.TrimSpace(string(HighestAmountsJSON))),
			UserJSON:           template.JS(strings.TrimSpace(string(UserJSON))),
		}

		// Parse and execute the dashboard.html template
		tmpl, err := template.ParseFiles("dashboard.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		err = tmpl.Execute(w, data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

}

func manipulateData() {

	var response Response
	response = res1
	var responsetwo Response
	responsetwo = res2

	fmt.Println("Your very first submission was", responsetwo.Data.Users[0].Transactions[0].Path, "on", responsetwo.Data.Users[0].Transactions[0].CreatedAt, "aww, how cute! Look at you now!")

	// Access the data using the Response struct fields
	if len(response.Data.Users) > 0 {
		user := response.Data.Users[0]
		userdata = User{
			ID:         user.ID,
			Login:      user.Login,
			AuditRatio: user.AuditRatio,
			Campus:     user.Campus,
		}

		//create a struct for transactions which have "skill" in the type
		type SkillTransaction struct {
			Path      string    `json:"path"`
			CreatedAt time.Time `json:"createdAt"`
			Amount    float64   `json:"amount"`
			Type      string    `json:"type"`
		}
		//extract the transactions with "skill" in the path
		var skillTransactions []SkillTransaction
		// Iterate through the transactions
		for _, transaction := range user.Transactions {
			// Check if the transaction path contains the word "skill"
			if strings.Contains(transaction.Type, "skill") {
				// Append the transaction to the skillTransactions slice
				skillTransactions = append(skillTransactions, SkillTransaction{
					Path:      transaction.Path,
					CreatedAt: transaction.CreatedAt,
					Amount:    transaction.Amount,
					Type:      transaction.Type,
				})
			}
		}
		//range over the skillTransactions, find out the highest amount for each type and save only the highest amounts in highestAmounts
		for _, skillTransaction := range skillTransactions {
			//check if the type is already in the map
			if _, ok := highestAmounts[skillTransaction.Type]; ok {
				//if it is, check if the current amount is higher than the one in the map
				if skillTransaction.Amount > highestAmounts[skillTransaction.Type] {
					//if it is, replace the value in the map with the current amount
					highestAmounts[skillTransaction.Type] = skillTransaction.Amount
				}
			} else {
				//if it is not, add the type and amount to the map
				highestAmounts[skillTransaction.Type] = skillTransaction.Amount
			}
		}

		//create structs for transactions of type "up" and "down"
		// UP = I did an audit
		// DOWN = I got audited
		type UpTransaction struct {
			Path      string    `json:"path"`
			CreatedAt time.Time `json:"createdAt"`
			Amount    float64   `json:"amount"`
			Type      string    `json:"type"`
		}

		type DownTransaction struct {
			Path      string    `json:"path"`
			CreatedAt time.Time `json:"createdAt"`
			Amount    float64   `json:"amount"`
			Type      string    `json:"type"`
		}

		//extract the transactions with "up" in the type, type has to be exact as "up" could be in other types as well
		var upTransactions []UpTransaction
		// Iterate through the transactions
		for _, transaction := range user.Transactions {
			// Check if the transaction type equals "up"
			if transaction.Type == "up" {
				// Append the transaction to the skillTransactions slice
				upTransactions = append(upTransactions, UpTransaction{
					Path:      transaction.Path,
					CreatedAt: transaction.CreatedAt,
					Amount:    transaction.Amount,
					Type:      transaction.Type,
				})
			}
		}

		//extract the transactions with "down" in the type, type has to be exact as "down" could be in other types as well
		var downTransactions []DownTransaction
		// Iterate through the transactions
		for _, transaction := range user.Transactions {
			// Check if the transaction type equals "down"
			if transaction.Type == "down" {
				// Append the transaction to the skillTransactions slice
				downTransactions = append(downTransactions, DownTransaction{
					Path:      transaction.Path,
					CreatedAt: transaction.CreatedAt,
					Amount:    transaction.Amount,
					Type:      transaction.Type,
				})
			}
		}

		//extract the transactions with "/gritlab/school-curriculum/" in the path
		var schoolTransactions []SchoolTransaction
		// Iterate through the transactions
		for _, transaction := range user.Transactions {
			// Check if the transaction path contains the word "/gritlab/school-curriculum/" && type is "xp" && path doesn't contain "checkpoint" or "piscine" && auditID is < 1
			if strings.Contains(transaction.Path, "/gritlab/school-curriculum/") && transaction.Type == "xp" && !strings.Contains(transaction.Path, "piscine") && int(transaction.Attrs.AuditId) < 1 {
				// Append the transaction to the skillTransactions slice
				schoolTransactions = append(schoolTransactions, SchoolTransaction{
					Path:      transaction.Path,
					CreatedAt: transaction.CreatedAt,
					Amount:    transaction.Amount,
					Type:      transaction.Type,
					Attrs:     transaction.Attrs,
				})
			}
			//if type is "xp" && path contains "piscine" && amount is 70000 add the transaction to the schoolTransactions slice
			if transaction.Type == "xp" && strings.Contains(transaction.Path, "piscine") && transaction.Amount == 70000 {
				schoolTransactions = append(schoolTransactions, SchoolTransaction{
					Path:      transaction.Path,
					CreatedAt: transaction.CreatedAt,
					Amount:    transaction.Amount,
					Type:      transaction.Type,
				})
			}
		}
		//sort the schoolTransactions in ascending order by createdAt
		sort.Slice(schoolTransactions, func(i, j int) bool {
			return schoolTransactions[i].CreatedAt.Before(schoolTransactions[j].CreatedAt)
		})

		//calculate sum of all school transactions
		var sumSchoolTransactions float64
		for _, schoolTransaction := range schoolTransactions {
			sumSchoolTransactions += schoolTransaction.Amount
		}
		//populate xpTransactions with SchoolTransactions
		xpTransactions = append(xpTransactions, schoolTransactions...)
	}
}
