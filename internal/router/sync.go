package router

import (
	"archive/tar"
	"bufio"
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"log/slog"

	"github.com/sergeizaitcev/gophkeeper/internal/vault"
	"github.com/sergeizaitcev/gophkeeper/pkg/randutil"
	"github.com/sergeizaitcev/gophkeeper/pkg/workdir"
)

const usersDirName = "users" // Директория с пользовательскими данными.

// sync обрабатывает входящие запросы на синхронизацию данных пользователей.
func (router *Router) sync(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		router.syncData(w, r)
		return
	}

	ctx := r.Context()
	token, _ := ctx.Value(ctxToken).(string)

	s := newTarService(router.storage)

	src, exists, err := s.Get(ctx, token)
	if err != nil {
		router.log.Debug(err.Error(), slog.String("token", token))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	defer src.Close()

	w.Header().Set("Content-Type", "application/x-tar")
	w.WriteHeader(http.StatusOK)

	_, _ = io.Copy(w, src)
}

func (router *Router) syncData(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/x-tar" {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}

	ctx := r.Context()
	token, _ := ctx.Value(ctxToken).(string)

	s := newTarService(router.storage)

	src, created, err := s.Sync(ctx, token, r.Body)
	if err != nil {
		router.log.Debug(err.Error(), slog.String("token", token))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if created {
		w.WriteHeader(http.StatusCreated)
		return
	}
	defer src.Close()

	w.Header().Set("Content-Type", "application/x-tar")
	w.WriteHeader(http.StatusOK)

	_, _ = io.Copy(w, src)
}

type tarService struct {
	files Files
}

func newTarService(s Storage) *tarService {
	return &tarService{files: s}
}

func (s *tarService) Sync(ctx context.Context, token string, src io.Reader) (
	rc io.ReadCloser, created bool, err error,
) {
	path, err := s.files.Get(ctx, token)
	if err != nil {
		return nil, false, err
	}
	if path != "" {
		rc, err = merge(path, src)
		return rc, false, err
	}

	path, err = save(src)
	if err != nil {
		return nil, false, err
	}

	return nil, true, s.files.Save(ctx, token, path)
}

func (s *tarService) Get(ctx context.Context, token string) (io.ReadCloser, bool, error) {
	path, err := s.files.Get(ctx, token)
	if err != nil {
		return nil, false, err
	}
	if path == "" {
		return nil, false, nil
	}

	f, err := open(path)
	if err != nil {
		return nil, false, err
	}

	return f, true, nil
}

func merge(name string, src io.Reader) (io.ReadCloser, error) {
	f, err := os.OpenFile(name, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	current, replacement := tar.NewReader(bufio.NewReader(f)), tar.NewReader(src)

	files, err := mergeFiles(current, replacement)
	if err != nil {
		return nil, err
	}

	temp, err := os.CreateTemp(filepath.Dir(name), "temp-*.tar")
	if err != nil {
		return nil, err
	}
	defer temp.Close()

	buf := bufio.NewWriter(temp)
	dst := tar.NewWriter(buf)

	if err = writeFiles(dst, files); err != nil {
		return nil, err
	}
	if err = copyDataBy(dst, current, files, false); err != nil {
		return nil, err
	}
	if err = copyDataBy(dst, replacement, files, true); err != nil {
		return nil, err
	}

	if err = dst.Close(); err != nil {
		return nil, err
	}
	if err = buf.Flush(); err != nil {
		return nil, err
	}
	if err = switchFile(f, temp); err != nil {
		return nil, err
	}

	return os.OpenFile(name, os.O_RDONLY, 0)
}

func save(src io.Reader) (string, error) {
	f, err := create()
	if err != nil {
		return "", err
	}
	defer f.Close()

	buf := bufio.NewWriter(f)

	tw := tar.NewWriter(buf)
	tr := tar.NewReader(src)

	files, err := readFiles(tr)
	if err != nil {
		return "", err
	}
	if err = writeFiles(tw, files); err != nil {
		return "", err
	}
	if err = copyDataBy(tw, tr, files, false); err != nil {
		return "", err
	}
	if err = tw.Close(); err != nil {
		return "", err
	}
	if err = buf.Flush(); err != nil {
		return "", err
	}

	return f.Name(), nil
}

func switchFile(dst, src *os.File) error {
	newpath, oldpath := dst.Name(), src.Name()
	if err := src.Close(); err != nil {
		return err
	}
	if err := dst.Close(); err != nil {
		return err
	}
	return os.Rename(oldpath, newpath)
}

func mergeFiles(current, replacement *tar.Reader) (vault.Files, error) {
	currentFiles, err := readFiles(current)
	if err != nil {
		return nil, err
	}
	replacementFiles, err := readFiles(replacement)
	if err != nil {
		return nil, err
	}
	return replacementFiles.Merge(currentFiles), nil
}

func readFiles(tr *tar.Reader) (vault.Files, error) {
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if hdr.Name == vault.FilesName {
			var fs vault.Files
			if _, err = fs.ReadFrom(io.LimitReader(tr, hdr.Size)); err != nil {
				return nil, err
			}
			return fs, nil
		}
	}
	return nil, ErrNotFound
}

func writeFiles(tw *tar.Writer, files vault.Files) error {
	var buf bytes.Buffer

	n, err := files.WriteTo(&buf)
	if err != nil {
		return err
	}

	hdr := &tar.Header{
		Name:     vault.FilesName,
		Typeflag: tar.TypeReg,
		Size:     n,
		Mode:     workdir.FileMode,
	}

	if err = tw.WriteHeader(hdr); err != nil {
		return err
	}
	if _, err = buf.WriteTo(tw); err != nil {
		return err
	}

	return nil
}

func copyDataBy(dst *tar.Writer, src *tar.Reader, files vault.Files, skipDir bool) error {
	for {
		hdr, err := src.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if hdr.Typeflag == tar.TypeDir {
			if !skipDir {
				if err = dst.WriteHeader(hdr); err != nil {
					return err
				}
			}
			continue
		}

		file, i := files.Lookup(filepath.Base(hdr.Name))
		if i < 0 || file.IsDeleted {
			continue
		}

		if err = dst.WriteHeader(hdr); err != nil {
			return err
		}
		if _, _ = io.CopyN(dst, src, hdr.Size); err != nil {
			return err
		}
	}
	return nil
}

func create() (*os.File, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	dir, err := workdir.Dir(pwd).Dir(usersDirName)
	if err != nil {
		return nil, err
	}

	return dir.Create(randutil.Hex(32))
}

func open(path string) (*os.File, error) {
	dir, filename := filepath.Dir(path), filepath.Base(path)
	return workdir.Dir(dir).Open(filename)
}
