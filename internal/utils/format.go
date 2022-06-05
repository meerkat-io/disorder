package utils

import (
	"encoding/json"
	"fmt"
)

func PrintObject(value interface{}) {
	data, _ := json.MarshalIndent(value, "", "    ")
	fmt.Println(string(data))
}
