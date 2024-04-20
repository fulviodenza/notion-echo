package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"strings"
)

func SplitString(s string) []string {
	if len(s) <= 0 {
		return nil
	}

	maxGroupLen := MAX_LEN_MESSAGE - 1
	if len(s) < maxGroupLen {
		maxGroupLen = len(s)
	}
	group := s[:maxGroupLen]
	return append([]string{group}, SplitString(s[maxGroupLen:])...)
}

func EncryptString(plainText string, encryptionKey []byte) (string, error) {
	plaintext := []byte(plainText)

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", err
	}

	cipherText := make([]byte, aes.BlockSize+len(plaintext))
	iv := cipherText[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], plaintext)

	return base64.URLEncoding.EncodeToString(cipherText), nil
}

func DecryptString(cryptoText string, key []byte) (string, error) {
	cipherText, err := base64.URLEncoding.DecodeString(cryptoText)
	if err != nil {
		return "", err
	}

	if len(cipherText) < aes.BlockSize {
		return "", errors.New("ciphertext block size is too short")
	}

	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(cipherText, cipherText)

	return string(cipherText), nil
}

func GetExt(path string) string {
	comps := strings.Split(path, ".")
	if len(comps) == 0 || comps[0] == "" {
		return ""
	}
	return comps[len(comps)-1]
}

func EscapeString(input string) string {
	specialChars := []string{"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"}

	for _, char := range specialChars {
		input = strings.ReplaceAll(input, char, "\\"+char)
	}

	return input
}
