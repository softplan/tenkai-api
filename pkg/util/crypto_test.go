package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateHash(t *testing.T) {
	hash := createHash("abacaxi")
	assert.NotEmpty(t, hash)
}
func TestEncryptDecrypt(t *testing.T) {
	passKey := "ABCDEFGHIJKLMNO"
	password := "Minha senha"
	encryptPassword := Encrypt([]byte(password), passKey)
	assert.NotEmpty(t, encryptPassword)
	decriptPassword, error := Decrypt(encryptPassword, passKey)
	assert.Nil(t, error)
	assert.Equal(t, password, string(decriptPassword))
}
