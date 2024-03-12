package router

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"log/slog"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/sergeizaitcev/gophkeeper/internal/router/mocks"
	"github.com/sergeizaitcev/gophkeeper/pkg/randutil"
)

func TestLogin(t *testing.T) {
	storage := mocks.NewMockStorage()

	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	slog.SetDefault(log)

	testRegister(t, storage, randutil.String(32), randutil.String(32))
}

func testRegister(t *testing.T, storage *mocks.MockStorage, username, password string) {
	data := LoginRequest{
		Username: username,
		Password: password,
	}

	b, err := json.Marshal(data)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/login", bytes.NewReader(b))

	req.Header.Add("Accep", "application/json")
	req.Header.Add("Content-Type", "application/json")

	storage.On("Register", mock.Anything, username, password).Return("", nil)

	New(storage, slog.Default()).ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}
