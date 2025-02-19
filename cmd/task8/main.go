package main

import (
	"fmt"
)

func main() {
	// 1. Найти 20 и заменить на 200 (только первое вхождение)
	numbers := []int{10, 20, 30, 20, 50}
	for i, v := range numbers {
		if v == 20 {
			numbers[i] = 200
			break
		}
	}
	fmt.Println("После замены первого 20 на 200:", numbers)

	// 2. Удалить пустые строки из списка
	stringsList := []string{"hello", "", "world", "", "go", ""}
	filteredStrings := []string{}
	for _, str := range stringsList {
		if str != "" {
			filteredStrings = append(filteredStrings, str)
		}
	}
	fmt.Println("После удаления пустых строк:", filteredStrings)

	// 3. Превратить список чисел в их квадраты
	numList := []int{1, 2, 3, 4, 5}
	squaredList := make([]int, len(numList))
	for i, num := range numList {
		squaredList[i] = num * num
	}
	fmt.Println("Список квадратов:", squaredList)

	// 4. Удалить все вхождения 20 из списка
	nums := []int{10, 20, 30, 20, 50, 20}
	filteredNums := []int{}
	for _, num := range nums {
		if num != 20 {
			filteredNums = append(filteredNums, num)
		}
	}
	fmt.Println("После удаления всех 20:", filteredNums)
}
