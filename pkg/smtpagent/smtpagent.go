package smtpagent

import (
	"log"
	"net/smtp"
)

var email string
var password string

func SetCredentials(e string, p string) {
	email = e
	password = p
}

func SendMail(recipients []string, msg []byte) {
	hostname := "mail.example.com"
	auth := smtp.PlainAuth("", email, password, hostname)

	err := smtp.SendMail(hostname+":25", auth, email, recipients, msg)
	if err != nil {
		log.Fatal(err)
	}
}
