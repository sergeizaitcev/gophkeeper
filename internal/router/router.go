package router

import (
	"net/http"

	"log/slog"
)

// Router определяет обработчик HTTP-запросов, принимающий запросы от клиентов
// gophkeeper.
type Router struct {
	mux     *http.ServeMux
	storage Storage
	log     *slog.Logger
}

// New возвращает новый обработчик HTTP-запросов.
func New(s Storage, log *slog.Logger) *Router {
	r := &Router{
		mux:     http.NewServeMux(),
		storage: s,
		log:     log,
	}
	r.initHandlers()
	return r
}

// ServeHTTP реализует интерфейс http.Handler.
func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router.mux.ServeHTTP(w, r)
}

func (router *Router) initHandlers() {
	login := use(router.login, post, router.logger)
	sync := use(
		router.sync,
		allowed(http.MethodPost, http.MethodGet),
		router.compress,
		router.logger,
		router.auth,
	)

	router.mux.HandleFunc("/api/v1/login", login)
	router.mux.HandleFunc("/api/v1/sync", sync)
}
