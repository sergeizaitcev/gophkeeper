package gophkeeper

import (
	"context"
	"errors"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/sergeizaitcev/gophkeeper/internal/client"
	"github.com/sergeizaitcev/gophkeeper/internal/vault"
)

var (
	flagUsername string // Имя пользователя.
	flagPassword string // Пароль пользователя.
)

// Login выполняет вход в систему на удалённом сервере.
func Login([]string) error {
	if flagUsername == "" {
		return errors.New("username must not be blank")
	}
	if flagPassword == "" {
		return errors.New("password must not be blank")
	}

	v, err := vault.NewVault()
	if err != nil {
		return err
	}

	remote := v.GetRemote()
	if remote.Address == "" {
		return errors.New("remote server address must not be blank")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGQUIT)
	defer cancel()

	token, err := client.New(remote.Address).Login(ctx, flagUsername, flagPassword)
	if err != nil {
		return err
	}

	if err = v.SetRemoteToken(token); err != nil {
		return err
	}

	fmt.Println("successful login")

	return nil
}
