package services

import (
	"fmt"
	"github.com/sentrionic/OlympusGin/config"
	"net/smtp"
)

type ResetInput struct {
	Email string
	Token string
}

type MailService interface {
	SendResetEmail(in ResetInput)
}

type mailService struct {
	user     string
	password string
	origin   string
}

func NewMailService(c *config.Config) MailService {
	cfg := c.Get()
	user := cfg.GetString("gmail.user")
	password := cfg.GetString("gmail.password")
	origin := cfg.GetString("app.origin")

	return &mailService{
		user:     user,
		password: password,
		origin:   origin,
	}
}

func (ms *mailService) SendResetEmail(in ResetInput) {
	from := ms.user
	pass := ms.password

	msg := "From: " + from + "\n" +
		"To: " + in.Email + "\n" +
		"Subject: Reset Email\n\n" +
		fmt.Sprintf("<a href=\"%s/reset-password/%s\">Reset Password</a>", ms.origin, in.Token)

	err := smtp.SendMail("smtp.gmail.com:587",
		smtp.PlainAuth("", from, pass, "smtp.gmail.com"),
		from, []string{in.Email}, []byte(msg))

	if err != nil {
		fmt.Println(err)
		return
	}
}
