# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.7.0] - 2026-07-30

### Added
- **OUI**: Dedicated `pkg/oui` package providing an offline IEEE OUI lookup database (`Lookup`, `LookupWithPrefix`) ([#106](https://github.com/catnet-io/engine/issues/106)).
- **Fingerprint**: Heuristic OS detection using TCP Window Size analysis (`GuessOSFromTCPParams`), cross-banner correlation (Windows, Linux, Printers, NAS, Routers), and `AssetClass` taxonomy (`network`, `compute`, `iot`, `mobile`, `storage`, `unknown`) ([#107](https://github.com/catnet-io/engine/issues/107)).
- **Results / Exporter**: `AssetClass` field added to `DeviceInfo` and XML/CSV exports.
- **Results / Ports**: `WellKnownServiceName` for mapping TCP port numbers to standard service names ([#141](https://github.com/catnet-io/engine/issues/141)).
- **Discovery**: `TestPingCancellation` cancellation test with leak detection ([#140](https://github.com/catnet-io/engine/issues/140)).

### Changed
- **CI**: Upgraded `actions/setup-go` to `v7.0.0` ([#139](https://github.com/catnet-io/engine/issues/139)) and `actions/checkout` to `v7.0.1` ([#138](https://github.com/catnet-io/engine/issues/138)).
- **Docs**: API stability contract updates for Phase 2 domain expansions.

## [0.6.0] - 2026-07-24

### Added
- **Targets**: Local network interface auto-detection support via `targets.DetectLocalRange` and `ParseRange("auto")` ([#136](https://github.com/catnet-io/engine/issues/136)).
- **Fingerprint**: Active RDP negotiation probe and OS heuristics (`pkg/fingerprint`).
- **Export**: Optimized CSV export memory allocations.

### Changed
- **BREAKING CHANGE**: `Ping`, `ReverseDNS`, `GetMAC`, and `ScanPorts` wrappers in `pkg/scan` and `pkg/discovery` now accept `context.Context` as their first argument ([#132](https://github.com/catnet-io/engine/issues/132)).
- **Discovery**: `osGetMAC` in `pkg/discovery/os_posix.go` and `pkg/discovery/os_windows.go` properly respect context cancellation.
- **Governance**: Updated Hard Rule #4 in `AGENTS.md` mandating English for all repository artifacts, PR review comments, descriptions, and commit messages.

### Fixed
- **Governance**: Added Hard Rule #8 in `AGENTS.md` forbidding `squash and merge` to maintain DevSecOps traceability, commit provenance, and auditability.

### Security
- **CI**: Added explicit primary branch safety guards (`main`, `master`, `develop`) and automated unit test suite (`scripts/cleanup-branches_test.sh`) for branch cleanup workflow to prevent catastrophic branch deletion.


## [0.5.1] - 2026-07-16

### Removed
- **Engine**: Removed `pkg/store` and `pkg/diff` from engine repository (moved to consumers to keep engine light and free of database dependencies).

## [0.5.0] - 2026-07-16

### Added
- **Engine**: Asynchronous event dispatcher (`asyncDispatcher`) with 512-event buffer to decouple scan workers from slow callback executions.
- **Engine**: `TestSlowCallbackDoesNotStallWorkers` and `TestNoGoroutineLeakOnPrematureCancel` tests.
- **Docs**: API stability and scan lifecycle events contract updates.

### Deprecated
- **Engine**: `pkg/engine.StartScan` deprecated in favor of `pkg/scan.Engine.ScanStream`.

## [0.4.0] - 2026-07-16

### Added
- **Fingerprint**: `BannerGrabConfig.AggressiveSMB` allows opt-in SMB negotiate probe (off by default).
- **Fingerprint**: `FingerprintWithConfig` accepts `BannerGrabConfig` for fine-grained control.
- **Fingerprint**: `BannerConcurrency` constant (`5`) limits simultaneous banner grab connections.
- **Fingerprint + Discovery**: Fuzz targets for `VendorFromMAC`, `sanitizeBanner`, `parseProcNetArp`.
- **Engine**: `TestScanReportCompleteness` ensures `len(Devices) == Total` always.
- **Engine**: `TestNoGoroutineLeakOnCancel` verifies no goroutine leak after cancellation.
- **Topology**: `TestBuildGraph_EdgeLimit` verifies `maxEdgesPerSubnet = 200` is respected.
- **Docs**: `pkg/fingerprint`, `pkg/topology`, `pkg/coreerr` added to README packages table.
- **Docs**: `docs/examples/integration_examples.md` section 4 with topology example.
- **Docs**: `api-stability.md` lists `pkg/coreerr` as stable.

### Changed
- **BREAKING CHANGE**: `GrabBanners` signature changed — now accepts `BannerGrabConfig` as last parameter.
- **BREAKING CHANGE**: `Fingerprint` delegates to `FingerprintWithConfig` with default config.
- **Engine `StartScan`**: Unified alive/dead host paths into a single goroutine per IP, eliminating the nested goroutine for alive hosts. This fixes a race condition in `report.Devices` ordering and ensures `EventLifecycleCancel` is always the last event emitted. Port scan + fingerprint now run sequentially in the same worker goroutine.
- **Export CSV/XML**: Added `OS`, `DeviceType`, `Vendor` columns/fields. `OSFamily` omitted by design (redundant with `OS`).
- **Config**: `ScanConfig` godoc expanded with default values and sanitize limits for all fields.

### Fixed
- **discovery/parseProcNetArp**: Guard against infinite loop when `eol == 0` (empty line). `dataToSearch` now advances by at least 1 byte.
- **fingerprint/doc.go**: Package comment moved before `package` line so `go doc` renders it correctly.
- **README**: Removed `pkg/scanner` (no longer exists). Added `pkg/fingerprint`, `pkg/topology`, `pkg/coreerr`.

### Security
- **SMB probe now opt-in**: `BannerGrabConfig.AggressiveSMB=false` by default. Set to `true` only if permitted by engagement rules.

## [0.3.0] - 2026-06-24

### Added
- **pkg/results**: `HostResult` â€” tipo canÃ´nico de domÃ­nio com `Alive` e `OpenPorts` e JSON tags `alive`/`open_ports`. Substitui `DeviceInfo` como formato de intercÃ¢mbio para a API event-driven.
- **pkg/results**: `HostResult.ToDeviceInfo()` â€” conversÃ£o para compatibilidade com `DeviceInfo`.
- **pkg/profile**: `ScanProfile`, `DefaultProfile`, `Sanitize` â€” configuraÃ§Ã£o de varredura com `DefaultPorts`, `Concurrency`, `TimeoutMs` e JSON tags.
- **pkg/events**: `Event`, `EventType` (string-based), `HostDiscoveredData`, `ProgressData` â€” sistema de eventos assÃ­ncrono via channel.
- **pkg/export**: `ExportJSON`, `ExportCSV` para `[]results.HostResult` com sanitizaÃ§Ã£o CSV.
- **pkg/scan**: `Engine`, `NewEngine`, `ScanStream`, `Stop` â€” orquestrador event-driven que delega a lÃ³gica de varredura para `pkg/engine` e traduz eventos.
- **pkg/scan**: `Ping`, `ReverseDNS`, `GetMAC`, `ScanPorts` â€” wraerts delegando para `pkg/discovery` e `pkg/ports`.
- **Tests**: Testes para todos os novos pacotes (`pkg/results`, `pkg/events`, `pkg/profile`, `pkg/export`, `pkg/scan`).

### Changed
- **pkg/results/HostResult**: Mudou de type alias para `DeviceInfo` para struct independente com campos `Alive`/`OpenPorts` e JSON tags distintas (`alive`/`open_ports`). CÃ³digo existente que usa `DeviceInfo` nÃ£o Ã© afetado.
- **pkg/events/EventType**: Mudou de `int` (iota) para `string` para facilitar serializaÃ§Ã£o nos frontends Wails e TUI.

### Notes
- `pkg/engine`, `pkg/discovery`, `pkg/ports`, `pkg/exporter`, `pkg/fingerprint`, `pkg/topology`, `pkg/coreerr` preservados sem alteraÃ§Ã£o.
- Nenhuma dependÃªncia externa adicionada â€” apenas stdlib Go.

## [0.2.0] - 2026-06-21

### Added
- `FingerprintProvider` interface injetÃ¡vel em `ScanConfig`.
- `ScanPorts` refatorada para retornar `<-chan int` (canal desacopla port scan do worker).
- `parseProcNetArp` extraÃ­da como funÃ§Ã£o testÃ¡vel.
- `sanitizeBanner` para output seguro de banners.
- `ExportCSV` otimizada com `strconv.AppendInt` + reuse de slice.
- `maxEdgesPerSubnet = 200` em topology para limitar explosÃ£o O(NÂ²).
- **Topology Graph Support** (`pkg/topology/`): Complete network topology graph builder with gateway identification, device clustering by subnet, and graph export capabilities.
- **OS/Device Fingerprinting Enhancements**:
  - OUI (Organizationally Unique Identifier) module with zero-allocation lookup for MAC vendor identification.
  - TTL (Time To Live) fingerprinting module for OS detection.
  - Banner fingerprinting support for service identification.
  - Comprehensive test suites and benchmarks for all fingerprinting modules.
- **DevSecOps Documentation**: Added `docs/devsecops.md` with security practices, vulnerability scanning procedures, and fuzzing guidelines.
- **Enhanced CI/CD Infrastructure**:
  - Golangci-lint configuration (`.golangci.yml`) with strict linting rules.
  - Dependabot automation for dependency updates (`dependabot.yml`).
  - Fuzz testing CI workflow for automatic fuzzing on pull requests.
  - Govulncheck PR workflow for automated vulnerability scanning.
  - GPG-signed commit workflow for supply chain security.
  - Automated developâ†’main merge workflow with GPG signing.

### Changed
- **BREAKING CHANGE**: `ScanPorts` nÃ£o retorna mais `[]int` â€” retorna `<-chan int`. Consumidores de `pkg/ports` diretamente (fora do engine) precisam ler do canal.
- **Performance Optimizations (Zero-Allocation Patterns)**:
  - âš¡ ARP table parsing in MAC discovery now uses zero-allocation parsing.
  - âš¡ VendorFromMAC lookup optimized for minimal memory allocations with dedicated benchmarks.
  - âš¡ Subnet extraction (`/24`) in topology builder uses zero-allocation string slicing.
  - âš¡ Topology graph edge keys replaced with zero-allocation struct keys, eliminating string concatenation overhead.
- **Port Scanner Improvements**: Enhanced port scanning logic with better concurrency handling and timeout calculations.
- **CI Workflow Updates**: Refactored CI workflows for better separation of concerns and improved maintainability.
- **Dependency Updates**:
  - `actions/checkout`: Bumped from v4 to v7 for enhanced features and security fixes.
  - `actions/setup-go`: Bumped from v5 to v6 for improved Go version handling.

### Fixed
- **POSIX `osPing` Enhancement**: Improved timeout parameter handling to properly convert milliseconds to whole seconds for Linux (`-W` flag) and milliseconds for macOS.
- **Golangci-lint Version Mismatch**: Fixed CI compilation issue by compiling golangci-lint from source when necessary.

### Security
- All commits are now GPG-signed via automated workflow for supply chain integrity.
- Added regular vulnerability scanning with `govulncheck`.
- Enhanced CI linting with golangci-lint and security-focused configuration.

## [0.1.2] - 2026-06-06

### Added
- Topology graph builder foundation.
- Initial fingerprinting module structure.

### Changed
- Port scanner refactoring (PR #40 integration).

## [0.1.1] - 2026-06-06

### Added
- Added package-level documentation (`doc.go`) to all public packages and `internal/netutil`.
- Added mockable, deterministic unit tests for input validation in `pkg/discovery`.
- Added `macos-latest` to GitHub Actions CI workflow to validate POSIX builds.
- Added fuzz testing (`FuzzParseRange`) to target parsing logic to ensure resilience against malformed inputs.
- Added `os_posix_test.go` with OS-specific validations for fallback ping execution.

### Changed
- **BREAKING CHANGE**: Removed `OpenPortsCount` field from `DeviceInfo` struct to eliminate redundancy and potential state corruption. Use the new `PortCount()` method instead.
- Improved defensive timeout calculation in `StartScan` by considering concurrent port scan batches, avoiding overestimated timeouts.
- Updated minimum Go version requirement to `1.26.4` in `go.mod` and `1.26.x` in CI workflows.
- Improved `sanitizeCSVField` to block newline (`\n` / `\r`) injections and control characters within the body of CSV fields.
- Context is now explicitly propagated to blocking I/O calls (`discovery.Ping` and `ports.ScanPorts`) for immediate hard-cancellation support.

### Fixed
- Fixed pointer memory aliasing bug in `EventCallback` within `StartScan` loop, ensuring consumers receive a safe, distinct copy of `DeviceInfo`.
- Fixed `osPing` in both Windows and POSIX implementations to handle `timeoutMs <= 0` gracefully with a safe default (1000ms).
- Fixed POSIX `osPing` ignoring `timeoutMs` parameter, properly converting to whole seconds for the `-W` flag on Linux, and using milliseconds on macOS.

## [0.1.0] - 2026-06-02

### Added
- `pkg/scanner`: `DeviceInfo`, `ScanConfig`, `DefaultConfig`, `Sanitize`.
- `pkg/scanner`: `validateIPv4`, `Ping`, `ReverseDNS`, `GetMAC`, `ScanPorts`.
- `pkg/scanner`: `ParseRange` com suporte a CIDR e dash range.
- `pkg/scanner`: `StartScan`, `StopScan` com goroutines e cancelamento por context.
- `pkg/scanner`: build tags separados para Windows (`SendARP`) e POSIX (`arp`).
- `pkg/exporter`: `ExportJSON`, `ExportCSV`, `ExportXML` com sanitizaÃ§Ã£o de injeÃ§Ã£o CSV.
- CI: `go vet` + `go test -race` no GitHub Actions.
- CI: `govulncheck` semanal.

[Unreleased]: https://github.com/catnet-io/engine/compare/v0.7.0...HEAD
[0.7.0]: https://github.com/catnet-io/engine/compare/v0.6.0...v0.7.0
[0.6.0]: https://github.com/catnet-io/engine/compare/v0.5.1...v0.6.0
[0.5.1]: https://github.com/catnet-io/engine/compare/v0.5.0...v0.5.1
[0.5.0]: https://github.com/catnet-io/engine/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/catnet-io/engine/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/catnet-io/engine/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/catnet-io/engine/compare/v0.1.2...v0.2.0
[0.1.2]: https://github.com/catnet-io/engine/releases/tag/v0.1.2
[0.1.1]: https://github.com/catnet-io/engine/releases/tag/v0.1.1
[0.1.0]: https://github.com/catnet-io/engine/releases/tag/v0.1.0
