package models

import "time"

type Comment struct {
	BaseModel
	Body      string
	Author    User
	AuthorID  uint
	Article   Article
	ArticleID uint
}

type CommentResponse struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Body      string    `json:"body"`
	Author    Profile   `json:"author"`
}
