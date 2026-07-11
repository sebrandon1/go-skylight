package lib

import "testing"

func TestPathSeg(t *testing.T) {
	if pathSeg("abc") != "abc" {
		t.Fatalf("plain id: %q", pathSeg("abc"))
	}
	got := pathSeg("../x")
	if got == "../x" {
		t.Fatalf("expected escape, got %q", got)
	}
	if pathSeg("a/b") != "a%2Fb" {
		t.Fatalf("slash: %q", pathSeg("a/b"))
	}
}
