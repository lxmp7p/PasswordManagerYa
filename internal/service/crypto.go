package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
)

type CryptoService struct {
	aead cipher.AEAD
}

func NewCryptoService(key string) (CryptoServiceInterface, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return &CryptoService{
		aead: aead,
	}, nil
}

func (c *CryptoService) Encrypt(data []byte) ([]byte, error) {
	nonce := make([]byte, c.aead.NonceSize())

	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	return c.aead.Seal(nonce, nonce, data, nil), nil
}

func (c *CryptoService) Decrypt(data []byte) ([]byte, error) {
	size := c.aead.NonceSize()

	if len(data) < size {
		return nil, errors.New("invalid encrypted data")
	}

	nonce := data[:size]
	ciphertext := data[size:]

	return c.aead.Open(nil, nonce, ciphertext, nil)
}
