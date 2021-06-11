package controllers

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gosimple/slug"
	"github.com/sentrionic/OlympusGin/models"
	"github.com/sentrionic/OlympusGin/models/apperrors"
	"github.com/sentrionic/OlympusGin/services"
	"github.com/sentrionic/OlympusGin/utils"
	log "github.com/sirupsen/logrus"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
)

type ArticleController interface {
	GetArticles(c *gin.Context)
	CreateArticle(c *gin.Context)
	GetFeed(c *gin.Context)
	GetBookmarked(c *gin.Context)
	GetTags(c *gin.Context)
	GetBySlug(c *gin.Context)
	UpdateArticle(c *gin.Context)
	DeleteArticle(c *gin.Context)
	FavoriteArticle(c *gin.Context)
	UnfavoriteArticle(c *gin.Context)
	BookmarkArticle(c *gin.Context)
	UnbookmarkArticle(c *gin.Context)
}

const LIMIT = 20

var validOrderTypes = map[string]bool{
	"DESC": true,
	"ASC":  true,
	"TOP":  true,
}

type articleController struct {
	as services.ArticleService
	fs services.FileService
}

func NewArticleController(as services.ArticleService, fs services.FileService) ArticleController {
	return &articleController{
		as,
		fs,
	}
}

func (ac *articleController) GetArticles(c *gin.Context) {
	page := 0
	pageQuery := c.Query("p")
	if pageQuery != "" {
		p, err := strconv.Atoi(pageQuery)
		if err != nil {
			c.JSON(utils.CreateApiError(http.StatusBadRequest, errors.New("invalid page query parameter")))
			return
		}
		page = p
	}

	limit := LIMIT
	limitQuery := c.Query("limit")
	if limitQuery != "" {
		l, err := strconv.Atoi(limitQuery)
		if err != nil {
			c.JSON(utils.CreateApiError(http.StatusBadRequest, errors.New("invalid limit query parameter")))
			return
		}
		if l < LIMIT {
			limit = l
		}
	}

	order := "DESC"
	orderQuery := strings.ToUpper(c.Query("order"))
	if orderQuery != "" {
		_, exists := validOrderTypes[orderQuery]

		if !exists {
			c.JSON(utils.CreateApiError(http.StatusBadRequest, errors.New("invalid order parameter")))
			return
		}
		order = orderQuery
	}

	limitPlusOne := limit + 1

	search := c.Query("search")
	tag := c.Query("tag")
	author := c.Query("author")
	cursor := c.Query("cursor")
	favorited := c.Query("favorited")

	query := services.ListQuery{
		Limit:     limitPlusOne,
		Page:      page,
		Cursor:    cursor,
		Tag:       tag,
		Author:    author,
		Favorited: favorited,
		Order:     order,
		Search:    search,
	}

	articles, err := ac.as.List(query)

	if err != nil {
		c.JSON(utils.ErrorFromDatabase(err))
		return
	}

	current := utils.GetUser(c)

	response := make([]models.ArticleResponse, 0)

	if len(*articles) > 0 {
		for i, a := range *articles {
			if i != limit {
				article := ArticleSerializer(&a, current)
				response = append(response, article)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"articles": response,
		"hasMore":  len(*articles) == limitPlusOne,
	})
	return
}

type articleInput struct {
	Title       string                `form:"title" binding:"required,gte=10,lte=100"`
	Description string                `form:"description" binding:"required,gte=10,lte=150"`
	Body        string                `form:"body" binding:"required"`
	Image       *multipart.FileHeader `form:"image" binding:"omitempty"`
	TagList     []string              `form:"tagList" binding:"required"`
}

func (ac *articleController) CreateArticle(c *gin.Context) {
	var req articleInput
	if valid := bindData(c, &req); !valid {
		return
	}

	if len(req.TagList) > 5 {
		c.JSON(http.StatusBadRequest, gin.H{
			"field": "tagList",
			"error": "at most 5 tags",
		})
		return
	}

	for _, tag := range req.TagList {
		length := len(tag)
		if length < 3 || length > 15 {
			c.JSON(http.StatusBadRequest, gin.H{
				"field": "tagList",
				"error": "minimum 3 characters, maximum 15",
			})
			return
		}
	}

	authUser := c.MustGet("user").(*models.User)

	seed := utils.RandomString(10)
	slg := fmt.Sprintf("%s-%s", slug.Make(req.Title), seed)
	a := models.Article{
		Slug:        slg,
		Title:       req.Title,
		Description: req.Description,
		Body:        req.Body,
		Author:      *authUser,
		Image:       fmt.Sprintf("https://picsum.photos/seed/%s/1080", seed),
	}

	if req.Image != nil {

		mimeType := req.Image.Header.Get("Content-Type")

		// Validate image mime-type is allowable
		if valid := isAllowedImageType(mimeType); !valid {
			log.Println("Image is not an allowable mime-type")
			e := apperrors.NewBadRequest("imageFile must be 'image/jpeg' or 'image/png'")
			c.JSON(e.Status(), gin.H{
				"error": e,
			})
			return
		}

		directory := fmt.Sprintf("gin/users/%d", authUser.ID)
		url, err := ac.fs.UploadImage(req.Image, directory)

		if err != nil {
			c.JSON(500, gin.H{
				"error": err,
			})
			return
		}

		a.Image = url
	}

	err := ac.as.SetArticleTags(req.TagList, &a)

	if err != nil {
		c.JSON(utils.ErrorFromDatabase(err))
		return
	}

	article, err := ac.as.Create(a)

	if err != nil {
		c.JSON(utils.ErrorFromDatabase(err))
		return
	}

	c.JSON(http.StatusCreated, ArticleSerializer(article, authUser))
	return
}

func (ac *articleController) GetFeed(c *gin.Context) {
	current := c.MustGet("user").(*models.User)
	cursor := c.Query("cursor")

	page := 0
	pageQuery := c.Query("p")
	if pageQuery != "" {
		p, err := strconv.Atoi(pageQuery)
		if err != nil {
			c.JSON(utils.CreateApiError(http.StatusBadRequest, errors.New("invalid page query parameter")))
			return
		}
		page = p
	}

	limit := LIMIT
	limitQuery := c.Query("limit")
	if limitQuery != "" {
		l, err := strconv.Atoi(limitQuery)
		if err != nil {
			c.JSON(utils.CreateApiError(http.StatusBadRequest, errors.New("invalid limit query parameter")))
			return
		}

		if l < LIMIT {
			limit = l
		}
	}

	limitPlusOne := limit + 1

	articles, err := ac.as.Feed(current.ID, limitPlusOne, cursor, page)

	if err != nil {
		c.JSON(utils.ErrorFromDatabase(err))
		return
	}

	response := make([]models.ArticleResponse, 0)

	if len(*articles) > 0 {
		for i, a := range *articles {
			if i != limit {
				article := ArticleSerializer(&a, current)
				response = append(response, article)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"articles": response,
		"hasMore":  len(*articles) == limitPlusOne,
	})
	return
}

func (ac *articleController) GetBookmarked(c *gin.Context) {
	current := c.MustGet("user").(*models.User)
	cursor := c.Query("cursor")

	page := 0
	pageQuery := c.Query("p")
	if pageQuery != "" {
		p, err := strconv.Atoi(pageQuery)
		if err != nil {
			c.JSON(utils.CreateApiError(http.StatusBadRequest, errors.New("invalid page query parameter")))
			return
		}
		page = p
	}

	limit := LIMIT
	limitQuery := c.Query("limit")
	if limitQuery != "" {
		l, err := strconv.Atoi(limitQuery)
		if err != nil {
			c.JSON(utils.CreateApiError(http.StatusBadRequest, errors.New("invalid limit query parameter")))
			return
		}

		if l < LIMIT {
			limit = l
		}
	}

	limitPlusOne := limit + 1

	articles, err := ac.as.Bookmarked(current.ID, limitPlusOne, cursor, page)

	if err != nil {
		c.JSON(utils.ErrorFromDatabase(err))
		return
	}

	response := make([]models.ArticleResponse, 0)

	if len(*articles) > 0 {
		for i, a := range *articles {
			if i != limit {
				article := ArticleSerializer(&a, current)
				response = append(response, article)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"articles": response,
		"hasMore":  len(*articles) == limitPlusOne,
	})
	return
}

func (ac *articleController) GetTags(c *gin.Context) {
	tags, err := ac.as.GetTags()

	if err != nil {
		c.JSON(utils.ErrorFromDatabase(err))
		return
	}

	response := make([]string, 0)

	if len(*tags) > 0 {
		for _, t := range *tags {
			response = append(response, t.Tag)
		}
	}

	c.JSON(http.StatusOK, response)
	return
}

func (ac *articleController) GetBySlug(c *gin.Context) {

	slg := c.Param("slug")

	if slg == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "did not specify a valid slug",
		})
		return
	}

	article, err := ac.as.GetArticleBySlug(slg)

	if err != nil {
		c.JSON(utils.ErrorFromDatabase(err))
		return
	}

	if article.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "no article with that slug",
		})
		return
	}

	current := utils.GetUser(c)

	c.JSON(http.StatusOK, ArticleSerializer(article, current))
	return
}

func (ac *articleController) UpdateArticle(c *gin.Context) {
	var req articleInput
	if valid := bindData(c, &req); !valid {
		return
	}

	if len(req.TagList) > 5 {
		c.JSON(http.StatusBadRequest, gin.H{
			"field": "tagList",
			"error": "at most 5 tags",
		})
		return
	}

	for _, tag := range req.TagList {
		length := len(tag)
		if length < 3 || length > 15 {
			c.JSON(http.StatusBadRequest, gin.H{
				"field": "tagList",
				"error": "minimum 3 characters, maximum 15",
			})
			return
		}
	}

	authUser := c.MustGet("user").(*models.User)

	slg := c.Param("slug")
	article, err := ac.as.GetArticleBySlug(slg)

	if err != nil {
		c.JSON(utils.ErrorFromDatabase(err))
		return
	}

	if article.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "no article with that slug",
		})
		return
	}

	if authUser.ID != article.AuthorId {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "only the owner of the article is allowed to delete",
		})
		return
	}

	article.Title = req.Title
	article.Description = req.Description
	article.Body = req.Body
	article.Author = *authUser

	if req.Image != nil {

		mimeType := req.Image.Header.Get("Content-Type")

		// Validate image mime-type is allowable
		if valid := isAllowedImageType(mimeType); !valid {
			log.Println("Image is not an allowable mime-type")
			e := apperrors.NewBadRequest("imageFile must be 'image/jpeg' or 'image/png'")
			c.JSON(e.Status(), gin.H{
				"error": e,
			})
			return
		}

		directory := fmt.Sprintf("gin/users/%d", authUser.ID)
		url, err := ac.fs.UploadImage(req.Image, directory)

		if err != nil {
			c.JSON(500, gin.H{
				"error": err,
			})
			return
		}

		_ = ac.fs.DeleteImage(article.Image)
		article.Image = url
	}

	err = ac.as.SetArticleTags(req.TagList, article)

	if err != nil {
		c.JSON(utils.ErrorFromDatabase(err))
		return
	}

	err = ac.as.UpdateArticle(*article)

	if err != nil {
		c.JSON(utils.ErrorFromDatabase(err))
		return
	}

	c.JSON(http.StatusCreated, ArticleSerializer(article, authUser))
	return
}

func (ac *articleController) DeleteArticle(c *gin.Context) {
	slg := c.Param("slug")

	if slg == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "did not specify a valid slug",
		})
		return
	}

	article, err := ac.as.GetArticleBySlug(slg)

	if err != nil {
		c.JSON(utils.ErrorFromDatabase(err))
		return
	}

	if article.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "no article with that slug",
		})
		return
	}

	current := c.MustGet("user").(*models.User)

	if current.ID != article.AuthorId {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "only the owner of the article is allowed to delete",
		})
		return
	}

	if err := ac.as.DeleteArticle(article.ID); err != nil {
		c.JSON(utils.ErrorFromDatabase(err))
		return
	}

	_ = ac.fs.DeleteImage(article.Image)

	c.JSON(http.StatusOK, ArticleSerializer(article, current))
	return
}

func (ac *articleController) FavoriteArticle(c *gin.Context) {

	slg := c.Param("slug")
	article, err := ac.as.GetArticleBySlug(slg)

	if err != nil {
		c.JSON(utils.ErrorFromDatabase(err))
		return
	}

	if article.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "no article with that slug",
		})
		return
	}

	current := c.MustGet("user").(*models.User)

	if !isFavorited(article, current) {
		err := ac.as.Favorite(article, current)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": err,
			})
			return
		}
	}

	article, err = ac.as.GetArticleBySlug(slg)

	c.JSON(http.StatusOK, ArticleSerializer(article, current))
	return
}

func (ac *articleController) UnfavoriteArticle(c *gin.Context) {

	slg := c.Param("slug")
	article, err := ac.as.GetArticleBySlug(slg)

	if err != nil {
		c.JSON(utils.ErrorFromDatabase(err))
		return
	}

	if article.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "no article with that slug",
		})
		return
	}

	current := c.MustGet("user").(*models.User)

	if isFavorited(article, current) {
		err := ac.as.Unfavorite(article, current)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": err,
			})
			return
		}
	}

	article, err = ac.as.GetArticleBySlug(slg)

	c.JSON(http.StatusOK, ArticleSerializer(article, current))
	return
}

func (ac *articleController) BookmarkArticle(c *gin.Context) {

	slg := c.Param("slug")
	article, err := ac.as.GetArticleBySlug(slg)

	if err != nil {
		c.JSON(utils.ErrorFromDatabase(err))
		return
	}

	if article.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "no article with that slug",
		})
		return
	}

	current := c.MustGet("user").(*models.User)

	if !isBookmarked(article, current) {
		err := ac.as.Bookmark(article, current)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": err,
			})
			return
		}
	}

	article, err = ac.as.GetArticleBySlug(slg)

	c.JSON(http.StatusOK, ArticleSerializer(article, current))
	return
}

func (ac *articleController) UnbookmarkArticle(c *gin.Context) {

	slg := c.Param("slug")
	article, err := ac.as.GetArticleBySlug(slg)

	if err != nil {
		c.JSON(utils.ErrorFromDatabase(err))
		return
	}

	if article.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "no article with that slug",
		})
		return
	}

	current := c.MustGet("user").(*models.User)

	if isBookmarked(article, current) {
		err := ac.as.Unbookmark(article, current)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": err,
			})
			return
		}
	}

	article, err = ac.as.GetArticleBySlug(slg)

	c.JSON(http.StatusOK, ArticleSerializer(article, current))
	return
}

func ArticleSerializer(article *models.Article, current *models.User) models.ArticleResponse {

	var tagList []string
	for _, tag := range article.Tags {
		tagList = append(tagList, tag.Tag)
	}

	return models.ArticleResponse{
		ID:             article.ID,
		CreatedAt:      article.CreatedAt,
		UpdatedAt:      article.UpdatedAt,
		Slug:           article.Slug,
		Title:          article.Title,
		Description:    article.Description,
		Body:           article.Body,
		Image:          article.Image,
		TagList:        tagList,
		Favorited:      isFavorited(article, current),
		Bookmarked:     isBookmarked(article, current),
		FavoritesCount: len(article.Favorites),
		Author:         ProfileSerializer(&article.Author, current),
	}
}

func isFavorited(article *models.Article, current *models.User) bool {
	if current == nil {
		return false
	}

	for _, v := range article.Favorites {
		if v.ID == current.ID {
			return true
		}
	}
	return false
}

func isBookmarked(article *models.Article, current *models.User) bool {
	if current == nil {
		return false
	}

	for _, v := range article.Bookmarks {
		if v.ID == current.ID {
			return true
		}
	}
	return false
}
