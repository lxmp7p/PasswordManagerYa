package service

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCryptoService_EncryptDecrypt_Success(t *testing.T) {
	crypto, err := NewCryptoService(
		"01234567890123456789012345678901",
	)

	require.NoError(t, err)

	original := []byte("my secret password")

	encrypted, err := crypto.Encrypt(original)

	require.NoError(t, err)
	require.NotEmpty(t, encrypted)

	decrypted, err := crypto.Decrypt(encrypted)

	require.NoError(t, err)
	require.Equal(t, original, decrypted)
}

func TestCryptoService_Encrypt_DifferentCiphertext(t *testing.T) {
	crypto, err := NewCryptoService(
		"01234567890123456789012345678901",
	)

	require.NoError(t, err)

	data := []byte("same secret")

	first, err := crypto.Encrypt(data)
	require.NoError(t, err)

	second, err := crypto.Encrypt(data)
	require.NoError(t, err)

	require.False(
		t,
		bytes.Equal(first, second),
	)
}

func TestCryptoService_Decrypt_InvalidData(t *testing.T) {
	crypto, err := NewCryptoService(
		"01234567890123456789012345678901",
	)

	require.NoError(t, err)

	_, err = crypto.Decrypt([]byte("bad"))

	require.Error(t, err)
}

func TestCryptoService_Decrypt_TamperedData(t *testing.T) {
	crypto, err := NewCryptoService(
		"01234567890123456789012345678901",
	)

	require.NoError(t, err)

	encrypted, err := crypto.Encrypt(
		[]byte("secret"),
	)

	require.NoError(t, err)

	encrypted[len(encrypted)-1] ^= 1

	_, err = crypto.Decrypt(encrypted)

	require.Error(t, err)
}

func TestCryptoService_WrongKey(t *testing.T) {
	firstCrypto, err := NewCryptoService(
		"01234567890123456789012345678901",
	)

	require.NoError(t, err)

	secondCrypto, err := NewCryptoService(
		"11111111111111111111111111111111",
	)

	require.NoError(t, err)

	encrypted, err := firstCrypto.Encrypt(
		[]byte("secret"),
	)

	require.NoError(t, err)

	_, err = secondCrypto.Decrypt(encrypted)

	require.Error(t, err)
}

func TestNewCryptoService_InvalidKey(t *testing.T) {
	_, err := NewCryptoService(
		"short-key",
	)

	require.Error(t, err)
}
