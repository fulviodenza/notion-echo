package utils

import (
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"os"
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

func SplitFirstOccurrence(s string, sep string) (string, string) {
	var part1, part2 string
	if i := strings.Index(s, sep); i >= 0 {
		part1, part2 = s[:i], s[i:]
	} else {
		return "", ""
	}
	return part1, part2
}

func CompressFile(source, target string) error {
	reader, err := os.Open(source)
	if err != nil {
		return err
	}
	defer reader.Close()

	writer, err := os.Create(target)
	if err != nil {
		return err
	}
	defer writer.Close()

	gzipWriter := gzip.NewWriter(writer)
	defer gzipWriter.Close()

	_, err = io.Copy(gzipWriter, reader)
	return err
}
