# VST Monster

A three-part system for cataloging audio plugins (VST/VST3/AU/CLAP) and
installing them from a desktop client. Per `VISION.md`, the goal is a
comprehensive plugin registry with one-click installs; what exists is a
working-shape local pipeline built in agent sessions during June 2026.

## Components

- **`crawler/` (Go)** — a Colly-based scraper for plugin sources (KVR,
  GitHub, Plugin Boutique per `HANDOFF.md`) with `scraper/`, `parser/`,
  `queue/`, and `proxy/` packages. Downloads are hashed (SHA-256, streamed
  to disk) and upserted into PostgreSQL via `lib/pq`
  (`ON CONFLICT (name, developer) DO UPDATE`). Includes a test file
  (`downloader_test.go`).
- **`registry/` (TypeScript)** — an Express + PostgreSQL REST API
  (`src/index.ts`, `schema.sql` with `plugins`, `plugin_releases`,
  `plugin_distributions` tables). Serves `GET /plugins` on port 3000.
- **`client/` (Tauri + Svelte 5)** — desktop app whose UI fetches the
  registry API and hands downloads to a Rust/Tauri install engine
  (`client/src-tauri/`).

`start.sh` at the root ties the pieces together; `HANDOFF.md` is the
detailed (agent-written) session log and the most accurate map of what was
built and what was planned next (installer strategies, more crawler
targets).

## Status

Prototype. The components were developed and run locally; there is no
hosted registry, no release, and the only recent commits are dependabot
bumps (August 2026). The README this replaces contained nothing but a
banner and a link to this repo — the description above is reconstructed
from the code and `HANDOFF.md`.
