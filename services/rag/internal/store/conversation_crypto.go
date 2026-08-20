package store

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
)

func encryptConversation(secret, aad []byte, plaintext string) ([]byte, error) {
	aead, err := conversationAEAD(secret)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("create conversation nonce: %w", err)
	}

	return aead.Seal(nonce, nonce, []byte(plaintext), aad), nil
}

func decryptConversation(secret, aad, ciphertext []byte) (string, error) {
	aead, err := conversationAEAD(secret)
	if err != nil {
		return "", err
	}

	nonceSize := aead.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("invalid conversation ciphertext")
	}

	plaintext, err := aead.Open(nil, ciphertext[:nonceSize], ciphertext[nonceSize:], aad)
	if err != nil {
		return "", fmt.Errorf("open conversation ciphertext: %w", err)
	}

	return string(plaintext), nil
}

func conversationAEAD(secret []byte) (cipher.AEAD, error) {
	key := sha256.Sum256(append([]byte("godarai-managed-rag-conversation-v1:"), secret...))

	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, fmt.Errorf("create conversation cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create conversation AEAD: %w", err)
	}

	return aead, nil
}
