package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

type GraphQLRequest struct {
	Query string `json:"query"`
}

func getJWTToken(authType, identifier, password string) (string, error) {
	signinURL := "https://01.gritlab.ax/api/auth/signin"

	// Prepare Basic authentication header
	authValue := fmt.Sprintf("%s:%s", identifier, password)
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(authValue))
	authHeader := fmt.Sprintf("Basic %s", encodedAuth)

	// Create POST request
	req, err := http.NewRequest("POST", signinURL, nil)
	if err != nil {
		return "", err
	}

	// Set headers
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Check if the response is successful
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to obtain JWT token: %s", string(bodyBytes))
	}

	// Read JWT token from the response
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	//trim the token and remove the double quotes
	bodyBytes = bodyBytes[1 : len(bodyBytes)-1]
	fmt.Println(string(bodyBytes))
	return strings.TrimSpace(string(bodyBytes)), nil

}

func queryGraphQL(jwtToken, query string) (*Response, error) {
	graphqlURL := "https://01.gritlab.ax/api/graphql-engine/v1/graphql"

	requestBody, err := json.Marshal(GraphQLRequest{Query: query})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", graphqlURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwtToken))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to query GraphQL: %s", string(bodyBytes))
	}

	var result Response
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})
	//handle /login
	http.HandleFunc("/login", loginHandler)
	port := "8080"
	fmt.Printf("Starting server at http://localhost:%s\n", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
	}

}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	// Serve the login HTML page for GET requests
	if r.Method == "GET" {
		http.ServeFile(w, r, "login.html")
		return
	}

	// Process the login form for POST requests
	if r.Method == "POST" {
		// Parse the form data from the request
		err := r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Get the identifier and password from the form data
		identifier := r.Form.Get("identifier")
		password := r.Form.Get("password")
		// fmt.Println("identifier:", identifier)
		// fmt.Println("password:", password)
		// Call the getJWTToken function to get the JWT token
		token, err := getJWTToken("email", identifier, password)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		// Set the JWT token as a cookie and redirect to the dashboard page
		http.SetCookie(w, &http.Cookie{
			Name:  "jwt_token",
			Value: token,
			Path:  "/",
		})
		queryWithJWTToken(token)
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	} else {
		// If the request method is not GET or POST, return a 405 Method Not Allowed error
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
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

type Transaction struct {
	Path      string    `json:"path"`
	CreatedAt time.Time `json:"createdAt"`
	Amount    float64   `json:"amount"`
	Type      string    `json:"type"`
}

func queryWithJWTToken(jwtToken string) {
	// Example usage with username:password

	// Example usage with email:password
	// jwtToken, err = getJWTToken("email", "your_email@example.com", "your_password")
	// if err != nil {
	// 	fmt.Println("Error:", err)
	// } else {
	// 	fmt.Println("JWT Token (email):", jwtToken)
	// }

	query := `
	{
		user {
		  id
		  login
		  auditRatio 
		  campus
	  
		  transactions {
			path 
			createdAt
			amount
			type
		  }
		}
	  }
	`
	response, err := queryGraphQL(jwtToken, query)
	if err != nil {
		log.Fatal(err)
	}

	// Access the data using the Response struct fields
	if len(response.Data.Users) > 0 {
		user := response.Data.Users[0]
		fmt.Println("User ID:", user.ID)
		fmt.Println("User Login:", user.Login)
		fmt.Println("User Audit Ratio:", user.AuditRatio)
		fmt.Println("User Campus:", user.Campus)
		//	fmt.Println("User Transactions:", user.Transactions)
		fmt.Println("")
		// ... and so on

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
				//print current transaction
				skillTransactions = append(skillTransactions, SkillTransaction{
					Path:      transaction.Path,
					CreatedAt: transaction.CreatedAt,
					Amount:    transaction.Amount,
					Type:      transaction.Type,
				})
			}
		}
		//range over the skillTransactions, find out the highest amount for each type and save only the highest amounts in a new slice
		//create a map of type string and float64
		var highestAmounts = make(map[string]float64)
		//iterate over the skillTransactions
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
		//print the map
		fmt.Println("Highest Amounts: ", highestAmounts)
		fmt.Println("")

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
		//fmt.Println("Up Transactions: ", upTransactions)
		fmt.Println("")

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
		//fmt.Println("Down Transactions: ", downTransactions)

		//create a struct for transactions which contain "/gritlab/school-curriculum/" in the path
		type SchoolTransaction struct {
			Path      string    `json:"path"`
			CreatedAt time.Time `json:"createdAt"`
			Amount    float64   `json:"amount"`
			Type      string    `json:"type"`
		}
		//extract the transactions with "/gritlab/school-curriculum/" in the path
		var schoolTransactions []SchoolTransaction
		// Iterate through the transactions
		for _, transaction := range user.Transactions {
			// Check if the transaction path contains the word "/gritlab/school-curriculum/" && type is "xp" && path doesn't contain "checkpoint" or "piscine"
			if strings.Contains(transaction.Path, "/gritlab/school-curriculum/") && transaction.Type == "xp" && !strings.Contains(transaction.Path, "checkpoint") && !strings.Contains(transaction.Path, "piscine") {
				// Append the transaction to the skillTransactions slice
				schoolTransactions = append(schoolTransactions, SchoolTransaction{
					Path:      transaction.Path,
					CreatedAt: transaction.CreatedAt,
					Amount:    transaction.Amount,
					Type:      transaction.Type,
				})
			}
		}
		fmt.Println("School Transactions: ", schoolTransactions)
		fmt.Println("")
		//calculate sum of all school transactions
		var sumSchoolTransactions float64
		for _, schoolTransaction := range schoolTransactions {
			sumSchoolTransactions += schoolTransaction.Amount
		}
		fmt.Println("Sum of School Transactions: ", sumSchoolTransactions)
		fmt.Println("")
	}
}
