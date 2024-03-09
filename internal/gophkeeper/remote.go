package gophkeeper

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/sergeizaitcev/gophkeeper/internal/client"
	"github.com/sergeizaitcev/gophkeeper/internal/vault"
)

// RemoteSet устанавливает адрес удалённого репозитория.
func RemoteSet(args []string) error {
	if len(args) < 1 {
		return errArgsTooSmall
	}

	v, err := vault.NewVault()
	if err != nil {
		return err
	}

	return v.SetRemoteAddress(args[0])
}

// RemoteShow выводит в консоль адрес удалённого репозитория.
func RemoteShow([]string) error {
	v, err := vault.NewVault()
	if err != nil {
		return err
	}

	remote := v.GetRemote()
	if remote.Address != "" {
		fmt.Println(remote.Address)
	}

	return nil
}

// Sync синхронизирует данные в хранилище с данными из удалённого репозитория.
func Sync([]string) error {
	v, err := vault.NewVault()
	if err != nil {
		return err
	}

	remote := v.GetRemote()
	if remote.Address == "" {
		return errors.New("remote server address must not be blank")
	}
	if remote.Token == "" {
		return errors.New("not authorized")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGQUIT)
	defer cancel()

	if v.IsEmpty() {
		err = getData(ctx, v, remote)
	} else {
		err = syncData(ctx, v, remote)
	}
	if err != nil {
		return err
	}

	fmt.Println("sync up-to-date")

	return nil
}

func getData(ctx context.Context, v *vault.Vault, remote vault.Remote) error {
	c := client.New(remote.Address)

	src, err := c.GetData(ctx, remote.Token)
	if err != nil {
		return err
	}
	if src == nil {
		return nil
	}
	defer func() {
		_, _ = io.Copy(io.Discard, src)
		_ = src.Close()
	}()

	return v.Unpack(src)
}

func syncData(ctx context.Context, v *vault.Vault, remote vault.Remote) error {
	archive, err := v.Pack()
	if err != nil {
		return err
	}
	defer func() {
		name := archive.Name()
		_ = archive.Close()
		_ = os.RemoveAll(name)
	}()

	c := client.New(remote.Address)

	src, err := c.SyncData(ctx, remote.Token, archive)
	if err != nil {
		return err
	}
	if src == nil {
		return nil
	}
	defer func() {
		_, _ = io.Copy(io.Discard, src)
		_ = src.Close()
	}()

	return v.Unpack(src)
}
