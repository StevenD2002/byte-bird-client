package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/manifoldco/promptui"
)

const baseURL = "http://localhost:8080"

type User struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Post struct {
	AuthorName string `json:"authorName"`
	Content    string `json:"content"`
}

func main() {
	// auth token that we use with every request
	token := ""
	for {
		action := promptForAction()

		switch strings.ToLower(action) {
		case "register":
			registerUser(&token)
		case "login":
			loginUser(&token)
		case "post":
			checkForToken(&token)
			makePost(&token)
		case "get_posts":
			checkForToken(&token)
			getPosts(&token)
		case "exit":
			fmt.Println("Exiting...")
			os.Exit(0)
		default:
			fmt.Println("Invalid action. Supported actions: register, login, post, get_posts, exit")
		}
	}
}

func checkForToken(token *string) {
	if *token == "" {
		fmt.Println("You must be logged in to perform this action")
		os.Exit(1)
	}
}

func promptForAction() string {
	prompt := promptui.Select{
		Label: "Select Action",
		Items: []string{"register", "login", "post", "get_posts", "exit"},
	}

	_, result, err := prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		os.Exit(1)
	}

	return result
}

func promptForInput(label string) string {
	prompt := promptui.Prompt{
		Label: label,
	}

	result, err := prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		os.Exit(1)
	}

	return result
}

func registerUser(token *string) {
	// Create a User struct with the registration data
	user := User{
		Name:     promptForInput("Enter your name:"),
		Email:    promptForInput("Enter your email:"),
		Password: promptForInput("Enter your password:"),
	}

	// Convert the User struct to JSON
	jsonData, err := json.Marshal(user)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return
	}

	// Make a POST request to the registration endpoint
	url := fmt.Sprintf("%s/register", baseURL)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error making POST request:", err)
		return
	}

	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusCreated {
		fmt.Printf("Registration failed. Status code: %d\n", resp.StatusCode)
		return
	}

	// parse the token from the json response
	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		fmt.Println("Error parsing JSON:", err)
		return
	}

	// check if the token is present in the response
	if tokenValue, ok := data["token"].(string); ok {
		*token = tokenValue
	} else {
		fmt.Println("Token not present in response")
	}

	fmt.Println("Registration successful!")
}

func loginUser(token *string) {
	email := promptForInput("Enter email address:")
	password := promptForInput("Enter password:")

	// Implement the logic for logging in
	fmt.Printf("Logging in with username: %s, password: %s\n", email, password)

	user := User{
		Email:    email,
		Password: password,
	}

	jsonData, err := json.Marshal(user)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return
	}

	url := fmt.Sprintf("%s/login", baseURL)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error making POST request:", err)
		return
	}
	defer resp.Body.Close()

	// check the response status code
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Login failed. Status code: %d\n", resp.StatusCode)
		return
	}

	// parse the token from the json response
	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		fmt.Println("Error parsing JSON:", err)
		return
	}

	// check if the token is present in the response
	if tokenValue, ok := data["token"].(string); ok {
		*token = tokenValue
	} else {
		fmt.Println("Token not present in response")
	}

	fmt.Println("Login successful!")
}

func makePost(token *string) {
	// Implement the logic for making a post
	// Create a Post struct with the post data
	post := Post{
		Content: promptForInput("Enter your post content:"),
	}


	// Convert the Post struct to JSON
	jsonData, err := json.Marshal(post)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return
	}

	// Make a POST request to the posts endpoint with authentication header
	url := fmt.Sprintf("%s/createPost", baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error creating HTTP request:", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+*token)

	// Use the http.DefaultClient to make the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error making POST request:", err)
		return
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusCreated {
		fmt.Printf("Post creation failed. Status code: %d\n", resp.StatusCode)
		return
	}

	fmt.Println("Post creation successful!")

	// Your post creation logic here
}

// Update the getPosts function
func getPosts(token *string) {
	// Make a GET request to the posts endpoint with authentication header
	url := fmt.Sprintf("%s/posts", baseURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating HTTP request:", err)
		return
	}

	req.Header.Set("Authorization", "Bearer "+*token)

	// Use the http.DefaultClient to make the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error making GET request:", err)
		return
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Failed to fetch posts. Status code: %d\n", resp.StatusCode)
		return
	}

	// Read the response body into a byte slice
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}


	// Manually trim any extra characters or line breaks from the raw response
	trimmedResponse := strings.TrimSpace(string(body))

	// Parse the trimmed response into a struct
	var response struct {
		Posts  []Post `json:"posts"`
		Status string `json:"status"`
	}

	if err := json.Unmarshal([]byte(trimmedResponse), &response); err != nil {
		fmt.Println("Error parsing JSON:", err)
		return
	}

	// Display posts in the terminal
	fmt.Println("Posts:")
	for _, p := range response.Posts {
		fmt.Printf("[%s] : %s\n", p.AuthorName, p.Content)
		fmt.Println("-------------")
	}
}
