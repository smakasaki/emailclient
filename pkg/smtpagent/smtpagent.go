package smtpagent

import (
	"log"
	"net/smtp"
	"strings"
)

var email string
var password string

func SetCredentials(e string, p string) {
	log.Println("SMTP credentials initialized")
	email = e
	password = p
}

func CreateMessage(to []string, subject string, body string) []byte {
	// Преобразование списка получателей в строку, разделенную запятыми
	toHeader := strings.Join(to, ", ")

	// Формирование заголовков
	headers := []string{
		"From: " + email,
		"To: " + toHeader,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=\"utf-8\"",
		"",
		"",
	}

	// Соединение заголовков и тела сообщения
	message := strings.Join(headers, "\r\n") + body

	return []byte(message)
}

func SendMail(recipients []string, msg []byte) error {
	hostname := "smtp.gmail.com"
	auth := smtp.PlainAuth("", email, password, hostname)

	err := smtp.SendMail(hostname+":587", auth, email, recipients, msg)
	if err != nil {
		log.Printf("Failed to send email: %v", err)
		return err
	}
	log.Printf("Email sent successfully to %v", recipients)
	return err
}
