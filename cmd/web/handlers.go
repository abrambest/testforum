package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"text/template"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/saparaly/snippentbox/db"
	"github.com/saparaly/snippentbox/pkg/models"
)

// userIdPost is loged in users id
var (
	userIdPost int
	userName   string
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		ErrorHandler(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		app.clienrError(w, http.StatusMethodNotAllowed)
		return
	}
	if r.URL.Path != "/" {
		// app.notFound(w)
		ErrorHandler(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	// fmt.Println(userid)

	s, err := app.posts.Latest()
	if err != nil {
		// fmt.Println("444")
		app.serverError(w, err)
		ErrorHandler(w, http.StatusText(http.StatusInternalServerError), 500)
		return
	}
	// fmt.Println("555")

	files := []string{
		"./ui/templates/index.html",
		"./ui/templates/header.html",
		"./ui/templates/footer.html",
		"./ui/templates/buttons.html",
		"./ui/templates/category-topik.html",
	}
	// fmt.Println("222")

	// Use the new render helper.
	userid := app.GetUserIDForUse(w, r)
	app.render(w, r, files, &templateData{
		Posts:  s,
		UserID: userid,
	})
	// fmt.Println("666")
}

func (app *application) showSnippet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		ErrorHandler(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		app.clienrError(w, http.StatusMethodNotAllowed)
		return
	}
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}

	post, err := app.posts.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			// app.notFound(w)
			ErrorHandler(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		} else {
			ErrorHandler(w, http.StatusText(http.StatusInternalServerError), 500)
			// app.serverError(w, err)
		}
		return
	}

	if post == nil {
		// app.notFound(w)
		ErrorHandler(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	comments, err := app.posts.GetComments(post.Id)
	if err != nil {
		// app.serverError(w, err)
		ErrorHandler(w, http.StatusText(http.StatusInternalServerError), 500)
		return
	}

	files := []string{
		"./ui/templates/category-page.html",
		"./ui/templates/header.html",
		"./ui/templates/footer.html",
		"./ui/templates/comment.html",
		"./ui/templates/buttons.html",
		"./ui/templates/category-topik.html",
	}

	userid := app.GetUserIDForUse(w, r)
	// Use the new render helper.
	app.render(w, r, files, &templateData{
		Post:     post,
		Comments: comments,
		UserID:   userid,
	})
}

func (app *application) createSnippetForm(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			ErrorHandler(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			app.clienrError(w, http.StatusMethodNotAllowed)
			return
		}
		s, err := app.posts.Latest()
		if err != nil {
			app.serverError(w, err)
			ErrorHandler(w, http.StatusText(http.StatusInternalServerError), 500)
			return
		}
		files := []string{
			"./ui/templates/write.html",
			"./ui/templates/header.html",
			"./ui/templates/footer.html",
		}

		userid := app.GetUserIDForUse(w, r)
		app.render(w, r, files, &templateData{
			Posts:  s,
			UserID: userid,
		})
	case http.MethodPost:
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			app.clienrError(w, http.StatusMethodNotAllowed)
			ErrorHandler(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
		err := r.ParseForm()
		if err != nil {
			ErrorHandler(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			app.clienrError(w, http.StatusBadRequest)
			return
		}

		userModel := app.GetUserIDForUse(w, r)
		userId := userModel.Id
		title := r.PostForm.Get("title")
		content := r.PostForm.Get("content")
		tags := r.PostForm.Get("category")

		errors := make(map[string]string)
		if userId == 0 {
			errors["title"] = "log in first"
		}
		if strings.TrimSpace(title) == "" {
			errors["title"] = "This field cannot be blank"
		} else if utf8.RuneCountInString(title) > 100 {
			errors["title"] = "This field is too long (maximum is 100 characters)"
		}

		if strings.TrimSpace(content) == "" {
			errors["content"] = "This field cannot be blank"
		}

		if len(errors) > 0 {
			files := []string{
				"./ui/templates/write.html",
				"./ui/templates/header.html",
				"./ui/templates/footer.html",
			}
			app.render(w, r, files, &templateData{
				FormErrors: errors,
				FormData:   r.PostForm,
			})
			return
		}

		id, err := app.posts.Insert(userId, title, content, userModel.Username, tags)
		if err != nil {
			app.serverError(w, err)
			ErrorHandler(w, http.StatusText(http.StatusInternalServerError), 500)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/snippet?id=%d", id), http.StatusSeeOther)
	}
}

func (app *application) createComment(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			ErrorHandler(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			app.clienrError(w, http.StatusMethodNotAllowed)
			return
		}
		app.showSnippet(w, r)
	case "POST":
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			ErrorHandler(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			app.clienrError(w, http.StatusMethodNotAllowed)
			return
		}
		id, err := strconv.Atoi(r.URL.Query().Get("id"))
		if err != nil || id < 1 {
			// app.notFound(w)
			ErrorHandler(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		err = r.ParseForm()
		if err != nil {
			// app.clienrError(w, http.StatusBadRequest)
			ErrorHandler(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		commentContent := r.PostForm.Get("commentContent")

		if strings.TrimSpace(commentContent) == "" {
			http.Redirect(w, r, fmt.Sprintf("/snippet?id=%d", id), http.StatusSeeOther)
			return
		}

		userModel := app.GetUserIDForUse(w, r)
		comment := &models.Comment{
			PostId:   id,
			Text:     commentContent,
			UserName: userModel.Username,
			UserId:   userModel.Id,
		}
		err = app.posts.InsertComment(comment)
		if err != nil {
			// app.serverError(w, err)
			ErrorHandler(w, http.StatusText(http.StatusInternalServerError), 500)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/snippet?id=%d", id), http.StatusSeeOther)
	}
}

func Signup(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			ErrorHandler(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
		renderTemplate(w, "signup", nil)
	} else if r.Method == "POST" {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			ErrorHandler(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
		email := r.FormValue("email")
		username := r.FormValue("username")
		password := r.FormValue("password")

		if strings.TrimSpace(email) == "" {
			http.Redirect(w, r, "/signup", http.StatusSeeOther)
			return
		}
		if strings.TrimSpace(username) == "" {
			http.Redirect(w, r, "/signup", http.StatusSeeOther)
			return
		}
		if strings.TrimSpace(password) == "" {
			http.Redirect(w, r, "/signup", http.StatusSeeOther)
			return
		}
		// Check if the user already exists
		db, err := db.CreateDB()
		if err != nil {
			fmt.Println("Could not connect to database:", err)
			ErrorHandler(w, http.StatusText(http.StatusInternalServerError), 500)
			return
		}
		defer db.Close()

		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ? OR name = ?", email, username).Scan(&count)
		if err != nil {
			fmt.Println("Could not query the database:", err)
			ErrorHandler(w, http.StatusText(http.StatusInternalServerError), 500)
			return
		}

		if count > 0 {
			// user already exists, redirect to signin page
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}

		stmt, err := db.Prepare("INSERT INTO users(name, email, password) values(?,?,?)")
		if err != nil {
			fmt.Println("Could not prepare insert statement:", err)
			ErrorHandler(w, http.StatusText(http.StatusInternalServerError), 500)
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(username, email, password)
		if err != nil {
			fmt.Println("Could not insert new user:", err)
			ErrorHandler(w, http.StatusText(http.StatusInternalServerError), 500)
			return
		}

		http.Redirect(w, r, "/signin", http.StatusSeeOther)
	}
}

func Signin(w http.ResponseWriter, r *http.Request) {
	dataBase, err := db.CreateDB()
	if err != nil {
		fmt.Println("Could not connect to database:", err)
		ErrorHandler(w, http.StatusText(http.StatusInternalServerError), 500)
		// http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if r.URL.Query().Get("message") == "loginfirst" {
		tmpl, _ := template.ParseFiles("./ui/templates/signin.html", "./ui/templates/header.html", "./ui/templates/footer.html")
		txt := "Please login first."
		err = tmpl.Execute(w, txt)
		return

	}
	defer dataBase.Close()
	switch r.Method {
	case "GET":
		if r.Method != http.MethodGet {

			w.Header().Set("Allow", http.MethodGet)
			ErrorHandler(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		tmpl, err := template.ParseFiles("./ui/templates/signin.html", "./ui/templates/header.html", "./ui/templates/footer.html")
		err = tmpl.Execute(w, nil)
		if err != nil {
			fmt.Println("Could :", err)
			// w.WriteHeader(http.StatusInternalServerError)
			ErrorHandler(w, http.StatusText(http.StatusInternalServerError), 500)
			return
		}
	case "POST":
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			ErrorHandler(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
		username := r.FormValue("username")
		password := r.FormValue("password")
		// Check if the user exists in the database
		// user, err := db.GetUserByUsername(dataBase, username)
		// fmt.Println(user)

		if strings.TrimSpace(username) == "" {
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}
		if strings.TrimSpace(password) == "" {
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}

		if username != "" && password != "" {
			user, err := db.GetUserByUsername(dataBase, username)
			// fmt.Println(user)
			if err != nil {
				http.Redirect(w, r, "/signup", http.StatusSeeOther)
				return
			}
			if user.Password != password {
				// w.WriteHeader(http.StatusUnauthorized)
				ErrorHandler(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			sessionToken := uuid.NewString()

			expiresAt := time.Now().Add(30 * time.Minute)

			err = db.CreateSession(dataBase, models.Session{
				UserID:         user.Id,
				Token:          sessionToken,
				ExpirationDate: expiresAt,
			})

			// this needed to user could create post
			// userIdPost = user.Id
			// userName = username

			if err != nil {
				ErrorHandler(w, http.StatusText(http.StatusInternalServerError), 500)
				// w.WriteHeader(http.StatusInternalServerError)
				return
			}

			http.SetCookie(w, &http.Cookie{
				Name:    "session_token",
				Value:   sessionToken,
				Expires: expiresAt,
				Path:    "/",
			})

			// fmt.Println(sessionToken)
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
	}
}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	t, err := template.ParseFiles("./ui/templates/signup.html", "./ui/templates/header.html", "./ui/templates/footer.html")
	if err != nil {
		ErrorHandler(w, http.StatusText(http.StatusInternalServerError), 500)
		// http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = t.Execute(w, data)
	if err != nil {
		ErrorHandler(w, http.StatusText(http.StatusInternalServerError), 500)
		// http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		ErrorHandler(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	// Get the session token from the cookie
	sessionToken, err := r.Cookie("session_token")
	if err != nil {
		// Redirect to the home page if there is no session token
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	// Delete the session from the database
	dataBase, err := db.CreateDB()
	if err != nil {
		fmt.Println("Could not connect to database:", err)
		// http.Error(w, "Internal server error", http.StatusInternalServerError)
		ErrorHandler(w, http.StatusText(http.StatusInternalServerError), 500)
		return
	}
	defer dataBase.Close()

	err = db.DeleteSession(dataBase, sessionToken.Value)
	if err != nil {
		fmt.Println(err)
		// w.WriteHeader(http.StatusInternalServerError)
		ErrorHandler(w, http.StatusText(http.StatusInternalServerError), 500)
		return
	}

	// Clear the session token cookie on the client side
	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   "",
		Expires: time.Now().Add(-time.Minute),
	})

	// Redirect to the home page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) likePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		ErrorHandler(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	postID, err := strconv.Atoi(r.PostFormValue("post_id"))
	if err != nil {
		app.errorLog.Println(err)
		app.clienrError(w, http.StatusBadRequest)
		ErrorHandler(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	action := r.PostFormValue("action")
	userModel := app.GetUserIDForUse(w, r)

	if action == "like" {
		err = app.posts.LikePost(postID, userModel.Id)
	} else if action == "dislike" {
		err = app.posts.DislikePost(postID, userModel.Id)
	}

	if err != nil {
		app.errorLog.Println(err)
		ErrorHandler(w, http.StatusText(http.StatusInternalServerError), 500)
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/snippet?id=%d", postID), http.StatusSeeOther)
}

func (app *application) likeComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		ErrorHandler(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	commentID, err := strconv.Atoi(r.PostFormValue("comment_id"))
	postID, err := strconv.Atoi(r.PostFormValue("post_idc"))
	// fmt.Println(commentID)
	if err != nil {
		app.errorLog.Println(err)
		ErrorHandler(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		app.clienrError(w, http.StatusBadRequest)
		return
	}
	action := r.PostFormValue("actionc")
	userModel := app.GetUserIDForUse(w, r)
	if action == "likec" {
		err = app.posts.LikeComment(commentID, userModel.Id)
	} else if action == "dislikec" {
		err = app.posts.DislikeComment(commentID, userModel.Id)
	}

	if err != nil {
		app.errorLog.Println(err)
		app.serverError(w, err)
		ErrorHandler(w, http.StatusText(http.StatusInternalServerError), 500)
		return
	}

	// Redirect back to the comment page.
	http.Redirect(w, r, fmt.Sprintf("/snippet?id=%d", postID), http.StatusSeeOther)
}

func (app *application) likedPosts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		ErrorHandler(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	userModel := app.GetUserIDForUse(w, r)

	s, err := app.posts.GetLikedPosts(userModel.Id)
	if err != nil {
		// app.serverError(w, err)
		ErrorHandler(w, http.StatusText(http.StatusInternalServerError), 500)
		return
	}
	files := []string{
		"./ui/html/liked_posts.html",
		"./ui/html/base.layout.html",
		"./ui/html/footer.html",
	}

	userid := app.GetUserIDForUse(w, r)
	app.render(w, r, files, &templateData{
		Posts:  s,
		UserID: userid,
	})
}

func (app *application) filterByTag(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		ErrorHandler(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	tag := r.FormValue("gettag")

	s, err := app.posts.GetPostsByTag(tag)
	if err != nil {
		// app.serverError(w, err)
		ErrorHandler(w, http.StatusText(http.StatusInternalServerError), 500)
		return
	}
	files := []string{
		"./ui/templates/select-category.html",
		"./ui/templates/header.html",
		"./ui/templates/footer.html",
		"./ui/templates/buttons.html",
		"./ui/templates/category-topik.html",
	}

	userid := app.GetUserIDForUse(w, r)
	app.render(w, r, files, &templateData{
		Posts:  s,
		UserID: userid,
	})
}

func getCookieValue(r *http.Request, cookieName string) (string, error) {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

func (app *application) GetUserIDForUse(w http.ResponseWriter, r *http.Request) *models.User {
	sessionToken, err := getCookieValue(r, "session_token")
	if err != nil {
		// app.clienrError(w, http.StatusBadRequest)
		// http.Redirect(w, r, "/", http.StatusSeeOther)
		// ErrorHandler(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return nil
	}
	userIdFrom, err := app.posts.GetUserIDByToken(sessionToken)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			// app.notFound(w)
			fmt.Println("123")
			http.Redirect(w, r, "/logout", http.StatusSeeOther)

			ErrorHandler(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		} else {
			// app.serverError(w, err)
			ErrorHandler(w, http.StatusText(http.StatusInternalServerError), 500)
		}
		return nil
	}
	return userIdFrom
}

func (app *application) userPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		ErrorHandler(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	userModel := app.GetUserIDForUse(w, r)

	s, err := app.posts.GetUserCreatedPost(userModel.Id)
	if err != nil {
		// app.serverError(w, err)
		ErrorHandler(w, http.StatusText(http.StatusInternalServerError), 500)
		return
	}
	files := []string{
		"./ui/html/userpost.html",
		"./ui/html/base.layout.html",
		"./ui/html/footer.html",
	}

	userid := app.GetUserIDForUse(w, r)
	app.render(w, r, files, &templateData{
		Posts:  s,
		UserID: userid,
	})
}
