package cryptio

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"

	"github.com/sergeizaitcev/gophkeeper/pkg/randutil"
)

var (
	_ json.Marshaler   = (*Meta)(nil)
	_ json.Unmarshaler = (*Meta)(nil)
)

// Meta определяет метаданные зашифрованного потока.
type Meta []byte

func (m Meta) MarshalJSON() ([]byte, error) {
	return json.Marshal(hex.EncodeToString(m))
}

func (m *Meta) UnmarshalJSON(data []byte) error {
	if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
		return errors.New("data must be a string")
	}

	data = data[1 : len(data)-1]
	*m = make(Meta, hex.DecodedLen(len(data)))

	if _, err := hex.Decode(*m, data); err != nil {
		return err
	}

	return nil
}

// Encrypter определяет потоковый шифратор данных.
type Encrypter struct {
	src    io.Reader
	block  cipher.Block
	stream cipher.Stream
	iv     []byte
}

// Decrypter определяет потоковый дешифратор данных.
type Decrypter struct {
	src    io.Reader
	block  cipher.Block
	stream cipher.Stream
	meta   Meta
}

// NewEncrypter возвращает новый экземпляр Encrypter.
func NewEncrypter(src io.Reader, key string) (*Encrypter, error) {
	block, err := aes.NewCipher(hash32(key))
	if err != nil {
		return nil, err
	}

	iv := make([]byte, block.BlockSize())
	if _, err = randutil.Rand.Read(iv); err != nil {
		return nil, err
	}

	stream := cipher.NewCTR(block, iv)

	return &Encrypter{
		src:    src,
		block:  block,
		stream: stream,
		iv:     iv,
	}, nil
}

// NewDecrypter возвращает новый экземпляр Decrypter.
func NewDecrypter(src io.Reader, key string, meta Meta) (*Decrypter, error) {
	block, err := aes.NewCipher(hash32(key))
	if err != nil {
		return nil, err
	}

	stream := cipher.NewCTR(block, []byte(meta))

	return &Decrypter{
		src:    src,
		block:  block,
		stream: stream,
		meta:   meta,
	}, nil
}

// Meta возвращает метаданные зашифрованного потока, необходимые при расшифровке.
// Метод следует вызывать только после завершения шифрования потока.
func (enc *Encrypter) Meta() Meta {
	meta := make(Meta, len(enc.iv))
	copy(meta, enc.iv)
	return meta
}

// Read считывает байты из src, а затем шифрует их.
func (enc *Encrypter) Read(p []byte) (n int, err error) {
	n, err = enc.src.Read(p)
	if err != nil || n == 0 {
		return n, err
	}
	enc.stream.XORKeyStream(p[:n], p[:n])
	return n, nil
}

// Read считывает байты из src, а затем расшифровывает их.
func (dec *Decrypter) Read(p []byte) (n int, err error) {
	n, err = dec.src.Read(p)
	if err != nil || n == 0 {
		return n, err
	}
	dec.stream.XORKeyStream(p[:n], p[:n])
	return n, nil
}

func hash32(s string) []byte {
	hash := md5.New()
	hash.Write([]byte(s))
	return hash.Sum(nil)
}
