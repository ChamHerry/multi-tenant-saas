// Package crypto provides AES-256-GCM encryption/decryption for sensitive config values.
package crypto

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

var encryptionKey []byte // 32 bytes for AES-256

// InitEncryption reads the encryptionKey from YAML config and initializes the global key.
// Must be called once at startup, before any Encrypt/Decrypt operations.
func InitEncryption(ctx context.Context) error {
	key := g.Cfg().MustGet(ctx, "encryptionKey", "").String()
	if key == "" {
		return gerror.New("encryptionKey is not configured in config.yaml")
	}

	keyBytes := []byte(key)
	if len(keyBytes) < 32 {
		padded := make([]byte, 32)
		copy(padded, keyBytes)
		keyBytes = padded
	} else if len(keyBytes) > 32 {
		keyBytes = keyBytes[:32]
	}

	encryptionKey = keyBytes
	g.Log().Info(ctx, "[crypto] AES-256-GCM encryption initialized")
	return nil
}

// Encrypt encrypts plaintext using AES-256-GCM and returns a Base64-encoded string.
func Encrypt(plaintext string) (string, error) {
	if len(encryptionKey) == 0 {
		return "", gerror.New("encryption not initialized")
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", gerror.Wrap(err, "aes.NewCipher")
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", gerror.Wrap(err, "cipher.NewGCM")
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", gerror.Wrap(err, "generate nonce")
	}

	// nonce is prepended to ciphertext
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts a Base64-encoded ciphertext produced by Encrypt.
func Decrypt(encoded string) (string, error) {
	if len(encryptionKey) == 0 {
		return "", gerror.New("encryption not initialized")
	}

	ciphertext, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", gerror.Wrap(err, "base64 decode")
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", gerror.Wrap(err, "aes.NewCipher")
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", gerror.Wrap(err, "cipher.NewGCM")
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", gerror.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", gerror.Wrap(err, "gcm.Open")
	}

	return string(plaintext), nil
}
