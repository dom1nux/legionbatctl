# Agent Guidelines for legionbatctl

## Build, Test, and Service Tasks

All tasks live in `mise.toml` (Go 1.25.0 is pinned in `[tools]`).

| Task | Purpose | Root? |
|---|---|---|
| `mise run build` | Build `build/legionbatctl` with version ldflags | no |
| `mise run test` | `go test ./...` | no |
| `mise run vet` | `go vet ./...` | no |
| `mise run fmt` | `gofmt -w .` | no |
| `mise run clean` | Remove `build/` + `go clean` | no |
| `mise run dev` | Build, then run `build/legionbatctl status` | no |
| `mise run status` | `systemctl status` + CLI status | no |
| `mise run logs` | `journalctl -u legionbatctl.service -f` | no |
| `mise run install` | Install binary + systemd unit, enable & start | prompts for password |
| `mise run uninstall` | Stop, disable, remove artifacts | prompts for password |
| `mise run restart` | Restart the daemon | prompts for password |

Manual build (requires Go 1.25+ on PATH): `go build -o build/legionbatctl ./cmd/legionbatctl`.
Single test: `go test ./internal/state -v -run TestNewManager`.

## Project Structure

- `cmd/legionbatctl/` — main entry point
- `internal/cli/commands/` — Cobra commands (one per subcommand)
- `internal/client/` — Unix-socket client, `CommandResult` type, result formatters
- `internal/daemon/` — battery-monitoring daemon, request handlers, server loop
- `internal/protocol/` — JSON request/response message types and codec
- `internal/state/` — thread-safe `Manager` with atomic JSON persistence
- `systemd/legionbatctl.service` — systemd unit

## Code Style

`gofmt` is mandatory (`mise run fmt` enforces it). Conventions:

- **Imports**: stdlib first, then third-party, then internal `github.com/dom1nux/legionbatctl/...`. No blank lines between groups.
- **Naming**: PascalCase exported, camelCase unexported, UPPER_SNAKE_CASE for package-level constants. Acronyms capitalized once (`Xml`, not `XML`). Interfaces end in `-er`.
- **Types**: add JSON tags on serialized structs; document non-obvious fields.
- **Errors**: define package-level vars (`ErrInvalidThreshold`); wrap with `fmt.Errorf("…: %w", err)`. Never `panic` in app code.
- **Concurrency**: `sync.RWMutex` for read-heavy state; always `defer mu.Unlock()`. Channels for goroutine signaling.
- **State**: return copies from getters; atomic file writes (write `.tmp`, then `os.Rename`).
- **Configuration**: package-level `const` blocks; env vars (`SOCKET_PATH`, `STATE_PATH`) for runtime overrides; `time.Duration` not magic numbers.
- **Comments**: godoc on exported funcs starting with the function name. Skip "what" comments for self-explanatory code.
- **Testing**: `t.TempDir()` for tmp paths; table-driven with `t.Run(name, …)`; names `TestX`; cover success and failure paths.

```go
// GetState returns a copy of the current state (thread-safe).
func (m *Manager) GetState() State {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return *m.state
}
```

## Git Identity

**Always commit with the user's git identity, never a placeholder.** Read it in this order and stop at the first hit:

1. `git config user.name` / `user.email` (per-repo).
2. `git config --global user.name` / `user.email`.
3. If neither is set: stop and ask the user — do not commit, do not invent.

Use `-c` flags so the user's config is not modified:

```bash
git -c user.name="$(git config --get user.name || git config --get --global user.name)" \
    -c user.email="$(git config --get user.email || git config --get --global user.email)" \
    commit -m "…"
```

If a commit was made with a wrong author, alert the user **before** rewriting history (`git rebase --exec 'git commit --amend --no-edit --author="<correct>"'`) and force-pushing — both are destructive and require explicit consent.

## Git Commit Messages

Keep them brief and concise.

- **Subject**: 50-72 chars, imperative mood, no trailing period. Use Conventional Commits prefixes (`feat:`, `fix:`, `chore:`, `docs:`, `refactor:`, `test:`).
- **Body**: 1-4 short lines, separated by a blank line. Explain the *why*, not the *what* — the diff already shows the *what*.
- **Anti-patterns**: multi-paragraph essays, "previously X, now Y" treatises, bodies that just restate the subject in different words.

The diff shows *what* changed. The message should answer *why*.
