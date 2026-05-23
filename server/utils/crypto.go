package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"os"
)

const dbKeySecretsPath = "/run/secrets/db_encryption.key" // #nosec G101 -- file path, not a credential

func getEncryptionKey() ([]byte, error) {
	if raw, err := os.ReadFile(dbKeySecretsPath); err == nil {
		return parseKey(bytes.TrimSpace(raw))
	}

	keyEnv := os.Getenv("DB_ENCRYPTION_KEY")
	if keyEnv == "" {
		return nil, errors.New("PHI encryption key not found: set DB_ENCRYPTION_KEY or provide " + dbKeySecretsPath)
	}
	return parseKey([]byte(keyEnv))
}

func parseKey(raw []byte) ([]byte, error) {
	key, err := base64.StdEncoding.DecodeString(string(raw))
	if err != nil {
		key = raw
	}
	switch len(key) {
	case 16, 24, 32:
		return key, nil
	default:
		return nil, errors.New("DB encryption key must be 16, 24 or 32 bytes (or base64 of that length)")
	}
}

func EncryptString(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	key, err := getEncryptionKey()
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
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

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func DecryptString(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}

	key, err := getEncryptionKey()
	if err != nil {
		return "", err
	}

	cipherData, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(cipherData) < nonceSize {
		return "", errors.New("invalid ciphertext")
	}

	nonce, payload := cipherData[:nonceSize], cipherData[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, payload, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
