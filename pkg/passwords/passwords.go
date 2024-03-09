package passwords

import (
	"golang.org/x/crypto/bcrypt"
)

const cost = bcrypt.MinCost

// Hash возвращает хеш пароля. Длина пароля должна составлять не более
// 72 символов.
func Hash(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	return string(hashedPassword), err
}

// Compare сравнивает пароль и его хеш.
func Compare(hashedPassword string, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
