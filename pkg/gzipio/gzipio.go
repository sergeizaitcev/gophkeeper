package gzipio

import (
	"compress/gzip"
	"io"
	"net/http"
)

// NewCompressingReader возвращает средство чтения, которое считывает данные
// из src после того, как они будут сжаты.
func NewCompressingReader(src io.Reader) io.Reader {
	pr, pw := io.Pipe()
	go func() {
		gw := gzip.NewWriter(pw)
		_, _ = io.Copy(gw, src)
		gw.Close()
		pw.Close()
	}()
	return pr
}

var _ io.ReadCloser = (*DecompressingReader)(nil)

// DecompressingReader определяет средство чтения, которое считывает данные из
// src после того, как они будут разжаты.
type DecompressingReader struct {
	src io.Reader
	gr  *gzip.Reader
}

// NewDecompressingReader возвращает новый экземпляр DecompressingReader.
func NewDecompressingReader(src io.Reader) (*DecompressingReader, error) {
	gr, err := gzip.NewReader(src)
	if err != nil {
		return nil, err
	}
	return &DecompressingReader{src: src, gr: gr}, nil
}

func (r *DecompressingReader) Read(p []byte) (int, error) {
	return r.gr.Read(p)
}

func (r *DecompressingReader) Close() error {
	var firstErr error
	if closer, ok := r.src.(io.Closer); ok {
		firstErr = closer.Close()
	}
	if err := r.gr.Close(); err != nil && firstErr == nil {
		firstErr = err
	}
	return firstErr
}

var _ http.ResponseWriter = (*ResponseWriter)(nil)

// ResponseWriter возвращает средство записи, которое записывает данные
// в http.ResponseWriter после того, как они будут сжаты.
type ResponseWriter struct {
	http.ResponseWriter
	gw *gzip.Writer
}

// NewResponseWriter возвращает новый экземпляр ResponseWriter.
func NewResponseWriter(dst http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{
		ResponseWriter: dst,
		gw:             gzip.NewWriter(dst),
	}
}

func (w *ResponseWriter) Write(p []byte) (int, error) {
	return w.gw.Write(p)
}

func (w *ResponseWriter) Close() error {
	return w.gw.Close()
}
