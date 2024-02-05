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
	"github.com/mnako/letters"
)

const limit uint32 = 30

var messages []letters.Email

var offset uint32 = 0

var a fyne.App

func RunAppGUI() {
	a = app.New()
	w := a.NewWindow("Gmail Client")
	w.Resize(fyne.NewSize(600, 600))

	c, err := imapagent.Connect("imap.gmail.com:993", nil)
	if err != nil {
		log.Printf("Failed to dial IMAP server: %v", err)
		showPopUp(w, "Ошибка подключения к IMAP Google", 3*time.Second)
		w.Close()
	}
	defer c.Close()

	log.Printf("Connected to IMAP server")
	showLoginPage(w, c)
	// showMainContent(w, c)
	w.ShowAndRun()
}

func updateMessages(c *imapclient.Client) {
	msgs, err := imapagent.FetchInboxMessages(c, limit, offset)
	if err != nil {
		log.Printf("Failed to fetch inbox messages: %v", err)
	}
	offset += 31

	parsedMessages := parseMessagesBody(msgs)
	messages = append(messages, parsedMessages...)
}

func imapView(c *imapclient.Client) *fyne.Container {
	if len(messages) == 0 {
		updateMessages(c)
	}

	list := widget.NewList(
		func() int {
			return len(messages)
		},
		func() fyne.CanvasObject {
			fromLabel := widget.NewLabel("")
			subjectLabel := widget.NewLabel("")
			dateLabel := widget.NewLabel("")
			container := container.NewVBox(container.NewHBox(fromLabel, layout.NewSpacer(), dateLabel),
				subjectLabel)
			// Возвращаем контейнер как элемент списка
			return container
		},
		func(id widget.ListItemID, object fyne.CanvasObject) {
			container := object.(*fyne.Container)
			hbox := container.Objects[0].(*fyne.Container) // Первый дочерний элемент — это HBox
			fromLabel := hbox.Objects[0].(*widget.Label)   // Первый элемент в HBox
			fromLabel.TextStyle = fyne.TextStyle{Bold: true}
			dateLabel := hbox.Objects[2].(*widget.Label) // Третий элемент в HBox (после Spacer)

			// Теперь получаем доступ к subjectLabel, который является вторым дочерним элементом VBox
			subjectLabel := container.Objects[1].(*widget.Label)

			// Устанавливаем текст для лейблов
			subjectLabel.SetText(trimSpaces(truncateMessage(messages[id].Headers.Subject, 50))) // Пример использования функций trimSpaces и truncateMessage
			if len(messages[id].Headers.From) > 0 {
				fromLabel.SetText(trimSpaces(messages[id].Headers.From[0].Name))
			} else {
				fromLabel.SetText("Неизвестный отправитель")
			}
			dateLabel.SetText(messages[id].Headers.Date.Local().Format("02 Jan 2006"))
		},
	)

	list.OnSelected = func(id widget.ListItemID) {
		message := messages[id]
		showFullMessageContent(message)
	}

	showMoreButton := widget.NewButton("Показать ещё", func() {
		updateMessages(c)
		list.Refresh()
	})

	content := container.NewBorder(nil, showMoreButton, nil, nil, list)

	return content

}

func showFullMessageContent(message letters.Email) {
	detailWindow := a.NewWindow("Message")
	subjectLabel := widget.NewLabelWithStyle("Тема:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	subjectContentLabel := widget.NewLabel(message.Headers.Subject)
	fromLabel := widget.NewLabelWithStyle("Отправитель:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	fromContentLabel := widget.NewLabel(message.Headers.From[0].Address)
	textLabel := widget.NewLabelWithStyle("Содержимое письма:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	textContentLabel := widget.NewLabel(message.Text)
	textContentLabel.Wrapping = fyne.TextWrapWord // Обеспечивает перенос слов
	scrollContainer := container.NewScroll(textContentLabel)
	scrollContainer.SetMinSize(fyne.NewSize(600, 600)) // Устанавливаем минимальный размер окна

	// Собираем все виджеты вместе и добавляем их в окно
	content := container.NewVBox(subjectLabel, subjectContentLabel, fromLabel,
		fromContentLabel, textLabel, scrollContainer)
	detailWindow.SetContent(content)
	detailWindow.Show()
}

func showMainContent(w fyne.Window, c *imapclient.Client) {
	mailIconRes, err := fyne.LoadResourceFromPath("../../assets/email_icon.png")
	if err != nil {
		log.Printf("Failed to load icon: %v", err)
	}

	sendIconRes, err := fyne.LoadResourceFromPath("../../assets/send_icon.png")
	if err != nil {
		log.Printf("Failed to load icon: %v", err)
	}

	imapComponent := imapView(c)

	tabs := container.NewAppTabs(
		container.NewTabItemWithIcon(" IMAP", mailIconRes, imapComponent),
		container.NewTabItemWithIcon(" SMTP", sendIconRes, widget.NewLabel("Отправка сообщения")),
	)
	tabs.SetTabLocation(container.TabLocationTop)

	w.SetContent(tabs)
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

func showPopUp(w fyne.Window, message string, delay time.Duration) {
	popUp := widget.NewPopUp(widget.NewLabel(message), w.Canvas())
	popUp.Show()

	time.AfterFunc(delay, func() {
		popUp.Hide()
	})
}
