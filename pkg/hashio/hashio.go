package hashio

import (
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"io"
)

type HashReader struct {
	src  io.Reader
	hash hash.Hash
}

type HashWriter struct {
	dst  io.Writer
	hash hash.Hash
}

func NewHashReader(src io.Reader) *HashReader {
	return &HashReader{
		src:  src,
		hash: sha256.New(),
	}
}

func NewHashWriter(dst io.Writer) *HashWriter {
	return &HashWriter{
		dst:  dst,
		hash: sha256.New(),
	}
}

func (r *HashReader) Checksum() string {
	return string(hex.EncodeToString(r.hash.Sum(nil)))
}

func (w *HashWriter) Checksum() string {
	return string(hex.EncodeToString(w.hash.Sum(nil)))
}

func (r *HashReader) Read(p []byte) (int, error) {
	n, err := r.src.Read(p)
	if err != nil || n == 0 {
		return n, err
	}
	r.hash.Write(p[:n])
	return n, nil
}

func (w *HashWriter) Write(p []byte) (int, error) {
	n, err := w.dst.Write(p)
	if err != nil || n == 0 {
		return n, err
	}
	w.hash.Write(p[:n])
	return n, nil
}
