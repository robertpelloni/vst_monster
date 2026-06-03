# MEMORY: VST Monster Architectural Observations

## Codebase Traits
- **Crawler**: Built in Go for high concurrency and performance. Target sources: KVR, GitHub, developer sites.
- **Registry**: Node.js/TypeScript backend with PostgreSQL and Redis.
- **Client**: Tauri/Rust for native OS access and minimal footprint.

## Design Preferences
- De-coupled microservices architecture.
- JSONB in PostgreSQL for flexible metadata.
- SHA-256 hashing for all binaries.
- Strategy-based installation engine.

## Discovered Optimizations
- Using Go's Colly for efficient scraping.
- Using Tauri to avoid heavy Electron overhead and gain native file system access.
