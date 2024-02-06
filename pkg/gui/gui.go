package gui

import (
	"emailclient/pkg/imapagent"
	"emailclient/pkg/smtpagent"
	"log"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/mnako/letters"
)

const limit uint32 = 30

var (
	messages []letters.Email
	offset   uint32 = 0
	a        fyne.App
)

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
	smtpComponent := smtpView(w)

	tabs := container.NewAppTabs(
		container.NewTabItemWithIcon(" IMAP", mailIconRes, imapComponent),
		container.NewTabItemWithIcon(" SMTP", sendIconRes, smtpComponent),
	)
	tabs.SetTabLocation(container.TabLocationTop)

	w.SetContent(tabs)
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

func smtpView(w fyne.Window) *fyne.Container {
	emailLabel := widget.NewLabelWithStyle("Получатели:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	emailEntry := widget.NewEntry()
	emailEntry.SetPlaceHolder("Email получателей, разделенные запятой")

	subjectLabel := widget.NewLabelWithStyle("Тема письма:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	subjectEntry := widget.NewEntry()
	subjectEntry.SetPlaceHolder("Тема письма")

	bodyLabel := widget.NewLabelWithStyle("Тело письма:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	bodyEntry := widget.NewMultiLineEntry()

	sendButton := widget.NewButton("Отправить", func() {
		// Преобразование списка получателей из строки в слайс
		recipients := strings.Split(emailEntry.Text, ",")
		for i, r := range recipients {
			recipients[i] = strings.TrimSpace(r)
		}

		subject := subjectEntry.Text
		body := bodyEntry.Text

		// Предполагаем, что email отправителя уже установлен в SetCredentials
		msg := smtpagent.CreateMessage(recipients, subject, body)
		err := smtpagent.SendMail(recipients, msg)
		if err != nil {
			dialog.ShowError(err, w)
		} else {
			dialog.ShowInformation("Отправка письма", "Письмо успешно отправлено", w)
			emailEntry.Text = ""
			emailEntry.Refresh()
			subjectEntry.Text = ""
			subjectEntry.Refresh()
			bodyEntry.Text = ""
			bodyEntry.Refresh()
		}

	})

	topContent := container.NewVBox(
		emailLabel,
		emailEntry,
		subjectLabel,
		subjectEntry,
		bodyLabel,
	)

	// Использование container.NewBorder для размещения кнопки внизу и остального содержимого сверху
	mainContent := container.NewBorder(topContent, sendButton, nil, nil, bodyEntry)

	return mainContent
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
			subjectLabel.SetText(trimSpaces(truncateMessage(messages[id].Headers.Subject, 100))) // Пример использования функций trimSpaces и truncateMessage
			if len(messages[id].Headers.From) > 0 {
				fromLabel.SetText("✉ " + trimSpaces(messages[id].Headers.From[0].Name))
			} else {
				fromLabel.SetText("✉ Неизвестный отправитель")
			}

			now := time.Now()
			msgDate := messages[id].Headers.Date.Local()
			// Переводим текущее время в UTC+2
			loc, _ := time.LoadLocation("Europe/Kiev")
			nowInLoc := now.In(loc)
			msgDateInLoc := msgDate.In(loc)

			// Проверяем, является ли дата сообщения сегодняшним днем
			if nowInLoc.Format("02 Jan 2006") == msgDateInLoc.Format("02 Jan 2006") {
				// Если сегодняшняя дата, отображаем время
				dateLabel.SetText(msgDateInLoc.Format("15:04"))
			} else {
				// В других случаях оставляем дату
				dateLabel.SetText(msgDateInLoc.Format("02 Jan 2006"))
			}
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
