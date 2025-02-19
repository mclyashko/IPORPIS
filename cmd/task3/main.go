package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/mclyashko/IPORPIS/internal/email"
	"golang.org/x/exp/rand"
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
		createBatchEmailUI(a, w, sender)
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

// Второй этап: Ввод данных для батчевой отправки
func createBatchEmailUI(_ fyne.App, w fyne.Window, sender email.Sender) {
	csvPathEntry := widget.NewEntry()
	csvPathEntry.SetPlaceHolder("Выберите CSV файл")

	chooseFileButton := widget.NewButton("Выбрать CSV", func() {
		dialog.NewFileOpen(func(file fyne.URIReadCloser, err error) {
			if err == nil && file != nil {
				csvPathEntry.SetText(file.URI().Path())
			}
		}, w).Show()
	})

	sendButton := widget.NewButton("Отправить", func() {
		csvPath := csvPathEntry.Text
		if csvPath == "" {
			dialog.ShowError(fmt.Errorf("ошибка: Путь к CSV файлу не указан"), w)
			return
		}

		// Открываем CSV файл
		file, err := os.Open(csvPath)
		if err != nil {
			dialog.ShowError(fmt.Errorf("ошибка: Не удалось открыть файл: %v", err), w)
			return
		}
		defer file.Close()

		reader := csv.NewReader(file)
		records, err := reader.ReadAll()
		if err != nil {
			dialog.ShowError(fmt.Errorf("ошибка: Не удалось прочитать CSV файл: %v", err), w)
			return
		}

		// Обработка записей из CSV
		for _, record := range records {
			delay := rand.Intn(10) + 5 // случайное число от 5 до 14
			time.Sleep(time.Duration(delay) * time.Second)
			log.Println("Задержка пройдена!")

			if len(record) < 3 {
				dialog.ShowError(fmt.Errorf("ошибка: Неправильный формат строки в CSV"), w)
				continue
			}

			recipient := record[0]
			subject := record[1]
			message := record[2]
			var attachments []string
			if len(record) > 3 {
				for _, attachment := range record[3:] {
					if attachment != "" { // Добавляем только непустые вложения
						attachments = append(attachments, attachment)
					}
				}
			}

			if err := sender.Send(recipient, subject, message, attachments); err != nil {
				dialog.ShowError(fmt.Errorf("ошибка при отправке письма для %s: %v", recipient, err), w)
			}
		}

		dialog.ShowInformation("Успех", "Все письма успешно отправлены", w)
	})

	content := container.NewVBox(
		widget.NewLabel("Путь до CSV файла:"),
		csvPathEntry,
		chooseFileButton,
		sendButton,
	)

	w.SetContent(content)
	w.Resize(fyne.NewSize(400, 300))
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
