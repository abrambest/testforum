package models

import (
	"errors"
	"time"
)

var ErrNoRecord = errors.New("models: no matching record found")

type Snippet struct {
	ID      int
	Title   string
	Content string
	Created time.Time
}

type User struct {
	Id       int
	Email    string
	Username string
	Password string
	Role     string
}

type Like struct {
	ID        int
	PostID    int
	CommentID int
	UserID    int
}

type Dislike struct {
	ID        int
	PostID    int
	CommentID int
	UserID    int
}

type Comment struct {
	Id       int
	UserId   int
	PostId   int
	UserName string
	Date     string
	Text     string
	Like     int
	Dislike  int
}

type Post struct {
	Id           int
	AuthorId     int
	CategoriesId int
	UserName     string
	Title        string
	Description  string
	Like         int
	Dislike      int
	Tags         string
	Created      time.Time
}

type Categories struct {
	Id   int
	Name string
}

type Session struct {
	ID             int
	UserID         int
	Token          string
	ExpirationDate time.Time
}

type ErrorMsg struct {
	Status int
	Msg    string
}
