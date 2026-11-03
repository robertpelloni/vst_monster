# CHANGELOG

## [0.2.0] - 2026-11-03
### Added
- System Tray Integration: Added native desktop tray icon with context menu actions (Show Dashboard, Hide Dashboard, Quit VST Monster & Servers) in `client/src-tauri/src/lib.rs`.
- Consolidated Command Dashboard UI: Redesigned `client/src/routes/+page.svelte` into a unified single-page layout featuring live health indicators, catalog stats, search and multi-criteria filters, real-time crawler console output, interactive tooltips, and native installer buttons.
- Extended Registry Endpoints: Added `/health`, `/stats`, `/crawler/run`, `/crawler/logs`, and `/system/shutdown` in `registry/src/index.ts`.
- Subprocess Crawler Pipeline Triggering: Registry API now supports starting crawler jobs on demand and capturing log streams directly for frontend display.
- Enhanced Startup Engine: Upgraded `start.sh` with automated service checks for PostgreSQL/Redis, database migration readiness, safe trap handling, and archived old script in `scripts/archive/start.sh.bak`.

## [0.1.2] - 2026-06-15
### Added
- SHA-256 binary validation during crawler downloads.
- PostgreSQL upsert schema handling with unique constraints.
- Initial Svelte 5 frontend with basic grid.
