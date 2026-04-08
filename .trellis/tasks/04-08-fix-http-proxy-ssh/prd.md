# Fix HTTP Proxy CONNECT Handshake for SSH

## Goal

Make HTTP proxy tunneling work reliably for SSH and other non-HTTP protocols by fixing CONNECT handshake timing and buffered tunnel data handling.

## Requirements

- Allow the HTTP proxy dialer to establish a CONNECT tunnel even when the client reads before its first write.
- Preserve any bytes buffered while parsing the proxy's CONNECT response.
- Keep existing HTTP request proxying behavior unchanged.
- Add focused regression tests for SSH-like tunnel behavior.

## Acceptance Criteria

- [ ] CONNECT handshake can be triggered from `Read()` when no application write has happened yet.
- [ ] Bytes arriving immediately after the proxy's `200 Connection Established` response are still readable by the client.
- [ ] Existing HTTP request path still works.
- [ ] `go test ./dialer/http ./dialer/transport/httpproxy ./...` passes.

## Technical Approach

- Refactor `dialer/transport/httpproxy/conn.go` to centralize handshake state in helper methods.
- Trigger CONNECT from `Read()` when the tunnel is not yet established.
- Reuse a `bufio.Reader` for post-CONNECT reads so buffered tunneled bytes are not dropped.
- Add package-local tests that simulate a proxy replying with CONNECT success and immediate SSH banner bytes.

## Out of Scope

- Changing HTTP proxy authentication semantics
- Reworking unrelated tracer or TCP relay code
- Adding SOCKS-specific behavior

## Technical Notes

- Relevant files: `dialer/transport/httpproxy/conn.go`, `dialer/http/http_test.go`
- The issue likely affects protocols like SSH that may read before writing or receive server bytes immediately after CONNECT succeeds.
