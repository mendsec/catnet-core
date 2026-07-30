# AGENTS.md — catnet-io/engine

This file provides persistent context for AI coding agents (Antigravity, Jules, OpenHands,
Claude Code) working in the `catnet-io/engine` repository.

---

## What this repository is

`catnet-io/engine` is the shared Go scanning engine for the CatNet ecosystem.
It is a pure Go library — no binary, no main package, no UI code.
All network scanning logic lives exclusively here.
Consumer repositories (`catnet-io/catnet`, `catnet-io/app`, `catnet-io/tui`) are
pure frontends that import this module via `go.mod`.

**Module path:** `github.com/catnet-io/engine`  
**Go version:** 1.26.4  
**Current stable tag:** v0.7.0

---

## Architecture

### Package map

| Package | Responsibility | Stability |
|---|---|---|
| `pkg/engine` | Callback-based scan API (`StartScan`, `EventCallback`, `ScanConfig`) | Stable |
| `pkg/scan` | Channel-based scan API (`Engine`, `ScanStream`, `Stop`) — canonical for GUI/TUI consumers | Stable |
| `pkg/events` | Event types for the channel API (`Event`, `EventType`, `HostDiscoveredData`, `ProgressData`) | Stable |
| `pkg/profile` | Scan configuration for channel API (`ScanProfile`, `DefaultProfile`, `Sanitize`) | Stable |
| `pkg/results` | Domain types (`ScanReport`, `DeviceInfo`, `HostResult`) | Stable |
| `pkg/discovery` | ICMP ping, ARP MAC lookup, reverse DNS | Stable |
| `pkg/ports` | TCP port scanner, returns `<-chan int` | Stable |
| `pkg/fingerprint` | OS/device fingerprinting via TTL, banner, OUI | Stable |
| `pkg/targets` | IP range parsing (CIDR, dash range) | Stable |
| `pkg/topology` | Network graph builder, gateway identification | Stable |
| `pkg/export` | JSON/CSV export for `[]results.HostResult` | Stable |
| `pkg/exporter` | JSON/CSV/XML export for `*results.ScanReport` | Stable |
| `pkg/store` | SQLite scan history — **scheduled for removal** (move to consumers in Sprint 3) | Deprecated |
| `pkg/diff` | Scan comparison — **scheduled for removal** (move to consumers in Sprint 3) | Deprecated |
| `pkg/coreerr` | Typed error taxonomy (`ErrTimeout`, `ErrCancelled`) | Stable |
| `internal/netutil` | Internal network utilities — not part of public API | Internal |

### Two scan APIs — read before editing

There are currently two scan APIs in this repository. This is known architectural drift
being resolved in Milestone 5:

- **`pkg/engine.StartScan`** — synchronous callback API. Used by `catnet-io/catnet` (CLI).
  Each `EventCallback` is called synchronously by the scan worker goroutine.
  Milestone 5 will add an internal async dispatcher to decouple callback latency from worker throughput.

- **`pkg/scan.Engine.ScanStream`** — asynchronous channel API. Used by `catnet-io/app` (GUI).
  `ScanStream` delegates to `pkg/engine` internally and translates events to the channel.
  This is the **preferred API for new consumers**.

Do not merge or remove either API without explicit instruction. Do not add a third scan API.

---

## Hard rules — never violate

1. **Zero CGO.** No CGO in any file in this repository. `modernc.org/sqlite` (pure Go) is
   the only exception and is scheduled for removal.
2. **Zero UI code.** No Wails bindings, no Bubble Tea imports, no terminal output.
   This is a library — it has no `main` package.
3. **No scanning logic in consumers.** If you find yourself adding discovery, port scanning,
   or fingerprinting logic outside this repository, stop and add it here instead.
4. **English only everywhere.** All code, comments, godoc, log messages, error strings, PR descriptions, PR review comments, commit messages, and documentation across the entire repository must be in English. Portuguese and non-English text are strictly forbidden.
5. **No local `replace` directives committed to `main`.** Use `scripts/dev-replace.sh on/off`
   to toggle during local development.
6. **`pkg/store` and `pkg/diff` are deprecated.** Do not add new functionality to these
   packages. Do not create new callers inside this repository.
7. **Immutable GitHub Action Pinning.** All GitHub Actions in workflow files must be pinned to full 40-character commit SHAs. Never use unpinned tags (e.g. `@v4`).
8. **No `squash and merge`.** Never use squash merges (`gh pr merge --squash` or GitHub UI squash) for PRs in this repository. All PR merges must preserve atomic commit history via merge commits (`gh pr merge --merge`) or rebase merges (`gh pr merge --rebase`) to maintain DevSecOps traceability, commit provenance, and auditability.

---

## Conventions

### Commit messages — Conventional Commits

```
feat(engine): add async event dispatcher
fix(fingerprint): guard against nil BannerGrabConfig
chore(deps): update golang.org/x/sys to v0.45.0
test(ports): add race detector test for ScanPorts cancellation
docs(contracts): document EventCallback deprecation timeline
refactor(scan): extract asyncDispatcher to internal package
perf(topology): reduce allocations in edge key generation
```

Scope must match the package or area modified: `engine`, `scan`, `events`, `profile`,
`results`, `discovery`, `ports`, `fingerprint`, `targets`, `topology`, `export`,
`exporter`, `store`, `diff`, `coreerr`, `netutil`, `deps`, `ci`, `contracts`.

### Changelog — Keep a Changelog

Every PR that changes behavior must update `CHANGELOG.md` under `[Unreleased]`.
Sections: `Added`, `Changed`, `Fixed`, `Security`, `Deprecated`, `Removed`.
Breaking changes must start with `**BREAKING CHANGE**:`.

### Testing

- All new public functions must have unit tests.
- Concurrency-sensitive code must be tested with `-race`.
- New parsing functions must have fuzz tests in `_test.go` files with `Fuzz` prefix.
- Integration tests live in `tests/integration_test.go`.
- Do not use `t.Parallel()` in tests that touch shared network state.

### Go style

- `gofmt` and `goimports` on all files.
- `golangci-lint` must pass (see `.golangci.yml` for enabled linters).
- Zero allocation patterns for hot paths — see existing benchmarks in
  `pkg/fingerprint/oui_bench_test.go` and `pkg/topology/builder_bench_test.go`.
- Prefer `context.Context` as first parameter in all public functions that do I/O.
- `ScanConfig.Sanitize()` must be called before any scan — the engine calls it defensively
  in `StartScan`, but document this if exposing new entry points.

---

## Current milestone: M4 complete, M5 in progress

**Milestone 5 tasks (next):**
1. Implement internal async dispatcher in `pkg/engine` (buffered channel, separate goroutine)
2. Add `TestSlowCallbackDoesNotStallWorkers` — slow callback must not delay scan > 5%
3. Add `TestNoGoroutineLeakOnPrematureCancel` — zero goroutine leak after context cancel
4. Deprecate `pkg/engine.StartScan` in godoc (prefer `pkg/scan.Engine.ScanStream`)
5. Remove `pkg/store` and `pkg/diff` (coordinate with `catnet-io/app` Sprint 3)

See `docs/milestones/milestone5_plan.md` for full specification.

---

## API stability contracts

Read `docs/contracts/api-stability.md` before modifying any public API.
Read `docs/contracts/events.md` before modifying `pkg/events`.
Read `docs/contracts/compatibility.md` before introducing breaking changes.

Stable packages require a BREAKING CHANGE entry in CHANGELOG and a major version bump
if the change removes or renames exported symbols.

---

## CI requirements — all must pass before merge

- `go build ./...`
- `go test -race ./...`
- `go vet ./...`
- `golangci-lint run` (via `golangci-lint.yml` workflow)
- `govulncheck ./...` (via `govulncheck.yml` workflow)

---

## Legacy/Local Model Configuration (Migrated from Continue)

The following settings were migrated from `.continue/agents/lmstudio.yaml` and serve as reference for configuring local developer assistants:

- **Name:** Fabio LM Studio
- **Schema:** v1
- **Supported Models (via LM Studio):**
  - **LM Studio Auto** (Provider: `lmstudio`, Model: `AUTODETECT`, API Base: `http://localhost:1234/v1`)
  - **Qwen 2.5 Coder 7B** (Provider: `lmstudio`, Model: `qwen2.5-coder-7b-instruct`, API Base: `http://localhost:1234/v1`)
  - **Qwen 2.5 Coder 14B** (Provider: `lmstudio`, Model: `qwen2.5-coder-14b-instruct`, API Base: `http://localhost:1234/v1`)
  - **Qwen 3.6 27B** (Provider: `lmstudio`, Model: `qwen3.6-27b`, API Base: `http://localhost:1234/v1`)
- **Default Context Providers:**
  - `code`
  - `docs`
  - `diff`
  - `terminal`
  - `problems`
  - `folder`
