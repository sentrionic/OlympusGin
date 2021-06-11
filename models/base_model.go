package models

import "time"

type BaseModel struct {
	ID        uint      `json:"id" gorm:"primary_key"`
	CreatedAt time.Time `json:"createdAt" gorm:"default:now()"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"default:now()"`
}
