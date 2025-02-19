package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/Knetic/govaluate"
)

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Калькулятор")

	// Создаем элементы интерфейса
	input := widget.NewEntry()
	resultLabel := widget.NewLabel("Результат: ")

	calculateButton := widget.NewButton("Вычислить", func() {
		expression := input.Text
		result, err := evaluateExpression(expression)
		if err != nil {
			resultLabel.SetText("Ошибка: " + err.Error())
		} else {
			resultLabel.SetText(fmt.Sprintf("Результат: %.2f", result))
		}
	})

	// Размещаем элементы в контейнере
	content := container.NewVBox(
		input,
		calculateButton,
		resultLabel,
	)

	myWindow.SetContent(content)
	myWindow.Resize(fyne.NewSize(300, 200))
	myWindow.ShowAndRun()
}

// evaluateExpression вычисляет результат математического выражения
func evaluateExpression(expression string) (float64, error) {
	expr, err := govaluate.NewEvaluableExpression(expression)
	if err != nil {
		return 0, err
	}
	result, err := expr.Evaluate(nil)
	if err != nil {
		return 0, err
	}
	return result.(float64), nil
}
