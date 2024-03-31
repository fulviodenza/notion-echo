package utils

import (
	"io"
	"log"
	"os"
)

const MAX_LEN_MESSAGE = 4096

func Read(filename string, dst *[]byte) error {
	f, err := os.Open(filename)
	if err != nil {
		log.Fatal("[parse]: ", err)
		return err
	}

	bytes, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	*dst = bytes
	return nil
}

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
