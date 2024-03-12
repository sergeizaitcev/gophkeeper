package vault

import (
	"archive/tar"
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sergeizaitcev/gophkeeper/pkg/hashio"
	"github.com/sergeizaitcev/gophkeeper/pkg/workdir"
)

const (
	DirName     = ".gophkeeper" // Рабочая директория.
	DataDirName = "data"        // Директория с зашифрованными файлами.
)

type readWriter interface {
	io.ReaderFrom
	io.WriterTo
}

// Vault определяет хранилище зашифрованных файлов.
type Vault struct {
	root   workdir.Dir
	data   workdir.Dir
	remote Remote
	files  Files
}

var homedir = workdir.Home // для тестов.

// NewVault возвращает новый экземпляр Vault.
func NewVault() (*Vault, error) {
	root, err := homedir(DirName)
	if err != nil {
		return nil, err
	}

	files, err := root.Dir(DataDirName)
	if err != nil {
		return nil, err
	}

	v := &Vault{root: root, data: files}
	if err = v.init(); err != nil {
		return nil, err
	}

	return v, nil
}

func (v *Vault) init() error {
	saveOrLoad := func(name string, rw readWriter) error {
		if v.root.Exists(name) {
			return v.load(name, rw)
		}
		return v.save(name, rw)
	}

	errc := make(chan error, 2)

	go func() { errc <- saveOrLoad(FilesName, &v.files) }()
	go func() { errc <- saveOrLoad(RemoteName, &v.remote) }()

	for i := 0; i < 2; i++ {
		if err := <-errc; err != nil {
			return err
		}
	}

	return nil
}

// SetRemoteAddress устанавливает адрес удалённого сервера.
func (v *Vault) SetRemoteAddress(address string) error {
	v.remote.Address = address
	return v.save(RemoteName, v.remote)
}

// SetRemoteToken устанавливает токен авторизации на удалённом сервере.
func (v *Vault) SetRemoteToken(token string) error {
	v.remote.Token = token
	return v.save(RemoteName, v.remote)
}

// GetRemote возвращает данные для удалённого подключения.
func (v *Vault) GetRemote() Remote {
	return v.remote
}

// IsEmpty возвращает true, если зашифрованные файлы в хранилище отсутствуют.
func (v *Vault) IsEmpty() bool {
	return len(v.files) == 0
}

// Walk итеративно обходит зашифрованные файлы и вызывает fn.
func (v *Vault) Walk(fn func(File) error) error {
	for _, file := range v.files {
		if err := fn(file); err != nil {
			return err
		}
	}
	return nil
}

// Add добавляет зашифрованный файл в хранилище.
func (v *Vault) Add(description string, src io.Reader) error {
	return v.add(description, TypeBinary, src)
}

// AddLoginPassword добавляет зашифрованные данные банковской карты в хранилище.
func (v *Vault) AddBankCard(description string, card BankCard) error {
	data, err := card.MarshalBinary()
	if err != nil {
		return err
	}
	src := bytes.NewReader(data)
	return v.add(description, TypeCard, src)
}

// AddLoginPassword добавляет зашифрованные данные для авторизации пользователя
// в хранилище.
func (v *Vault) AddLoginPassword(description string, logpass UsernamePassword) error {
	data, err := logpass.MarshalBinary()
	if err != nil {
		return err
	}
	src := bytes.NewReader(data)
	return v.add(description, TypeLogpass, src)
}

func (v *Vault) add(description string, typ Type, src io.Reader) error {
	enc, err := newEncrypter(src)
	if err != nil {
		return err
	}

	id := generateID()

	f, err := v.data.Create(id)
	if err != nil {
		return err
	}
	defer f.Close()

	hw := hashio.NewHashWriter(f)
	buf := bufio.NewWriter(hw)

	if _, err = io.Copy(buf, enc); err != nil {
		return err
	}
	if err = buf.Flush(); err != nil {
		return err
	}

	file, i := v.files.Lookup(id)

	file.ID = id
	file.Type = typ
	file.Description = description
	file.SHA256 = hw.Checksum()
	file.Meta = enc.Meta()
	file.LastUpdate = time.Now().UTC()

	if i < 0 {
		v.files = append(v.files, file)
	} else {
		v.files[i] = file
	}

	return v.save(FilesName, v.files)
}

// Get возвращает дешифрованный файл по ID.
func (v *Vault) Get(id string) (rc io.ReadCloser, err error) {
	file, i := v.files.Lookup(id)
	if i < 0 {
		return nil, fmt.Errorf("%s not found", id)
	}
	if file.IsDeleted {
		return nil, fmt.Errorf("%s has been deleted", id)
	}

	f, err := v.data.Open(id)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = f.Close()
		}
	}()

	dec, err := newDecrypter(f, file.Meta)
	if err != nil {
		return nil, err
	}

	switch file.Type {
	case TypeCard:
		b, err := io.ReadAll(dec)
		if err != nil {
			return nil, err
		}

		var card BankCard

		if err = card.UnmarshalBinary(b); err != nil {
			return nil, err
		}
		if err = card.Validate(); err != nil {
			return nil, err
		}

		rc = io.NopCloser(strings.NewReader(card.String() + "\n"))
	case TypeLogpass:
		b, err := io.ReadAll(dec)
		if err != nil {
			return nil, err
		}

		var up UsernamePassword

		if err = up.UnmarshalBinary(b); err != nil {
			return nil, err
		}
		if err = up.Validate(); err != nil {
			return nil, err
		}

		rc = io.NopCloser(strings.NewReader(up.String() + "\n"))
	default:
		rc = &decryptCloser{
			Decrypter: dec,
			Closer:    f,
		}
	}

	return rc, nil
}

// Del удаляет зашифрованный файл из хранилища по id.
func (v *Vault) Del(id string) error {
	file, i := v.files.Lookup(id)
	if i < 0 {
		return fmt.Errorf("%s not found", id)
	}
	if file.IsDeleted {
		return fmt.Errorf("%s has been deleted", id)
	}

	if err := v.data.Remove(id); err != nil {
		return err
	}

	file.IsDeleted = true
	v.files[i] = file

	return v.save(FilesName, v.files)
}

// Clear очищает хранилище от лишних файлов.
func (v *Vault) Clear() error {
	return v.data.Walk(func(entry fs.DirEntry) error {
		if _, i := v.files.Lookup(entry.Name()); i >= 0 {
			return nil
		}
		return v.data.Remove(entry.Name())
	})
}

// Pack упаковывает содержимое хранилища в архив tar.
func (v *Vault) Pack() (*os.File, error) {
	temp, err := v.root.Temp("temp-*.tar")
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			name := temp.Name()
			_ = temp.Close()
			_ = os.RemoveAll(name)
		}
	}()

	buf := bufio.NewWriter(temp)
	tw := tar.NewWriter(buf)

	if err = v.packFiles(tw); err != nil {
		return nil, err
	}
	if err = v.packData(tw); err != nil {
		return nil, err
	}
	if err = tw.Close(); err != nil {
		return nil, err
	}
	if err = buf.Flush(); err != nil {
		return nil, err
	}
	if _, err = temp.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	return temp, nil
}

func (v *Vault) packFiles(tw *tar.Writer) error {
	files, err := v.root.Open(FilesName)
	if err != nil {
		return err
	}
	defer files.Close()

	return v.pack(tw, files, "")
}

func (v *Vault) packData(tw *tar.Writer) error {
	err := tw.WriteHeader(&tar.Header{
		Typeflag: tar.TypeDir,
		Name:     DataDirName,
		Mode:     int64(v.data.Mode()),
	})
	if err != nil {
		return err
	}

	return v.Walk(func(file File) error {
		if file.IsDeleted {
			return nil
		}

		f, err := v.data.Open(file.ID)
		if err != nil {
			return err
		}
		defer f.Close()

		return v.pack(tw, f, DataDirName)
	})
}

func (v *Vault) pack(tw *tar.Writer, f *os.File, baseDir string) error {
	info, err := f.Stat()
	if err != nil {
		return err
	}

	hdr, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return err
	}

	hdr.Name = filepath.Join(baseDir, filepath.Base(info.Name()))

	if err = tw.WriteHeader(hdr); err != nil {
		return err
	}
	if _, err = io.Copy(tw, f); err != nil {
		return err
	}

	return nil
}

// Unpack распаковывает содержимое архива tar в хранилище.
func (v *Vault) Unpack(src io.Reader) error {
	tr := tar.NewReader(src)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if hdr.Typeflag == tar.TypeDir {
			continue
		}

		if hdr.Name == FilesName {
			var fs Files
			if _, err = fs.ReadFrom(io.LimitReader(tr, hdr.Size)); err != nil {
				return err
			}
			v.files = fs.Merge(v.files)
			continue
		}

		if err = v.unpack(hdr, tr); err != nil {
			return err
		}
	}

	return v.save(FilesName, v.files)
}

func (v *Vault) unpack(hdr *tar.Header, tr *tar.Reader) error {
	file, i := v.files.Lookup(filepath.Base(hdr.Name))
	if i < 0 {
		return nil
	}

	f, err := v.data.Create(file.ID)
	if err != nil {
		return err
	}
	defer f.Close()

	buf := bufio.NewWriter(f)
	hw := hashio.NewHashWriter(buf)

	if _, err = io.CopyN(hw, tr, hdr.Size); err != nil {
		return err
	}
	if err = buf.Flush(); err != nil {
		return err
	}

	if file.SHA256 != hw.Checksum() {
		return fmt.Errorf("chechsum is invalid for %s", file.ID)
	}

	return nil
}

func (v *Vault) load(name string, r io.ReaderFrom) error {
	f, err := v.root.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err = r.ReadFrom(bufio.NewReader(f)); err != nil {
		return err
	}

	return nil
}

func (v *Vault) save(name string, w io.WriterTo) error {
	f, err := v.root.Create(name)
	if err != nil {
		return err
	}
	defer f.Close()

	buf := bufio.NewWriter(f)

	if _, err = w.WriteTo(buf); err != nil {
		return err
	}

	return buf.Flush()
}
