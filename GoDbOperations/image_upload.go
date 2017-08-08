package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

// Global sql.DB to access the database by all handlers
var db *sql.DB
var err error

type user struct {
	Username string
	Name     string
	Surname  string
	Gender   string
	Email    string
	Photo    string
}

func homePage(res http.ResponseWriter, req *http.Request) {
	http.ServeFile(res, req, "index.html")
}

func main() {

	// Create an sql.DB and check for errors
	db, err = sql.Open("mysql", "root:toor@/db")
	if err != nil {
		panic(err.Error())
	}
	// sql.DB should be long lived "defer" closes it once this function ends
	defer db.Close()

	// Test the connection to the database
	err = db.Ping()
	if err != nil {
		panic(err.Error())
	}
	//http.HandleFunc("/user", userPage)
	http.HandleFunc("/login", login)
	http.HandleFunc("/signup", singupPage)
	http.HandleFunc("/", homePage)
	fs := http.FileServer(http.Dir("public"))
	http.Handle("/public/", http.StripPrefix("/public/", fs))
	http.ListenAndServe(":8080", nil)
}

func login(res http.ResponseWriter, req *http.Request) {
	var u user
	// If method is GET serve an html login page
	if req.Method != "POST" {
		http.ServeFile(res, req, "login.html")
		return
	}

	// Grab the username/password from the submitted post form
	username := req.FormValue("username")
	password := req.FormValue("password")

	// Grab from the database
	var databaseUsername string
	var databasePassword string

	// Search the database for the username provided
	// If it exists grab the password for validation
	err := db.QueryRow("SELECT username, password FROM USER WHERE username=?", username).Scan(&databaseUsername, &databasePassword)
	// If not then redirect to the login page
	if err != nil {
		http.Redirect(res, req, "/login", 301)
		return
	}

	// Validate the password
	err = bcrypt.CompareHashAndPassword([]byte(databasePassword), []byte(password))
	// If wrong password redirect to the login
	if err != nil {
		http.Redirect(res, req, "/login", 301)
		return
	}

	e := db.QueryRow("SELECT username,name,surname,photo FROM USER WHERE username=?", username).Scan(
		&u.Username, &u.Name, &u.Surname, &u.Photo)
	if e != nil {
		fmt.Println("Query Error")
	}
	// If the login succeeded
	//res.Write([]byte("Hello " + databaseUsername))

	template, err := template.ParseFiles("user.html")
	if err != nil {
		fmt.Println("template acilamadi")
	}
	er := template.Execute(res, u)
	if er != nil {
		fmt.Println(er, "hata")
	}

}

func singupPage(res http.ResponseWriter, req *http.Request) {
	var u user
	req.ParseMultipartForm(0)
	// Serve signup.html to get requests to /signup
	if req.Method != "POST" {
		http.ServeFile(res, req, "signup.html")
		return
	}
	name := req.FormValue("name")
	surname := req.FormValue("surname")
	username := req.FormValue("username")
	password := req.FormValue("password")

	var user string

	err := db.QueryRow("SELECT username FROM USER WHERE username=?", username).Scan(&user)

	switch {
	// Username is available
	case err == sql.ErrNoRows:
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

		if err != nil {
			http.Error(res, "Server error, unable to create your account.", 500)
			return
		}
		//////////////////////////////////////

		file, handler, errorr := req.FormFile("photo") // _ : handler
		if errorr != nil {
			fmt.Println(err)
			return
		}
		//fmt.Println(handler.Header)
		defer file.Close()

		f, err1 := os.OpenFile("public/image/photos/"+username+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
		if err1 != nil {
			fmt.Println(err1)
			return
		}
		photo := "public/image/photos/" + username + handler.Filename

		io.Copy(f, file)
		defer f.Close()

		//////////////////////////////////////

		_, err = db.Exec("INSERT INTO USER(username, name, surname, password, photo) VALUES(?, ?, ?, ?, ?)", username, name, surname, hashedPassword, photo)

		if err != nil {
			http.Error(res, "Server error, unable to create your account.", 500)
			return
		}

		e := db.QueryRow("SELECT username,name,surname,photo FROM USER WHERE username=?", username).Scan(
			&u.Username, &u.Name, &u.Surname, &u.Photo)
		if e != nil {
			fmt.Println("Query Error")
		}

		template, err := template.ParseFiles("user.html")
		if err != nil {
			fmt.Println("template acilamadi")
		}
		template.Execute(res, u)
		//res.Write([]byte("User created!"))

	case err != nil:
		http.Error(res, "Server error, unable to create your account.", 500)
		return
	default:
		http.Redirect(res, req, "/", 301)
	}
}

func userPage(res http.ResponseWriter, req *http.Request) {

}
