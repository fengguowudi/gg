# Database Guidelines

> Database patterns and conventions for this project.

---

## Overview

This project does not have a database layer, ORM, or migration system.

The only persisted state is configuration written to TOML files through Viper-backed helpers. Runtime lookup data is kept in memory and rebuilt on startup.

Treat this file as the persistence guide for the project, not as an instruction to invent a database abstraction that does not exist.

---

## Query Patterns

- Read persisted settings through the typed config path: `getConfig()` in `cmd/config.go` loads Viper, `config.NewBinder(v).Bind(...)` applies defaults, and `v.Unmarshal(&config.ParamsObj)` materializes the typed struct.
- Validate config writes against the typed struct before mutating the backing map. `cmd/config.go` uses `config.SetValueHierarchicalStruct(...)` before `config.SetValueHierarchicalMap(...)`.
- Persist config with `WriteConfig(...)`, which encodes TOML and writes with `0600` permissions.
- For high-frequency runtime lookups, prefer specialized in-memory structures instead of fake persistence layers. Examples: `infra/trie/trie.go`, `infra/ip_mtu_trie/ip_mtu_trie.go`, and the address-domain mappers in `proxy/`.

Examples:
- `cmd/config.go`: config file discovery, read, validate, and write path.
- `config/config.go`: typed schema for all persisted keys.
- `config/bind.go`: hierarchical binding and default application.

---

## Migrations

There are no schema migrations today.

Compatibility is handled by keeping config parsing tolerant and default-driven:

- Add new persisted fields to `config/config.go` with `mapstructure` tags and `default` tags where appropriate.
- Keep config loading backward-compatible through the binder instead of adding one-off rewrite scripts.
- Use tolerant decoding where external formats are inconsistent. The project already does this for JSON string-number ambiguity via `common/jsoniter.go` and `extra.RegisterFuzzyDecoders()` in `main.go`.

If a future feature truly needs durable structured storage, document that change explicitly and update this file instead of copying generic ORM guidance.

---

## Naming Conventions

- Persisted config keys are snake_case and dot-separated by hierarchy: `subscription.link`, `subscription.cache_last_node`, `cache.subscription.last_node`, `test_node_before_use`.
- Config structs use exported Go field names with `mapstructure` tags that match persisted key names. See `config/config.go`.
- Config file locations are explicit:
  - `${XDG_CONFIG_HOME}/gg/config.toml`
  - `${HOME}/.ggconfig.toml`
  - `/etc/ggconfig.toml`

---

## Common Mistakes

- Do not introduce repository or ORM layers for this codebase unless the product actually grows a database requirement.
- Do not write config files by concatenating strings. Use `WriteConfig(...)` so the encoding and permissions stay consistent.
- Do not mutate arbitrary Viper keys without adding the typed field and validation path first.
- Do not treat runtime maps as durable state. Structures in `proxy/`, `tracer/`, and `infra/` are process-local caches, not storage.

---

## Examples

- `cmd/config.go`: primary persistence boundary for config files.
- `config/config.go`: canonical list of persisted settings and defaults.
- `config/bind.go`: typed hierarchical binding and validation helpers.
- `proxy/proxy.go`: example of runtime-only state that should not be persisted.
