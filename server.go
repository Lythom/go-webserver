package main

import (
	"./dao"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
)

var key = securecookie.GenerateRandomKey(32)
var store = sessions.NewCookieStore([]byte(key))

type User struct {
	Username string
	Password string
	Email    string
}

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	connectionString := fmt.Sprintf(
		"%s:%s@tcp(%s:3306)/%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_NAME"),
	)
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		log.Print(err.Error())
	}
	defer db.Close()

	http.HandleFunc("/random", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%f", rand.Float64())
	})

	http.HandleFunc("/subscribe", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			var user User
			dec := json.NewDecoder(r.Body)
			err := dec.Decode(&user)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if user.Username == "" || user.Password == "" || user.Email == "" {
				http.Error(w, "please provide email, password and username", http.StatusBadRequest)
				return
			}
			dao.UserExists(db, user.Username, user.Password, func(exists bool, _ bool, err error) {
				if exists {
					http.Error(w, "User already exists, please use another login", http.StatusInternalServerError)
					return
				}
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				// create user in database
				dao.CreateUser(db, user.Username, user.Password, user.Email, func(err error) {
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
					fmt.Fprintf(w, "OK")
				})
			})
		}
	})

	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			var user User
			dec := json.NewDecoder(r.Body)
			err := dec.Decode(&user)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if user.Username == "" || user.Password == "" {
				http.Error(w, "Please provide username and password", http.StatusBadRequest)
				return
			}

			dao.UserExists(db, user.Username, user.Password, func(exists bool, verifiedPassword bool, err error) {
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				if !verifiedPassword || !exists {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}

				session, _ := store.Get(r, "auth")
				session.Values["authenticated"] = true
				session.Save(r, w)

				fmt.Fprintf(w, "OK")
			})

		} else {
			http.Error(w, "Bad Request", http.StatusBadRequest)
		}
	})

	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "auth")
		auth, ok := session.Values["authenticated"].(bool)
		if !auth || !ok {
			fmt.Fprintf(w, "Visiteur")
		} else {
			fmt.Fprintf(w, "Authentifié")
		}
	})

	http.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "auth")
		session.Values["authenticated"] = false
		session.Save(r, w)
		fmt.Fprintf(w, "Vous avez été deconnecté")
	})

	http.HandleFunc("/save", func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "auth")
		session.Values["authenticated"] = false
		session.Save(r, w)
		fmt.Fprintf(w, "Vous avez été deconnecté")
	})

	http.ListenAndServe(":1337", nil)
}
