package main

import (
	"github.com/sentrionic/OlympusGin/config"
	"github.com/sentrionic/OlympusGin/controllers"
	"github.com/sentrionic/OlympusGin/database"
	"github.com/sentrionic/OlympusGin/routes"
	"github.com/sentrionic/OlympusGin/services"
)

func main() {

	// Config
	c := config.NewConfig()
	r := routes.NewRouter(c)
	mail := services.NewMailService(c)
	file := services.NewFileService(c)
	conn := database.NewDatabaseConnection(c)
	redis := database.NewRedisConnection(c)

	// Services
	rs := services.NewRedisService(redis)
	aus := services.NewAuthService(conn)
	us := services.NewUserService(conn)
	ps := services.NewProfileService(conn)
	ars := services.NewArticleService(conn)
	cs := services.NewCommentService(conn)

	// Controllers
	au := controllers.NewAuthController(aus, rs, mail)
	uc := controllers.NewUserController(us, file)
	pc := controllers.NewProfileController(ps)
	ac := controllers.NewArticleController(ars, file)
	cc := controllers.NewCommentController(cs, ars)

	// Routes
	r.RegisterAuthRoutes(au)
	r.RegisterUserRoutes(uc, aus)
	r.RegisterProfileRoutes(pc, aus)
	r.RegisterArticleRoutes(ac, aus)
	r.RegisterCommentRoutes(cc, aus)

	if err := r.Serve(); err != nil {
		panic("error serving routes")
	}
}
