package models

import "time"

type Article struct {
	BaseModel
	Slug        string `gorm:"uniqueIndex"`
	Title       string
	Description string
	Body        string `gorm:"type:text"`
	Image       string
	Author      User
	AuthorId    uint
	Tags        []Tag  `gorm:"many2many:article_tags"`
	Favorites   []User `gorm:"many2many:article_favorites"`
	Bookmarks   []User `gorm:"many2many:article_bookmarks"`
}

type ArticleResponse struct {
	ID             uint      `json:"id"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
	Slug           string    `json:"slug"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	Body           string    `json:"body"`
	Image          string    `json:"image"`
	TagList        []string  `json:"tagList"`
	Favorited      bool      `json:"favorited"`
	Bookmarked     bool      `json:"bookmarked"`
	FavoritesCount int       `json:"favoritesCount"`
	Author         Profile   `json:"author"`
}
