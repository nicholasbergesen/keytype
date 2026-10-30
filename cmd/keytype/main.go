// Command keytype types text and presses keys through a USB board running the
// keytype firmware, so the computer sees a real keyboard.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/nicholasbergesen/keytype"
)

const usage = `keytype types on a computer through a USB board running the keytype firmware.

Usage:
  keytype [-port NAME] type [-delay 5ms] [-enter] [text ...]
  keytype [-port NAME] keys COMBO [COMBO ...]
  keytype [-port NAME] ping
  keytype ports

Commands:
  type   Type the text given as arguments, or read from standard input when
         there are none. Only US-ASCII can be typed; newlines press Enter.
  keys   Press each key combination in turn, e.g. ctrl+alt+delete, win+r,
         enter, f5, ctrl+shift+esc.
  ping   Check that the board answers.
  ports  List USB serial ports, to find the board's port name.

The board is found automatically. Use -port (or the KEYTYPE_PORT environment
variable) to choose one, e.g. -port COM5 or -port /dev/ttyACM0.
`

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "keytype:", err)
		os.Exit(1)
	}
}

// board is what the commands need from a keytype.Device.
type board interface {
	Ping() (string, error)
	SetKeyDelay(time.Duration) error
	Type(string) error
	Press(string) error
	Close() error
}

// openBoard is replaced in tests.
var openBoard = func(port string) (board, error) {
	if port != "" {
		return keytype.Open(port)
	}
	return keytype.Find()
}

func run(args []string, stdin io.Reader, stdout io.Writer) error {
	global := flag.NewFlagSet("keytype", flag.ContinueOnError)
	global.SetOutput(io.Discard)
	port := global.String("port", os.Getenv("KEYTYPE_PORT"), "serial port of the board")
	if err := global.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fmt.Fprint(stdout, usage)
			return nil
		}
		return fmt.Errorf("%w\n\n%s", err, usage)
	}
	if global.NArg() == 0 {
		return errors.New("no command given\n\n" + usage)
	}
	cmd, rest := global.Arg(0), global.Args()[1:]

	switch cmd {
	case "ports":
		return listPorts(stdout)
	case "ping":
		return withBoard(*port, func(b board) error {
			version, err := b.Ping()
			if err == nil {
				fmt.Fprintln(stdout, version)
			}
			return err
		})
	case "type":
		return typeText(*port, rest, stdin)
	case "keys":
		if len(rest) == 0 {
			return errors.New("keys needs at least one key combination, e.g. keytype keys ctrl+alt+delete")
		}
		for _, combo := range rest {
			if _, err := keytype.ParseCombo(combo); err != nil {
				return err
			}
		}
		return withBoard(*port, func(b board) error {
			for _, combo := range rest {
				if err := b.Press(combo); err != nil {
					return err
				}
			}
			return nil
		})
	case "help":
		fmt.Fprint(stdout, usage)
		return nil
	default:
		return fmt.Errorf("unknown command %q\n\n%s", cmd, usage)
	}
}

func typeText(port string, args []string, stdin io.Reader) error {
	fs := flag.NewFlagSet("type", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	delay := fs.Duration("delay", 5*time.Millisecond, "pause after each key")
	enter := fs.Bool("enter", false, "press Enter after the text")
	if err := fs.Parse(args); err != nil {
		return err
	}
	var text string
	if fs.NArg() > 0 {
		text = strings.Join(fs.Args(), " ")
	} else {
		data, err := io.ReadAll(stdin)
		if err != nil {
			return fmt.Errorf("read standard input: %w", err)
		}
		// Windows PowerShell prepends a UTF-8 byte order mark when it pipes
		// text to a program; it marks the encoding and is never meant as a key.
		text = strings.TrimPrefix(string(data), "\ufeff")
	}
	if *enter {
		text += "\n"
	}
	// Check the text before touching the board, so nothing is half-typed.
	if _, err := keytype.CheckText(text); err != nil {
		return err
	}
	return withBoard(port, func(b board) error {
		if err := b.SetKeyDelay(*delay); err != nil {
			return err
		}
		return b.Type(text)
	})
}

func withBoard(port string, fn func(board) error) error {
	b, err := openBoard(port)
	if err != nil {
		return err
	}
	defer b.Close()
	return fn(b)
}

func listPorts(stdout io.Writer) error {
	ports, err := keytype.ListPorts()
	if err != nil {
		return err
	}
	if len(ports) == 0 {
		fmt.Fprintln(stdout, "no USB serial ports found")
		return nil
	}
	for _, p := range ports {
		fmt.Fprintf(stdout, "%s\t%s:%s\t%s\n", p.Name, p.VID, p.PID, p.Product)
	}
	return nil
}
