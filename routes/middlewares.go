package routes

import (
	"errors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/sentrionic/OlympusGin/services"
)

func AuthUser(as services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		id := session.Get("userId")

		if id == nil {
			err := errors.New("provided session is invalid")
			c.JSON(401, gin.H{
				"error": err,
			})
			c.Abort()
			return
		}

		userId := id.(uint)

		user, err := as.GetById(userId)

		if err != nil {
			c.JSON(401, gin.H{
				"error": err,
			})
			c.Abort()
			return
		}

		c.Set("user", user)

		c.Next()
	}
}

func OptionalAuth(as services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		id := session.Get("userId")

		if id == nil {
			c.Next()
			return
		}

		userId := id.(uint)

		user, err := as.GetById(userId)

		if err != nil {
			c.Next()
			return
		}

		c.Set("user", user)

		c.Next()
	}
}
