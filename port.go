package keytype

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"go.bug.st/serial"
	"go.bug.st/serial/enumerator"
)

// boardVIDs are the USB vendor IDs of ATmega32U4 boards that keytype runs on:
// Arduino (2341, 2A03), SparkFun (1B4F) and DFRobot (3343).
var boardVIDs = []string{"2341", "2A03", "1B4F", "3343"}

// Open opens the board on the named serial port (for example "COM5",
// "/dev/ttyACM0" or "/dev/cu.usbmodem1101") and checks that it answers.
func Open(portName string) (*Device, error) {
	port, err := serial.Open(portName, &serial.Mode{
		BaudRate: 115200,
		// The board only sends replies while the host asserts DTR.
		InitialStatusBits: &serial.ModemOutputBits{DTR: true, RTS: true},
	})
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", portName, err)
	}
	if err := port.SetReadTimeout(50 * time.Millisecond); err != nil {
		port.Close()
		return nil, fmt.Errorf("set read timeout on %s: %w", portName, err)
	}
	_ = port.ResetInputBuffer()
	d := New(port)
	if _, err := d.Ping(); err != nil {
		port.Close()
		return nil, fmt.Errorf("%s: %w", portName, err)
	}
	return d, nil
}

// Port describes a USB serial port that may be a keytype board.
type Port struct {
	Name    string
	VID     string
	PID     string
	Product string
}

// ListPorts returns the USB serial ports on this computer.
func ListPorts() ([]Port, error) {
	details, err := enumerator.GetDetailedPortsList()
	if err != nil {
		return nil, fmt.Errorf("list serial ports: %w", err)
	}
	var ports []Port
	for _, p := range details {
		if p.IsUSB {
			ports = append(ports, Port{Name: p.Name, VID: strings.ToUpper(p.VID), PID: strings.ToUpper(p.PID), Product: p.Product})
		}
	}
	return ports, nil
}

// ErrNotFound is returned by Find when no board answers.
var ErrNotFound = errors.New("no keytype board found")

// Find opens the first keytype board it finds. It only tries ports whose USB
// vendor is a known ATmega32U4 board maker, so it never writes to unrelated
// devices; pass a port name to Open for anything else.
func Find() (*Device, error) {
	ports, err := ListPorts()
	if err != nil {
		return nil, err
	}
	var tried []string
	for _, p := range ports {
		if !knownBoard(p.VID) {
			continue
		}
		d, err := Open(p.Name)
		if err == nil {
			return d, nil
		}
		tried = append(tried, err.Error())
	}
	if len(tried) == 0 {
		return nil, fmt.Errorf("%w: no Arduino-compatible USB serial port is connected", ErrNotFound)
	}
	return nil, fmt.Errorf("%w: %s", ErrNotFound, strings.Join(tried, "; "))
}

func knownBoard(vid string) bool {
	for _, v := range boardVIDs {
		if strings.EqualFold(v, vid) {
			return true
		}
	}
	return false
}
