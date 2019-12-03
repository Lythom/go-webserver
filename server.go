package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
)

var key = securecookie.GenerateRandomKey(32)
var store = sessions.NewCookieStore([]byte(key))

func main() {

	type User struct {
		Username string
		Password string
	}

	http.HandleFunc("/random", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%f", rand.Float64())
	})

	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			var user User
			dec := json.NewDecoder(r.Body)
			err := dec.Decode(&user)
			if err != nil {
				http.Error(w, "Bad Request", http.StatusBadRequest)
				fmt.Fprintf(w, "%s", err.Error())
			} else {
				if user.Username == "theuser" && user.Password == "thepassword" {
					session, _ := store.Get(r, "auth")
					session.Values["authenticated"] = true
					session.Save(r, w)
					fmt.Fprintf(w, "OK")
				} else if user.Username == "" || user.Password == "" {
					http.Error(w, "Bad Request", http.StatusBadRequest)
				} else {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
				}

			}
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

	http.ListenAndServe(":1337", nil)
}
