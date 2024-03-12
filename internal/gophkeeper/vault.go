package gophkeeper

import (
	"fmt"
	"io"
	"os"

	"github.com/rodaine/table"

	"github.com/sergeizaitcev/gophkeeper/internal/vault"
	"github.com/sergeizaitcev/gophkeeper/pkg/workdir"
)

var (
	flagDescription string // Описание данных.
	flagOutput      string // Вывод данных.
)

// AddBankCard добавляет данные банковской карты в хранилище.
func AddBankCard(args []string) error {
	if len(args) < 1 {
		return errArgsTooSmall
	}

	card := vault.NewBankCard(args[0])

	err := card.Validate()
	if err != nil {
		return err
	}

	v, err := vault.NewVault()
	if err != nil {
		return err
	}

	if err = v.AddBankCard(flagDescription, card); err != nil {
		return err
	}

	fmt.Println("the data has been successfully added")

	return nil
}

// AddUsernamePassword добавляет учетные данные пользователя в хранилище.
func AddUsernamePassword(args []string) error {
	if len(args) < 2 {
		return errArgsTooSmall
	}

	logpass := vault.NewUsernamePassword(args[0], args[1])
	err := logpass.Validate()
	if err != nil {
		return err
	}

	v, err := vault.NewVault()
	if err != nil {
		return err
	}

	if err = v.AddLoginPassword(flagDescription, logpass); err != nil {
		return err
	}

	fmt.Println("the data has been successfully added")

	return nil
}

// AddFile добавляет содержимое файла в хранилище.
func AddFile(args []string) error {
	if len(args) < 1 {
		return errArgsTooSmall
	}

	f, err := os.Open(args[0])
	if err != nil {
		return err
	}
	defer f.Close()

	v, err := vault.NewVault()
	if err != nil {
		return err
	}

	if err = v.Add(flagDescription, f); err != nil {
		return err
	}

	fmt.Println("the data has been successfully added")

	return nil
}

// Remove удаляет данные из хранилища.
func Remove(args []string) error {
	if len(args) < 1 {
		return errArgsTooSmall
	}

	v, err := vault.NewVault()
	if err != nil {
		return err
	}

	if err = v.Del(args[0]); err != nil {
		return err
	}

	fmt.Println("the data was successfully deleted")

	return nil
}

// List выводит список всех защищённых данных.
func List([]string) error {
	v, err := vault.NewVault()
	if err != nil {
		return err
	}

	tb := table.New("ID", "TYPE", "DESCRIPTION")

	_ = v.Walk(func(file vault.File) error {
		if !file.IsDeleted {
			tb.AddRow(file.ID, file.Type, file.Description)
		}
		return nil
	})

	tb.Print()

	return nil
}

// Show показывает содержимое данных.
func Show(args []string) error {
	if len(args) < 1 {
		return errArgsTooSmall
	}

	v, err := vault.NewVault()
	if err != nil {
		return err
	}

	src, err := v.Get(args[0])
	if err != nil {
		return err
	}
	defer src.Close()

	dst := os.Stdout
	if flagOutput != "" {
		dst, err = os.OpenFile(flagOutput, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, workdir.FileMode)
		if err != nil {
			return err
		}
		defer dst.Close()
	}

	if _, err = io.Copy(dst, src); err != nil {
		return err
	}

	return nil
}
