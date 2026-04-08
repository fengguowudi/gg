# Directory Structure

> How backend code is organized in this project.

---

## Overview

This repository is a Go CLI backend, not a layered web service.

The execution flow is intentionally thin at the top:

```text
main.go -> cmd/ -> config/ + dialer/ + tracer/ + proxy/
```

Keep process startup and exit logic in `main.go` and `cmd/`. Keep reusable behavior in packages that return errors instead of exiting the process.

---

## Directory Layout

```text
.
|-- main.go                  # Process entrypoint and global initialization
|-- cmd/                     # Cobra commands, flag parsing, CLI orchestration
|   |-- infra/               # OS-specific command helpers
|-- config/                  # Typed config schema and Viper binding helpers
|-- dialer/                  # Protocol parsers and dialer implementations
|   |-- transport/           # Lower-level transport adapters
|   |-- sagernet/            # Shared TLS and REALITY helpers
|-- proxy/                   # Local TCP/UDP proxy runtime
|-- tracer/                  # ptrace-based interception and socket rewriting
|-- infra/                   # Generic low-level utilities
|-- common/                  # Small shared helpers
|-- completion/              # Shell completion scripts
|-- release/                 # Release scripts and asset metadata
```

Tests live next to the package they cover as `*_test.go`.

---

## Module Organization

- `main.go` should stay small. It currently registers fuzzy JSON decoders, sets the default HTTP timeout, and delegates to `cmd.Execute()`.
- `cmd/` owns user-facing behavior: Cobra commands, prompts, configuration loading, node selection, runtime startup, and fatal exits. Examples: `cmd/cmd.go`, `cmd/config.go`, `cmd/dialer.go`.
- `config/` owns the typed configuration schema and hierarchical key helpers. New persisted config fields belong here first, not scattered across command code.
- `dialer/` owns protocol-specific parsing and adapter construction. Each protocol gets its own subpackage and self-registers in `init()`. Examples: `dialer/http/http.go`, `dialer/v2ray/v2ray.go`, `dialer/trojan/trojan.go`.
- `proxy/` and `tracer/` hold the runtime engine. They should not become dumping grounds for CLI parsing or config persistence code.
- `infra/` is reserved for low-level reusable helpers that are not domain-specific, such as `infra/trie/trie.go` and `infra/ip_mtu_trie/ip_mtu_trie.go`.
- `common/` is for very small general helpers. If code becomes protocol-specific, tracer-specific, or config-specific, move it out of `common/`.

---

## Naming Conventions

- Package and directory names are short, lowercase, and responsibility-based: `cmd`, `config`, `proxy`, `tracer`, `dialer/http`, `dialer/v2ray`.
- File names usually match one concern or one platform variant: `http.go`, `reality.go`, `capability_unix.go`, `capability_stub.go`, `tracer_unsupported.go`.
- Constructors use `NewX` naming. Parsers use `ParseX` naming. Shared conversions expose explicit methods such as `Dialer()`, `URL()`, and `ExportToURL()`.
- OS- and architecture-specific behavior is split into separate files with build tags instead of large runtime branches. Examples: `tracer/tracer.go`, `tracer/tracer_unsupported.go`, `tracer/syscall_linux_amd64.go`.
- Keep tests in the same package directory as the implementation, for example `dialer/http/http_test.go` and `infra/trie/trie_test.go`.

---

## Examples

- `main.go`: thin entrypoint that performs global setup and delegates immediately to the command package.
- `cmd/cmd.go`: root command wiring, logger configuration, and process lifecycle orchestration.
- `config/config.go` plus `config/bind.go`: typed config schema plus hierarchical binding helpers.
- `dialer/dialer.go`: shared registry and wrapper abstraction used by protocol subpackages.
- `proxy/proxy.go` plus `tracer/tracer.go`: long-running runtime components kept separate from CLI and config code.

---

## Anti-Patterns

- Do not add `logrus.Fatal`, `os.Exit`, or user prompts in reusable packages such as `dialer/`, `proxy/`, or `tracer/`. Keep those in `cmd/` and `main.go`.
- Do not mix protocol-specific parsing into `dialer/dialer.go`. Shared registration stays in the registry package, while parsing lives in protocol subpackages.
- Do not place Linux-only logic in generic files when build-tagged files already exist for the same concern.
