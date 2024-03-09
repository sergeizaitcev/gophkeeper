package passwords_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sergeizaitcev/gophkeeper/pkg/passwords"
	"github.com/sergeizaitcev/gophkeeper/pkg/randutil"
)

func TestPasswords(t *testing.T) {
	testCases := []struct {
		name      string
		password  string
		wantError bool
	}{
		{
			name:      "small",
			password:  randutil.String(10),
			wantError: false,
		},
		{
			name:      "medium",
			password:  randutil.String(50),
			wantError: false,
		},
		{
			name:      "large",
			password:  randutil.String(73),
			wantError: true,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			hashedPassword, err := passwords.Hash(tc.password)

			if tc.wantError {
				require.Error(t, err)
			} else {
				require.True(t, passwords.Compare(hashedPassword, tc.password))
			}
		})
	}
}
