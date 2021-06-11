package services

import (
	"github.com/sentrionic/OlympusGin/database"
	"github.com/sentrionic/OlympusGin/models"
	"github.com/sentrionic/OlympusGin/utils"
	"gorm.io/gorm"
	"strings"
	"time"
)

type UserService interface {
	GetById(id uint) (*models.User, error)
	Edit(user models.User) (*models.User, error)
	ChangePassword(id uint, password string) error
	GetByEmail(email string) (*models.User, error)
	GetByUsername(username string) (*models.User, error)
}

type userService struct {
	db *gorm.DB
}

func NewUserService(conn database.Connection) UserService {
	return &userService{db: conn.Get()}
}

func (us *userService) GetById(id uint) (*models.User, error) {
	var u models.User
	result := us.db.Where("id = ?", id).First(&u)
	return &u, result.Error
}

func (us *userService) Edit(user models.User) (*models.User, error) {
	result := us.db.Save(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func (us *userService) GetByEmail(email string) (*models.User, error) {
	var u models.User
	result := us.db.Where("email = ?", email).FirstOrInit(&u)
	return &u, result.Error
}

func (us *userService) GetByUsername(username string) (*models.User, error) {
	var u models.User
	result := us.db.Where("LOWER(username) = ?", strings.ToLower(username)).FirstOrInit(&u)
	return &u, result.Error
}

func (us *userService) ChangePassword(id uint, password string) error {

	hash, _ := utils.HashPassword(password)

	result := us.db.Table("users").Where("id = ?", id).Updates(map[string]interface{}{
		"password":   hash,
		"updated_at": time.Now(),
	})

	if result.Error != nil {
		return result.Error
	}
	return nil
}
