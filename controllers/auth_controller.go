package controllers

import (
	"errors"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/sentrionic/OlympusGin/models"
	"github.com/sentrionic/OlympusGin/services"
	"github.com/sentrionic/OlympusGin/utils"
	"net/http"
	"strings"
)

type AuthController interface {
	Register(c *gin.Context)
	Login(c *gin.Context)
	Logout(ctx *gin.Context)
	ForgotPassword(ctx *gin.Context)
	ResetPassword(ctx *gin.Context)
}

type authController struct {
	as    services.AuthService
	redis services.RedisService
	mail  services.MailService
}

func NewAuthController(as services.AuthService, rs services.RedisService, ms services.MailService) AuthController {
	return &authController{
		as:    as,
		redis: rs,
		mail:  ms,
	}
}

type registerRequest struct {
	Username string `json:"username" binding:"required,gte=3,lte=30"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,gte=6"`
}

func (ac *authController) Register(c *gin.Context) {

	var req registerRequest
	if valid := bindData(c, &req); !valid {
		return
	}

	u := models.User{
		Username:     req.Username,
		Email:        strings.ToLower(req.Email),
		PasswordHash: req.Password,
		Followers:    nil,
		Followee:     nil,
	}

	exists, err := ac.as.GetByUsername(u.Username)

	if err != nil {
		fmt.Println(err)
		c.JSON(utils.CreateApiError(http.StatusBadRequest, errors.New("something went wrong")))
		return
	}

	if exists.ID != 0 {
		c.JSON(utils.CreateApiError(http.StatusBadRequest, errors.New("a user with that name already exists")))
		return
	}

	exists, err = ac.as.GetByEmail(u.Email)

	if err != nil {
		c.JSON(utils.CreateApiError(http.StatusBadRequest, errors.New("something went wrong")))
		return
	}

	if exists.ID != 0 {
		c.JSON(utils.CreateApiError(http.StatusBadRequest, errors.New("a user with that email already exists")))
		return
	}

	user, err := ac.as.Register(u)
	if err != nil {
		c.JSON(utils.ErrorFromDatabase(err))
		return
	}

	setUserSession(c, user.ID)
	c.JSON(http.StatusCreated, user)
	return
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (ac *authController) Login(c *gin.Context) {
	var req loginRequest
	if valid := bindData(c, &req); !valid {
		return
	}

	user, err := ac.as.Login(strings.ToLower(req.Email), req.Password)
	if err != nil {
		c.JSON(utils.ErrorFromDatabase(err))
		return
	}

	setUserSession(c, user.ID)
	c.JSON(http.StatusCreated, user)
	return
}

func (ac *authController) Logout(c *gin.Context) {
	c.Set("user", nil)

	session := sessions.Default(c)
	session.Set("userId", "")
	session.Clear()
	session.Options(sessions.Options{Path: "/", MaxAge: -1})
	err := session.Save()

	if err != nil {
		fmt.Printf("error clearing session: %v", err)
	}

	c.JSON(http.StatusOK, true)
	return
}

type forgotRequest struct {
	Email string `json:"email" binding:"required,email"`
}

func (ac *authController) ForgotPassword(c *gin.Context) {
	var req forgotRequest
	if valid := bindData(c, &req); !valid {
		return
	}

	user, err := ac.as.GetByEmail(req.Email)

	if err != nil {
		c.JSON(utils.ErrorFromDatabase(err))
		return
	}

	if user.ID == 0 {
		c.JSON(http.StatusOK, true)
		return
	}

	ctx := c.Request.Context()
	token, err := ac.redis.SetResetToken(ctx, user.ID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong. Try again later",
		})
		return
	}

	in := services.ResetInput{
		Email: user.Email,
		Token: token,
	}
	ac.mail.SendResetEmail(in)

	c.JSON(http.StatusCreated, true)
	return
}

type resetRequest struct {
	Token           string `json:"token" binding:"required"`
	Password        string `json:"newPassword" binding:"required"`
	ConfirmPassword string `json:"confirmNewPassword" binding:"required"`
}

func (ac *authController) ResetPassword(c *gin.Context) {
	var req resetRequest

	if valid := bindData(c, &req); !valid {
		return
	}

	if req.Password != req.ConfirmPassword {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Passwords do not match",
		})
	}

	ctx := c.Request.Context()
	id, err := ac.redis.GetIdFromToken(ctx, req.Token)

	user, err := ac.as.GetById(id)

	if err != nil {
		c.JSON(utils.ErrorFromDatabase(err))
		return
	}

	if user.ID == 0 {
		c.JSON(http.StatusOK, true)
		return
	}

	err = ac.as.ChangePassword(*user, req.Password)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong. Try again later",
		})
		return
	}

	setUserSession(c, user.ID)

	c.JSON(http.StatusOK, user)
	return
}

func setUserSession(c *gin.Context, id uint) {
	session := sessions.Default(c)
	session.Set("userId", id)
	if err := session.Save(); err != nil {
		fmt.Println(err)
	}
}
