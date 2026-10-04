package base62

import (
	"errors"
	"strings"
)

const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func Encode(n uint64) string {
	if n == 0 {
		return string(alphabet[0])
	}

	var b []byte
	for n > 0 {
		b = append(b, alphabet[n%62])
		n /= 62
	}

	for i , j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

func Decode(s string) (uint64, error) {
	var n uint64
	for _, c := range s {
		i := strings.IndexRune(alphabet, c)
		if i == -1 {
			return 0, errors.New("Invalid Character")
		}
		n = n*62 + uint64(i)
	}
	return n, nil
}
