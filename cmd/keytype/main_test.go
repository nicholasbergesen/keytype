package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"
)

type fakeBoard struct {
	port    string
	delay   time.Duration
	typed   string
	pressed []string
	closed  bool
}

func (f *fakeBoard) Ping() (string, error)             { return "keytype 1", nil }
func (f *fakeBoard) SetKeyDelay(d time.Duration) error { f.delay = d; return nil }
func (f *fakeBoard) Type(s string) error               { f.typed += s; return nil }
func (f *fakeBoard) Press(c string) error              { f.pressed = append(f.pressed, c); return nil }
func (f *fakeBoard) Close() error                      { f.closed = true; return nil }

// useFake makes run talk to a fake board and returns it, or nil until opened.
func useFake(t *testing.T) **fakeBoard {
	t.Helper()
	var fb *fakeBoard
	old := openBoard
	openBoard = func(port string) (board, error) {
		fb = &fakeBoard{port: port}
		return fb, nil
	}
	t.Cleanup(func() { openBoard = old })
	return &fb
}

func TestTypeArgs(t *testing.T) {
	fb := useFake(t)
	if err := run([]string{"-port", "COM7", "type", "-delay", "20ms", "-enter", "hello", "world"}, nil, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	b := *fb
	if b.port != "COM7" || b.typed != "hello world\n" || b.delay != 20*time.Millisecond || !b.closed {
		t.Fatalf("board state %+v", b)
	}
}

func TestTypeStdin(t *testing.T) {
	fb := useFake(t)
	if err := run([]string{"type"}, strings.NewReader("line one\r\nline two\n"), &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if (*fb).typed != "line one\r\nline two\n" {
		t.Fatalf("typed %q", (*fb).typed)
	}
}

func TestTypeRejectsBadTextWithoutOpeningBoard(t *testing.T) {
	fb := useFake(t)
	if err := run([]string{"type", "안녕"}, nil, &bytes.Buffer{}); err == nil {
		t.Fatal("typed Hangul")
	}
	if *fb != nil {
		t.Fatal("opened the board for text it can't type")
	}
}

func TestKeys(t *testing.T) {
	fb := useFake(t)
	t.Setenv("KEYTYPE_PORT", "")
	if err := run([]string{"keys", "win+r", "enter"}, nil, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if got := strings.Join((*fb).pressed, ","); got != "win+r,enter" {
		t.Fatalf("pressed %s", got)
	}
}

func TestKeysValidatesEveryComboFirst(t *testing.T) {
	fb := useFake(t)
	if err := run([]string{"keys", "win+r", "nonsense"}, nil, &bytes.Buffer{}); err == nil {
		t.Fatal("accepted an unknown key")
	}
	if *fb != nil {
		t.Fatal("pressed keys before finding the bad combination")
	}
}

func TestPortFromEnvironment(t *testing.T) {
	t.Setenv("KEYTYPE_PORT", "/dev/ttyACM0")
	fb := useFake(t)
	var out bytes.Buffer
	if err := run([]string{"ping"}, nil, &out); err != nil {
		t.Fatal(err)
	}
	if (*fb).port != "/dev/ttyACM0" || strings.TrimSpace(out.String()) != "keytype 1" {
		t.Fatalf("port %q, output %q", (*fb).port, out.String())
	}
}

func TestOpenError(t *testing.T) {
	old := openBoard
	openBoard = func(string) (board, error) { return nil, errors.New("no keytype board found") }
	t.Cleanup(func() { openBoard = old })
	if err := run([]string{"keys", "enter"}, nil, &bytes.Buffer{}); err == nil || !strings.Contains(err.Error(), "no keytype board") {
		t.Fatalf("error = %v", err)
	}
}

func TestUsageErrors(t *testing.T) {
	for _, args := range [][]string{nil, {"dance"}, {"keys"}} {
		if err := run(args, nil, &bytes.Buffer{}); err == nil {
			t.Errorf("run(%q) succeeded", args)
		}
	}
	var out bytes.Buffer
	if err := run([]string{"-h"}, nil, &out); err != nil || !strings.Contains(out.String(), "Usage:") {
		t.Fatalf("-h: %v, %q", err, out.String())
	}
}
