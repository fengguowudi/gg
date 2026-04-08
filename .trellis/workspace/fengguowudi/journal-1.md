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
