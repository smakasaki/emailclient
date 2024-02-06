package gui

import (
	"bytes"
	"emailclient/pkg/imapagent"
	"emailclient/pkg/smtpagent"
	"log"
	"regexp"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/mnako/letters"
)

func updateMessages(c *imapclient.Client) {
	msgs, err := imapagent.FetchInboxMessages(c, limit, offset)
	if err != nil {
		log.Printf("Failed to fetch inbox messages: %v", err)
	}
	offset += 31

	parsedMessages := parseMessagesBody(msgs)
	messages = append(messages, parsedMessages...)
}

func loginFunc(w fyne.Window, c *imapclient.Client, username string, password string) {
	if username == "" || password == "" {
		log.Printf("Login failed: empty username or password")
		showPopUp(w, "Пустое поле", 1500*time.Millisecond)
		return
	}

	err := imapagent.Login(username, password, c)
	if err != nil {
		dialog.ShowError(err, w)
		return
	}

	smtpagent.SetCredentials(username, password)
	showMainContent(w, c)
}

func showPopUp(w fyne.Window, message string, delay time.Duration) {
	popUp := widget.NewPopUp(widget.NewLabel(message), w.Canvas())
	popUp.Show()

	time.AfterFunc(delay, func() {
		popUp.Hide()
	})
}

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
