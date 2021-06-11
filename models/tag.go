package models

type Tag struct {
	BaseModel
	Tag     string    `gorm:"uniqueIndex"`
	Article []Article `gorm:"many2many:article_tags;"`
}
