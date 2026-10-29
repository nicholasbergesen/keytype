package keytype

import (
	"bytes"
	"encoding/hex"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"
)

// fakeBoard acts like the firmware: it answers each command line the way the
// board would and records what it typed and pressed.
type fakeBoard struct {
	in       bytes.Buffer
	out      bytes.Buffer
	commands []string
	typed    []byte
	pressed  [][]byte
	delayMs  string
	version  string
	fail     string // reply with this error to every T or K command
	silent   bool   // never reply
}

func newFakeBoard() *fakeBoard { return &fakeBoard{version: Version} }

func (f *fakeBoard) Write(p []byte) (int, error) {
	f.in.Write(p)
	for {
		line, err := f.in.ReadString('\n')
		if err != nil {
			f.in.WriteString(line) // incomplete line: keep it for the next write
			return len(p), nil
		}
		f.handle(strings.TrimSuffix(line, "\n"))
	}
}

func (f *fakeBoard) Read(p []byte) (int, error) {
	if f.out.Len() == 0 {
		return 0, nil // like a serial port whose read timed out
	}
	return f.out.Read(p)
}

func (f *fakeBoard) handle(cmd string) {
	f.commands = append(f.commands, cmd)
	if f.silent {
		return
	}
	reply := "OK"
	switch {
	case cmd == "P":
		reply = "OK " + f.version
	case strings.HasPrefix(cmd, "D"):
		f.delayMs = cmd[1:]
	case strings.HasPrefix(cmd, "T"), strings.HasPrefix(cmd, "K"):
		if f.fail != "" {
			reply = "ERR " + f.fail
			break
		}
		data, err := hex.DecodeString(cmd[1:])
		if err != nil || len(data) > 64 {
			reply = "ERR bad hex"
			break
		}
		if cmd[0] == 'T' {
			f.typed = append(f.typed, data...)
		} else {
			f.pressed = append(f.pressed, data)
		}
	default:
		reply = "ERR unknown command"
	}
	f.out.WriteString(reply + "\r\n")
}

func TestPing(t *testing.T) {
	fb := newFakeBoard()
	got, err := New(fb).Ping()
	if err != nil || got != Version {
		t.Fatalf("Ping() = %q, %v; want %q", got, err, Version)
	}
}

func TestPingRejectsOtherFirmware(t *testing.T) {
	fb := newFakeBoard()
	fb.version = "keytype 2"
	if _, err := New(fb).Ping(); err == nil {
		t.Fatal("Ping() accepted firmware keytype 2")
	}
}

func TestTypeSplitsLongTextIntoChunks(t *testing.T) {
	fb := newFakeBoard()
	text := strings.Repeat("The quick brown fox jumps over the lazy dog.\n", 5)
	if err := New(fb).Type(text); err != nil {
		t.Fatal(err)
	}
	if string(fb.typed) != text {
		t.Fatalf("typed %q, want %q", fb.typed, text)
	}
	if want := (len(text) + chunkSize - 1) / chunkSize; len(fb.commands) != want {
		t.Fatalf("sent %d commands, want %d", len(fb.commands), want)
	}
	for _, c := range fb.commands {
		if len(c) > 1+128 {
			t.Fatalf("command of %d bytes overflows the firmware's line buffer", len(c))
		}
	}
}

func TestTypeNormalisesCRLF(t *testing.T) {
	fb := newFakeBoard()
	if err := New(fb).Type("a\r\nb"); err != nil {
		t.Fatal(err)
	}
	if string(fb.typed) != "a\nb" {
		t.Fatalf("typed %q", fb.typed)
	}
}

func TestTypeRejectsNonASCIIBeforeSending(t *testing.T) {
	fb := newFakeBoard()
	err := New(fb).Type("café")
	if err == nil || !strings.Contains(err.Error(), "é") {
		t.Fatalf("Type(café) error = %v", err)
	}
	if len(fb.commands) != 0 {
		t.Fatalf("sent %v before rejecting the text", fb.commands)
	}
}

func TestTypeReportsBoardError(t *testing.T) {
	fb := newFakeBoard()
	fb.fail = "bad hex"
	err := New(fb).Type("hi")
	if !errors.Is(err, ErrBoard) || !strings.Contains(err.Error(), "bad hex") {
		t.Fatalf("error = %v, want ErrBoard with the reason", err)
	}
}

func TestTimeoutWhenBoardIsSilent(t *testing.T) {
	fb := newFakeBoard()
	fb.silent = true
	start := time.Now()
	_, err := New(fb).command("P", 100*time.Millisecond)
	if err == nil || !strings.Contains(err.Error(), "no reply") {
		t.Fatalf("error = %v, want a timeout", err)
	}
	if time.Since(start) > 2*time.Second {
		t.Fatal("timeout took far too long")
	}
}

func TestPress(t *testing.T) {
	fb := newFakeBoard()
	if err := New(fb).Press("ctrl+alt+delete"); err != nil {
		t.Fatal(err)
	}
	want := [][]byte{{KeyLeftCtrl, KeyLeftAlt, KeyDelete}}
	if !reflect.DeepEqual(fb.pressed, want) {
		t.Fatalf("pressed %x, want %x", fb.pressed, want)
	}
}

func TestSetKeyDelay(t *testing.T) {
	fb := newFakeBoard()
	d := New(fb)
	if err := d.SetKeyDelay(25 * time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if fb.delayMs != "25" {
		t.Fatalf("sent delay %q, want 25", fb.delayMs)
	}
	if err := d.SetKeyDelay(2 * time.Second); err == nil {
		t.Fatal("accepted a 2s delay")
	}
}

func TestRepliesSplitAcrossReads(t *testing.T) {
	fb := newFakeBoard()
	d := New(fb)
	// Two replies already waiting, as if a read returned both at once.
	fb.out.WriteString("OK\r\nOK keytype 1\r\n")
	if r, err := d.readLine(time.Second); err != nil || r != "OK" {
		t.Fatalf("first line %q, %v", r, err)
	}
	if r, err := d.readLine(time.Second); err != nil || r != "OK keytype 1" {
		t.Fatalf("second line %q, %v", r, err)
	}
}
