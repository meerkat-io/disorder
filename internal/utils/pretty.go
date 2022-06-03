package utils

import (
	"encoding/json"
	"fmt"
)

func PrettyPrint(value interface{}) {
	data, _ := json.MarshalIndent(value, "", "    ")
	fmt.Println(string(data))
}
