package models

type User struct {
	BaseModel
	Username     string  `gorm:"column:username;uniqueIndex" json:"username"`
	Email        string  `gorm:"column:email;uniqueIndex" json:"email"`
	Bio          string  `gorm:"column:bio;size:1024" json:"bio"`
	Image        string  `gorm:"column:image" json:"image"`
	PasswordHash string  `gorm:"column:password;not null" json:"-"`
	Followers    []*User `gorm:"many2many:followers" json:"-"`
	Followee     []*User `gorm:"many2many:followee" json:"-"`
}
