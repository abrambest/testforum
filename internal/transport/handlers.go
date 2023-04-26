package transport

import (
	"fmt"
	"html/template"
	"net/http"
	"testForum/internal/models"
	"testForum/internal/mysql"
	"testForum/internal/pkg"
)

type ContactDetails struct {
	Login         string
	Password      string
	Success       bool
	StorageAccess string
}

var (
	posts   map[string]*models.Post
	chTheme string
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("ui/templates/index.html", "ui/templates/header.html", "ui/templates/footer.html")
	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

	homePosts := make(map[string]*models.Post)

	for _, v := range posts {
		if v.Title != "" {
			homePosts[v.Id] = v
		} else {
			continue
		}
	}

	t.ExecuteTemplate(w, "index", homePosts)
}

func signHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("ui/templates/signin.html", "ui/templates/header.html", "ui/templates/footer.html")
	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

	// data:= ContactDetails{
	// 	Login:         r.FormValue()
	// Password:     string
	// Success:       bool
	// StorageAccess: string
	// }

	t.ExecuteTemplate(w, "signin", nil)
}

func writeHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("ui/templates/write.html", "ui/templates/header.html", "ui/templates/footer.html")
	if err != nil {
		fmt.Fprintf(w, err.Error())
	}
	t.ExecuteTemplate(w, "write", nil)
}

// func editHandler(w http.ResponseWriter, r *http.Request) {
// 	t, err := template.ParseFiles("ui/templates/write.html", "ui/templates/header.html", "ui/templates/footer.html")
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
	t, err := template.ParseFiles("ui/templates/write.html", "ui/templates/header.html", "ui/templates/footer.html")
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
	chTheme = r.FormValue("category")

	fmt.Printf("ffff - %v\n", chTheme)

	var post *models.Post

	if id != "" {
		post = posts[id]
		post.Title = title
		post.Content = content

	} else {
		id = pkg.GenerateId()

		// _, found := posts[chTheme]
		// if !found {
		// 	posts[chTheme] = make(map[string]*models.Post, 0)
		// }

		fmt.Printf("bbb - %v\n", chTheme)
		newPost := models.NewPost(id, title, content, chTheme)

		posts[newPost.Id] = newPost

		// if _, ok := posts[chTheme]; !ok {
		// 	fmt.Println("posts: AAAAAAAAAAAA posts")
		// 	http.Error(w, "Выберите категорию", http.StatusBadRequest)
		// 	return
		// }

	}
	tempTheme := fmt.Sprintf("/view?theme=%s", chTheme)
	fmt.Printf("posts: %v\n", posts)

	http.Redirect(w, r, tempTheme, 302)
}

func commentPostHandler(w http.ResponseWriter, r *http.Request) {
	title := ""
	content := r.FormValue("content")

	id := pkg.GenerateId()

	fmt.Printf("comment - %s\n", chTheme)
	newPost := models.NewPost(id, title, content, chTheme)

	posts[newPost.Id] = newPost

	tempTheme := fmt.Sprintf("/view?theme=%s", chTheme)

	http.Redirect(w, r, tempTheme, 302)
}

func signUpHandler(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	username := r.FormValue("username")
	passwd := r.FormValue("passwd")
	passhash, err := pkg.PassSecurity(passwd)

	erro := pkg.CheckUserInfo(*models.NewUser(username, email, passwd))
	if erro != nil {
		fmt.Fprint(w, erro)
		return
	}

	comngStruct := models.NewUser(email, username, passhash)

	err = mysql.SignUp(comngStruct)
	if err != nil {
		fmt.Fprint(w, "username is full")
		return
	}

	fmt.Println(comngStruct)

	http.Redirect(w, r, "/", http.StatusSeeOther)
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
	t, err := template.ParseFiles("ui/templates/view.html", "ui/templates/comment.html", "ui/templates/header.html", "ui/templates/footer.html")
	if err != nil {
		fmt.Fprintf(w, err.Error())
	}
	chTheme = r.URL.Query().Get("theme")
	fmt.Printf("view - %v\n", chTheme)

	arrPosts := make(map[string]*models.Post)

	for _, v := range posts {
		if v.Theme == chTheme {
			arrPosts[v.Id] = v
		} else {
			continue
		}
	}

	t.ExecuteTemplate(w, "view", arrPosts)
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("ui/templates/write.html", "ui/templates/header.html", "ui/templates/footer.html")
	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

	t.ExecuteTemplate(w, "write", nil)
}

func Handlers() {
	posts = make(map[string]*models.Post, 0)

	http.Handle("/assets/", http.StripPrefix("/assets", http.FileServer(http.Dir("./ui/assets/"))))

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/signin", signHandler)
	http.HandleFunc("/write", writeHandler)
	http.HandleFunc("/edit", editHandler)
	http.HandleFunc("/delete", deleteHandler)
	http.HandleFunc("/view", viewHandler)
	http.HandleFunc("/create", createHandler)
	http.HandleFunc("/CommentPost", commentPostHandler)
	http.HandleFunc("/SavePost", savePostHandler)
	http.HandleFunc("/SignUp", signUpHandler)

	fmt.Println("Listen port: http://localhost:3000")
}
