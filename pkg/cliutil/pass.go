package cliutil

import (
	"fmt"
	"syscall"

	"golang.org/x/term"
)

// ReadPassword считывает пароль из терминала и возвращает его.
func ReadPassword() (string, error) {
	fmt.Print("Password:")
	defer fmt.Println()

	pass, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}

	return string(pass), nil
}
