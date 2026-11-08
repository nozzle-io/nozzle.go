package nozzle

import "testing"

func TestSnapshotStringPreservesBoundedBytes(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{"empty", "", ""},
		{"leading_nul", "\x00ignored", ""},
		{"first_nul", "name\x00ignored", "name"},
		{"nonterminated", "full", "full"},
		{"utf8", "\u65e5\U0001f680", "\u65e5\U0001f680"},
		{"invalid_utf8", "a\xffb", "a\xffb"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			unsigned := []uint8(tt.raw)
			signed := make([]int8, len(unsigned))
			for i, ch := range unsigned {
				signed[i] = int8(ch)
			}
			if got := snapshotString(unsigned); got != tt.want {
				t.Errorf("unsigned bytes = %q, want %q", got, tt.want)
			}
			if got := snapshotString(signed); got != tt.want {
				t.Errorf("signed bytes = %q, want %q", got, tt.want)
			}
		})
	}
	if snapshotString[int8](nil) != "" || snapshotString[uint8](nil) != "" {
		t.Fatal("nil slices must produce empty strings")
	}
}

func TestSnapshotStringOwnsCopy(t *testing.T) {
	chars := []int8{'n', 'a', 'm', 'e', 0}
	retained := snapshotString(chars)
	for i := range chars {
		chars[i] = 0
	}
	if retained != "name" {
		t.Fatalf("retained bytes changed: %q", retained)
	}
}
