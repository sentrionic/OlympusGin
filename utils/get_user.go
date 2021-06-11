package utils

import (
	"github.com/gin-gonic/gin"
	"github.com/sentrionic/OlympusGin/models"
)

func GetUser(c *gin.Context) *models.User {
	var current *models.User
	value, exists := c.Get("user")

	if exists {
		current = value.(*models.User)
	}
	return current
}

