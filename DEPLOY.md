# DEPLOY: VST Monster

## Crawler
- Requires Go 1.21+.
- Environment variables for proxy rotation (optional).

## Registry
- Requires Node.js 20+, PostgreSQL 15+, Redis.
- Run locally with `cd registry && npm install && npm start`.
- Standard Docker Compose setup for production.

## Desktop Client
- Requires Rust/Cargo and Tauri dependencies.
- Build via `npm run tauri build`.
