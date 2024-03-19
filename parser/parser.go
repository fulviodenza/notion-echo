package parser

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/notion-echo/utils"
)

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

func ValidateSchedule(s, tz string) (string, error) {
	// times contains an array with two elements [Hours, Minutes]
	times := strings.SplitN(s, ":", -1)
	// in the crontab minutes come as first field
	if !utils.ValidateTime(times, tz) {
		return "", fmt.Errorf("error validating time")
	}
	if len(times) == 2 {
		return fmt.Sprintf("CRON_TZ=%s %s %s * * *", tz, times[1], times[0]), nil
	}
	return "", fmt.Errorf("not enough arguments: %v", times)
}

func ParseCategories(filename string) ([]string, error) {
	dst := make([]byte, 0)
	err := Read(filename, &dst)
	if err != nil {
		return nil, err
	}

	categories := strings.Split(string(dst), "\n")
	return categories, nil
}
