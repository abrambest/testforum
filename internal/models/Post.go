package models

type Post struct {
	Id      string
	Title   string
	Content string
	Theme   string
}

func NewPost(id, title, content, theme string) *Post {
	return &Post{id, title, content, theme}
}
