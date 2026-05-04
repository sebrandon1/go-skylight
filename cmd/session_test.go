package cmd

import (
	"regexp"
	"testing"
)

var uuidRe = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func TestNewUUID_Format(t *testing.T) {
	id := newUUID()
	if !uuidRe.MatchString(id) {
		t.Errorf("newUUID() = %q, want UUID v4 format", id)
	}
}

func TestNewUUID_Unique(t *testing.T) {
	a, b := newUUID(), newUUID()
	if a == b {
		t.Errorf("newUUID() returned identical values: %q", a)
	}
}
