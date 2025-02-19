package main

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/jackc/pgx/v5"
)

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("PostgreSQL SQL Executor")

	// Поля для ввода данных подключения
	hostEntry := widget.NewEntry()
	hostEntry.SetText("localhost")

	portEntry := widget.NewEntry()
	portEntry.SetText("5432")

	userEntry := widget.NewEntry()
	userEntry.SetText("user")

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetText("password")

	dbEntry := widget.NewEntry()
	dbEntry.SetText("task")

	// Поле для SQL-запроса
	sqlEntry := widget.NewMultiLineEntry()

	// Поле для результата
	resultArea := widget.NewMultiLineEntry()

	// Кнопка выполнения запроса
	executeButton := widget.NewButton("Выполнить запрос", func() {
		connStr := fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s",
			userEntry.Text, passwordEntry.Text, hostEntry.Text, portEntry.Text, dbEntry.Text,
		)
		result, err := executeSQL(connStr, sqlEntry.Text)
		if err != nil {
			dialog.ShowError(err, myWindow)
			return
		}
		resultArea.SetText(result)
	})

	// Интерфейс
	content := container.NewVBox(
		widget.NewLabel("Настройки подключения:"),
		container.NewGridWithColumns(2, widget.NewLabel("Хост:"), hostEntry),
		container.NewGridWithColumns(2, widget.NewLabel("Порт:"), portEntry),
		container.NewGridWithColumns(2, widget.NewLabel("Пользователь:"), userEntry),
		container.NewGridWithColumns(2, widget.NewLabel("Пароль:"), passwordEntry),
		container.NewGridWithColumns(2, widget.NewLabel("База данных:"), dbEntry),
		widget.NewLabel("Введите SQL-запрос:"),
		sqlEntry,
		executeButton,
		widget.NewLabel("Результат:"),
		resultArea,
	)

	myWindow.SetContent(content)
	myWindow.Resize(fyne.NewSize(600, 500))
	myWindow.ShowAndRun()
}

// Выполнение SQL-запроса
func executeSQL(connStr, query string) (string, error) {
	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		return "", fmt.Errorf("не удалось подключиться к базе данных: %w", err)
	}
	defer conn.Close(context.Background())

	rows, err := conn.Query(context.Background(), query)
	if err != nil {
		return "", fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer rows.Close()

	// Читаем заголовки
	fieldDescriptions := rows.FieldDescriptions()
	headers := ""
	for _, fd := range fieldDescriptions {
		headers += fd.Name + "\t"
	}
	headers += "\n"

	// Читаем данные
	var result string
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return "", fmt.Errorf("ошибка чтения результата: %w", err)
		}
		for _, val := range values {
			result += fmt.Sprintf("%v\t", val)
		}
		result += "\n"
	}

	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("ошибка обработки результатов: %w", err)
	}

	return headers + result, nil
}
