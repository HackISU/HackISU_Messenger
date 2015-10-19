package main

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"github.com/Januzellij/gotwilio"
	"github.com/gorilla/mux"
	"html/template"
	"net/http"
	"os"
	"strings"
)

// Your Twilio AccountSID and AuthToken
var ACCOUNTSID = ""
var AUTHTOKEN = ""

// Your Twilio phonenumber
var FROM = ""

// Creates TwilioClient to be used throughout program
var TWILIO = gotwilio.NewTwilioClient(ACCOUNTSID, AUTHTOKEN)

// Username and password for website
var USER = ""
var PASSWORD = ""

// TODO: Replace basic auth with bcrypt and gorilla sessions
// Basic user authentication
func BasicAuth(w http.ResponseWriter, r *http.Request) bool {
	s := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
	if len(s) != 2 {
		return false
	}

	b, err := base64.StdEncoding.DecodeString(s[1])
	if err != nil {
		return false
	}

	pair := strings.SplitN(string(b), ":", 2)
	if len(pair) != 2 {
		return false
	}

	return pair[0] == USER && pair[1] == PASSWORD
}

// Reads phone numbers from given filename and stores each number
// in returned array.
// File format should be one number per line, 10 digit numbers.
// e.g. 1234567890
func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// Website handler
// Tests for user authentication, and if user is authenticated serves request.
// If GET request, serves home.html to user
// If POST request, sends message to all numbers in file "numbers.txt"
func homeHandler(w http.ResponseWriter, r *http.Request) {
	// Test if user is authenticated
	if !BasicAuth(w, r) {
		w.Header().Set("WWW-Authenticate", `Basic realm="You must login to access this page. "`)
		w.WriteHeader(401)
		w.Write([]byte("401 Unauthorized\n"))
		return
	}
	if r.Method == "GET" {
		// Serve user "views/home.html"
		t, err := template.ParseFiles("views/home.html")
		if err != nil {
			fmt.Fprintln(w, err)
		}
		t.Execute(w, nil)
	} else {
		// Read all numbers from "numbers.txt" and store in array TO
		TO, err := readLines("numbers.txt")
		if err != nil {
			fmt.Println(err)
		}

		// Get message entered in HTML form
		message := r.FormValue("message")

		// Send message to all numbers
		for _, number := range TO {
			_, exception, err := TWILIO.SendSMS(FROM, "+1"+number, message, "", "")
			if exception != nil {
				fmt.Fprintln(w, *exception)
			}
			if err != nil {
				fmt.Fprintln(w, err)
			}
		}
	}
}

func main() {
	standardRouter := mux.NewRouter()
	standardRouter.HandleFunc("/", homeHandler)

	http.Handle("/", standardRouter)

	// Listen on port 8081
	http.ListenAndServe(":8081", nil)
}
