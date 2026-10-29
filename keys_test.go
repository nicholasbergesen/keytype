package keytype

import (
	"bytes"
	"testing"
)

func TestParseCombo(t *testing.T) {
	tests := []struct {
		combo string
		want  []byte
	}{
		{"ctrl+alt+delete", []byte{KeyLeftCtrl, KeyLeftAlt, KeyDelete}},
		{"Ctrl+C", []byte{KeyLeftCtrl, 'c'}},
		{"win+r", []byte{KeyLeftGUI, 'r'}},
		{"cmd+space", []byte{KeyLeftGUI, ' '}},
		{"ctrl+shift+esc", []byte{KeyLeftCtrl, KeyLeftShift, KeyEsc}},
		{"enter", []byte{KeyEnter}},
		{"f1", []byte{KeyF1}},
		{"F12", []byte{KeyF1 + 11}},
		{"ctrl+plus", []byte{KeyLeftCtrl, '+'}},
		{" alt + tab ", []byte{KeyLeftAlt, KeyTab}},
		{"/", []byte{'/'}},
	}
	for _, tt := range tests {
		got, err := ParseCombo(tt.combo)
		if err != nil {
			t.Errorf("ParseCombo(%q) error: %v", tt.combo, err)
			continue
		}
		if !bytes.Equal(got, tt.want) {
			t.Errorf("ParseCombo(%q) = %x, want %x", tt.combo, got, tt.want)
		}
	}
}

func TestParseComboErrors(t *testing.T) {
	for _, combo := range []string{"", "  ", "ctrl+", "+a", "hyper", "ctrl+é", "a+b+c+d+e+f+g"} {
		if got, err := ParseCombo(combo); err == nil {
			t.Errorf("ParseCombo(%q) = %x, want an error", combo, got)
		}
	}
}

func TestSixKeysPlusModifiersIsAllowed(t *testing.T) {
	if _, err := ParseCombo("ctrl+shift+alt+a+b+c+d+e+f"); err != nil {
		t.Fatal(err)
	}
}
