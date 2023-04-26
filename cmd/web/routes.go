package main

import (
	"net/http"
)

func (app *application) routers() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", app.home)
	mux.HandleFunc("/snippet/create", app.authMiddleware(app.createSnippetForm))
	mux.HandleFunc("/snippet", app.showSnippet)
	mux.HandleFunc("/createComment", app.authMiddleware(app.createComment))
	mux.HandleFunc("/signup", Signup)
	mux.HandleFunc("/signin", Signin)
	mux.HandleFunc("/logout", Logout)
	mux.HandleFunc("/post/like", app.authMiddleware(app.likePost))
	mux.HandleFunc("/comment/like", app.authMiddleware(app.likeComment))
	mux.HandleFunc("/liked", app.authMiddleware(app.likedPosts))
	mux.HandleFunc("/tag", app.filterByTag)
	mux.HandleFunc("/userpost", app.authMiddleware(app.userPost))

	fileServer := http.FileServer(http.Dir("./ui/assets/"))
	mux.Handle("/assets/", http.StripPrefix("/assets", fileServer))

	return app.recoverPanic((mux))
}
