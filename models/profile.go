package models

import "time"

type Profile struct {
	Id        uint      `json:"id"`
	Username  string    `json:"username"`
	Bio       string    `json:"bio"`
	Image     string    `json:"image"`
	Followers uint      `json:"followers"`
	Followee  uint      `json:"followee"`
	Following bool      `json:"following"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
