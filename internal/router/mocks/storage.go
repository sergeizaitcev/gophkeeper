package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockStorage struct {
	mock.Mock
	Filepath string
}

func NewMockStorage() *MockStorage {
	return new(MockStorage)
}

func (m *MockStorage) Register(ctx context.Context, username, password string) (string, error) {
	args := m.Called(ctx, username, password)
	return args.String(0), args.Error(1)
}

func (m *MockStorage) Check(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockStorage) Get(ctx context.Context, token string) (string, error) {
	args := m.Called(ctx, token)
	return args.String(0), args.Error(1)
}

func (m *MockStorage) Save(ctx context.Context, token string, filepath string) error {
	args := m.Called(ctx, token, filepath)
	m.Filepath = filepath
	return args.Error(0)
}
