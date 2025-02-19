package main

import (
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func main() {
	myApp := app.NewWithID("com.example.texteditor")
	myWindow := myApp.NewWindow("Текстовый редактор")

	// Создаем текстовое поле
	textArea := widget.NewMultiLineEntry()

	// Создаем кнопки
	openButton := widget.NewButton("Открыть", func() {
		openFileDialog(myWindow, textArea)
	})

	saveButton := widget.NewButton("Сохранить", func() {
		saveFileDialog(myWindow, textArea)
	})

	// Создаем контейнер для кнопок
	buttons := container.NewHBox(openButton, saveButton)

	// Создаем основной контейнер
	content := container.NewVBox(textArea, buttons)

	myWindow.SetContent(content)
	myWindow.Resize(fyne.NewSize(600, 400))
	myWindow.ShowAndRun()
}

// Функция для открытия файла
func openFileDialog(window fyne.Window, textArea *widget.Entry) {
	fileDialog := dialog.NewFileOpen(func(uc fyne.URIReadCloser, err error) {
		if err == nil && uc != nil {
			defer uc.Close()
			data, _ := os.ReadFile(uc.URI().Path())
			textArea.SetText(string(data))
		}
	}, window)

	fileDialog.Show()
}

// Функция для сохранения файла
func saveFileDialog(window fyne.Window, textArea *widget.Entry) {
	fileDialog := dialog.NewFileSave(func(uc fyne.URIWriteCloser, err error) {
		if err == nil && uc != nil {
			defer uc.Close()
			_ = os.WriteFile(uc.URI().Path(), []byte(textArea.Text), os.ModePerm)
		}
	}, window)

	fileDialog.Show()
}
