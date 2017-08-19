package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Global sql.DB to access the database by all handlers
var db *sql.DB
var err error
var dbSession = map[string]string{} // for cookie

type user struct {
	Username    string
	Given_name  string
	Family_name string
	Gender      string
	Email       string
	Picture     string
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
	http.HandleFunc("/GoogleLogin", handleGoogleLogin)
	http.HandleFunc("/Callback", handleGoogleCallback)
	http.HandleFunc("/user", userPage)
	http.HandleFunc("/User", GuserPage)
	http.HandleFunc("/login", login)
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/signup", singupPage)
	http.HandleFunc("/", homePage)
	fs := http.FileServer(http.Dir("public"))
	http.Handle("/public/", http.StripPrefix("/public/", fs))
	http.ListenAndServe(":8080", nil)
}

// randomString returns a random string with the specified length
func randomString(length int) (str string) {
	b := make([]byte, length)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

var u user

////////////OAuth Google API///////////
var (
	googleOauthConfig = &oauth2.Config{
		RedirectURL:  "",
		ClientID:     "", /////////////
		ClientSecret: "",                                                  //////////////////////
		Scopes: []string{"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/userinfo.email"},
		Endpoint: google.Endpoint,
	}
	// Some random string, random for each request
	oauthStateString = randomString(32)
)

func handleGoogleLogin(res http.ResponseWriter, req *http.Request) {
	//fmt.Println("handleGoogleLogin")
	url := googleOauthConfig.AuthCodeURL(oauthStateString)
	http.Redirect(res, req, url, http.StatusTemporaryRedirect)
}

func handleGoogleCallback(res http.ResponseWriter, req *http.Request) {
	//fmt.Println("handleGoogleCallback")
	state := req.FormValue("state")
	if state != oauthStateString {
		fmt.Printf("invalid oauth state, expected '%s', got '%s'\n", oauthStateString, state)
		http.Redirect(res, req, "/", http.StatusTemporaryRedirect)
		return
	}

	code := req.FormValue("code")
	token, err := googleOauthConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		fmt.Println("Code exchange failed with '%s'\n", err)
		http.Redirect(res, req, "/", http.StatusTemporaryRedirect)
		return
	}

	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)

	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)

	json.Unmarshal(contents, &u)

	sID := uuid.NewV4()
	c := &http.Cookie{
		Name:  "cookie",
		Value: sID.String(),
	}
	dbSession[c.Value] = u.Email // email ve username aynı sitemizde
	http.SetCookie(res, c)

	http.Redirect(res, req, "/User", http.StatusSeeOther)

}

func isLogin(req *http.Request) bool {
	caughtCookie, err := req.Cookie("cookie")
	if err != nil {
		return false
	}
	if _, ok := dbSession[caughtCookie.Value]; !ok {
		return false
	}
	return true

}

func login(res http.ResponseWriter, req *http.Request) {
	// If method is GET serve an html login page
	if isLogin(req) {
		http.Redirect(res, req, "/user", http.StatusSeeOther)
		//fmt.Println("sdvds")
		return
	}
	if req.Method != "POST" {
		http.ServeFile(res, req, "login.html")
		return
	}

	// Grab the username/password from the submitted post form
	email := req.FormValue("email")
	password := req.FormValue("password")

	// Grab from the database
	var databaseEmail string
	var databasePassword string

	// Search the database for the username provided
	// If it exists grab the password for validation
	err := db.QueryRow("SELECT email, password FROM USER WHERE email=?", email).Scan(&databaseEmail, &databasePassword)
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

	////////////// cookie işlemleri
	sID := uuid.NewV4()
	c := &http.Cookie{
		Name:  "cookie",
		Value: sID.String(),
	}
	dbSession[c.Value] = email
	http.SetCookie(res, c)
	/* Session'un kendine key'i ile Session'u oluşturulan kullanıcı arasında
	bir ilişki olması için aralarında map oluşturuldu.
	*/
	/////////////////////////

	http.Redirect(res, req, "/user", http.StatusSeeOther)

}

func logout(res http.ResponseWriter, req *http.Request) {
	//fmt.Println("Logout")
	clearSession(res)
	http.Redirect(res, req, "/", 302)
}

func clearSession(res http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:   "cookie",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(res, cookie)
}

func singupPage(res http.ResponseWriter, req *http.Request) {
	if isLogin(req) {
		http.Redirect(res, req, "/user", http.StatusSeeOther)
		//fmt.Println("sdvds")
		return
	}
	req.ParseMultipartForm(0)
	// Serve signup.html to get requests to /signup
	if req.Method != "POST" {
		http.ServeFile(res, req, "signup.html")
		return
	}
	name := req.FormValue("name")
	surname := req.FormValue("surname")
	email := req.FormValue("email")
	username := email // username == email
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
		///////////////// photo işlemleri /////////////////////

		file, handler, err := req.FormFile("photo") // _ : handler
		if err != nil {
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

		_, err = db.Exec("INSERT INTO USER(username, name, surname, email, password, photo) VALUES(?, ?, ?, ?, ?, ?)", username, name, surname, email, hashedPassword, photo)

		if err != nil {
			http.Error(res, "Server error, unable to create your account.", 500)
			return
		}

		///////////// cookie işlemleri
		sID := uuid.NewV4()
		c := &http.Cookie{
			Name:  "cookie",
			Value: sID.String(),
		}
		dbSession[c.Value] = username
		http.SetCookie(res, c)
		/* Session'un kendine has Value'si ile Session'u oluşturulan kullanıcı arasında
		bir ilişki olması için aralarında map oluşturuldu.
		*/
		/////////////////////////

		http.Redirect(res, req, "/user", http.StatusSeeOther)

	case err != nil:
		http.Error(res, "Server error, unable to create your account.", 500)
		return
	default:
		http.Redirect(res, req, "/", 301)
	}
}

func userPage(res http.ResponseWriter, req *http.Request) {
	if !isLogin(req) {
		http.Redirect(res, req, "/", http.StatusSeeOther)
		//fmt.Println("sdvds")
		return
	}

	var u user
	caughtCookie, err := req.Cookie("cookie")
	if err != nil {
		fmt.Println("cookie errorr")
	}

	e := db.QueryRow("SELECT username,name,surname,photo FROM USER WHERE username=?", dbSession[caughtCookie.Value]).Scan(
		&u.Username, &u.Given_name, &u.Family_name, &u.Picture)
	if e != nil {
		fmt.Println("11111Query Error")
	}

	template, err := template.ParseFiles("user.html")
	if err != nil {
		fmt.Println("template acilamadi")
	}
	template.Execute(res, u)

}
func GuserPage(res http.ResponseWriter, req *http.Request) {
	if !isLogin(req) {
		http.Redirect(res, req, "/", http.StatusSeeOther)
		//fmt.Println("sdvds")
		return
	}
	//fmt.Println("GuserPage")
	template, err := template.ParseFiles("user.html")
	if err != nil {
		fmt.Println("template acilamadi")
	}

	template.Execute(res, u)
}

/*

username := req.URL.Query().Get("username")


*/
