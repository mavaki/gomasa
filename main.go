package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sync"
	"io/ioutil"
	"strings"

	gomail "gopkg.in/mail.v2"
)

// global variables
var (
	title     string
	body      string
	recipient string
	lock      sync.Mutex
)

func getPassword(filepath string) (string, error) {
	// read from file
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return "", err
	}
	
	// parse the line (format: "KEY=value")
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "EMAIL_PASSWORD=") {
			return strings.TrimPrefix(line, "EMAIL_PASSWORD="), nil
		}
	}

	return "", nil
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	// redirect to edit if body is empty
	if body == "" {
		http.Redirect(w, r, "/edit/", http.StatusFound)
		return
	}
	
	// show the current content in view
	renderTemplate(w, "view", title, body, recipient)
}

func editHandler(w http.ResponseWriter, r *http.Request) {
	// show the current content in edit
	renderTemplate(w, "edit", title, body, recipient)
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
	// get values for title, body, and recipient
	lock.Lock() // lock to prevent race
	title = r.FormValue("title")
	body = r.FormValue("body")
	recipient = r.FormValue("recipient")
	lock.Unlock() // unlock after getting values

	// send email in goroutine
	go func() {
		if err := Deliver(body, recipient, title); err != nil {
			fmt.Printf("Failed to send email to %s: %v\n", recipient, err)
		}
	}()

	// redirect to the view page
	http.Redirect(w, r, "/view/", http.StatusFound)
}

var templates = template.Must(template.ParseFiles("edit.html", "view.html"))

func renderTemplate(w http.ResponseWriter, tmpl string, title string, body string, recipient string) {
	// use map to hold title, body, and recipient
	data := map[string]string{
		"Title":     title,
		"Body":      body,
		"Recipient": recipient,
	}
	
	// render the template
	err := templates.ExecuteTemplate(w, tmpl+".html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	// define url pages
	http.HandleFunc("/view/", viewHandler)
	http.HandleFunc("/edit/", editHandler)
	http.HandleFunc("/save/", saveHandler)

	// start server
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func Deliver(body string, recipient_email string, subject string) error {
	bot_email := "gomasamailbot@gmail.com"

	// get password from secret file
	password, err := getPassword("/etc/secrets/secrets.env")
	if err != nil {
		log.Fatalf("Failed to read password from secret file: %v", err)
	}

	// create new message
	message := gomail.NewMessage()

	// set headers
	message.SetHeader("From", bot_email)
	message.SetHeader("To", recipient_email)
	message.SetHeader("Subject", subject)

	// set body
	message.SetBody("text/plain", body)

	// set up dialer
	dialer := gomail.NewDialer("smtp.gmail.com", 587, bot_email, password)

	// send email
	if err := dialer.DialAndSend(message); err != nil {
		return err
	}

	// print success message
	fmt.Println("Email sent successfully to:", recipient_email)
	return nil
}

