package router

import (
	"encoding/json"
	"net/http"
)

// LoginRequest определяет запрос на регистрацию/авторизацию пользователя.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse определяет ответ с данными для авторизации пользователя.
type LoginResponse struct {
	Token string `json:"token"`
}

// login обрабатывает входящие запросы на регистрацию/авторизацию пользователей.
func (router *Router) login(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}

	var login LoginRequest

	err := json.NewDecoder(r.Body).Decode(&login)
	if err != nil {
		router.log.Debug(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if login.Username == "" || login.Password == "" {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	token, err := router.storage.Register(r.Context(), login.Username, login.Password)
	if err != nil {
		router.log.Debug(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(LoginResponse{Token: token})
}
