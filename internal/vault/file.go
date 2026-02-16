package vault

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
)

// FileVault stores API keys in an AES-256-GCM encrypted JSON file.
// This is the cross-platform fallback when macOS Keychain is unavailable.
type FileVault struct {
	path string
	key  []byte
}

// NewFileVault creates a vault backed by an encrypted file.
func NewFileVault() *FileVault {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir == "" {
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, ".config")
	}

	return &FileVault{
		path: filepath.Join(dir, "tamr", "vault.enc"),
		key:  deriveKey(),
	}
}

func deriveKey() []byte {
	hostname, _ := os.Hostname()
	username := os.Getenv("USER")
	if username == "" {
		username = os.Getenv("USERNAME")
	}

	seed := fmt.Sprintf("tamr-vault:%s:%s", hostname, username)
	hash := sha256.Sum256([]byte(seed))
	return hash[:]
}

func (f *FileVault) load() (map[string]string, error) {
	data, err := os.ReadFile(f.path)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]string), nil
		}
		return nil, err
	}

	plaintext, err := f.decrypt(data)
	if err != nil {
		return nil, fmt.Errorf("vault decrypt: %w", err)
	}

	var store map[string]string
	if err := json.Unmarshal(plaintext, &store); err != nil {
		return nil, fmt.Errorf("vault parse: %w", err)
	}
	return store, nil
}

func (f *FileVault) save(store map[string]string) error {
	plaintext, err := json.Marshal(store)
	if err != nil {
		return err
	}

	ciphertext, err := f.encrypt(plaintext)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(f.path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(f.path, ciphertext, 0o600)
}

func (f *FileVault) encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(f.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

func (f *FileVault) decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(f.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

func (f *FileVault) Set(key, value string) error {
	store, err := f.load()
	if err != nil {
		return err
	}
	store[key] = value
	return f.save(store)
}

func (f *FileVault) Get(key string) (string, error) {
	store, err := f.load()
	if err != nil {
		return "", err
	}
	val, ok := store[key]
	if !ok {
		return "", fmt.Errorf("key not found: %s", key)
	}
	return val, nil
}

func (f *FileVault) Delete(key string) error {
	store, err := f.load()
	if err != nil {
		return err
	}
	if _, ok := store[key]; !ok {
		return fmt.Errorf("key not found: %s", key)
	}
	delete(store, key)
	return f.save(store)
}

func (f *FileVault) List() ([]string, error) {
	store, err := f.load()
	if err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(store))
	for k := range store {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys, nil
}
