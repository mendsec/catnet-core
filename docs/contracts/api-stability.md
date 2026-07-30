# API Stability Policy

`catnet-core` is currently in a pre-`v1.0.0` stage. However, as the shared engine for the CatNet ecosystem, we are establishing clear expectations around API stability to allow downstream consumers (CLI, TUI, GUI) to build upon it reliably.

## Stable Packages

The following packages form the core stable surface of the ecosystem. Breaking changes to these packages will be avoided whenever possible, and if required, they will be communicated via explicit version bumps and detailed changelogs.

- **`pkg/scan`**: The canonical orchestrator. `Engine.ScanStream` is the standard way to run scans using the channel-based API.
- **`pkg/engine`**: Callback-based orchestrator. `StartScan` is deprecated in favor of `pkg/scan.Engine.ScanStream`.
- **`pkg/results`**: Defines the `DeviceInfo` model and other schemas. This is the canonical schema used across the ecosystem.
- **`pkg/targets`**: Target parsing utilities like `ParseRange`.
- **`pkg/discovery`**: Discovery tools for Liveness, DNS, and MAC resolution.
- **`pkg/ports`**: Port scanning utilities.
- **`pkg/exporter`**: Exporting capabilities (JSON, CSV, XML). JSON is the canonical format.
- **`pkg/oui`**: Offline IEEE OUI database lookup package (`Lookup`, `LookupWithPrefix`).
- **`pkg/coreerr`**: Structured error taxonomy. Sentinel errors (`ErrTimeout`, `ErrCancelled`, `ErrInvalidInput`, `ErrPermission`, `ErrExport`, `ErrPartial`) are stable and safe for use with `errors.Is`. New sentinel values may be added without breaking existing checks.

## Experimental / Internal Packages

- **`pkg/fingerprint`**: Heuristic and banner-based operating system and device type detection mechanisms. Experimental and subject to change.
- **`pkg/topology`**: Adjacency graph generation from scan reports. Experimental and subject to change.

Any package under `internal/` is strictly internal and offers **zero stability guarantees**. It may be changed or removed at any time without warning. Consumers must not import `internal/` packages.

## Deprecated Packages

- **`pkg/scanner`**: **REMOVED** since v0.2.0. The package no longer exists in the codebase. All consumers should already be migrated to `pkg/engine`, `pkg/discovery`, `pkg/ports`, `pkg/targets`, and `pkg/results`.

## Execution Model: Context

The preferred and official execution model uses `context.Context` to manage timeouts, deadlines, and cancellation. Legacy cancellation methods such as `StopScan()` are deprecated and should not be used in new integrations.
