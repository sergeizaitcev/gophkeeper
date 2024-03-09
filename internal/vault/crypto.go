package vault

import (
	"io"

	"github.com/sergeizaitcev/gophkeeper/pkg/cliutil"
	"github.com/sergeizaitcev/gophkeeper/pkg/cryptio"
	"github.com/sergeizaitcev/gophkeeper/pkg/randutil"
)

// getPass возвращает мастер-пароль пользователя.
var getpass = func() (string, error) {
	return cliutil.ReadPassword()
}

// decryptCloser определяет реализацию интерфейса io.ReadCloser для
// cryptio.Decrypter
type decryptCloser struct {
	*cryptio.Decrypter
	io.Closer
}

// generateID генерирует 12-символьную hex-строку.
func generateID() string {
	return randutil.Hex(12)
}

func newEncrypter(src io.Reader) (*cryptio.Encrypter, error) {
	key, err := getpass()
	if err != nil {
		return nil, err
	}
	return cryptio.NewEncrypter(src, key)
}

func newDecrypter(src io.Reader, meta cryptio.Meta) (*cryptio.Decrypter, error) {
	key, err := getpass()
	if err != nil {
		return nil, err
	}
	return cryptio.NewDecrypter(src, key, meta)
}
