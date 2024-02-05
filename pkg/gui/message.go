package gui

import (
	"bytes"
	"log"
	"regexp"
	"strings"

	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/mnako/letters"
)

func parseMessagesBody(messages []*imapclient.FetchMessageBuffer) []letters.Email {
	var emails []letters.Email
	for i := len(messages) - 1; i >= 0; i-- {
		var messageBody []byte
		for _, body := range messages[i].BodySection {
			messageBody = body
			break
		}
		reader := bytes.NewReader(messageBody)
		email, err := letters.ParseEmail(reader)
		if err != nil {
			log.Printf("failed to parse email: %v", err)
			continue
		}

		if email.Text == "" {
			email.Text = "Отсутствует текст сообщения..."
		} else {
			email.Text = trimSpaces(email.Text)
			email.Text = truncateMessage(email.Text, 3000)
		}

		emails = append(emails, email)
	}

	return emails
}

func truncateMessage(message string, maxLength int) string {
	if len(message) <= maxLength {
		return message
	}
	// Убедимся, что максимальная длина достаточно велика для добавления "..."
	if maxLength < 3 {
		return "..."
	}
	return message[:maxLength-3] + "..."
}

func trimSpaces(text string) string {
	// Сначала удаляем пробелы в начале и в конце строки
	text = strings.TrimSpace(text)
	// Затем используем регулярное выражение для замены множественных пробелов на один
	space := regexp.MustCompile(`\s+`)
	text = space.ReplaceAllString(text, " ")
	return text
}
