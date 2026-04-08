package secrets

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

// FileProvider is the encrypted-file fallback secret backend used when no
// platform-native store is available.
type FileProvider struct {
	mu   sync.Mutex
	path string
	key  []byte
}

// NewFileProvider tạo một FileProvider mới.
// Khóa được cung cấp phải dài 32 byte cho AES-256.
func NewFileProvider(path string, key []byte) (*FileProvider, error) {
	if path == "" {
		return nil, os.ErrInvalid
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("file secrets provider requires a 32-byte key")
	}
	return &FileProvider{
		path: path,
		key:  append([]byte(nil), key...),
	}, nil
}

// Store mã hóa và lưu một giá trị bí mật dưới khóa đã cho.
// Thao tác này an toàn cho luồng (thread-safe).
func (p *FileProvider) Store(_ context.Context, key string, value []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	store, err := p.loadStoreLocked()
	if err != nil {
		return err
	}

	encrypted, err := p.encrypt(value)
	if err != nil {
		return err
	}
	store[key] = encrypted
	return p.saveStoreLocked(store)
}

// Load truy xuất và giải mã một giá trị bí mật bằng khóa của nó.
// Nó trả về os.ErrNotExist nếu không tìm thấy khóa.
// Thao tác này an toàn cho luồng (thread-safe).
func (p *FileProvider) Load(_ context.Context, key string) ([]byte, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	store, err := p.loadStoreLocked()
	if err != nil {
		return nil, err
	}

	encrypted, ok := store[key]
	if !ok {
		return nil, os.ErrNotExist
	}
	return p.decrypt(encrypted)
}

// Delete xóa một bí mật khỏi kho lưu trữ.
// Thao tác này an toàn cho luồng (thread-safe).
func (p *FileProvider) Delete(_ context.Context, key string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	store, err := p.loadStoreLocked()
	if err != nil {
		return err
	}
	delete(store, key)
	return p.saveStoreLocked(store)
}

// loadStoreLocked đọc tệp bí mật đã mã hóa từ đĩa và giải tuần tự hóa nó.
// Nó trả về một map rỗng nếu tệp không tồn tại.
// Phương thức này không an toàn cho luồng và phải được gọi khi p.mu đã được khóa.
func (p *FileProvider) loadStoreLocked() (map[string]string, error) {
	store := map[string]string{}
	data, err := os.ReadFile(p.path)
	if err != nil {
		if os.IsNotExist(err) {
			return store, nil
		}
		return nil, err
	}
	if len(data) == 0 {
		return store, nil
	}
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, err
	}
	return store, nil
}

// saveStoreLocked tuần tự hóa map bí mật thành JSON và ghi nó vào đĩa
// một cách nguyên tử bằng cách sử dụng một tệp tạm thời.
// Phương thức này không an toàn cho luồng và phải được gọi khi p.mu đã được khóa.
func (p *FileProvider) saveStoreLocked(store map[string]string) error {
	if err := os.MkdirAll(filepath.Dir(p.path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return err
	}

	tempPath := p.path + ".tmp"
	if err := os.WriteFile(tempPath, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tempPath, p.path)
}

// encrypt mã hóa văn bản thuần túy bằng AES-256-GCM. Một nonce ngẫu nhiên mới được
// tạo cho mỗi lần mã hóa và được thêm vào đầu bản mã. Kết quả
// được mã hóa base64.
func (p *FileProvider) encrypt(plaintext []byte) (string, error) {
	block, err := aes.NewCipher(p.key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decrypt giải mã một bản mã được mã hóa base64 đã được mã hóa bằng
// AES-256-GCM. Nó mong đợi nonce được thêm vào đầu bản mã.
func (p *FileProvider) decrypt(encoded string) ([]byte, error) {
	block, err := aes.NewCipher(p.key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	ciphertext, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}
	if len(ciphertext) < gcm.NonceSize() {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce := ciphertext[:gcm.NonceSize()]
	data := ciphertext[gcm.NonceSize():]
	return gcm.Open(nil, nonce, data, nil)
}

var _ Provider = (*FileProvider)(nil)
