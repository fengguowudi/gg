# Quality Guidelines

> Code quality standards for backend development.

---

## Overview

Match the existing Go codebase before trying to improve it.

The dominant quality signals in this repository are:

- small, responsibility-focused packages
- explicit error returns in reusable code
- build-tagged platform separation for Linux-specific behavior
- protocol registration through package `init()`
- focused package-local tests for parsers and helpers

There is no visible dedicated lint workflow yet. The practical baseline is to keep the tree building and tests passing with idiomatic Go changes.

---

## Forbidden Patterns

- Do not call `logrus.Fatal`, `log.Fatal`, or `os.Exit` from library packages. Restrict process termination to `cmd/` and `main.go`.
- Do not add unregistered protocol implementations. New dialer packages must register themselves through `dialer.FromLinkRegister(...)` and, when applicable, `dialer.FromClashRegister(...)`.
- Do not bypass typed config plumbing by sprinkling raw string-key mutations throughout the codebase. Add fields to `config/config.go` and use the config helpers.
- Do not hide Linux-only or architecture-only behavior in generic files when build tags are the existing pattern. Follow files such as `tracer/tracer.go`, `tracer/tracer_unsupported.go`, `cmd/infra/capability_unix.go`, and `cmd/infra/capability_stub.go`.
- Do not copy the existing secret-leaking log behavior that prints full share-links. Treat it as debt, not as guidance.

---

## Required Patterns

- Keep `main.go` thin and push user-facing orchestration into `cmd/`.
- Use explicit constructor and parser names: `NewX`, `ParseX`, `Dialer()`, `ExportToURL()`. Existing examples include `NewHTTP`, `ParseHTTPURL`, `ParseVlessURL`, and `NewIPMTUTrieFromInterfaces`.
- Return wrapped errors from reusable code and let callers decide whether to continue, warn, or exit.
- Co-locate tests with the package they verify. Prefer small focused tests over giant integration-style files.
- Use `t.Parallel()` for independent unit tests when the package already follows that pattern. Examples: `dialer/http/http_test.go`, `dialer/v2ray/reality_test.go`, `dialer/anytls/anytls_test.go`.

---

## Testing Requirements

- Run `go test ./...` for backend changes.
- If you touch protocol parsing or URL export logic, add or update a package-local test next to that protocol implementation.
- If you touch low-level lookup structures or helpers, add deterministic unit coverage like `infra/trie/trie_test.go`.
- If you change Linux-specific runtime code, verify the relevant build-tagged package still compiles on the intended target even if the full behavior is difficult to exercise locally.
- If you change a stateful transport or proxy handshake, add regression tests for both `read-first` and `write-first` usage when applicable.
- If a protocol parser hands control back to a raw stream, add a test that verifies bytes buffered during parsing are still delivered to the caller.

Reference examples:
- `dialer/http/http_test.go`: focused behavioral test for HTTP proxy auth injection.
- `dialer/v2ray/reality_test.go`: parser coverage for new URL variants.
- `infra/trie/trie_test.go`: deterministic helper test without external dependencies.
- `dialer/transport/httpproxy/conn_test.go`: handshake-direction and buffered-tunnel-data coverage for HTTP CONNECT transport behavior.

---

## Code Review Checklist

- Is the code in the right package, or is CLI, config, and runtime logic being mixed together?
- If a new config field was added, was it added to `config/config.go` with the right `mapstructure` and default tags?
- If a new protocol or format was added, did it register itself and add a focused test?
- Are errors wrapped with enough context, and are fatal exits limited to the CLI boundary?
- Are logs informative without leaking secrets or producing info-level noise for packet-level details?
- If the change is Linux- or arch-specific, does it follow the existing build-tag split instead of runtime branching?
- For transport-layer fixes, did the review consider handshake ordering assumptions and buffered bytes crossing protocol boundaries?

---

## Examples

- `cmd/cmd.go`: thin orchestration boundary plus logger setup.
- `dialer/dialer.go`: shared abstractions and registration surface.
- `dialer/http/http_test.go` and `dialer/v2ray/reality_test.go`: the preferred style for focused parser and transport tests.
- `tracer/tracer.go` and `tracer/tracer_unsupported.go`: platform separation using build tags.
