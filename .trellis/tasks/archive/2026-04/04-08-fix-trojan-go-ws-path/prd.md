# Fix Trojan-go WebSocket Path Configuration

## Goal

Fix Trojan-go WebSocket URL handling so the WebSocket path is stored in the URL path segment instead of as a `path` query parameter, preventing misconfigured outbound requests that can produce `404` responses.

## Requirements

- Export Trojan-go WebSocket links with the WebSocket path in the URL path.
- Do not emit a redundant `path` query parameter for Trojan-go WebSocket links.
- Continue parsing existing links that still encode the WebSocket path in the query string for backward compatibility.
- Keep non-WebSocket Trojan and Trojan-go URL behavior unchanged.

## Acceptance Criteria

- [ ] `Trojan.ExportToURL()` writes Trojan-go WebSocket paths into `URL.Path`.
- [ ] Exported Trojan-go WebSocket links do not contain a `path` query parameter.
- [ ] `ParseTrojanURL()` accepts Trojan-go WebSocket links whose path is carried in `URL.Path`.
- [ ] Existing query-parameter form still parses for compatibility.
- [ ] `go test ./...` passes.

## Technical Approach

- Update `dialer/trojan/trojan.go` so Trojan-go WebSocket export uses `u.Path`.
- Parse Trojan-go WebSocket path from `t.EscapedPath()` / `t.Path` first, then fall back to `query.Get("path")`.
- Add focused parser/export regression tests in `dialer/trojan/trojan_test.go`.

## Out of Scope

- Changing Trojan-go gRPC URL encoding
- Refactoring Trojan transport setup outside the path-handling bug
- Changing unrelated dialer URL formats

## Technical Notes

- Relevant files: `dialer/trojan/trojan.go`, `dialer/trojan/trojan_test.go`, `dialer/trojan/sagernet.go`
- Existing pattern references: `dialer/v2ray/v2ray.go` already uses URL path/query fields deliberately for transport-specific export logic
