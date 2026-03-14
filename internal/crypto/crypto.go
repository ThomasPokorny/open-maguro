package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Encrypt encrypts plaintext using AES-256-GCM with the given 32-byte key.
// Returns nonce || ciphertext.
func Encrypt(plaintext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("new cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("new gcm: %w", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("generate nonce: %w", err)
	}
	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// Decrypt decrypts ciphertext produced by Encrypt using AES-256-GCM.
func Decrypt(ciphertext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("new cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("new gcm: %w", err)
	}
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}
	nonce, ct := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, ct, nil)
}

// GenerateKey generates a random 32-byte AES-256 key.
func GenerateKey() ([]byte, error) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, fmt.Errorf("generate key: %w", err)
	}
	return key, nil
}

// LoadOrGenerateKey reads a hex-encoded key from path, or generates one and
// writes it there if the file does not exist.
func LoadOrGenerateKey(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err == nil {
		return hex.DecodeString(string(data))
	}
	if !os.IsNotExist(err) {
		return nil, fmt.Errorf("read key file: %w", err)
	}
	key, err := GenerateKey()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, fmt.Errorf("create key directory: %w", err)
	}
	if err := os.WriteFile(path, []byte(hex.EncodeToString(key)), 0600); err != nil {
		return nil, fmt.Errorf("write key file: %w", err)
	}
	return key, nil
}
