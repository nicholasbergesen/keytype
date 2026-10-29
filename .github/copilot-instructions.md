# keytype

A USB keyboard board (ATmega32U4 Arduino firmware in `firmware/keytype`) and the Go package and CLI that drive it over its USB serial port (`keytype` package at the root, CLI in `cmd/keytype`).

- The firmware and the Go package share a line protocol, documented at the top of `keytype.ino` and in the README's Protocol table. Change all three together, and bump the version string (`keytype 1`) in both the firmware and `device.go` when a change is not backward compatible.
- The firmware's line buffer holds 64 bytes of hex; `chunkSize` in `device.go` must stay below that.
- Go code has no cgo of its own. `go.bug.st/serial/enumerator` needs cgo on macOS only.
- Tests use a fake board (`device_test.go`) and a fake `openBoard` (`cmd/keytype/main_test.go`); no test needs hardware.
- Text is not hard-wrapped in Markdown: one line per paragraph or list item.

Build and test: `go vet ./... && go test ./...`. Firmware: `arduino-cli compile --fqbn arduino:avr:leonardo firmware/keytype`. The `pr-ready` skill has the merge procedure.
