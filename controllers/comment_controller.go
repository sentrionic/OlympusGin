package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/sentrionic/OlympusGin/models"
	"github.com/sentrionic/OlympusGin/services"
	"github.com/sentrionic/OlympusGin/utils"
	"net/http"
	"strconv"
)

type CommentController interface {
	CreateComment(c *gin.Context)
	GetArticleComments(c *gin.Context)
	DeleteComment(c *gin.Context)
}

type commentController struct {
	cs services.CommentService
	as services.ArticleService
}

func NewCommentController(cs services.CommentService, as services.ArticleService) CommentController {
	return &commentController{cs, as}
}

type commentRequest struct {
	Body string `json:"body" binding:"required,gte=3,lte=250"`
}

func (cc *commentController) CreateComment(c *gin.Context) {
	var req commentRequest
	if valid := bindData(c, &req); !valid {
		return
	}
	authUser := c.MustGet("user").(*models.User)

	slg := c.Param("slug")
	article, err := cc.as.GetArticleBySlug(slg)

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

	nc := models.Comment{
		Body:    req.Body,
		Author:  *authUser,
		Article: *article,
	}

	comment, err := cc.cs.Create(nc)

	if err != nil {
		c.JSON(utils.ErrorFromDatabase(err))
		return
	}

	c.JSON(http.StatusOK, CommentSerializer(comment, authUser))
	return
}

func (cc *commentController) GetArticleComments(c *gin.Context) {
	current := utils.GetUser(c)
	slg := c.Param("slug")

	comments, err := cc.cs.List(slg)

	if err != nil {
		c.JSON(utils.ErrorFromDatabase(err))
		return
	}

	response := make([]models.CommentResponse, 0)

	if len(*comments) > 0 {
		for _, c := range *comments {
			comment := CommentSerializer(&c, current)
			response = append(response, comment)
		}
	}

	c.JSON(http.StatusOK, response)
	return
}

func (cc *commentController) DeleteComment(c *gin.Context) {
	authUser := c.MustGet("user").(*models.User)

	slg := c.Param("slug")
	article, err := cc.as.GetArticleBySlug(slg)

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

	param := c.Param("id")
	id, _ := strconv.Atoi(param)

	comment, err := cc.cs.Get(uint(id))

	if err != nil {
		c.JSON(utils.ErrorFromDatabase(err))
		return
	}

	if comment.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "no comment with that id",
		})
		return
	}

	if comment.AuthorID != authUser.ID {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "not the owner of the comment",
		})
		return
	}

	err = cc.cs.Delete(*comment)

	if err != nil {
		c.JSON(utils.ErrorFromDatabase(err))
		return
	}

	c.JSON(http.StatusOK, CommentSerializer(comment, authUser))
	return
}

func CommentSerializer(comment *models.Comment, current *models.User) models.CommentResponse {
	return models.CommentResponse{
		ID:        comment.ID,
		CreatedAt: comment.CreatedAt,
		UpdatedAt: comment.UpdatedAt,
		Body:      comment.Body,
		Author:    ProfileSerializer(&comment.Author, current),
	}
}
