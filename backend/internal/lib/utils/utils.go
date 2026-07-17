package utils

import (
	"encoding/json"
	"fmt"
)

func PrintJson(v interface{}) {
	json, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return
	}

	fmt.Println(string(json))
}
