---
name: pr-ready
description: Take a keytype branch from "code written" to "merged" - go vet and test locally, compile the firmware if it changed, push once, wait for the hosted CI, squash-merge. Use when finishing an issue or PR in nicholasbergesen/keytype, or when asked to validate, ready, or merge a branch here.
---

# Getting a keytype branch merged

1. **Focused local run.** `go vet ./... && go test ./...` (the whole suite takes a few seconds and needs no hardware, so running all of it is fine). If `firmware/` changed: `arduino-cli compile --fqbn arduino:avr:leonardo --warnings all firmware/keytype`, with no warnings from `keytype.ino`.
2. **Affected-tests selector:** none; the suite is small.
3. **Never run locally:** nothing is off limits. Flashing a board (`arduino-cli upload`) needs one plugged in and is a manual check, not a gate.
4. **CI checks:** workflow `CI`. `go (ubuntu-latest)`, `go (windows-latest)` and `go (macos-latest)` are three legs of one matrix, plus `firmware`. Wait for all four.
5. **Queue:** GitHub-hosted runners (public repository, free). No self-hosted runners.
6. **Known flakes:** none recorded.
7. **Merge policy:** autonomous merge is authorized (mypilot's standing rule). Squash-merge, put the reasoning in the PR body, delete the branch. A protocol change must update the firmware, `device.go` and the README together, or it does not merge.
8. **Decisions:** in the PR body; durable conventions go in `.github/copilot-instructions.md`.
9. **Test tiers:** none; every test is fast.
10. **Deployment:** none. Releases are `go install ...@<tag>`; tag only when the user asks.
11. **Cadence:** CI on every push to a PR and to `main`. Nothing scheduled.
12. **Never run at all:** nothing repository-specific.
