package cmd

import (
	"strings"
	"testing"
)

func TestDefaultFingerprint_NonEmpty(t *testing.T) {
	fp := defaultFingerprint()
	if fp == "" {
		t.Error("expected non-empty fingerprint")
	}
	if !strings.Contains(fp, "-") {
		t.Errorf("expected fingerprint to look like a UUID, got %s", fp)
	}
}

func TestDefaultFingerprint_Stable(t *testing.T) {
	fp1 := defaultFingerprint()
	fp2 := defaultFingerprint()
	if fp1 != fp2 {
		t.Errorf("fingerprint is not stable: %s vs %s", fp1, fp2)
	}
}

func TestFnv32_Deterministic(t *testing.T) {
	h1 := fnv32("hello")
	h2 := fnv32("hello")
	if h1 != h2 {
		t.Errorf("fnv32 is not deterministic: %d vs %d", h1, h2)
	}
}

func TestFnv32_Distinct(t *testing.T) {
	h1 := fnv32("hello")
	h2 := fnv32("world")
	if h1 == h2 {
		t.Errorf("fnv32 produced same hash for different inputs")
	}
}
