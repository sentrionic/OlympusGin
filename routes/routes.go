package routes

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/sentrionic/OlympusGin/controllers"
	"github.com/sentrionic/OlympusGin/services"
	"github.com/sentrionic/OlympusGin/utils"
	"net/http"
)

func setupDefaults(r *gin.Engine) {
	r.Use(gin.Recovery())

	r.GET("/api", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"health": "OK"})
		return
	})
	r.NoRoute(func(c *gin.Context) {
		c.JSON(utils.CreateApiError(http.StatusNotFound, errors.New("no route found")))
		return
	})
}

func (r *router) RegisterAuthRoutes(c controllers.AuthController) {
	rg := r.Group("/api")
	rg.POST("/users", c.Register)
	rg.POST("/users/login", c.Login)
	rg.POST("/users/logout", c.Logout)
	rg.POST("/users/forgot-password", c.ForgotPassword)
	rg.POST("/users/reset-password", c.ResetPassword)
}

func (r *router) RegisterUserRoutes(c controllers.UserController, as services.AuthService) {
	r.PUT("/api/users/change-password", AuthUser(as), c.ChangePassword)
	rg := r.Group("/api")
	rg.Use(AuthUser(as))
	rg.GET("/user", c.Current)
	rg.PUT("/user", c.Edit)
}

func (r *router) RegisterProfileRoutes(c controllers.ProfileController, as services.AuthService) {
	rg := r.Group("/api")
	rg.Use(OptionalAuth(as))
	rg.GET("/profiles", c.GetProfiles)
	rg.GET("/profiles/:username", c.GetProfileByUsername)

	rg.Use(AuthUser(as))
	rg.POST("/profiles/:username/follow", c.FollowProfile)
	rg.DELETE("/profiles/:username/follow", c.UnfollowProfile)
}

// RegisterArticleRoutes Issue: https://github.com/gin-gonic/gin/issues/2682
func (r *router) RegisterArticleRoutes(c controllers.ArticleController, as services.AuthService) {
	rg := r.Group("/api")
	rg.Use(OptionalAuth(as))
	rg.GET("/articles", c.GetArticles)
	rg.GET("/articles/:slug", c.GetBySlug)
	rg.GET("/articles/tags", c.GetTags)

	rg.Use(AuthUser(as))
	rg.POST("/articles", c.CreateArticle)
	rg.POST("/articles/:slug/favorite", c.FavoriteArticle)
	rg.DELETE("/articles/:slug/favorite", c.UnfavoriteArticle)
	rg.POST("/articles/:slug/bookmark", c.BookmarkArticle)
	rg.DELETE("/articles/:slug/bookmark", c.UnbookmarkArticle)
	rg.GET("/articles/feed", c.GetFeed)
	rg.GET("/articles/bookmarked", c.GetBookmarked)
	rg.PUT("/articles/:slug", c.UpdateArticle)
	rg.DELETE("/articles/:slug", c.DeleteArticle)
}

func (r *router) RegisterCommentRoutes(c controllers.CommentController, as services.AuthService) {
	rg := r.Group("/api")
	rg.Use(OptionalAuth(as))
	rg.GET("/articles/:slug/comments", c.GetArticleComments)

	rg.Use(AuthUser(as))
	rg.POST("/articles/:slug/comments", c.CreateComment)
	rg.DELETE("/articles/:slug/comments/:id", c.DeleteComment)
}