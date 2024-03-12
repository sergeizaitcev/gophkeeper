package vault

import (
	"encoding/json"
	"io"
	"sort"
	"time"

	"github.com/sergeizaitcev/gophkeeper/pkg/cryptio"
)

const (
	FilesName  = "files"  // Наименование файла для files.
	RemoteName = "remote" // Наименование файла для remote.
)

type counter struct {
	io.Writer
	io.Reader
	n int
}

func (c *counter) Write(p []byte) (int, error) {
	n, err := c.Writer.Write(p)
	c.n += n
	return n, err
}

func (c *counter) Read(p []byte) (int, error) {
	n, err := c.Reader.Read(p)
	c.n += n
	return n, err
}

var (
	_ io.WriterTo   = (*Remote)(nil)
	_ io.ReaderFrom = (*Remote)(nil)
)

// Remote определяет конфигурацию удалённого подключения.
type Remote struct {
	Address string `json:"address"` // Адрес удалённого сервера.
	Token   string `json:"token"`   // Токен авторизации.
}

func (r *Remote) ReadFrom(src io.Reader) (int64, error) {
	c := &counter{Reader: src}
	err := json.NewDecoder(c).Decode(r)
	return int64(c.n), err
}

func (r Remote) WriteTo(dst io.Writer) (int64, error) {
	c := &counter{Writer: dst}
	enc := json.NewEncoder(c)
	enc.SetIndent("", "  ")
	err := enc.Encode(&r)
	return int64(c.n), err
}

// File определяет конфигурацию файла с зашифрованными данными.
type File struct {
	ID          string       `json:"id"`          // Уникальный идентификатор.
	Type        Type         `json:"type"`        // Тип зашифрованных данных.
	Description string       `json:"description"` // Описание данных.
	SHA256      string       `json:"sha256"`      // Хеш-строка.
	Meta        cryptio.Meta `json:"meta"`        // Метаданные.
	LastUpdate  time.Time    `json:"last_update"` // Последнее изменение файла.
	IsDeleted   bool         `json:"is_deleted"`  // Флаг удаления.
}

// After возвращает true, если дата последнего изменения файла позже чем в x.
func (f File) After(x File) bool {
	return f.LastUpdate.After(x.LastUpdate)
}

var (
	_ sort.Interface = (*Files)(nil)
	_ io.ReaderFrom  = (*Files)(nil)
	_ io.WriterTo    = (*Files)(nil)
)

// Files определяет конфигурацию файлов с зашифрованными данными.
type Files []File

func (fs Files) Len() int           { return len(fs) }
func (fs Files) Less(i, j int) bool { return fs[i].LastUpdate.After(fs[j].LastUpdate) }
func (fs Files) Swap(i, j int)      { fs[i], fs[j] = fs[j], fs[i] }

func (fs *Files) ReadFrom(src io.Reader) (int64, error) {
	c := &counter{Reader: src}
	err := json.NewDecoder(c).Decode(fs)
	return int64(c.n), err
}

func (fs Files) WriteTo(dst io.Writer) (int64, error) {
	c := &counter{Writer: dst}
	enc := json.NewEncoder(c)
	enc.SetIndent("", "  ")
	err := enc.Encode(fs)
	return int64(c.n), err
}

// Lookup выполняет поиск зашифрованного файла по ID.
func (fs Files) Lookup(id string) (File, int) {
	for i, file := range fs {
		if file.ID == id {
			return file, i
		}
	}
	return File{}, -1
}

// Clone возвращает полную копию Files.
func (fs Files) Clone() Files {
	fs2 := make(Files, len(fs))
	copy(fs2, fs)
	return fs2
}

// Merge объединяет конфигурацию файлов в одну и возвращает её.
func (fs Files) Merge(x Files) Files {
	if len(fs) == 0 {
		return x.Clone()
	}
	if len(x) == 0 {
		return fs.Clone()
	}
	if len(x) < len(fs) {
		x, fs = fs, x
	}

	merged := make(Files, 0, len(fs)+len(x))

	set := make(map[string]int, len(fs))
	for i, file := range fs {
		set[file.ID] = i
	}

	for _, file := range x {
		i, ok := set[file.ID]
		if !ok {
			if !file.IsDeleted {
				merged = append(merged, file)
			}
			continue
		}

		matched := fs[i]

		if file.LastUpdate.After(matched.LastUpdate) {
			matched = file
		}
		if !matched.IsDeleted {
			merged = append(merged, matched)
		}
	}

	if len(merged) > 0 {
		sort.Sort(merged)
		merged = merged[:len(merged):len(merged)]
	}

	return merged
}
