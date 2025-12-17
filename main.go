// main.go
package main

import (
	"fmt"
	"os"

	"yamlvalidator/validator"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <yaml-file>\n", os.Args[0])
		os.Exit(1)
	}

	filename := os.Args[1]

	err := validator.ValidatePodYAML(filename)
	if err != nil {
		// Ошибки уже выведены в stderr внутри ValidatePodYAML
		os.Exit(1)
	}
	// Успех — код 0
	os.Exit(0)
}