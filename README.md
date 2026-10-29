# keytype

Type on a computer from a program, through a real USB keyboard.

keytype is a tiny USB board plus a command-line tool. The board plugs into a computer and shows up as two things at once: a serial port and a keyboard. Whatever a program sends to the serial port, the board types as key presses. To the computer it is indistinguishable from someone typing on a plugged-in keyboard, so it works where software key injection doesn't: login screens, UAC prompts, remote-desktop and VM consoles, games, and apps that block simulated input.

```text
your script ──► keytype CLI ──► USB serial ──► board ──► USB keyboard ──► the same computer
```

One USB connection carries both directions, so the board types into the computer it's plugged into.

## Hardware

Any board with an **ATmega32U4** (native USB) works. Tested target:

- **DFRobot Beetle** (DFR0282): about 20 × 35 mm with a USB-A plug on the board, so it plugs straight into a port. No soldering: the pin headers that come with it are not needed.

Also works: Arduino Leonardo, Arduino Micro, SparkFun Pro Micro and the many 32U4 "Pro Micro" clones.

## 1. Flash the firmware (once)

The firmware is a small Arduino sketch in [`firmware/keytype`](firmware/keytype/keytype.ino).

**With the Arduino IDE:** open `firmware/keytype/keytype.ino`, choose **Tools → Board → Arduino AVR Boards → Arduino Leonardo**, pick the board's port, and click **Upload**.

**With [arduino-cli](https://arduino.github.io/arduino-cli/):**

```sh
arduino-cli core install arduino:avr
arduino-cli lib install Keyboard
arduino-cli compile --fqbn arduino:avr:leonardo firmware/keytype
arduino-cli upload  --fqbn arduino:avr:leonardo -p COM5 firmware/keytype   # or /dev/ttyACM0
```

After that, the board needs nothing else. Re-flashing works the same way at any time.

## 2. Install the command-line tool

With Go 1.25 or newer:

```sh
go install github.com/nicholasbergesen/keytype/cmd/keytype@latest
```

Or clone this repository and run `go build ./cmd/keytype`. The result is a single executable with no runtime dependencies. On macOS, build with cgo enabled (the default when Xcode's command-line tools are installed), because listing USB ports needs it.

## 3. Use it

```sh
keytype ping                              # check the board answers
keytype type "Hello, world"               # type some text
keytype type -enter "dir"                 # type, then press Enter
keytype keys win+r                        # press a key combination
keytype keys ctrl+alt+delete
keytype keys ctrl+a ctrl+c                # several combinations in turn
cat notes.txt | keytype type              # type a file (from standard input)
keytype type -delay 30ms "slow typing"    # slow down for apps that drop keys
keytype ports                             # list USB serial ports
```

The board is found automatically among Arduino, SparkFun and DFRobot USB devices. To pick one yourself, pass `-port COM5` (Windows), `-port /dev/ttyACM0` (Linux) or `-port /dev/cu.usbmodem1101` (macOS), or set `KEYTYPE_PORT`.

Key names for `keys` (case-insensitive, joined with `+`): `ctrl`, `shift`, `alt`, `win` (also `cmd`, `gui`, `super`), the right-hand `rctrl`, `rshift`, `ralt`/`altgr`, `rwin`; `enter`, `esc`, `tab`, `space`, `backspace`, `delete`, `insert`, `home`, `end`, `pageup`, `pagedown`, `up`, `down`, `left`, `right`, `f1`–`f12`, `capslock`, `printscreen`, `scrolllock`, `pause`, `menu`; any single printable character such as `a`, `5` or `/`; and `plus` for `+`.

## Use it from Go

```go
dev, err := keytype.Find() // or keytype.Open("COM5")
if err != nil {
	log.Fatal(err)
}
defer dev.Close()
dev.Type("Hello from Go\n")
dev.Press("ctrl+s")
```

## Limits

- **US-ASCII only.** A keyboard sends keys, not characters, and the computer's keyboard layout decides which character a key makes. keytype types as a US layout, so set the target computer's layout to US English. Accented letters, Hangul, emoji and other non-ASCII text are refused with an error rather than typed wrongly. (To enter Hangul, switch the target's input method to Korean and type the romanized keys, for example `dkssud` for 안녕.)
- Roughly 100 characters a second by default. Raise `-delay` if an app drops characters.

## Protocol

Plain text lines over the serial port, one reply line per command, so any language can drive it:

| Command | Meaning | Reply |
| --- | --- | --- |
| `P` | Ping | `OK keytype 1` |
| `T<hex>` | Type these bytes (up to 64): ASCII, `\n` is Enter, `\t` is Tab | `OK` |
| `K<hex>` | Press these key codes together, then release all | `OK` |
| `D<ms>` | Pause after each key, 0–1000 ms (default 5) | `OK` |
| `R` | Release all keys | `OK` |

Errors reply `ERR <reason>`. Key codes are the [Arduino Keyboard library's](https://docs.arduino.cc/language-reference/en/functions/usb/Keyboard/keyboardModifiers/): modifiers `0x80`–`0x87`, special keys from `0xB0`, and ASCII for everything else. Open the port with DTR asserted, or the board won't reply.

## Use it responsibly

keytype is a keyboard. Only plug it into computers you own or are allowed to use, and remember that anything that can send it text can type on that computer.

## License

[MIT](LICENSE)
