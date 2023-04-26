package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/saparaly/snippentbox/db"
)

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				// app.serverError(w, fmt.Errorf("%s", err))
				ErrorHandler(w, http.StatusText(http.StatusInternalServerError), 500)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (app *application) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err != nil {
			http.Redirect(w, r, "/signin?message=loginfirst", http.StatusFound)
			w.Write([]byte("login first"))
			return
		}
		token := cookie.Value
		session, err := db.GetSessionByToken(app.posts.DB, token)
		if err != nil {
			log.Println(err.Error())
			ErrorHandler(w, http.StatusText(http.StatusInternalServerError), 500)
			return
		}
		if session.ExpirationDate.Before(time.Now()) {
			if session != nil {
				err = db.DeleteSession(app.posts.DB, token)
				if err != nil {
					fmt.Println("Error deleting session:", err)
				}
			}
			http.Redirect(w, r, "/signin", http.StatusFound)
			return
		}
		// Check if user is trying to access login or signup page
		if session != nil && (r.URL.Path == "/signin" || r.URL.Path == "/signup") {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		next.ServeHTTP(w, r)
	})
}
