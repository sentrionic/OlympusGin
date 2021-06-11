package services

import (
	"github.com/sentrionic/OlympusGin/database"
	"github.com/sentrionic/OlympusGin/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CommentService interface {
	Create(comment models.Comment) (*models.Comment, error)
	List(slug string) (*[]models.Comment, error)
	Delete(comment models.Comment) error
	Get(id uint) (*models.Comment, error)
}

type commentService struct {
	db *gorm.DB
}

func NewCommentService(conn database.Connection) CommentService {
	return &commentService{db: conn.Get()}
}

func (cs *commentService) Create(comment models.Comment) (*models.Comment, error) {
	result := cs.db.Create(&comment)
	return &comment, result.Error
}

func (cs *commentService) List(slug string) (*[]models.Comment, error) {
	var c []models.Comment
	result := cs.db.
		Preload(clause.Associations).
		Joins("LEFT JOIN \"articles\" on \"articles\".id = comments.article_id").
		Where("\"articles\".slug = ?", slug).
		Find(&c)

	return &c, result.Error
}

func (cs *commentService) Delete(comment models.Comment) error {
	result := cs.db.Delete(&comment)
	return result.Error
}

func (cs *commentService) Get(id uint) (*models.Comment, error) {
	var c models.Comment
	result := cs.db.First(&c, "id = ?", id)
	return &c, result.Error
}
