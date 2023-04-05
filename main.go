package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
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

func queryGraphQL(jwtToken, query string) (map[string]interface{}, error) {
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

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	return result, nil
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
		fmt.Println("identifier:", identifier)
		fmt.Println("password:", password)
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
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	} else {
		// If the request method is not GET or POST, return a 405 Method Not Allowed error
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func queryWithJWTToken() {
	// Example usage with username:password
	jwtToken, err := getJWTToken("username", "villepat", "!Lsdh85m9022gri")
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("JWT Token (username):", jwtToken)
	}

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
	    profile
	    attrs
	    auditRatio
	    audits {
	      id
	    }
	  }
	}
	`

	result, err := queryGraphQL(jwtToken, query)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Query result:")
		prettyJSON, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(prettyJSON))
	}
}
