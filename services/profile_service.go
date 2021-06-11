package services

import (
	"github.com/sentrionic/OlympusGin/database"
	"github.com/sentrionic/OlympusGin/models"
	"gorm.io/gorm"
)

type ProfileService interface {
	SearchByUsername(username string) (*[]models.User, error)
	GetByUsername(username string) (*models.User, error)
	FollowUser(user models.User, current models.User) error
	UnfollowUser(user models.User, current models.User) error
}

type profileService struct {
	db *gorm.DB
}

func NewProfileService(conn database.Connection) ProfileService {
	return &profileService{db: conn.Get()}
}

func (ps *profileService) SearchByUsername(username string) (*[]models.User, error) {
	var u []models.User
	result := ps.db.
		Preload("Followers").
		Preload("Followee").
		Where("LOWER(username) LIKE ?", "%"+username+"%").
		Find(&u)

	return &u, result.Error
}

func (ps *profileService) GetByUsername(username string) (*models.User, error) {
	var u models.User
	result := ps.db.
		Preload("Followers").
		Preload("Followee").
		First(&u, "LOWER(username) = ?", username)
	return &u, result.Error
}

func (ps *profileService) FollowUser(user models.User, current models.User) error {
	err := ps.db.Table("followers").Create(map[string]interface{}{
		"user_id":     user.ID,
		"follower_id": current.ID,
	}).Table("followee").Create(map[string]interface{}{
		"followee_id": user.ID,
		"user_id":     current.ID,
	}).Error
	return err
}

func (ps *profileService) UnfollowUser(user models.User, current models.User) error {
	err := ps.db.
		Exec("DELETE FROM followers WHERE user_id = ? AND follower_id = ?", user.ID, current.ID).
		Exec("DELETE FROM followee WHERE followee_id = ? AND user_id = ?", user.ID, current.ID).
		Error
	return err
}
