package vault

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sergeizaitcev/gophkeeper/pkg/workdir"
)

func testHomedir(t *testing.T) func(string) (workdir.Dir, error) {
	return func(string) (workdir.Dir, error) {
		return workdir.Dir(t.TempDir()), nil
	}
}

func testGetpass(_ *testing.T) func() (string, error) {
	return func() (string, error) {
		return "password", nil
	}
}

func TestVault(t *testing.T) {
	homedir = testHomedir(t)
	getpass = testGetpass(t)

	v, err := NewVault()
	require.NoError(t, err)

	require.True(t, v.root.Exists(FilesName))
	require.True(t, v.root.Exists(RemoteName))

	card := NewBankCard("4720-4755-3562-9559")
	logpass := NewUsernamePassword("user", "pass")
	binary := []byte("some data")

	err = v.Add("description", bytes.NewReader(binary))
	require.NoError(t, err)

	err = v.AddBankCard("description", card)
	require.NoError(t, err)

	err = v.AddLoginPassword("description", logpass)
	require.NoError(t, err)

	var n int

	err = v.Walk(func(file File) error {
		rc, err := v.Get(file.ID)
		if err != nil {
			return err
		}
		defer rc.Close()

		b, err := io.ReadAll(rc)
		if err != nil {
			return err
		}

		switch file.Type {
		case TypeBinary:
			if string(binary) != string(b) {
				return errors.New("binary is not equals")
			}
		case TypeCard:
			if string(b[:len(b)-1]) != card.String() {
				return errors.New("cards is not equals")
			}
		case TypeLogpass:
			if string(b[:len(b)-1]) != logpass.String() {
				return errors.New("logpass is not equals")
			}
		default:
			return errors.New("unexpected type")
		}

		n++

		return nil
	})

	require.NoError(t, err)
	require.Equal(t, 3, n)
}
