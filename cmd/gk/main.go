package main

import (
	"fmt"
	"os"

	"github.com/sergeizaitcev/gophkeeper/internal/gophkeeper"
)

func main() {
	if err := gophkeeper.Command.Execute(); err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}
}
