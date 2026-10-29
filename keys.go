package keytype

import (
	"fmt"
	"strings"
)

// Key codes understood by the firmware. They are the Arduino Keyboard
// library's codes: modifiers at 0x80-0x87, special keys at 0xB0 and up, and
// plain ASCII for everything else.
const (
	KeyLeftCtrl   byte = 0x80
	KeyLeftShift  byte = 0x81
	KeyLeftAlt    byte = 0x82
	KeyLeftGUI    byte = 0x83
	KeyRightCtrl  byte = 0x84
	KeyRightShift byte = 0x85
	KeyRightAlt   byte = 0x86
	KeyRightGUI   byte = 0x87

	KeyEnter       byte = 0xB0
	KeyEsc         byte = 0xB1
	KeyBackspace   byte = 0xB2
	KeyTab         byte = 0xB3
	KeyCapsLock    byte = 0xC1
	KeyF1          byte = 0xC2
	KeyPrintScreen byte = 0xCE
	KeyScrollLock  byte = 0xCF
	KeyPause       byte = 0xD0
	KeyInsert      byte = 0xD1
	KeyHome        byte = 0xD2
	KeyPageUp      byte = 0xD3
	KeyDelete      byte = 0xD4
	KeyEnd         byte = 0xD5
	KeyPageDown    byte = 0xD6
	KeyRight       byte = 0xD7
	KeyLeft        byte = 0xD8
	KeyDown        byte = 0xD9
	KeyUp          byte = 0xDA
	KeyMenu        byte = 0xED
)

var keyNames = map[string]byte{
	"ctrl": KeyLeftCtrl, "control": KeyLeftCtrl, "lctrl": KeyLeftCtrl, "rctrl": KeyRightCtrl,
	"shift": KeyLeftShift, "lshift": KeyLeftShift, "rshift": KeyRightShift,
	"alt": KeyLeftAlt, "lalt": KeyLeftAlt, "ralt": KeyRightAlt, "altgr": KeyRightAlt, "option": KeyLeftAlt,
	"win": KeyLeftGUI, "gui": KeyLeftGUI, "cmd": KeyLeftGUI, "command": KeyLeftGUI, "super": KeyLeftGUI, "meta": KeyLeftGUI,
	"rwin": KeyRightGUI,

	"enter": KeyEnter, "return": KeyEnter,
	"esc": KeyEsc, "escape": KeyEsc,
	"backspace":   KeyBackspace,
	"tab":         KeyTab,
	"space":       ' ',
	"capslock":    KeyCapsLock,
	"printscreen": KeyPrintScreen, "prtsc": KeyPrintScreen,
	"scrolllock": KeyScrollLock,
	"pause":      KeyPause,
	"insert":     KeyInsert, "ins": KeyInsert,
	"delete": KeyDelete, "del": KeyDelete,
	"home": KeyHome, "end": KeyEnd,
	"pageup": KeyPageUp, "pgup": KeyPageUp,
	"pagedown": KeyPageDown, "pgdn": KeyPageDown,
	"up": KeyUp, "down": KeyDown, "left": KeyLeft, "right": KeyRight,
	"menu": KeyMenu,
	"plus": '+',
}

func init() {
	for i := 0; i < 12; i++ {
		keyNames[fmt.Sprintf("f%d", i+1)] = KeyF1 + byte(i)
	}
}

// maxComboKeys is how many keys a USB keyboard reports at once: up to six
// ordinary keys plus the modifiers.
const maxComboKeys = 6

// ParseCombo turns a key combination such as "ctrl+alt+delete", "win+r" or
// "f5" into the key codes to hold down together. Names are case-insensitive;
// a single printable character stands for its own key (use "plus" for '+').
func ParseCombo(combo string) ([]byte, error) {
	if strings.TrimSpace(combo) == "" {
		return nil, fmt.Errorf("empty key combination")
	}
	var codes []byte
	ordinary := 0
	for _, part := range strings.Split(combo, "+") {
		name := strings.ToLower(strings.TrimSpace(part))
		if name == "" {
			return nil, fmt.Errorf("key combination %q has an empty key", combo)
		}
		code, ok := keyNames[name]
		if !ok {
			if len(name) != 1 || name[0] < 0x21 || name[0] > 0x7E {
				return nil, fmt.Errorf("unknown key %q in %q", part, combo)
			}
			code = name[0]
		}
		if code < KeyLeftCtrl || code > KeyRightGUI {
			ordinary++
		}
		codes = append(codes, code)
	}
	if ordinary > maxComboKeys {
		return nil, fmt.Errorf("key combination %q holds more than %d non-modifier keys", combo, maxComboKeys)
	}
	return codes, nil
}
