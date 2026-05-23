package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
)

func loadKey() ([]byte, error) {
	hexKey := os.Getenv("PHI_ENCRYPTION_KEY")
	if hexKey == "" {
		return nil, errors.New("PHI_ENCRYPTION_KEY environment variable not set")
	}
	k, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, fmt.Errorf("PHI_ENCRYPTION_KEY must be valid hex: %w", err)
	}
	if len(k) != 32 {
		return nil, fmt.Errorf("PHI_ENCRYPTION_KEY must decode to 32 bytes (64 hex chars), got %d bytes", len(k))
	}
	return k, nil
}

// encrypts plaintext with AES-256-GCM and returns base64
func Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}
	key, err := loadKey()
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("aes cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("aes-gcm: %w", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("nonce generation: %w", err)
	}
	sealed := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

// decrypts a base64-encoded AES-256-GCM ciphertext produced by Encrypt.
func Decrypt(encoded string) (string, error) {
	if encoded == "" {
		return "", nil
	}
	key, err := loadKey()
	if err != nil {
		return "", err
	}
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("base64 decode: %w", err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("aes cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("aes-gcm: %w", err)
	}
	ns := gcm.NonceSize()
	if len(data) < ns {
		return "", errors.New("ciphertext too short")
	}
	plaintext, err := gcm.Open(nil, data[:ns], data[ns:], nil)
	if err != nil {
		return "", fmt.Errorf("decrypt: %w", err)
	}
	return string(plaintext), nil
}