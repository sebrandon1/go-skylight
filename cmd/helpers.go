package cmd

import (
	"encoding/json"
	"fmt"
	"os"
)

func printJSON(data interface{}) {
	output, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(output))
}
