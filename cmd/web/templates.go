package main

import (
	"net/url"

	"github.com/saparaly/snippentbox/pkg/models"
)

type templateData struct {
	CurrentYear int
	FormData    url.Values
	FormErrors  map[string]string
	Post        *models.Post
	Posts       []*models.Post
	Comments    []*models.Comment
	UserID      *models.User
	Likes       map[int]int
	Dislikes    map[int]int
}
