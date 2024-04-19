package utils

import (
	"crypto/aes"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestRead(t *testing.T) {
	filename := "testfile.txt"
	content := []byte("hello world")
	os.WriteFile(filename, content, 0644)
	defer os.Remove(filename)

	var dst []byte
	err := Read(filename, &dst)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if diff := cmp.Diff(dst, content); diff != "" {
		t.Errorf("Expected %v, got %v", content, dst)
	}
}

func TestSplitString(t *testing.T) {
	tests := []struct {
		s      string
		expect []string
	}{
		{"hello", []string{"hello"}},
		{"hello world", []string{"hello world"}},
		{"", nil},
	}

	for _, tc := range tests {
		result := SplitString(tc.s)
		if diff := cmp.Diff(result, tc.expect); diff != "" {
			t.Errorf("Expected %v, got %v", tc.expect, result)
		}
	}
}

func TestEncryptDecrypt(t *testing.T) {
	encryptionKey := make([]byte, aes.BlockSize)
	plainText := "Hello, world!"

	encryptedText, err := EncryptString(plainText, encryptionKey)
	if err != nil {
		t.Fatalf("EncryptString failed: %v", err)
	}

	decryptedText, err := DecryptString(encryptedText, encryptionKey)
	if err != nil {
		t.Fatalf("DecryptString failed: %v", err)
	}

	if decryptedText != plainText {
		t.Errorf("Expected %v, got %v", plainText, decryptedText)
	}
}

func TestGetExt(t *testing.T) {
	testcases := map[string]string{
		"a.pdf":  "pdf",
		"b.png":  "png",
		"c.jpeg": "jpeg",
		"d.":     "",
		"e.jpg":  "jpg",
		"f.g.a":  "a",
		"":       "",
	}
	for testcase, want := range testcases {
		got := GetExt(testcase)
		if strings.Compare(got, want) != 0 {
			t.Errorf("testcase: %s\nwant: %s, got: %s", testcase, want, got)
		}
	}
}

func TestAll(t *testing.T) {
	t.Run("Read", TestRead)
	t.Run("SplitString", TestSplitString)
	t.Run("EncryptDecrypt", TestEncryptDecrypt)
	t.Run("GetExt", TestGetExt)
}
