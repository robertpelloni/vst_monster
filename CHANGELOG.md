# CHANGELOG: VST Monster

## [0.1.2] - Crawler PostgreSQL Integration
- Wired the Go Colly crawler to directly ingest and populate the PostgreSQL `plugins` and `plugin_releases` schema tables.
- Included struct mapping definitions inside Go wrapper for standard extraction rules.

## [0.1.1] - Registry & UI Wiring
- Built an Express.js backend for the Registry API with PostgreSQL integration.
- Wired the Tauri/SvelteKit frontend to fetch and display plugins from the Registry.

## [0.1.0] - Initial Setup
- Initialized project vision, roadmap, and documentation.
- Defined architecture for Crawler, Registry, and Client.
