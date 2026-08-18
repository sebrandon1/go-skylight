package cmd

import (
	"strings"
	"testing"
)

func TestNewUUID_Version4Bits(t *testing.T) {
	u := newUUID()
	parts := strings.Split(u, "-")
	if len(parts) != 5 {
		t.Fatalf("expected 5 hyphen-separated groups, got %d: %s", len(parts), u)
	}
	// version nibble: 3rd group must start with '4'
	if parts[2][0] != '4' {
		t.Errorf("expected version 4 UUID (3rd group starts with '4'), got %q", parts[2])
	}
	// variant bits: 4th group must start with '8', '9', 'a', or 'b'
	c := parts[3][0]
	if c != '8' && c != '9' && c != 'a' && c != 'b' {
		t.Errorf("expected RFC 4122 variant bits (4th group starts with 8/9/a/b), got %q", parts[3])
	}
}
