package main

import (
	"fmt"
	"html/template"
	"net/http"
	"testForum/models"
)

var (
	posts   map[string]map[string]*models.Post
	chTheme string
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/index.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

	t.ExecuteTemplate(w, "index", posts)
}

func writeHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/write.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		fmt.Fprintf(w, err.Error())
	}
	t.ExecuteTemplate(w, "write", nil)
}

// func editHandler(w http.ResponseWriter, r *http.Request) {
// 	t, err := template.ParseFiles("templates/write.html", "templates/header.html", "templates/footer.html")
// 	if err != nil {
// 		fmt.Fprintf(w, err.Error())
// 	}
// 	id := r.FormValue("id")
// 	post, found := posts[id]
// 	if !found {
// 		http.NotFound(w, r)
// 		return
// 	}
// 	t.ExecuteTemplate(w, "write", post)
// }

func editHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/write.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		fmt.Fprintf(w, err.Error())
	}
	id := r.FormValue("id")
	post, found := posts[id]
	if !found {
		http.NotFound(w, r)
		return
	}
	t.ExecuteTemplate(w, "write", post)
}

func savePostHandler(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	title := r.FormValue("title")
	content := r.FormValue("content")

	var post *models.Post

	if id != "" {
		post = posts[chTheme][id]
		post.Title = title
		post.Content = content
	} else {
		id = GenerateId()
		posts[chTheme] = make(map[string]*models.Post)
		newPost := models.NewPost(id, title, content)
		posts[chTheme][newPost.Id] = newPost

	}
	fmt.Printf("posts: %v\n", posts)

	http.Redirect(w, r, "/", 302)
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")

	if id == "" {
		http.NotFound(w, r)
	}
	delete(posts, id)

	http.Redirect(w, r, "/", 302)
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/view.html", "templates/write.html", "templates/index.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		fmt.Fprintf(w, err.Error())
	}
	chTheme = r.URL.Query().Get("theme")

	postTheme, found := posts[chTheme]
	if !found {
		// fmt.Printf("tut: %s\n", chTheme)
		t.ExecuteTemplate(w, "view", nil)
		return
	}
	// fmt.Printf("tut222222: %s\n", chTheme)

	t.ExecuteTemplate(w, "view", postTheme)
}

func main() {
	posts = make(map[string]map[string]*models.Post, 0)

	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets"))))
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/write", writeHandler)
	http.HandleFunc("/edit", editHandler)
	http.HandleFunc("/delete", deleteHandler)
	http.HandleFunc("/view", viewHandler)
	http.HandleFunc("/SavePost", savePostHandler)

	fmt.Println("Listen port: http://localhost:3000")

	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		fmt.Println(err)
	}
}
