package cmd

import "testing"

func TestFilterEmptyStrings(t *testing.T) {
	got := filterEmptyStrings([]string{"a", "", "b", "", "c"})
	want := []string{"a", "b", "c"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("got %v, want %v", got, want)
		}
	}
}

func TestFilterEmptyStrings_AllEmpty(t *testing.T) {
	got := filterEmptyStrings([]string{"", "", ""})
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %v", got)
	}
}

func TestFilterEmptyStrings_NoneEmpty(t *testing.T) {
	got := filterEmptyStrings([]string{"a", "b"})
	if len(got) != 2 {
		t.Errorf("expected 2 items, got %v", got)
	}
}
