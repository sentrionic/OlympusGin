package routes

import (
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"github.com/sentrionic/OlympusGin/config"
	"github.com/sentrionic/OlympusGin/controllers"
	"github.com/sentrionic/OlympusGin/services"
	"net/http"
)

type Router interface {
	gin.IRouter
	Serve() error
	RegisterAuthRoutes(c controllers.AuthController)
	RegisterUserRoutes(c controllers.UserController, as services.AuthService)
	RegisterProfileRoutes(c controllers.ProfileController, as services.AuthService)
	RegisterArticleRoutes(c controllers.ArticleController, as services.AuthService)
	RegisterCommentRoutes(c controllers.CommentController, as services.AuthService)
}

type router struct {
	*gin.Engine
	c *config.Config
}

func NewRouter(c *config.Config) Router {
	cfg := c.Get()
	r := gin.New()

	origin := cfg.GetString("app.origin")
	r.Use(CORS(origin))

	prod := cfg.GetString("ENVIRONMENT") == "prod"
	if prod {
		gin.SetMode(gin.ReleaseMode)
	}

	if cfg.GetBool("app.log") {
		r.Use(gin.Logger())
	}
	setupDefaults(r)

	host := cfg.GetString("redis.host")
	port := cfg.GetInt("redis.port")
	secret := cfg.GetString("app.secret")
	sessionName := cfg.GetString("app.sessionKey")

	store, _ := redis.NewStore(10, "tcp", fmt.Sprintf("%s:%d", host, port), "", []byte(secret))

	domain := cfg.GetString("app.domain")

	store.Options(sessions.Options{
		Domain:   domain,
		MaxAge:   1000 * 60 * 60 * 24 * 7, // 7 days
		Secure:   prod,
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})

	r.Use(sessions.Sessions(sessionName, store))

	return &router{Engine: r, c: c}
}

func (r *router) Serve() error {
	port := r.c.Get().GetString("app.port")
	return r.Run(":" + port)
}

func CORS(origin string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
