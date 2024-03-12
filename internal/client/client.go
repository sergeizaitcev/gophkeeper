package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/sergeizaitcev/gophkeeper/internal/router"
	"github.com/sergeizaitcev/gophkeeper/pkg/gzipio"
)

// Client определяет клиент для gophkeeper.
type Client struct {
	addr   string
	client *http.Client
}

// New возвращает новый экземпляр Client.
func New(addr string) *Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return &Client{
		addr: addr,
		client: &http.Client{
			Transport: transport,
			Timeout:   5 * time.Second,
		},
	}
}

// Login выполняет регистрацию/авторизацию на удалённом сервере.
func (c *Client) Login(ctx context.Context, username, password string) (string, error) {
	b, err := json.Marshal(router.LoginRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		return "", err
	}

	uri := url.URL{
		Scheme: "https",
		Host:   c.addr,
		Path:   "/api/v1/login",
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri.String(), bytes.NewReader(b))
	if err != nil {
		return "", err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	res, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() {
		_, _ = io.Copy(io.Discard, res.Body)
		_ = res.Body.Close()
	}()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	var login router.LoginResponse
	if err = json.NewDecoder(res.Body).Decode(&login); err != nil {
		return "", err
	}

	return login.Token, nil
}

// GetData возвращает данные из удалённого репозитория.
func (c *Client) GetData(ctx context.Context, token string) (rc io.ReadCloser, err error) {
	uri := url.URL{
		Scheme: "https",
		Host:   c.addr,
		Path:   "/api/v1/sync",
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri.String(), http.NoBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", token)
	req.Header.Set("Accept", "application/x-tar")
	req.Header.Set("Accept-Encoding", "gzip")

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_, _ = io.Copy(io.Discard, res.Body)
			_ = res.Body.Close()
		}
	}()

	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	return gzipio.NewDecompressingReader(res.Body)
}

// SyncData синхронизирует данные в хранилище с данными из удалённого репозитория.
func (c *Client) SyncData(
	ctx context.Context,
	token string,
	src io.Reader,
) (io.ReadCloser, error) {
	uri := url.URL{
		Scheme: "https",
		Host:   c.addr,
		Path:   "/api/v1/sync",
	}

	src = gzipio.NewCompressingReader(src)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri.String(), src)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", token)
	req.Header.Set("Accept", "application/x-tar")
	req.Header.Set("Content-Type", "application/x-tar")
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Content-Encoding", "gzip")

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_, _ = io.Copy(io.Discard, res.Body)
			_ = res.Body.Close()
		}
	}()

	if res.StatusCode == http.StatusCreated {
		return nil, nil
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	return gzipio.NewDecompressingReader(res.Body)
}
