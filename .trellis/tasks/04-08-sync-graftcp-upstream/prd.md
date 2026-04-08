# Brainstorm: Sync Graftcp Upstream

## Goal

Sync relevant recent patches and features from the upstream `hmgle/graftcp` project into `gg`, while preserving `gg`'s Go-native architecture and existing CLI behavior where possible.

## What I already know

- `gg` is inspired by `graftcp` but is a pure Go implementation with additional protocol support.
- This repo currently has no active Trellis task and this work has been captured as `04-08-sync-graftcp-upstream`.
- The request explicitly asks for the latest upstream patches and features, so the upstream state must be verified against the current `hmgle/graftcp` repository before implementation.
- `gg` uses a Go CLI architecture centered around `cmd/`, `dialer/`, `proxy/`, and `tracer/`.
- Upstream latest `master` HEAD currently resolves to commit `6b8e7e6` dated 2026-04-02.
- Upstream latest tag is `v0.7.4` dated 2024-07-30, while GitHub's latest published release page still shows `v0.7.1` from 2023-12-17.

## Assumptions (temporary)

- The user wants practical feature parity where upstream changes map cleanly onto `gg`, not a literal line-by-line port.
- Some upstream changes may not apply because `graftcp` and `gg` have different implementations and dependency stacks.
- The work may need to be scoped to a subset of upstream changes if the latest delta is large.

## Open Questions

- None. The scope is locked to low-risk compatible sync items only.

## Requirements (evolving)

- Identify the current latest upstream state of `hmgle/graftcp`.
- Compare upstream recent patches and features against this repository.
- Propose a concrete set of changes that can be safely ported to `gg`.
- Implement the chosen subset in a way that follows existing project conventions.
- Preserve `gg`'s Go-native architecture instead of copying upstream C or split-process design literally.
- Implement the `connect(2)` destination-restore fix in `gg`'s tracer path.
- Align TCP relay behavior with upstream's half-close preservation semantics where it maps directly to `gg`.
- Add regression tests for directly synced behavior and for already-equivalent address filtering behavior.

## Acceptance Criteria (evolving)

- [x] Upstream latest commits or release state are documented in this task.
- [x] A concrete compatibility and scope decision is recorded before implementation.
- [ ] Selected upstream-compatible fixes are implemented in `gg`.
- [ ] Regression tests cover the synced behavior.
- [ ] `go test ./...` passes after the changes.

## Definition of Done (team quality bar)

- Tests added or updated where appropriate
- `go test ./...` passes
- Docs or notes updated if user-facing behavior changes
- Risky or incompatible upstream items are explicitly called out

## Out of Scope (explicit)

- Blindly mirroring every upstream file or implementation detail
- Rewriting `gg` into graftcp's architecture
- Pulling in incompatible behavior without verifying it fits `gg`
- Importing `graftcp-local` / `mgraftcp` proxy-selection features such as `select_proxy_mode`, `http_proxy` fallback, or SOCKS5-HTTP direct selection modes
- Adding IP blacklist or whitelist file support in this task

## Technical Approach

- `tracer/stop_handler.go`
  - Save the original `connect(2)` sockaddr pointer, bytes, and length after rewriting the syscall arguments.
  - Restore the original sockaddr and argument length during the syscall exit stop.
- `proxy/tcp.go`
  - Preserve half-closed TCP streams by only forcing deadlines on error paths, not on clean EOF-driven half-closes.
- Tests
  - Add a runnable proxy regression test for half-closed relay behavior.
  - Add Linux-only tracer tests for connect restore state handling and address-filtering behavior.

## Decision (ADR-lite)

**Context**: Upstream `graftcp` has continued to receive tracer and relay fixes, but `gg` differs architecturally and should not inherit incompatible CLI or local-proxy behavior wholesale.

**Decision**: Use the compatible-sync-only approach. Port only the upstream fixes that map directly onto `gg`'s tracer and relay internals, and add regression tests around those paths.

**Consequences**:

- `gg` gains safer tracer behavior without changing the CLI surface.
- User-visible parity features from `graftcp-local` and `mgraftcp` remain out of scope.
- Linux-only tracer tests may need compile-only verification on non-Linux development hosts.

## Implementation Plan

- Step 1: Add connect redirection restore state to the tracer entry and exit flow.
- Step 2: Update `RelayTCP` to preserve half-closed streams without imposing deadlines on clean EOF.
- Step 3: Add regression tests for relay half-close behavior and Linux tracer behavior.
- Step 4: Run `go test ./...` and a Linux tracer compile check.

## Technical Notes

- Initial local references: `README.md`, `README_zh.md`, `tracer/tracer.go`
- Relevant local packages likely include `cmd/`, `proxy/`, `tracer/`, and low-level Linux support files

## Relevant Specs

- `.trellis/spec/backend/directory-structure.md`: keep changes within existing `proxy/` and `tracer/` boundaries
- `.trellis/spec/backend/error-handling.md`: preserve error-return boundaries and avoid fatal exits in reusable code
- `.trellis/spec/backend/logging-guidelines.md`: avoid introducing noisy or secret-bearing logs
- `.trellis/spec/backend/quality-guidelines.md`: add focused package-local tests and preserve build-tag separation

## Code Patterns Found

- Tracer syscall rewriting flow: `tracer/stop_handler.go`
- TCP relay flow and connection shutdown behavior: `proxy/tcp.go`
- Linux build-tag split for tracer internals: `tracer/tracer.go`, `tracer/syscall_linux_amd64.go`, `tracer/tracer_unsupported.go`

## Files to Modify

- `tracer/stop_handler.go`: add `connect(2)` original sockaddr restoration
- `proxy/tcp.go`: preserve half-closed TCP streams on clean EOF
- `proxy/tcp_test.go`: add relay regression test
- `tracer/stop_handler_test.go`: add Linux-only tracer regression tests

## Research Notes

### Upstream state verified

- Upstream repository: `https://github.com/hmgle/graftcp`
- Latest `master` commit inspected: `6b8e7e6` on 2026-04-02
- Latest tags inspected locally from upstream clone:
  - `v0.7.4` on 2024-07-30
  - `v0.7.3` on 2024-07-18
  - `v0.7.2` on 2024-02-22
- GitHub release page still shows `v0.7.1` on 2023-12-17 as the latest published release entry

### Recent upstream changes that appear relevant

- `9e6355b`: preserve half-closed TCP connections in the local relay path
- `75a1105`: bypass IPv6 loopback `::1` when ignore-local mode is enabled
- `d155856`: auto mode should try HTTP proxy before falling back to direct
- `214f8b8`: fix IP file loading leak and trim behavior
- `52d79e2`: restore original `connect(2)` destination address after syscall exits
- `f1dc319`: handle IPv4-mapped IPv6 addresses in ignore filtering
- `8cc6dfe` and `546b62c`: low-level ptrace utility correctness fixes in the C implementation

### Constraints from our repo

- `gg` has no `graftcp-local` split process; it uses an in-process Go proxy plus tracer.
- Some upstream features such as `select_proxy_mode`, `http_proxy` fallback mode, and IP allow-deny config files do not map 1:1 onto `gg`'s dialer and config model.
- `gg` already appears to cover some upstream fixes functionally:
  - IPv6 loopback and IPv4-mapped IPv6 handling are already accounted for in `tracer/stop_handler.go`
  - Go's own `syscall.PtracePokeData` already handles partial trailing writes correctly, so the upstream C tail-write fix is probably not applicable directly
  - `proxy/RelayTCP` already performs `CloseWrite()` on half-closed copies, though this should still be verified by tests

### Feasible approaches here

**Approach A: Compatible sync only** (Recommended)

- How it works:
  - Port only the upstream fixes and features that map cleanly onto `gg`'s current architecture
  - Likely includes the `connect(2)` destination restore fix and targeted regression tests
  - Add tests to lock in already-equivalent behavior where upstream recently fixed bugs
- Pros:
  - Lowest risk
  - Delivers real upstream compatibility value quickly
  - Keeps CLI surface stable
- Cons:
  - Does not attempt broader feature parity with newer graftcp UX options

**Approach B: Broader parity pass**

- How it works:
  - Include Approach A plus selected user-facing parity features that `gg` does not currently expose, such as `--user`-style execution or other compatible tracer options
- Pros:
  - Closer feature parity with modern graftcp
  - More visible user-facing gains
- Cons:
  - More design work, more OS-specific edge cases, and more risk of changing `gg` semantics

**Approach C: Full parity-minded sync**

- How it works:
  - Try to cover both low-level patches and most recent user-facing graftcp features, including configuration and selection behavior
- Pros:
  - Maximum upstream coverage
- Cons:
  - Highest scope and risk
  - Several upstream concepts do not map cleanly to `gg`, so this likely turns into a partial redesign rather than a patch sync
