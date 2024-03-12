package router

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"log/slog"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/sergeizaitcev/gophkeeper/internal/router/mocks"
	"github.com/sergeizaitcev/gophkeeper/pkg/gzipio"
	"github.com/sergeizaitcev/gophkeeper/pkg/randutil"
)

var (
	tdBefore = filepath.Join("testdata", "before.tar")
	tdAfter  = filepath.Join("testdata", "after.tar")
)

func TestSync(t *testing.T) {
	storage := mocks.NewMockStorage()

	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	slog.SetDefault(log)

	testCreated(t, storage, randutil.Hex(16))
	testMerge(t, storage, randutil.Hex(16))
}

func testCreated(t *testing.T, storage *mocks.MockStorage, token string) {
	f, err := os.OpenFile(tdBefore, os.O_RDONLY, 0)
	require.NoError(t, err)
	defer f.Close()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/sync", gzipio.NewCompressingReader(f))

	req.Header.Add("Authorization", token)
	req.Header.Add("Accept-Encoding", "gzip")
	req.Header.Add("Content-Encoding", "gzip")
	req.Header.Add("Content-Type", "application/x-tar")

	storage.On("Check", mock.Anything, token).Return(nil)
	storage.On("Get", mock.Anything, token).Return("", nil)
	storage.On("Save", mock.Anything, token, mock.Anything).Return(nil)

	New(storage, slog.Default()).ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)
}

func testMerge(t *testing.T, storage *mocks.MockStorage, token string) {
	f, err := os.OpenFile(tdAfter, os.O_RDONLY, 0)
	require.NoError(t, err)
	defer f.Close()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/sync", gzipio.NewCompressingReader(f))

	req.Header.Add("Authorization", token)
	req.Header.Add("Accept-Encoding", "gzip")
	req.Header.Add("Content-Encoding", "gzip")
	req.Header.Add("Content-Type", "application/x-tar")

	storage.On("Check", mock.Anything, token).Return(nil)
	storage.On("Get", mock.Anything, token).Return(storage.Filepath, nil)

	New(storage, slog.Default()).ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}
