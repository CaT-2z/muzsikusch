package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"git.sch.bme.hu/kszk/opensource/authsch-go"
	"github.com/gorilla/sessions"
)

var AuthHandler http.Handler

var AuthRequest func(handler http.Handler) http.HandlerFunc

func SetupAuthSCH() {

	defer func() {
		//Ha Gebasz van, elindul basicauth-val
		if err := recover(); err != nil {
			fmt.Println("Using legacy auth: Unable to connect to AuthSCH:", err)
			passwords, err := os.Open("passwords.json")
			if err != nil {
				panic(err)
			}

			validator, err := NewBasicPasswordValidator(passwords)
			if err != nil {
				panic(err)
			}

			auth := NewBasicAuthenticator("muzsikusch", validator)

			AuthRequest = auth.AuthRequest

			AuthHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, "/", 302)
			})

		}
	}()

	if os.Getenv("AUTHSCH_ID") == "" || os.Getenv("AUTHSCH_TOKEN") == "" {
		panic("No token/id provided")
	}

	Auth := authsch.CreateClient(os.Getenv("AUTHSCH_ID"), os.Getenv("AUTHSCH_TOKEN"), []string{"basic"})

	// AuthHandler for handling the authsch login
	AuthHandler = Auth.GetLoginHandler(func(details *authsch.AccDetails, w http.ResponseWriter, r *http.Request) {
		_, err := json.MarshalIndent(details, "", "    ")
		if err != nil {
			panic(err)
		}

		session, _ := Store.Get(r, "session")
		session.Values["id"] = details.InternalID
		session.Save(r, w)

		http.Redirect(w, r, "/", 302)
	},
		func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("Couldn't log you in"))
		})

	//Checks if logged in
	AuthRequest = func(handler http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {

			session, _ := Store.Get(r, "session")

			_, ok := session.Values["id"]
			if !ok {
				http.Redirect(w, r, Auth.GetAuthURL(), http.StatusTemporaryRedirect)
				return
			}

			handler.ServeHTTP(w, r)
		}
	}
}

// TODO: put these in a different package
// Store for sessions
var Store *sessions.CookieStore

// TODO: Don't do this
var secret = os.Getenv("SESSION_SECRET")

// Init the session obj
func SessionsInit() {
	Store = sessions.NewCookieStore([]byte(secret))
	Store.MaxAge(60 * 60 * 24 * 7)
}

// GetUserSessionID get the user id from session
func GetUserSessionID(r *http.Request) string {
	session, err := Store.Get(r, "session")
	id, ok := session.Values["id"]
	if err != nil || !ok {
		log.Println("Not logged in")
		return ""
	}
	return fmt.Sprintf("%s", id)
}

// DeleteSession for logout
func DeleteSession(w http.ResponseWriter, r *http.Request) {
	session, _ := Store.Get(r, "session")
	session.Options.MaxAge = -1
	session.Save(r, w)
}
