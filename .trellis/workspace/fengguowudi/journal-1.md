# Journal - fengguowudi (Part 1)

> AI development session journal
> Started: 2026-04-08

---



## Session 1: Bootstrap Trellis backend guidelines

**Date**: 2026-04-08
**Task**: Bootstrap Trellis backend guidelines

### Summary

(Add summary)

### Main Changes

| Area | Description |
|------|-------------|
| Trellis bootstrap | Initialized Trellis workflow files, skills, workspace metadata, and project-level agent instructions. |
| Backend specs | Replaced the backend guideline templates with project-specific guidance derived from the `gg` Go codebase. |
| Conventions captured | Documented actual package layout, config-file persistence, error handling, logging patterns, testing expectations, and known anti-patterns. |
| Task tracking | Archived the completed `00-bootstrap-guidelines` task after the guideline work was finished. |

**Key Files**:
- `.trellis/spec/backend/index.md`
- `.trellis/spec/backend/directory-structure.md`
- `.trellis/spec/backend/database-guidelines.md`
- `.trellis/spec/backend/error-handling.md`
- `.trellis/spec/backend/logging-guidelines.md`
- `.trellis/spec/backend/quality-guidelines.md`
- `.trellis/tasks/archive/2026-04/00-bootstrap-guidelines/task.json`

**Verification**:
- `go test ./...`


### Git Commits

| Hash | Message |
|------|---------|
| `f1c5377` | (see git log) |

### Testing

- [OK] (Add test results)

### Status

[OK] **Completed**

### Next Steps

- None - task complete


## Session 2: Sync compatible graftcp upstream fixes

**Date**: 2026-04-08
**Task**: Sync compatible graftcp upstream fixes

### Summary

(Add summary)

### Main Changes

| Area | Description |
|------|-------------|
| Upstream research | Verified `hmgle/graftcp` current master and tags, then scoped the work to directly mappable low-risk fixes only. |
| Tracer compatibility sync | Ported the compatible `connect(2)` destination-restore behavior so the original sockaddr and length are restored on syscall exit after tracer rewriting. |
| Relay behavior | Aligned `gg` TCP relay behavior with upstream half-close preservation semantics on clean EOF paths. |
| Regression tests | Added focused regression tests for TCP half-close relay behavior and Linux tracer address-handling behavior. |
| Task tracking | Archived the completed `04-08-sync-graftcp-upstream` Trellis task. |

**Updated Files**:
- `proxy/tcp.go`
- `proxy/tcp_test.go`
- `tracer/stop_handler.go`
- `tracer/stop_handler_test.go`
- `.trellis/tasks/archive/2026-04/04-08-sync-graftcp-upstream/prd.md`
- `.trellis/tasks/archive/2026-04/04-08-sync-graftcp-upstream/task.json`

**Verification**:
- `go test ./...`
- `GOOS=linux GOARCH=amd64 go test -c ./tracer`


### Git Commits

| Hash | Message |
|------|---------|
| `6230e18` | (see git log) |

### Testing

- [OK] (Add test results)

### Status

[OK] **Completed**

### Next Steps

- None - task complete


## Session 3: Fix Trojan-go WebSocket path configuration

**Date**: 2026-04-08
**Task**: Fix Trojan-go WebSocket path configuration

### Summary

(Add summary)

### Main Changes

| Area | Description |
|------|-------------|
| URL export fix | Updated Trojan-go WebSocket URL export to place the WebSocket path in the URL path segment instead of emitting a `path` query parameter. |
| Parser compatibility | Updated Trojan-go URL parsing to read the WebSocket path from `URL.Path` first while keeping the legacy `path` query form as a compatibility fallback. |
| Regression tests | Added focused Trojan dialer tests covering path-based parsing, legacy query fallback, and export behavior without redundant `path` query output. |
| Task tracking | Archived the completed `04-08-fix-trojan-go-ws-path` Trellis task. |

**Updated Files**:
- `dialer/trojan/trojan.go`
- `dialer/trojan/trojan_test.go`
- `.trellis/tasks/archive/2026-04/04-08-fix-trojan-go-ws-path/prd.md`
- `.trellis/tasks/archive/2026-04/04-08-fix-trojan-go-ws-path/task.json`

**Verification**:
- `go test ./dialer/trojan`
- `go test ./...`


### Git Commits

| Hash | Message |
|------|---------|
| `33e70a6` | (see git log) |

### Testing

- [OK] (Add test results)

### Status

[OK] **Completed**

### Next Steps

- None - task complete


## Session 4: Add Linux/386 support and Linux-only release workflow

**Date**: 2026-04-08
**Task**: Add Linux/386 support and Linux-only release workflow

### Summary

(Add summary)

### Main Changes

| Area | Description |
|------|-------------|
| Linux/386 runtime support | Added Linux/386 tracer syscall adaptation, including socketcall decoding, so the project now builds correctly for `GOOS=linux GOARCH=386`. |
| Tracer structure | Refactored tracer syscall dispatch so architecture-specific decoding is routed through a common interface while keeping existing non-386 Linux behavior intact. |
| Release automation | Updated GitHub Actions release workflow to build Linux targets only, explicitly including `linux/386`, and removed Windows and other non-Linux targets from automated builds. |
| Documentation | Updated README support lists to mark `Linux/386` as supported. |
| Verification | Confirmed host tests still pass and verified Linux/386 build and tracer compile targets. |
| Task tracking | Archived the completed `04-08-add-linux-386-support` Trellis task. |

**Updated Files**:
- `.github/workflows/release.yml`
- `README.md`
- `README_zh.md`
- `tracer/stop_handler.go`
- `tracer/stop_handler_test.go`
- `tracer/syscall_kind.go`
- `tracer/syscall_decode_default.go`
- `tracer/syscall_linux_386.go`
- `.trellis/tasks/archive/2026-04/04-08-add-linux-386-support/prd.md`
- `.trellis/tasks/archive/2026-04/04-08-add-linux-386-support/task.json`

**Verification**:
- `go test ./...`
- `GOOS=linux GOARCH=386 go test -c ./tracer`
- `GOOS=linux GOARCH=386 go build ./...`


### Git Commits

| Hash | Message |
|------|---------|
| `889f9d5` | (see git log) |

### Testing

- [OK] (Add test results)

### Status

[OK] **Completed**

### Next Steps

- None - task complete


## Session 5: Support configurable DNS parameter

**Date**: 2026-04-08
**Task**: Support configurable DNS parameter

### Summary

(Add summary)

### Main Changes

| Area | Description |
|------|-------------|
| CLI and config | Added a user-facing `--dns` parameter and a persisted `dns` config key so DNS upstream selection can be controlled temporarily or stored with `gg config`. |
| DNS forwarding path | Replaced the hardcoded DNS upstream fallback in the proxy UDP path with configurable selection while preserving existing default behavior when `dns` is unset. |
| Normalization | Added DNS server normalization so values like `1.1.1.1`, `dns.google`, or IPv6 literals are normalized to include port `53` when omitted. |
| Regression tests | Added focused proxy tests covering DNS target fallback and DNS server normalization. |
| Documentation | Updated README config examples to show the new `dns` config entry. |
| Task tracking | Archived the completed `04-08-support-configurable-dns` Trellis task. |

**Updated Files**:
- `config/config.go`
- `cmd/cmd.go`
- `cmd/config.go`
- `proxy/proxy.go`
- `proxy/udp.go`
- `proxy/udp_test.go`
- `tracer/tracer.go`
- `tracer/tracer_unsupported.go`
- `tracer/stop_handler_test.go`
- `README.md`
- `README_zh.md`
- `.trellis/tasks/archive/2026-04/04-08-support-configurable-dns/prd.md`
- `.trellis/tasks/archive/2026-04/04-08-support-configurable-dns/task.json`

**Verification**:
- `go test ./proxy`
- `go test ./...`


### Git Commits

| Hash | Message |
|------|---------|
| `8a763a7` | (see git log) |

### Testing

- [OK] (Add test results)

### Status

[OK] **Completed**

### Next Steps

- None - task complete


## Session 6: Fix HTTP proxy CONNECT handshake for SSH

**Date**: 2026-04-08
**Task**: Fix HTTP proxy CONNECT handshake for SSH

### Summary

(Add summary)

### Main Changes

| Area | Description |
|------|-------------|
| HTTP proxy tunnel fix | Fixed HTTP CONNECT handshake behavior so SSH and other non-HTTP protocols can establish tunnels reliably even when reads happen before the first application write. |
| Buffered tunnel data | Preserved bytes buffered while parsing the proxy `200 Connection Established` response, preventing early tunnel payload such as SSH banners from being lost. |
| Regression tests | Added focused transport tests covering read-first CONNECT startup and immediate tunneled payload delivery after CONNECT success. |
| Break-loop documentation | Captured the bug class into backend quality guidance and the cross-layer thinking guide so future transport-layer fixes explicitly consider handshake ordering and buffered-reader ownership. |
| Template sync | Synced the updated spec guidance into `src/templates/markdown/spec/` templates. |
| Task tracking | Archived the completed `04-08-fix-http-proxy-ssh` Trellis task. |

**Updated Files**:
- `dialer/transport/httpproxy/conn.go`
- `dialer/transport/httpproxy/conn_test.go`
- `.trellis/spec/backend/quality-guidelines.md`
- `.trellis/spec/guides/cross-layer-thinking-guide.md`
- `src/templates/markdown/spec/backend/quality-guidelines.md`
- `src/templates/markdown/spec/guides/cross-layer-thinking-guide.md`
- `.trellis/tasks/archive/2026-04/04-08-fix-http-proxy-ssh/prd.md`
- `.trellis/tasks/archive/2026-04/04-08-fix-http-proxy-ssh/task.json`

**Verification**:
- `go test ./dialer/transport/httpproxy`
- `go test ./dialer/http`
- `go test ./...`


### Git Commits

| Hash | Message |
|------|---------|
| `5fa7d8d` | (see git log) |

### Testing

- [OK] (Add test results)

### Status

[OK] **Completed**

### Next Steps

- None - task complete
