# Logging Guidelines

> How logging is done in this project.

---

## Overview

The project uses `github.com/sirupsen/logrus`.

`cmd.NewLogger()` creates the logger and maps verbosity flags to levels:

- no `-v`: `WarnLevel`
- `-v`: `InfoLevel`
- `-vv` or more: `TraceLevel`

The codebase mostly uses plain message logging (`Infof`, `Warnf`, `Tracef`) instead of structured fields. Keep new logging consistent with that style unless there is a compelling reason to standardize more broadly.

---

## Log Levels

- `Trace`: noisy diagnostics for parsing attempts, config source selection, node test results, packet flow, and syscall tracing. Examples: `cmd/subscription.go`, `cmd/config.go`, `proxy/udp.go`, `tracer/tracer.go`.
- `Info`: important runtime milestones and recoverable operational notes. Examples: selected node messages in `cmd/dialer.go`, UDP capability note in `cmd/cmd.go`, listener read errors in `proxy/proxy.go`.
- `Warn`: degraded-but-continuing situations that need user attention. Examples: unexpected select mode fallback in `cmd/dialer.go`, connection handling warnings in `proxy/proxy.go`.
- `Error` or `Fatal`: unrecoverable CLI failures or unusual signals. Fatal logging is concentrated in `cmd/cmd.go`, `cmd/config.go`, and `cmd/infra/su.go`.

There is no separate debug level in practice. Use `Trace` for verbose diagnostics.

---

## Structured Logging

The current codebase is not using `WithFields` or a JSON log formatter. Messages are human-readable strings.

When adding logs:

- include the key values directly in the message text
- prefer concise messages that identify the operation and the target
- keep packet-by-packet or syscall-by-syscall details at `Trace`
- do not mix multiple logging styles inside one subsystem

Examples:
- `cmd/config.go`: `Using config file: ...`, `Config:\n...`
- `cmd/dialer.go`: `Use the node: ...`, `Test nodes: ...`
- `tracer/stop_handler.go`: trace logs include pid, fd, syscall, and target address directly in the message

---

## What to Log

- Configuration source and important derived decisions, especially when behavior changes based on flags or config. Example: `cmd/config.go`.
- Node selection, subscription pull progress, and connectivity-test progress. Example: `cmd/dialer.go`.
- Recoverable runtime errors for individual sockets or connections. Examples: `proxy/proxy.go`, `tracer/tracer.go`.
- Protocol parsing fallback attempts and reasons for skipping invalid entries at `Trace`. Example: `cmd/subscription.go`.

---

## What NOT to Log

- Do not log full share-links, proxy credentials, passwords, or raw secret-bearing URLs at `Info`, `Warn`, or `Error`.
- Do not log raw packet payloads or byte buffers above `Trace`. `proxy/udp.go` already contains very noisy trace logging; treat that as the upper bound.
- Do not add fatal logs in reusable packages.

Important project-specific caveat:

- `cmd/dialer.go` currently logs `d.Link()` when caching the last subscription node. That exposes the full share-link and should be treated as legacy behavior, not as a pattern to copy into new code.

---

## Examples

- `cmd/cmd.go`: verbosity-to-level mapping and fatal CLI failures.
- `cmd/config.go`: trace logging for config discovery and effective settings.
- `cmd/dialer.go`: info and warn logging for node selection flow.
- `proxy/proxy.go` and `tracer/tracer.go`: runtime diagnostics for recoverable operational issues.
