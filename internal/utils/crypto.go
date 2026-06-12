package utils

import (
	"crypto/aes"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const legacyAESPassword = "Maybewewonseethesunrise"

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func VerifyPassword(stored, password string) (matched bool, shouldUpgrade bool) {
	if strings.HasPrefix(stored, "$2a$") || strings.HasPrefix(stored, "$2b$") || strings.HasPrefix(stored, "$2y$") {
		return bcrypt.CompareHashAndPassword([]byte(stored), []byte(password)) == nil, false
	}

	plain, err := LegacyAESDecrypt(stored)
	if err != nil {
		return false, false
	}
	return plain == password, plain == password
}

func LegacyAESDecrypt(cipherText string) (string, error) {
	cipherBytes, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(legacyAESKey([]byte(legacyAESPassword)))
	if err != nil {
		return "", err
	}
	if len(cipherBytes)%block.BlockSize() != 0 {
		return "", errors.New("invalid legacy AES ciphertext length")
	}

	plain := make([]byte, len(cipherBytes))
	for bs, be := 0, block.BlockSize(); bs < len(cipherBytes); bs, be = bs+block.BlockSize(), be+block.BlockSize() {
		block.Decrypt(plain[bs:be], cipherBytes[bs:be])
	}

	plain, err = pkcs5Unpad(plain, block.BlockSize())
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

func LegacyAESEncrypt(plainText string) (string, error) {
	block, err := aes.NewCipher(legacyAESKey([]byte(legacyAESPassword)))
	if err != nil {
		return "", err
	}

	plain := pkcs5Pad([]byte(plainText), block.BlockSize())
	cipherBytes := make([]byte, len(plain))
	for bs, be := 0, block.BlockSize(); bs < len(plain); bs, be = bs+block.BlockSize(), be+block.BlockSize() {
		block.Encrypt(cipherBytes[bs:be], plain[bs:be])
	}
	return base64.StdEncoding.EncodeToString(cipherBytes), nil
}

func legacyAESKey(seed []byte) []byte {
	first := sha1.Sum(seed)
	second := sha1.Sum(first[:])
	return second[:16]
}

func pkcs5Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	out := make([]byte, len(data)+padding)
	copy(out, data)
	for i := len(data); i < len(out); i++ {
		out[i] = byte(padding)
	}
	return out
}

func pkcs5Unpad(data []byte, blockSize int) ([]byte, error) {
	if len(data) == 0 || len(data)%blockSize != 0 {
		return nil, errors.New("invalid padding size")
	}
	padding := int(data[len(data)-1])
	if padding == 0 || padding > blockSize || padding > len(data) {
		return nil, errors.New("invalid padding")
	}
	for _, v := range data[len(data)-padding:] {
		if int(v) != padding {
			return nil, errors.New("invalid padding")
		}
	}
	return data[:len(data)-padding], nil
}
