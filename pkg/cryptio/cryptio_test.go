package cryptio_test

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sergeizaitcev/gophkeeper/pkg/cryptio"
	"github.com/sergeizaitcev/gophkeeper/pkg/randutil"
)

func Test_EncryptDecrypt(t *testing.T) {
	key := randutil.String(64)
	text := randutil.Bytes(256)

	enc, err := cryptio.NewEncrypter(bytes.NewReader(text), key)
	require.NoError(t, err)

	encrypted, err := io.ReadAll(enc)
	require.NoError(t, err)

	dec, err := cryptio.NewDecrypter(bytes.NewReader(encrypted), key, enc.Meta())
	require.NoError(t, err)

	decrypted, err := io.ReadAll(dec)
	require.NoError(t, err)
	require.Equal(t, text, decrypted)
}

func TestMeta(t *testing.T) {
	type metaJSON struct {
		Meta cryptio.Meta `json:"meta"`
	}

	key := randutil.String(64)
	text := randutil.Bytes(256)

	enc, err := cryptio.NewEncrypter(bytes.NewReader(text), key)
	require.NoError(t, err)

	_, err = io.Copy(io.Discard, enc)
	require.NoError(t, err)

	want := metaJSON{Meta: enc.Meta()}

	jsontext, err := json.Marshal(&want)
	require.NoError(t, err)

	var got metaJSON
	require.NoError(t, json.Unmarshal(jsontext, &got))

	require.Equal(t, want, got)
}
