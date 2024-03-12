package vault

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBankCard_Validate(t *testing.T) {
	testCases := []struct {
		number string
		valid  bool
	}{
		{"4720-4755-3562-9559", true},
		{"4721-4755-3562-9559", false},
	}

	for _, tc := range testCases {
		t.Run(tc.number, func(t *testing.T) {
			card := NewBankCard(tc.number)
			require.Equal(t, tc.valid, card.Validate() == nil)
		})
	}
}

func TestUsernamePassword_Validate(t *testing.T) {
	testCases := []struct {
		login    string
		password string
		valid    bool
	}{
		{"login", "pass", true},
		{"login", "", false},
		{"", "pass", false},
		{"", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.login+":"+tc.password, func(t *testing.T) {
			logpass := NewUsernamePassword(tc.login, tc.password)
			require.Equal(t, tc.valid, logpass.Validate() == nil)
		})
	}
}
