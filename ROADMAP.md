# ROADMAP: VST Monster Milestones

## Milestone 1: Core Catalog Pipeline & Silent Engine (Completed)
- [x] Concurrent Colly Go scraper pipeline with proxy support.
- [x] PostgreSQL database persistence with unique constraints and relationships.
- [x] Tauri v2 desktop application running Svelte 5 Single-Page App.
- [x] Direct-to-disk streaming downloads preventing OOM with SHA-256 verification.

## Milestone 2: Central Command Dashboard & Process Orchestration (Completed v0.2.2)
- [x] Consolidated single-page dashboard with zero page fragmentation.
- [x] Real-time crawler process execution and live console log viewer.
- [x] Tauri System Tray icon with Hide/Show and Quit App & Servers capabilities.
- [x] High-value UI organization: Quick Search & Filter, Crawler Operations Hub, Real-time Metrics.
- [x] Manual plugin registration modal and REST API endpoints.
- [x] Automated ecosystem bootstrapper (`start.sh`) with dependency healthchecks.
- [x] Native cross-platform ZIP extraction to user audio directories (`~/.vst3`).
- [x] Native clean plugin uninstall and folder removal (`remove_plugin`).
- [x] Native installed plugin filesystem scanner (`scan_installed_plugins`) & dynamic UI badging.
- [x] Native plugin directory explorer (`open_plugin_directory`) with dashboard integration.
- [x] Native cryptographic SHA-256 binary integrity enforcement during silent installs.
- [x] Unified dynamic version string delivery across Registry API and dashboard UI.
- [x] Rich installed plugin scanning (`InstalledPluginInfo`), dynamic format badging, version diffing, and folder launch integration.

## Milestone 3: Advanced Installation Engines & Package Extraction
- [ ] Implement complete macOS `.dmg` automated extraction and `.vst3` / `.component` placement to `~/Library/Audio/Plug-Ins/`.
- [ ] Implement Linux `.so` / `.vst3` / `.clap` archive extraction and installation to `~/.vst3` and `~/.clap`.
- [ ] Implement Windows zip archive decompression for standalone `.dll` and `.vst3` files.
- [ ] Add sandbox malware scanning integration (VirusTotal API / ClamAV) prior to installation.

## Milestone 4: Distributed Community Sync & P2P Cache
- [ ] BitTorrent / IPFS content-addressed distribution mirrors for large open-source plugin collections.
- [ ] Community rating, compatibility tags (DAW support, OS versions), and user preset sharing.
