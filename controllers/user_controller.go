package controllers

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sentrionic/OlympusGin/models"
	"github.com/sentrionic/OlympusGin/services"
	"github.com/sentrionic/OlympusGin/utils"
	"mime/multipart"
	"net/http"
)

type UserController interface {
	Current(c *gin.Context)
	Edit(c *gin.Context)
	ChangePassword(c *gin.Context)
}

type userController struct {
	us services.UserService
	fs services.FileService
}

func NewUserController(us services.UserService, fs services.FileService) UserController {
	return &userController{
		us,
		fs,
	}
}

func (uc *userController) Current(c *gin.Context) {
	user, _ := c.Get("user")
	c.JSON(http.StatusOK, user)
	return
}

type editUserRequest struct {
	Username string                `form:"username" binding:"required,gte=3,lte=30"`
	Email    string                `form:"email" binding:"required,email"`
	Bio      string                `form:"bio" binding:"omitempty,lte=250"`
	Image    *multipart.FileHeader `form:"image" binding:"omitempty"`
}

func (uc *userController) Edit(c *gin.Context) {
	authUser := c.MustGet("user").(*models.User)

	var req editUserRequest

	if valid := bindData(c, &req); !valid {
		return
	}

	if authUser.Username != req.Username {
		exists, err := uc.us.GetByUsername(req.Username)

		if err != nil {
			fmt.Println(err)
			c.JSON(utils.CreateApiError(http.StatusBadRequest, errors.New("something went wrong")))
			return
		}

		if exists.ID != 0 {
			c.JSON(utils.CreateApiError(http.StatusBadRequest, errors.New("a user with that name already exists")))
			return
		}

	}

	if authUser.Email != req.Email {
		exists, err := uc.us.GetByEmail(req.Email)

		if err != nil {
			c.JSON(utils.CreateApiError(http.StatusBadRequest, errors.New("something went wrong")))
			return
		}

		if exists != nil {
			c.JSON(utils.CreateApiError(http.StatusBadRequest, errors.New("a user with that email already exists")))
			return
		}

	}

	authUser.Username = req.Username
	authUser.Bio = req.Bio
	authUser.Email = req.Email

	if req.Image != nil {
		directory := fmt.Sprintf("gin/users/%d", authUser.ID)
		url, err := uc.fs.UploadAvatar(req.Image, directory)

		if err != nil {
			c.JSON(500, gin.H{
				"error": err,
			})
			return
		}

		authUser.Image = url
	}

	user, err := uc.us.Edit(*authUser)

	if err != nil {
		c.JSON(500, gin.H{
			"error": err,
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

type changeRequest struct {
	CurrentPassword    string `json:"currentPassword" binding:"required"`
	NewPassword        string `json:"newPassword" binding:"required,gte=6"`
	ConfirmNewPassword string `json:"confirmNewPassword" binding:"required,gte=6"`
}

func (uc *userController) ChangePassword(c *gin.Context) {
	authUser := c.MustGet("user").(*models.User)

	var req changeRequest

	if valid := bindData(c, &req); !valid {
		return
	}

	if req.NewPassword != req.ConfirmNewPassword {
		c.JSON(utils.CreateApiError(http.StatusBadRequest, errors.New("passwords do not match")))
		return
	}

	err := uc.us.ChangePassword(authUser.ID, req.NewPassword)

	if err != nil {
		c.JSON(500, gin.H{
			"error": err,
		})
		return
	}

	c.JSON(http.StatusOK, authUser)
}
