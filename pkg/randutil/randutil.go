package randutil

import (
	"crypto/rand"
	"encoding/binary"
	"io"
	mathrand "math/rand"
	"unsafe"
)

var Rand *mathrand.Rand

func init() {
	buf := make([]byte, 8)
	_, err := io.ReadFull(rand.Reader, buf)
	if err != nil {
		panic(err)
	}
	src := mathrand.NewSource(int64(binary.LittleEndian.Uint64(buf)))
	Rand = mathrand.New(src)
}

const (
	ascii = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFJHIJKLMNOPQRSTUVWXYZ"
	hex   = "0123456789abcdef"
)

// Bytes генерирует случайную байтовую последовательность длинной n, состоящую
// из ASCII-символов.
//
// Если n <= 0, то возвращает nil.
func Bytes(n int) []byte {
	if n <= 0 {
		return nil
	}

	buf := make([]byte, 0, n)
	i := 0

	for len(buf) < n {
		idx := Rand.Intn(len(ascii) - 1)
		char := ascii[idx]
		if i == 0 && '0' <= char && char <= '9' {
			continue
		}
		buf = append(buf, char)
		i++
	}

	return buf
}

// String генерирует случайную ASCII-последовательность длинной n.
//
// Если n <= 0, то возвращает пустую строку.
func String(n int) string {
	b := Bytes(n)
	return unsafe.String(unsafe.SliceData(b), len(b))
}

// Hex генерирует случайную hex-последовательность длинной n.
//
// Если n <= 0, то возвращает пустую строку.
func Hex(n int) string {
	if n <= 0 {
		return ""
	}

	buf := make([]byte, 0, n)
	i := 0

	for len(buf) < n {
		idx := Rand.Intn(len(hex) - 1)
		char := hex[idx]
		if i == 0 && '0' <= char && char <= '9' {
			continue
		}
		buf = append(buf, char)
		i++
	}

	return string(buf)
}
