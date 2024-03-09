package workdir

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

const (
	DirMode  = 0o700 // Права доступа для директории.
	FileMode = 0o600 // Права доступа для файла.
)

// Dir определяет директорию.
type Dir string

// Home создает/открывает директорию в домашнем каталоге и возвращает её.
func Home(dir string) (Dir, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("search for the home directory: %w", err)
	}
	path := filepath.Join(homedir, dir)
	return Dir(path), mkdir(path)
}

// Dir создает/открывает поддиректорию и возвращает её.
func (d Dir) Dir(name string) (Dir, error) {
	path := filepath.Join(string(d), name)
	return Dir(path), mkdir(path)
}

// mkdir создает директорию, если её не существует.
func mkdir(dir string) error {
	if _, err := os.Stat(dir); err == nil {
		return nil
	}
	if err := os.Mkdir(dir, DirMode); err != nil {
		return fmt.Errorf("create a directory: %w", err)
	}
	return nil
}

// Mode возвращает права доступа директории.
func (d Dir) Mode() fs.FileMode {
	stat, err := os.Stat(string(d))
	if err != nil {
		return 0
	}
	return stat.Mode()
}

// Exists возвращает true, если файл или поддиректория существует.
func (d Dir) Exists(name string) bool {
	path := filepath.Join(string(d), name)
	_, err := os.Stat(path)
	return err == nil
}

// Walk обходит все файлы в директории и вызывает fn.
func (d Dir) Walk(fn func(entry fs.DirEntry) error) error {
	return filepath.WalkDir(string(d), func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return filepath.SkipDir
		}
		return fn(d)
	})
}

// Temp создает временный файл в директории и возвращает его.
func (d Dir) Temp(pattern string) (*os.File, error) {
	return os.CreateTemp(string(d), pattern)
}

// Open открывает файл в режиме чтения и возвращает его.
func (d Dir) Open(name string) (*os.File, error) {
	filename := filepath.Join(string(d), name)
	return os.OpenFile(filename, os.O_RDONLY, 0)
}

// Create создает файл в режиме записи и возвращает его; если файл уже
// существует, то он будет перезаписан.
func (d Dir) Create(name string) (*os.File, error) {
	filename := filepath.Join(string(d), name)
	return os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, FileMode)
}

// Remove удаляет файл или поддиректорию.
func (d Dir) Remove(name string) error {
	path := filepath.Join(string(d), name)
	return os.RemoveAll(path)
}
