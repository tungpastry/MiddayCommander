package local

import (
	"context"
	"io"
	iofs "io/fs"
	"os"
	"path/filepath"
	"strings"

	midfs "github.com/tungpastry/MiddayCommander/internal/fs"
)

// FS triển khai một midfs.FileSystem cho hệ thống tệp của hệ điều hành cục bộ.
type FS struct{}

// New trả về một FileSystem cục bộ mới.
func New() *FS {
	return &FS{}
}

// ID trả về định danh duy nhất cho loại hệ thống tệp này.
func (f *FS) ID() string {
	return "local"
}

// Scheme trả về lược đồ URI cho hệ thống tệp này.
func (f *FS) Scheme() midfs.Scheme {
	return midfs.SchemeFile
}

// Capabilities trả về tập hợp các hoạt động được hỗ trợ cho hệ thống tệp cục bộ.
func (f *FS) Capabilities() uint64 {
	return midfs.CapList | midfs.CapRead | midfs.CapWrite | midfs.CapMkdir | midfs.CapRename | midfs.CapRemove
}

// List đọc thư mục được đặt tên bởi URI và trả về một danh sách các mục trong thư mục.
func (f *FS) List(ctx context.Context, dir midfs.URI) ([]midfs.Entry, error) {
	dirEntries, err := os.ReadDir(dir.Path)
	if err != nil {
		return nil, err
	}

	entries := make([]midfs.Entry, 0, len(dirEntries))
	for _, dirEntry := range dirEntries {
		info, err := dirEntry.Info()
		if err != nil {
			return nil, err
		}
		entryURI := midfs.NewFileURI(filepath.Join(dir.Path, dirEntry.Name()))
		entries = append(entries, entryFromInfo(dirEntry.Name(), entryURI, info))
	}
	return entries, nil
}

// Stat trả về một midfs.Entry mô tả tệp hoặc thư mục được đặt tên.
// Nó sử dụng os.Lstat để tránh theo các liên kết tượng trưng.
func (f *FS) Stat(ctx context.Context, uri midfs.URI) (midfs.Entry, error) {
	info, err := os.Lstat(uri.Path)
	if err != nil {
		return midfs.Entry{}, err
	}
	return entryFromInfo(filepath.Base(uri.Path), uri, info), nil
}

// Mkdir tạo một thư mục mới với tên và các bit quyền được chỉ định.
func (f *FS) Mkdir(ctx context.Context, uri midfs.URI, perm os.FileMode) error {
	return os.Mkdir(uri.Path, perm)
}

// Rename đổi tên (di chuyển) một tệp hoặc thư mục.
func (f *FS) Rename(ctx context.Context, from midfs.URI, to midfs.URI) error {
	return os.Rename(from.Path, to.Path)
}

// Remove xóa tệp được đặt tên hoặc thư mục (nếu recursive là true).
func (f *FS) Remove(ctx context.Context, uri midfs.URI, recursive bool) error {
	if recursive {
		return os.RemoveAll(uri.Path)
	}
	return os.Remove(uri.Path)
}

// OpenReader mở tệp được đặt tên để đọc. Trình đọc được trả về hỗ trợ
// tìm kiếm đến một vị trí bù (offset).
func (f *FS) OpenReader(ctx context.Context, uri midfs.URI, opts midfs.OpenReadOptions) (io.ReadCloser, error) {
	file, err := os.Open(uri.Path)
	if err != nil {
		return nil, err
	}
	if opts.Offset <= 0 {
		return file, nil
	}
	if _, err := file.Seek(opts.Offset, io.SeekStart); err != nil {
		file.Close()
		return nil, err
	}
	return file, nil
}

// OpenWriter mở tệp được đặt tên để ghi.
// Nó hỗ trợ ghi nguyên tử bằng cách ghi vào một tệp tạm thời và sau đó đổi tên nó
// khi đóng. Nó cũng xử lý quyền tệp, ghi đè và tạo độc quyền.
func (f *FS) OpenWriter(ctx context.Context, uri midfs.URI, opts midfs.OpenWriteOptions) (io.WriteCloser, error) {
	perm := opts.Perm
	if perm == 0 {
		perm = 0o644
	}
	if opts.Atomic {
		tempPath := uri.Path + opts.TempExtension
		if opts.TempExtension == "" {
			tempPath = uri.Path + ".tmp"
		}
		file, err := os.OpenFile(tempPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
		if err != nil {
			return nil, err
		}
		if opts.Offset > 0 {
			if _, err := file.Seek(opts.Offset, io.SeekStart); err != nil {
				file.Close()
				_ = os.Remove(tempPath)
				return nil, err
			}
		}
		return &atomicWriteCloser{
			File:      file,
			target:    uri.Path,
			tempPath:  tempPath,
			overwrite: opts.Overwrite,
		}, nil
	}

	flags := os.O_CREATE | os.O_WRONLY
	switch {
	case opts.Offset > 0:
	case opts.Overwrite:
		flags |= os.O_TRUNC
	default:
		flags |= os.O_EXCL
	}

	file, err := os.OpenFile(uri.Path, flags, perm)
	if err != nil {
		return nil, err
	}
	if opts.Offset > 0 {
		if _, err := file.Seek(opts.Offset, io.SeekStart); err != nil {
			file.Close()
			return nil, err
		}
	}
	return file, nil
}

// Join nối bất kỳ số lượng phần tử đường dẫn nào thành một URI duy nhất, phân tách chúng
// bằng một Dấu phân cách cụ thể của HĐH.
func (f *FS) Join(base midfs.URI, elems ...string) midfs.URI {
	parts := []string{base.Path}
	parts = append(parts, elems...)
	return midfs.NewFileURI(filepath.Join(parts...))
}

// Parent trả về URI của thư mục cha.
func (f *FS) Parent(uri midfs.URI) midfs.URI {
	cleanPath := filepath.Clean(uri.Path)
	if midfs.IsRootFilePath(cleanPath) {
		return midfs.NewFileURI(cleanPath)
	}
	return midfs.NewFileURI(filepath.Dir(cleanPath))
}

// Clean trả về tên đường dẫn ngắn nhất tương đương với đường dẫn bằng cách xử lý
// hoàn toàn từ vựng. Nó cũng xóa mọi trường không liên quan đến đường dẫn khỏi URI.
func (f *FS) Clean(uri midfs.URI) midfs.URI {
	cleanURI := uri.Clone()
	cleanURI.Scheme = midfs.SchemeFile
	cleanURI.Path = filepath.Clean(uri.Path)
	cleanURI.Host = ""
	cleanURI.Port = 0
	cleanURI.User = ""
	cleanURI.Query = nil
	return cleanURI
}

// Close là một thao tác không làm gì cho hệ thống tệp cục bộ vì không có
// kết nối liên tục nào để quản lý.
func (f *FS) Close() error {
	return nil
}

// atomicWriteCloser bao bọc một *os.File để cung cấp các thao tác ghi nguyên tử. Nó ghi vào một
// tệp tạm thời và đổi tên nó thành đường dẫn đích khi Close() thành công.
type atomicWriteCloser struct {
	*os.File
	target    string
	tempPath  string
	overwrite bool
}

// Discard đóng và xóa tệp tạm thời, hủy bỏ thao tác ghi.
func (w *atomicWriteCloser) Discard() error {
	if w.File != nil {
		_ = w.File.Close()
	}
	return os.Remove(w.tempPath)
}

// Close hoàn tất việc ghi nguyên tử. Nó đóng tệp tạm thời và đổi tên nó
// thành đường dẫn đích cuối cùng. Nếu overwrite là false, nó sẽ trả về lỗi nếu
// đích đã tồn tại.
func (w *atomicWriteCloser) Close() error {
	if err := w.File.Close(); err != nil {
		_ = os.Remove(w.tempPath)
		return err
	}
	if !w.overwrite {
		if _, err := os.Stat(w.target); err == nil {
			_ = os.Remove(w.tempPath)
			return iofs.ErrExist
		}
	}
	return os.Rename(w.tempPath, w.target)
}

// entryFromInfo chuyển đổi một iofs.FileInfo thành một midfs.Entry.
func entryFromInfo(name string, uri midfs.URI, info iofs.FileInfo) midfs.Entry {
	entryType := midfs.EntryFile
	target := ""
	switch {
	case info.IsDir():
		entryType = midfs.EntryDir
	case info.Mode()&os.ModeSymlink != 0:
		entryType = midfs.EntrySymlink
		if linkTarget, err := os.Readlink(uri.Path); err == nil {
			target = linkTarget
		}
	}

	return midfs.Entry{
		Name:      name,
		Path:      uri.Path,
		URI:       uri,
		Type:      entryType,
		Size:      info.Size(),
		Mode:      info.Mode(),
		ModTime:   info.ModTime(),
		Target:    target,
		Readable:  info.Mode().Perm()&0o400 != 0,
		Writable:  info.Mode().Perm()&0o200 != 0,
		Hidden:    strings.HasPrefix(name, "."),
		IsArchive: false,
	}
}

var _ midfs.FileSystem = (*FS)(nil)
