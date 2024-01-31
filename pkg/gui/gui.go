package gui

import (
	"emailclient/pkg/imapagent"
	"emailclient/pkg/smtpagent"
	"log"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/emersion/go-imap/v2/imapclient"
)

func RunAppGUI() {
	a := app.New()
	w := a.NewWindow("Gmail Client")
	w.Resize(fyne.NewSize(600, 600))

	c, err := imapagent.Connect("imap.gmail.com:993", nil)
	if err != nil {
		log.Printf("Failed to dial IMAP server: %v", err)
		showPopUp(w, "Ошибка подключения к IMAP Google", 3*time.Second)
		w.Close()
	}

	log.Printf("Connected to IMAP server")
	showLoginPage(w, c)
	w.ShowAndRun()
}

func showLoginPage(w fyne.Window, c *imapclient.Client) {
	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("Введите e-mail")

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Введите пароль")

	loginButton := widget.NewButton("Login", func() {
		loginFunc(w, c, usernameEntry.Text, passwordEntry.Text)
	})

	usernameEntry.OnSubmitted = func(string) { loginFunc(w, c, usernameEntry.Text, passwordEntry.Text) }
	passwordEntry.OnSubmitted = func(string) { loginFunc(w, c, usernameEntry.Text, passwordEntry.Text) }

	form := container.NewVBox(
		widget.NewLabel("Почта"),
		usernameEntry,
		widget.NewLabel("Пароль"),
		passwordEntry,
		layout.NewSpacer(),
		loginButton,
	)

	w.SetContent(form)
}

func loginFunc(w fyne.Window, c *imapclient.Client, username string, password string) {
	if username == "" || password == "" {
		log.Printf("Login failed: empty username or password")
		showPopUp(w, "Пустое поле", 1500*time.Millisecond)
		return
	}

	err := imapagent.Login(username, password, c)
	if err != nil {
		showPopUp(w, "Неверный логин или пароль", 1500*time.Millisecond)
		return
	}

	smtpagent.SetCredentials(username, password)
	showMainContent(w, c)
}

func showMainContent(w fyne.Window, c *imapclient.Client) {
	welcomeLabel := widget.NewLabel("Welcome!")
	content := container.NewVBox(welcomeLabel)

	w.SetContent(content)
}

func showPopUp(w fyne.Window, message string, delay time.Duration) {
	popUp := widget.NewPopUp(widget.NewLabel(message), w.Canvas())
	popUp.Show()

	time.AfterFunc(delay, func() {
		popUp.Hide()
	})
}
