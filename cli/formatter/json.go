package formatter

import (
	"encoding/json"
	"fmt"
)

// PrintJSON renders data as formatted JSON.
func PrintJSON(v interface{}) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Println("error formatting JSON:", err)
		return
	}
	fmt.Println(string(data))
}
