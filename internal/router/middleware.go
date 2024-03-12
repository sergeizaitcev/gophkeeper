package router

import (
	"context"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"log/slog"

	"github.com/sergeizaitcev/gophkeeper/pkg/gzipio"
)

// ctxToken определяет ключ для передачи токена пользователя через контекст.
var ctxToken struct{}

type middleware = func(http.HandlerFunc) http.HandlerFunc

// use оборачивает next всеми промежуточными обработчиками и возвращает его.
func use(next http.HandlerFunc, mws ...middleware) http.HandlerFunc {
	for _, mw := range mws {
		next = mw(next)
	}
	return next
}

// auth проверяет наличие токена авторизации в запросе и передает его через
// контекст.
func (router *Router) auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		token := r.Header.Get("Authorization")
		if err := router.storage.Check(ctx, token); err != nil {
			router.log.Debug(err.Error(), slog.String("token", token))
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		r = r.WithContext(context.WithValue(ctx, ctxToken, token))
		next(w, r)
	}
}

// compress сжимает входящий и исходящий контент формата gzip.
func (router *Router) compress(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		encoding := r.Header.Get("Accept-Encoding")
		decoding := r.Header.Get("Content-Encoding")

		if encoding != "" && decoding != "" {
			if !strings.Contains(encoding, "gzip") || !strings.Contains(decoding, "gzip") {
				w.WriteHeader(http.StatusUnsupportedMediaType)
				return
			}

			dc, err := gzipio.NewDecompressingReader(r.Body)
			if err != nil {
				router.log.Debug(err.Error())
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			rw := gzipio.NewResponseWriter(w)
			defer rw.Close()

			rw.Header().Set("Content-Encoding", "gzip")
			r.Body = dc

			next(rw, r)

			return
		}

		if encoding != "" {
			if !strings.Contains(encoding, "gzip") {
				w.WriteHeader(http.StatusUnsupportedMediaType)
				return
			}

			rw := gzipio.NewResponseWriter(w)
			defer rw.Close()

			rw.Header().Set("Content-Encoding", "gzip")
			next(rw, r)

			return
		}

		if !strings.Contains(decoding, "gzip") {
			w.WriteHeader(http.StatusUnsupportedMediaType)
			return
		}

		dc, err := gzipio.NewDecompressingReader(r.Body)
		if err != nil {
			router.log.Debug(err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		r.Body = dc

		next(w, r)
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.statusCode = statusCode
}

// logger логирует результат обработки HTTP-запроса.
func (router *Router) logger(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		start := time.Now()

		next(rw, r)

		elapsed := time.Since(start)
		token, _ := r.Context().Value(ctxToken).(string)

		router.log.Info("",
			slog.String("token", hideToken(token)),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status_code", rw.statusCode),
			slog.Duration("elapsed", elapsed),
		)
	}
}

// hideToken скрывает первую половину токена символами '*'.
func hideToken(token string) string {
	hidden := make([]rune, 0, utf8.RuneCount([]byte(token)))
	for i, ch := range token {
		if i < len(token)/2 {
			ch = '*'
		}
		hidden = append(hidden, ch)
	}
	return string(hidden)
}

// post устанавливает поддержку POST-запросов для next.
func post(next http.HandlerFunc) http.HandlerFunc {
	return allowed(http.MethodPost)(next)
}

// allowed устанавливает поддержку запроса для next.
func allowed(methods ...string) func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			var n int
			for _, method := range methods {
				if r.Method == method {
					n++
				}
			}
			if n == 0 {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			next(w, r)
		}
	}
}
