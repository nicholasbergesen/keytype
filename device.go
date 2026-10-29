// Package keytype drives a board running the keytype firmware: a small USB
// device that shows up on a computer as both a serial port and a keyboard,
// and types whatever is sent to the serial port.
package keytype

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

// Version is the firmware version this package speaks to.
const Version = "keytype 1"

// chunkSize is how many characters go in one 'T' command. The firmware's line
// buffer holds 64 bytes of hex-encoded text.
const chunkSize = 60

// Device is a connection to a board running the keytype firmware.
type Device struct {
	port     io.ReadWriter
	keyDelay time.Duration
	pending  []byte
}

// New wraps an open connection to the board. Reads from port may return
// (0, nil) when nothing has arrived yet, as serial ports with a read timeout do.
func New(port io.ReadWriter) *Device {
	return &Device{port: port, keyDelay: 5 * time.Millisecond}
}

// Close closes the underlying port if it can be closed.
func (d *Device) Close() error {
	if c, ok := d.port.(io.Closer); ok {
		return c.Close()
	}
	return nil
}

// Ping checks that the board answers and runs a compatible firmware, and
// returns its version string.
func (d *Device) Ping() (string, error) {
	reply, err := d.command("P", time.Second)
	if err != nil {
		return "", err
	}
	if reply != Version {
		return reply, fmt.Errorf("board runs %q, want %q", reply, Version)
	}
	return reply, nil
}

// SetKeyDelay sets the pause after each key the board types (0 to 1s). Some
// programs drop keys that arrive too fast; raise it if characters go missing.
func (d *Device) SetKeyDelay(delay time.Duration) error {
	ms := delay.Milliseconds()
	if ms < 0 || ms > 1000 {
		return fmt.Errorf("key delay %v is outside 0-1s", delay)
	}
	if _, err := d.command("D"+strconv.FormatInt(ms, 10), time.Second); err != nil {
		return err
	}
	d.keyDelay = time.Duration(ms) * time.Millisecond
	return nil
}

// Type types text on the computer the board is plugged into, using a US
// keyboard layout. See CheckText for what text it accepts.
func (d *Device) Type(text string) error {
	data, err := CheckText(text)
	if err != nil {
		return err
	}
	for len(data) > 0 {
		n := min(len(data), chunkSize)
		if _, err := d.command("T"+hex.EncodeToString(data[:n]), d.timeoutFor(n)); err != nil {
			return err
		}
		data = data[n:]
	}
	return nil
}

// Press holds down a key combination such as "ctrl+alt+delete" and releases
// it. See ParseCombo for the key names.
func (d *Device) Press(combo string) error {
	codes, err := ParseCombo(combo)
	if err != nil {
		return err
	}
	_, err = d.command("K"+hex.EncodeToString(codes), d.timeoutFor(1))
	return err
}

// CheckText returns the bytes the board should type for text. The board types
// printable US-ASCII, newlines (as Enter) and tabs; "\r\n" becomes "\n".
// Anything else, such as accented letters or Hangul, is an error, because a
// keyboard can only send keys, and which character a key makes is decided by
// the computer's keyboard layout.
func CheckText(text string) ([]byte, error) {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	for i, r := range text {
		if r == '\n' || r == '\t' || (r >= 0x20 && r <= 0x7E) {
			continue
		}
		return nil, fmt.Errorf("can't type %q at byte %d: only US-ASCII letters, digits, symbols, spaces, tabs and newlines can be typed", r, i)
	}
	return []byte(text), nil
}

// timeoutFor allows for keys keys being typed, with room for USB latency.
func (d *Device) timeoutFor(keys int) time.Duration {
	return time.Second + time.Duration(keys)*(d.keyDelay+30*time.Millisecond)
}

// ErrBoard is returned, wrapped, when the board rejects a command.
var ErrBoard = errors.New("board error")

// command sends one command line and waits for its reply, returning the text
// after "OK".
func (d *Device) command(cmd string, timeout time.Duration) (string, error) {
	if _, err := io.WriteString(d.port, cmd+"\n"); err != nil {
		return "", fmt.Errorf("write to board: %w", err)
	}
	reply, err := d.readLine(timeout)
	if err != nil {
		return "", err
	}
	switch {
	case reply == "OK":
		return "", nil
	case strings.HasPrefix(reply, "OK "):
		return strings.TrimPrefix(reply, "OK "), nil
	case strings.HasPrefix(reply, "ERR"):
		return "", fmt.Errorf("%w: %s", ErrBoard, strings.TrimSpace(strings.TrimPrefix(reply, "ERR")))
	default:
		return "", fmt.Errorf("unexpected reply from board: %q", reply)
	}
}

// readLine reads one reply line, waiting at most timeout for it.
func (d *Device) readLine(timeout time.Duration) (string, error) {
	deadline := time.Now().Add(timeout)
	buf := make([]byte, 64)
	for {
		if i := bytes.IndexByte(d.pending, '\n'); i >= 0 {
			line := strings.TrimRight(string(d.pending[:i]), "\r")
			d.pending = d.pending[i+1:]
			return line, nil
		}
		if time.Now().After(deadline) {
			return "", fmt.Errorf("no reply from board within %v; is the keytype firmware flashed?", timeout)
		}
		n, err := d.port.Read(buf)
		d.pending = append(d.pending, buf[:n]...)
		if err != nil && !(errors.Is(err, io.EOF) && n > 0) {
			return "", fmt.Errorf("read from board: %w", err)
		}
	}
}
