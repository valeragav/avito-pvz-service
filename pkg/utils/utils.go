package utils

import (
	"encoding/json"
	"fmt"
)

func PrintJSON(data any) {
	jsonData, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		fmt.Printf("Ошибка при преобразовании в JSON: %v", err)
		return
	}
	fmt.Println(string(jsonData))
}
