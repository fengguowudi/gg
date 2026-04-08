# Error Handling

> How errors are handled in this project.

---

## Overview

Reusable packages return `error`. The CLI layer decides when to log fatally and exit.

The dominant pattern is:

1. Define sentinel errors at package scope when callers need to branch on them.
2. Wrap lower-level failures with `fmt.Errorf(... %w ...)`.
3. Use `errors.Is` or `errors.As` at the boundary where different handling is required.
4. Reserve `logrus.Fatal` and `os.Exit` for the command and entrypoint layer.

---

## Error Types

This codebase mostly uses sentinel errors instead of custom structs.

Examples:
- `dialer/dialer.go`: `ConnectivityTestFailedErr`, `UnexpectedFieldErr`, `InvalidParameterErr`
- `config/bind.go`: `ErrRequired`, `ErrMutualReference`, `ErrOverlayHierarchicalKey`
- `cmd/infra/capability_unix.go`: `ErrGetPtraceScope`, `ErrGetCapability`, `ErrBadPtraceScope`, `ErrBadCapability`
- `tracer/tracer_unsupported.go`: `ErrUnsupportedPlatform`
- `cmd/dialer.go`: `UnableToConnectErr`

Use package-level sentinels when the caller needs to distinguish one failure class from another. Otherwise return a wrapped descriptive error.

---

## Error Handling Patterns

- Constructors and parsers return wrapped errors instead of logging. Examples: `dialer/http/http.go`, `dialer/v2ray/v2ray.go`, `cmd/infra/capability_unix.go`.
- Long-running loops usually log recoverable per-connection errors and continue instead of crashing the process. Examples: `proxy/proxy.go`, `tracer/tracer.go`.
- Cancellation and timeout are handled explicitly with `errors.Is` and `errors.As` where needed. Examples: `cmd/cmd.go` ignores `context.Canceled`; `dialer/dialer.go` normalizes timeout failures via `errors.As(err, &netErr)`.
- Shared packages preserve the original cause when adding context. Prefer `fmt.Errorf("what failed: %w", err)` over string-only errors when there is an underlying error to keep.

Message style is simple and lower-case. Do not add trailing punctuation unless the string intentionally contains multi-line shell guidance.

---

## API Error Responses

This project is a CLI, not an HTTP API.

The user-facing contract is:

- library packages return `error`
- `cmd/` converts important failures into fatal log messages
- the process exits non-zero from `cmd/` or `main.go`

Examples:
- `cmd/cmd.go`: fatal exits for command lookup, dialer creation, tracer startup, and wait failures
- `cmd/config.go`: fatal exits for invalid flag combinations and invalid config writes
- `main.go`: exits with status `1` when `cmd.Execute()` returns an error

---

## Common Mistakes

- Do not call `logrus.Fatal`, `log.Fatal`, or `os.Exit` from reusable packages. That behavior belongs in `cmd/` and `main.go`.
- Do not replace wrapped errors with generic text when the original cause matters. Keep `%w` in the chain so callers can inspect it.
- Do not use `panic` for normal validation or user-input failures. The existing `panic("unexpected flag")` in `cmd/config.go` is an internal invariant and should not become a pattern.
- Do not silently swallow repeated runtime errors unless the surrounding loop intentionally degrades and logs them at an appropriate level.

---

## Examples

- `dialer/dialer.go`: sentinel errors plus wrapped timeout and connectivity errors.
- `config/bind.go`: validation errors that preserve both the key and the underlying failure.
- `cmd/infra/capability_unix.go`: wrapped system-call failures with caller-visible sentinel categories.
- `proxy/proxy.go`: recoverable runtime errors are logged and the server loop continues.
