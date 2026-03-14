package crypto

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestEncryptDecryptRoundtrip(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	plaintext := []byte("hello secret world")
	ct, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Decrypt(ct, key)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, plaintext) {
		t.Fatalf("expected %q, got %q", plaintext, got)
	}
}

func TestDecryptWrongKey(t *testing.T) {
	key1, _ := GenerateKey()
	key2, _ := GenerateKey()
	ct, _ := Encrypt([]byte("secret"), key1)
	_, err := Decrypt(ct, key2)
	if err == nil {
		t.Fatal("expected error decrypting with wrong key")
	}
}

func TestEncryptProducesDifferentCiphertext(t *testing.T) {
	key, _ := GenerateKey()
	plaintext := []byte("same input")
	ct1, _ := Encrypt(plaintext, key)
	ct2, _ := Encrypt(plaintext, key)
	if bytes.Equal(ct1, ct2) {
		t.Fatal("expected different ciphertext for each encryption")
	}
}

func TestGenerateKeyLength(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	if len(key) != 32 {
		t.Fatalf("expected 32 bytes, got %d", len(key))
	}
}

func TestLoadOrGenerateKey(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", "key")

	// First call generates
	key1, err := LoadOrGenerateKey(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(key1) != 32 {
		t.Fatalf("expected 32 bytes, got %d", len(key1))
	}

	// File should exist with restricted permissions
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatalf("expected 0600 permissions, got %o", info.Mode().Perm())
	}

	// Second call loads the same key
	key2, err := LoadOrGenerateKey(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(key1, key2) {
		t.Fatal("expected same key on reload")
	}
}
