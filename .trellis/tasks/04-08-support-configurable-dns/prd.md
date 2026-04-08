# Support Configurable DNS Parameter

## Goal

Allow users to specify the upstream DNS server used by `gg`'s DNS-handling path instead of relying on the current hardcoded fallback.

## Requirements

- Add a user-facing DNS parameter for temporary use from the CLI.
- Add a persisted config key so the DNS server can be stored with `gg config`.
- Preserve current behavior when the DNS parameter is not set.
- Normalize DNS server values so common inputs like `1.1.1.1` work as `1.1.1.1:53`.
- Route the configured DNS server into the proxy DNS forwarding path.

## Acceptance Criteria

- [ ] A `--dns` flag is available on the root command.
- [ ] A `dns` config key is available through `gg config`.
- [ ] When `dns` is unset, existing DNS behavior is preserved.
- [ ] When `dns` is set, DNS forwarding uses the configured server.
- [ ] DNS server values without an explicit port default to port `53`.
- [ ] `go test ./...` passes.

## Technical Approach

- Add `DNS string` to `config.Params`.
- Bind the new root flag in `cmd/config.go`.
- Pass the configured DNS server from `cmd/` into `tracer.New()` and `proxy.New()`.
- Replace the hardcoded DNS upstream handling in `proxy/udp.go` with a helper that prefers configured DNS and falls back to existing behavior.
- Add focused proxy tests for DNS target selection and normalization.

## Out of Scope

- Reworking DNS hijack semantics beyond the existing code paths
- Adding a full DNS resolver implementation
- Changing unrelated connectivity-test behavior

## Technical Notes

- Current hardcoded DNS target is in `proxy/udp.go`.
- Existing config and flag patterns are in `config/config.go`, `cmd/cmd.go`, and `cmd/config.go`.
