package main

import (
	"fmt"
	"log"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"golang.org/x/exp/rand"

	"github.com/mclyashko/IPORPIS/internal/email"
)

// Первый этап: Ввод данных для создания SMTP Sender
func createSenderUI(a fyne.App, w fyne.Window) {
	serverEntry := widget.NewSelect([]string{"smtp.rambler.ru"}, nil)
	serverEntry.SetSelected("smtp.rambler.ru")

	fromEntry := widget.NewEntry()
	fromEntry.SetPlaceHolder("Введите адрес отправителя")

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Введите пароль")

	continueButton := widget.NewButton("Продолжить", func() {
		server := serverEntry.Selected
		emailAddr := fromEntry.Text
		password := passwordEntry.Text

		if server == "" || emailAddr == "" || password == "" {
			dialog.ShowError(fmt.Errorf("ошибка: Все поля должны быть заполнены"), w)
			return
		}

		sender, err := email.NewSMTPSender(server, "465", emailAddr, password)
		if err != nil {
			dialog.ShowError(fmt.Errorf("ошибка: Не удалось создать SMTP-соединение"), w)
			log.Printf("Error creating SMTP sender: %v", err)
			return
		}

		// Переход ко второму этапу
		createEmailUI(a, w, sender)
	})

	content := container.NewVBox(
		widget.NewLabel("Сервер:"),
		serverEntry,
		widget.NewLabel("Адрес отправителя:"),
		fromEntry,
		widget.NewLabel("Пароль:"),
		passwordEntry,
		continueButton,
	)

	w.SetContent(content)
	w.Resize(fyne.NewSize(400, 300))
	w.Show()
}

// Второй этап: Ввод данных для отправки письма
func createEmailUI(_ fyne.App, w fyne.Window, sender email.Sender) {
	toEntry := widget.NewEntry()
	toEntry.SetPlaceHolder("Введите адрес получателя")

	subjectEntry := widget.NewEntry()
	subjectEntry.SetPlaceHolder("Введите тему")

	messageEntry := widget.NewMultiLineEntry()
	messageEntry.SetPlaceHolder("Введите текст сообщения")

	var attachments []string
	var fileList *widget.List
	fileList = widget.NewList(
		func() int { return len(attachments) },
		func() fyne.CanvasObject {
			// Каждый элемент списка состоит из горизонтального контейнера:
			// 1) Текстового поля с названием файла
			// 2) Кнопки "Удалить"
			hbox := container.NewHBox(
				widget.NewLabel(""),
				widget.NewButton("❌", nil),
			)
			return hbox
		},
		func(i int, obj fyne.CanvasObject) {
			// Получаем контейнер и извлекаем его компоненты
			hbox := obj.(*fyne.Container)
			label := hbox.Objects[0].(*widget.Label)
			deleteBtn := hbox.Objects[1].(*widget.Button)

			// Устанавливаем название файла
			label.SetText(attachments[i])

			// Обновляем поведение кнопки "Удалить"
			deleteBtn.OnTapped = func() {
				// Удаляем файл из списка
				attachments = append(attachments[:i], attachments[i+1:]...)
				fileList.Refresh() // Обновляем список
			}
		},
	)

	fileButton := widget.NewButton("Выбрать файлы", func() {
		dialog.NewFileOpen(func(file fyne.URIReadCloser, err error) {
			if err == nil && file != nil {
				attachments = append(attachments, file.URI().Path())
				fileList.Refresh()
			}
		}, w).Show()
	})

	sendButton := widget.NewButton("Отправить", func() {
		recipient := toEntry.Text
		subject := subjectEntry.Text
		message := messageEntry.Text

		if recipient == "" || subject == "" || message == "" {
			dialog.ShowError(fmt.Errorf("ошибка: Все поля должны быть заполнены"), w)
			return
		}

		if err := sender.Send(recipient, subject, message, attachments); err != nil {
			dialog.ShowError(fmt.Errorf("ошибка: Не удалось отправить письмо"), w)
			log.Printf("Error sending email: %v", err)
		} else {
			dialog.ShowInformation("Успех", "Письмо успешно отправлено", w)
		}
	})

	content := container.NewVBox(
		widget.NewLabel("Адрес получателя:"),
		toEntry,
		widget.NewLabel("Тема:"),
		subjectEntry,
		widget.NewLabel("Сообщение:"),
		messageEntry,
		fileButton,
		fileList,
		sendButton,
	)

	w.SetContent(content)
	w.Resize(fyne.NewSize(600, 400))
}

// Основная функция
func main() {
	rand.Seed(uint64(time.Now().UnixNano()))

	a := app.NewWithID("com.mclyashko.email_sender")
	w := a.NewWindow("Email Sender")

	// Начинаем с первого этапа
	createSenderUI(a, w)

	a.Run()
}
