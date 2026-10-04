package base62_test

import (
	"testing"
	"github.com/anktzz/ur-short/internal/base62"
)

func TestEncode(t *testing.T) {
	tests := []struct {
		in   uint64
		want string
	}{
		{0, "0"},
		{9, "9"},
		{10, "a"},
		{61, "Z"},
		{62, "10"},
		{125, "21"},
	}
	for _, tt := range tests {
		got := base62.Encode(tt.in)
		if got != tt.want {
			t.Errorf("Encode(%d) = %s, want %s", tt.in, got, tt.want)
		}
	}
}

func TestRoundTrip(t *testing.T) {
	for _, n := range []uint64{0, 1, 61, 62, 1000, 123456789} {
		got, err := base62.Decode(base62.Encode(n))
		if err != nil || got != n {
			t.Errorf("round trip failed for %d", n)
		}
	}
}

func TestDecodeInvalid(t *testing.T) {
	_, err := base62.Decode("ab-c")
	if err == nil {
		t.Error("expected error")
	}
}
