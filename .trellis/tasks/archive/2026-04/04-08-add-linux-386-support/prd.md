# Add Linux/386 Support and Linux-only Release Workflow

## Goal

Add working Linux/386 support to the project and update the GitHub Actions release workflow so automated builds target Linux variants only, including `linux/386`, while removing Windows targets.

## Requirements

- Make the project build successfully for `GOOS=linux GOARCH=386`.
- Add the missing tracer architecture support needed for Linux/386.
- Update the release workflow to build Linux targets only.
- Include `linux/386` in the release workflow matrix.
- Remove Windows targets from automated workflow builds.
- Update project documentation to reflect Linux/386 support.

## Acceptance Criteria

- [ ] `GOOS=linux GOARCH=386 go build ./...` passes.
- [ ] `GOOS=linux GOARCH=386 go test -c ./tracer` passes.
- [ ] GitHub Actions release workflow builds Linux targets only.
- [ ] GitHub Actions release workflow includes `linux/386`.
- [ ] README support list marks `Linux/386` as supported.
- [ ] `go test ./...` still passes on the current development host.

## Technical Approach

- Add a Linux/386 tracer syscall adaptation file under `tracer/`.
- Decode 32-bit x86 socket-related syscalls through `SYS_SOCKETCALL` where needed.
- Keep existing tracer logic architecture-independent by routing architecture-specific decoding through helper functions.
- Simplify the release matrix to Linux-only targets and keep the existing asset naming conventions.

## Out of Scope

- Adding non-Linux 386 targets
- Extending support to additional non-Linux operating systems
- Refactoring unrelated release automation behavior

## Technical Notes

- Current failing gap: `tracer` does not compile for Linux/386 because the project assumes direct socket syscalls (`SYS_SOCKET`, `SYS_CONNECT`, etc.), but 386 uses `SYS_SOCKETCALL`.
- Relevant files: `tracer/stop_handler.go`, `tracer/syscall_linux_*.go`, `.github/workflows/release.yml`, `README.md`, `README_zh.md`
