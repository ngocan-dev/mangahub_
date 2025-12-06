package formatter

import "fmt"

// PrintTable renders a placeholder table output.
func PrintTable(headers []string, rows [][]string) {
	fmt.Println("TABLE")
	fmt.Println(headers)
	for _, row := range rows {
		fmt.Println(row)
	}
}
