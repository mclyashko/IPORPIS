package main

import (
	"log"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"golang.org/x/exp/rand"

	"github.com/mclyashko/IPORPIS/internal/config"
	"github.com/mclyashko/IPORPIS/internal/email"
)

func getConfig() config.App {
	configLoader := &config.DotenvConfigLoader{}

	appConfig, err := configLoader.Load()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	return appConfig
}

func createUI(es email.Sender) {
	a := app.New()
	w := a.NewWindow("Email Sender")

	// Поле для ввода адреса получателя
	toEntry := widget.NewEntry()
	toEntry.SetPlaceHolder("Введите адрес получателя")

	// Поле для ввода темы
	subjectEntry := widget.NewEntry()
	subjectEntry.SetPlaceHolder("Введите тему")

	// Поле для ввода текста сообщения
	messageEntry := widget.NewMultiLineEntry()
	messageEntry.SetPlaceHolder("Введите текст сообщения")

	// Кнопка отправки
	sendButton := widget.NewButton("Отправить", func() {
		recipient := toEntry.Text
		subject := subjectEntry.Text
		message := messageEntry.Text

		if err := es.Send(recipient, subject, message, []string{}); err != nil {
			log.Printf("Error sending email: %v", err)
		} else {
			log.Println("Email sent!")
		}
	})

	// Организуем вертикальный контейнер
	content := container.NewVBox(
		widget.NewLabel("Адрес получателя:"),
		toEntry,
		widget.NewLabel("Тема:"),
		subjectEntry,
		widget.NewLabel("Сообщение:"),
		messageEntry,
		sendButton,
	)

	w.SetContent(content)

	// Устанавливаем размер окна
	w.Resize(fyne.NewSize(400, 300)) // Начальный размер окна

	// Показываем окно
	w.ShowAndRun()
}

func main() {
	rand.Seed(uint64(time.Now().UnixNano()))

	cfg := getConfig()
	es, err := email.NewSMTPSender(
		cfg.Email.Host,
		cfg.Email.Port,
		cfg.Email.Username,
		cfg.Email.Password,
	)
	if err != nil {
		log.Fatalf("Cant get SMTP sender: %v", err)
	}

	createUI(es)
}
