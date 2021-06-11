package controllers

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sentrionic/OlympusGin/models"
	"github.com/sentrionic/OlympusGin/services"
	"github.com/sentrionic/OlympusGin/utils"
	"net/http"
	"strings"
)

type ProfileController interface {
	GetProfiles(c *gin.Context)
	GetProfileByUsername(c *gin.Context)
	FollowProfile(c *gin.Context)
	UnfollowProfile(c *gin.Context)
}

type profileController struct {
	ps services.ProfileService
}

func NewProfileController(ps services.ProfileService) ProfileController {
	return &profileController{ps}
}

func (pc *profileController) GetProfiles(c *gin.Context) {
	username := c.Query("search")
	username = strings.ToLower(username)

	users, err := pc.ps.SearchByUsername(username)

	if err != nil {
		c.JSON(utils.CreateApiError(http.StatusNotFound, errors.New("no user with that name")))
		return
	}

	var current *models.User
	value, exists := c.Get("user")

	if exists {
		current = value.(*models.User)
	}

	profiles := make([]models.Profile, 0)

	if len(*users) > 0 {
		for _, user := range *users {
			profile := ProfileSerializer(&user, current)
			profiles = append(profiles, profile)
		}
	}

	c.JSON(http.StatusOK, profiles)
	return
}

func (pc *profileController) GetProfileByUsername(c *gin.Context) {
	username := c.Param("username")
	username = strings.ToLower(username)

	if username == "" {
		c.JSON(utils.CreateApiError(http.StatusNotFound, errors.New("forgot username")))
		return
	}

	user, err := pc.ps.GetByUsername(username)

	if err != nil {
		c.JSON(utils.CreateApiError(http.StatusNotFound, errors.New("no user with that name")))
		return
	}

	var current *models.User
	value, exists := c.Get("user")

	if exists {
		current = value.(*models.User)
	}

	c.JSON(http.StatusOK, ProfileSerializer(user, current))
	return
}

func (pc *profileController) FollowProfile(c *gin.Context) {
	username := c.Param("username")
	username = strings.ToLower(username)

	if username == "" {
		c.JSON(utils.CreateApiError(http.StatusNotFound, errors.New("forgot username")))
		return
	}

	user, err := pc.ps.GetByUsername(username)

	if err != nil {
		c.JSON(utils.CreateApiError(http.StatusNotFound, errors.New("no user with that name")))
		return
	}

	authUser := c.MustGet("user").(*models.User)

	if err := pc.ps.FollowUser(*user, *authUser); err != nil {
		fmt.Println(err)
		c.JSON(utils.CreateApiError(http.StatusBadRequest, errors.New("something went wrong")))
		return
	}

	user, _ = pc.ps.GetByUsername(username)

	c.JSON(http.StatusOK, ProfileSerializer(user, authUser))
	return
}

func (pc *profileController) UnfollowProfile(c *gin.Context) {
	username := c.Param("username")
	username = strings.ToLower(username)

	if username == "" {
		c.JSON(utils.CreateApiError(http.StatusNotFound, errors.New("forgot username")))
		return
	}

	user, err := pc.ps.GetByUsername(username)

	if err != nil {
		c.JSON(utils.CreateApiError(http.StatusNotFound, errors.New("no user with that username")))
		return
	}

	authUser := c.MustGet("user").(*models.User)

	if err := pc.ps.UnfollowUser(*user, *authUser); err != nil {
		fmt.Println(err)
		c.JSON(utils.CreateApiError(http.StatusBadRequest, errors.New("something went wrong")))
		return
	}

	user, _ = pc.ps.GetByUsername(username)

	c.JSON(http.StatusOK, ProfileSerializer(user, authUser))
	return
}

func ProfileSerializer(user *models.User, current *models.User) models.Profile {
	return models.Profile{
		Id:        user.ID,
		Username:  user.Username,
		Bio:       user.Bio,
		Image:     user.Image,
		Followers: uint(len(user.Followers)),
		Followee:  uint(len(user.Followee)),
		Following: isFollowing(user, current),
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

func isFollowing(user *models.User, current *models.User) bool {
	if current == nil {
		return false
	}

	for _, v := range user.Followers {
		if v.ID == current.ID {
			return true
		}
	}
	return false
}
