package services

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/sentrionic/OlympusGin/database"
	"github.com/sentrionic/OlympusGin/models"
	"github.com/sentrionic/OlympusGin/utils"
	"gorm.io/gorm"
	"strings"
	"time"
)

type AuthService interface {
	Register(u models.User) (*models.User, error)
	Login(email string, password string) (*models.User, error)
	ChangePassword(u models.User, password string) error
	GetById(id uint) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	GetByUsername(username string) (*models.User, error)
}

type authService struct {
	db *gorm.DB
}

func NewAuthService(conn database.Connection) AuthService {
	return &authService{
		db: conn.Get(),
	}
}

func (as *authService) Register(u models.User) (*models.User, error) {
	password, err := utils.HashPassword(u.PasswordHash)

	if err != nil {
		return nil, err
	}

	u.PasswordHash = password
	u.Image = fmt.Sprintf("https://gravatar.com/avatar/%s?d=identicon", getMD5Hash(u.Email))

	result := as.db.Create(&u)
	return &u, result.Error
}

func (as *authService) Login(email string, password string) (*models.User, error) {
	var result models.User
	if err := as.db.Where("email = ?", email).First(&result); err.Error != nil {
		return nil, errors.New("incorrect credentials")
	}

	if !utils.CheckPassword(password, result.PasswordHash) {
		return nil, errors.New("incorrect credentials")
	}
	return &result, nil
}

func (as *authService) GetById(id uint) (*models.User, error) {
	var u models.User
	result := as.db.Where("id = ?", id).First(&u)
	return &u, result.Error
}

func (as *authService) GetByEmail(email string) (*models.User, error) {
	var u models.User
	result := as.db.Where("email = ?", email).FirstOrInit(&u)
	return &u, result.Error
}

func (as *authService) GetByUsername(username string) (*models.User, error) {
	var u models.User
	result := as.db.Where("LOWER(username) = ?", strings.ToLower(username)).FirstOrInit(&u)
	return &u, result.Error
}

func (as *authService) ChangePassword(u models.User, password string) error {
	password, err := utils.HashPassword(password)

	if err != nil {
		return err
	}

	result := as.db.Model(&u).Where("id = ?", u.ID).Updates(map[string]interface{}{
		"password":   password,
		"updated_at": time.Now(),
	})

	if result.Error != nil {
		return result.Error
	}
	return nil
}

func getMD5Hash(email string) string {
	hash := md5.Sum([]byte(email))
	return hex.EncodeToString(hash[:])
}
