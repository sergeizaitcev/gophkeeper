package vault

import (
	"encoding/binary"
	"errors"
)

// Type определяет тип защищаемой информации.
type Type int8

const (
	TypeUnknown Type = iota

	TypeBinary
	TypeCard
	TypeLogpass
)

var typeValues = []string{
	"UNKNOWN",
	"BINARY",
	"CARD",
	"LOGPASS",
}

func (t Type) String() string {
	if int(t) < len(typeValues) {
		return typeValues[t]
	}
	return typeValues[0]
}

// BankCard определяет номер банковской карты.
type BankCard [16]byte

// NewBankCard конвертирует number в BankCard.
func NewBankCard(number string) BankCard {
	var c BankCard
	var n int
	for i := 0; i < len(number); i++ {
		v := number[i]
		if v >= '0' && v <= '9' && n < len(c) {
			c[n] = v
			n++
		}
	}
	return c
}

func (c BankCard) String() string {
	if c.Validate() != nil {
		return "<invalid>"
	}
	return string(c[0:4]) + " " + string(c[4:8]) + " " + string(c[8:12]) + " " + string(c[12:16])
}

// Validate проверяет по алгоритму Луна контрольную сумму номера банковской
// карты.
func (c BankCard) Validate() error {
	var sum int
	parity := len(c) % 2
	for i := 0; i < len(c); i++ {
		digit := int(c[i] - '0')
		if i%2 == parity {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
	}
	if sum%10 != 0 {
		return errors.New("bank card number is invalid")
	}
	return nil
}

func (c BankCard) MarshalBinary() ([]byte, error) {
	return c[:], nil
}

func (c *BankCard) UnmarshalBinary(data []byte) error {
	p := *c
	if len(data) != len(p) {
		return errors.New("data is invalid")
	}
	copy(p[:], data)
	*c = p
	return nil
}

// UsernamePassword определяет данные для авторизации пользователя.
type UsernamePassword struct {
	Username string
	Password string
}

// NewUsernamePassword конвертирует данные для авrоризации в UsernamePassword.
func NewUsernamePassword(login, password string) UsernamePassword {
	return UsernamePassword{Username: login, Password: password}
}

func (up UsernamePassword) String() string {
	if up.Validate() != nil {
		return "<invalid>"
	}
	return up.Username + ":" + up.Password
}

// Validate возвращает ошибку, если данные для авторизации не валидны.
func (up UsernamePassword) Validate() error {
	if up.Username == "" {
		return errors.New("username must not be blank")
	}
	if up.Password == "" {
		return errors.New("password must not be blank")
	}
	return nil
}

func (up UsernamePassword) MarshalBinary() ([]byte, error) {
	b := make([]byte, 0, len(up.Username)+len(up.Password)+4+4)
	b = binary.BigEndian.AppendUint32(b, uint32(len(up.Username)))
	b = append(b, up.Username...)
	b = binary.BigEndian.AppendUint32(b, uint32(len(up.Password)))
	b = append(b, up.Password...)
	return b, nil
}

func (up *UsernamePassword) UnmarshalBinary(data []byte) error {
	if len(data) < 8 {
		return errors.New("data is too short")
	}

	n := binary.BigEndian.Uint32(data[:4])
	data = data[4:]

	if len(data) < int(n) {
		return errors.New("login is corrupted")
	}

	p := *up
	p.Username = string(data[:n])
	data = data[n:]

	if len(data) < 4 {
		return errors.New("data is corrupted")
	}

	n = binary.BigEndian.Uint32(data[:4])
	data = data[4:]

	if len(data) < int(n) {
		return errors.New("password is corrupted")
	}

	p.Password = string(data[:n])
	*up = p

	return nil
}
