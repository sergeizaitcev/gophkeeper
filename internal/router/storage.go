package router

import (
	"context"
	"errors"
)

// ErrNotFound возвращается, когда данные не найдены.
var ErrNotFound = errors.New("not found")

// Storage описывает интерфейс хранилища gophkeeper.
type Storage interface {
	Users
	Files
}

// Users описывает интерфейс хранилища пользователей.
type Users interface {
	// Register регистрирует или авторизовывает пользователя и возвращает токен.
	Register(ctx context.Context, username, password string) (string, error)

	// Check проверяет наличие токена в хранилище.
	Check(ctx context.Context, token string) error
}

// Files описывает интерфейс хранилища файлов пользователей.
type Files interface {
	// Save сохраняет путь к файлу пользователя по токену.
	Save(ctx context.Context, token, filepath string) error

	// Get возвращает путь к файлу пользователя по токену.
	Get(ctx context.Context, token string) (string, error)
}
