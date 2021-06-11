package services

import (
	"fmt"
	"github.com/sentrionic/OlympusGin/database"
	"github.com/sentrionic/OlympusGin/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strings"
)

type ListQuery struct {
	Limit     int
	Page      int
	Cursor    string
	Tag       string
	Author    string
	Favorited string
	Order     string
	Search    string
}

type ArticleService interface {
	List(query ListQuery) (*[]models.Article, error)
	Feed(userId uint, limit int, cursor string, page int) (*[]models.Article, error)
	Bookmarked(userId uint, limit int, cursor string, page int) (*[]models.Article, error)
	Create(a models.Article) (*models.Article, error)
	GetTags() (*[]models.Tag, error)
	GetArticleBySlug(slug string) (*models.Article, error)
	UpdateArticle(a models.Article) error
	DeleteArticle(id uint) error
	SetArticleTags(tags []string, a *models.Article) error
	Favorite(a *models.Article, user *models.User) error
	Unfavorite(a *models.Article, user *models.User) error
	Bookmark(a *models.Article, user *models.User) error
	Unbookmark(a *models.Article, user *models.User) error
}

type articleService struct {
	db *gorm.DB
}

func NewArticleService(conn database.Connection) ArticleService {
	return &articleService{db: conn.Get()}
}

func (as *articleService) List(lq ListQuery) (*[]models.Article, error) {
	var a []models.Article
	offset := 0
	if lq.Page > 0 {
		offset = lq.Page - 1
	}

	query := as.db.
		Preload("Author.Followee").
		Preload("Author.Followers").
		Preload(clause.Associations)

	if lq.Order == "TOP" {
		query.Select(`
			"articles"."id",
			"articles"."created_at",
			"articles"."updated_at",       
			"articles"."slug",
			"articles"."title",
			"articles"."description",
			"articles"."body",
			"articles"."image",
			"articles"."author_id",
			(select count(*) from article_favorites where article_favorites.article_id = articles.id) as count
		`)
		query.Order("count DESC")
	} else {
		query.Order(fmt.Sprintf("created_at %s", lq.Order))
	}

	if lq.Cursor != "" {
		cursor := lq.Cursor[:len(lq.Cursor)-6]
		query.
			Where("created_at < ?", cursor)
	}

	if lq.Tag != "" {
		var t []models.Tag
		search := "%" + strings.ToLower(lq.Tag) + "%"
		as.db.Where("LOWER(tag) LIKE ?", search).Find(&t)

		var ids []uint
		for _, tag := range t {
			ids = append(ids, tag.ID)
		}

		query.Joins("JOIN article_tags ON article_tags.article_id = \"articles\".id").
			Where("article_tags.tag_id IN ?", ids)
	}

	if lq.Author != "" {
		var u models.User
		as.db.Where("LOWER(username) = ?", strings.ToLower(lq.Author)).First(&u)
		query.Where("articles.author_id = ?", u.ID)
	}

	if lq.Favorited != "" {
		var u models.User
		search := strings.ToLower(lq.Favorited)
		as.db.Where("LOWER(username) = ?", search).First(&u)

		query.Joins("JOIN article_favorites ON article_favorites.article_id = \"articles\".id").
			Where("article_favorites.user_id = ?", u.ID)
	}

	if lq.Search != "" {
		search := "%" + strings.ToLower(lq.Search) + "%"
		query.Where("LOWER(title) LIKE ? or LOWER(description) LIKE ?", search, search)
	}

	query.Limit(lq.Limit).
		Offset(offset * (lq.Limit - 1)).
		Find(&a)

	return &a, query.Error
}

func (as *articleService) Feed(userId uint, limit int, cursor string, page int) (*[]models.Article, error) {
	var a []models.Article
	offset := 0
	if page > 0 {
		offset = page - 1
	}

	query := as.db.
		Preload("Author.Followee").
		Preload("Author.Followers").
		Preload(clause.Associations).
		Joins("LEFT JOIN users u ON u.id = \"articles\".author_id").
		Joins("LEFT JOIN followee on u.id = followee.followee_id").
		Where("followee.user_id = ?", userId)

	if cursor != "" {
		cursor = cursor[:len(cursor)-6]
		query.
			Where("created_at < ?", cursor)
	}

	query.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset * (limit - 1)).
		Find(&a)

	return &a, query.Error
}

func (as *articleService) Bookmarked(userId uint, limit int, cursor string, page int) (*[]models.Article, error) {
	var a []models.Article
	offset := 0
	if page > 0 {
		offset = page - 1
	}

	query := as.db.
		Preload("Author.Followee").
		Preload("Author.Followers").
		Preload(clause.Associations).
		Joins("JOIN article_bookmarks ON article_bookmarks.article_id = \"articles\".id").
		Where("article_bookmarks.user_id = ?", userId)

	if cursor != "" {
		cursor = cursor[:len(cursor)-6]
		query.
			Where("created_at < ?", cursor)
	}

	query.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset * (limit - 1)).
		Find(&a)

	return &a, query.Error
}

func (as *articleService) Create(a models.Article) (*models.Article, error) {
	result := as.db.Create(&a)
	return &a, result.Error
}

func (as *articleService) GetTags() (*[]models.Tag, error) {
	var t []models.Tag
	result := as.db.
		Limit(10).
		Find(&t)

	return &t, result.Error
}

func (as *articleService) GetArticleBySlug(slug string) (*models.Article, error) {
	var a models.Article
	result := as.db.
		Preload("Author.Followee").
		Preload("Author.Followers").
		Preload(clause.Associations).
		Where("slug = ?", slug).
		FirstOrInit(&a)

	return &a, result.Error
}

func (as *articleService) UpdateArticle(a models.Article) error {
	result := as.db.Save(&a)
	return result.Error
}

func (as *articleService) DeleteArticle(id uint) error {
	err := as.db.Exec("DELETE FROM article_tags where article_id = ?", id).
		Exec("DELETE FROM article_favorites where article_id = ?", id).
		Exec("DELETE FROM article_bookmarks where article_id = ?", id).
		Exec("DELETE FROM comments where article_id = ?", id).
		Delete(&models.Article{}, id)
	return err.Error
}

func (as *articleService) SetArticleTags(tags []string, a *models.Article) error {
	var tagList []models.Tag

	for _, tag := range tags {
		var t models.Tag
		err := as.db.FirstOrCreate(&t, models.Tag{Tag: tag}).Error
		if err != nil {
			return err
		}
		tagList = append(tagList, t)
	}
	a.Tags = tagList

	return nil
}

func (as *articleService) Favorite(article *models.Article, current *models.User) error {
	err := as.db.Table("article_favorites").
		Create(map[string]interface{}{
			"user_id":    current.ID,
			"article_id": article.ID,
		}).Error
	return err
}

func (as *articleService) Unfavorite(article *models.Article, current *models.User) error {
	err := as.db.
		Exec("DELETE FROM article_favorites WHERE user_id = ? AND article_id = ?", current.ID, article.ID).
		Error
	return err
}

func (as *articleService) Bookmark(article *models.Article, current *models.User) error {
	err := as.db.Table("article_bookmarks").
		Create(map[string]interface{}{
			"user_id":    current.ID,
			"article_id": article.ID,
		}).Error
	return err
}

func (as *articleService) Unbookmark(article *models.Article, current *models.User) error {
	err := as.db.
		Exec("DELETE FROM article_bookmarks WHERE user_id = ? AND article_id = ?", current.ID, article.ID).
		Error
	return err
}
