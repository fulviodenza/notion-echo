package utils

import (
	"math/rand"
	"strconv"
	"time"
)

const MAX_LEN_MESSAGE = 4096

func MakeTimestamp(len int) int64 {
	rand.Seed(time.Now().Unix())
	return (time.Now().UnixNano() / int64(time.Millisecond)) % int64(len)
}

func Pick[K comparable, V any](m map[K]V) V {
	k := rand.Intn(len(m))
	for _, x := range m {
		if k == 0 {
			return x
		}
		k--
	}
	panic("unreachable")
}

func AggregateTags(tags []string) string {
	msg := ""
	for _, s := range tags {
		msg += "- " + s + "\n"
	}

	return msg
}

func ValidateTime(times []string, tz string) bool {
	// the only accepted format is HH:MM, so, with 2 elements in the times array
	if len(times) != 2 {
		return false
	}

	hours, err := strconv.Atoi(times[0])
	if err != nil {
		return false
	}
	minutes, err := strconv.Atoi(times[1])
	if err != nil {
		return false
	}

	// validate timezone
	_, err = time.LoadLocation(tz)
	if err != nil {
		return false
	}

	if hours < 0 || hours >= 24 {
		return false
	}
	if minutes < 0 || minutes >= 60 {
		return false
	}

	return true
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
