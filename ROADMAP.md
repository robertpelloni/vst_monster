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
- [x] Advanced installed state filtering, audio format filtering, status sorting, and interactive plugin detail inspector drawer.
- [x] Native plugin bundle validation (`verify_installed_plugin`), batch plugin update orchestration, and interactive filter count pills.
- [x] Custom audio scan paths configuration modal (`get_default_scan_paths`, custom paths) with persistent `localStorage` and standalone plugin manifest JSON exporter.
- [x] macOS `.dmg` automated mounting, bundle extraction (`.vst3`, `.component`, `.clap`), and detachment.
- [x] Real-time download progress streaming via Tauri events (`install_progress`).

## Milestone 3: Advanced Installation Engines & Package Extraction
- [x] Implement complete macOS `.dmg` automated extraction and `.vst3` / `.component` placement to `~/Library/Audio/Plug-Ins/`.
- [x] Discoverability tag extraction engine with interactive click-to-filter pills.
- [x] Automated audio format detection and badging across catalog.
- [x] Category discoverability chips for direct synthesizer and effect filtering.
- [x] Search match indicators and visual query clear triggers.
- [x] Expanded audio format discovery (LV2 / CLAP) and interactive format filter chips.
- [x] Crawler data ingestion stabilization, platform inference, and curated staple expansion.
- [x] Implement Linux `.deb` and `.tar.gz` archive extraction engine (`dpkg-deb`, `tar`) and reactive format count indicators.
- [x] Implement Linux direct `.so` standalone binary relocation and permission handling (`0o755`).
- [x] Format-Aware ZIP Unpack Engine with dynamic format bundle routing (`.lv2`, `.clap`, `.vst3`).
- [x] Audio Format & Host Compatibility Matrix breakdown and LV2 Open Audio Standard integration.
- [x] Tabbed Inspector Views (Overview & Specs, Filesystem & DAW, Raw Manifest JSON).
- [x] Local Scan Metadata Caching, Rescan Acceleration & Cache Flush Integration.
- [x] Audio Unit (AU) Discoverability Chip & CoreAudio Compatibility Integration (v0.4.8).
- [x] Pre-Flight Antivirus Sandbox Scanner & Cryptographic Verification UI (v0.4.9).
- [x] Dynamic DAW Theme Switching & UI Color Profiles (v0.5.0).
- [x] Deep Audio Bundle Structural Verification & Audit Diagnostics (v0.5.1).
- [x] CLAP Open Architecture Specification Callouts & Inspection Badging (v0.5.2).
- [x] Headless automated DAW plugin bundle fuzzing and validation test suite (v0.5.3).
- [x] Implement Windows zip archive decompression for standalone `.dll` and `.vst3` files (v0.5.4).
- [x] Exported DAW host entry point symbol verification (`GetPluginFactory`, `clap_entry`, `lv2_descriptor`, `VSTPluginMain`) (v0.5.4).
- [x] LV2 manifest Turtle (`.ttl`) version & URI extraction with expanded Windows VST2 discovery & legacy architecture callouts (v0.5.5).
- [x] Native DAW preset export engine & preset storage inspector integration (v0.5.6).
- [x] Multi-format audio capabilities detection & host feature matrix inspector (v0.5.7).
- [x] Dedicated LV2 GitHub scraper topics, release asset prioritization, and manifest architecture badges (v0.5.8).
- [x] Windows MSI silent installer engine, diagnostic log aggregation & exit code decoders (v0.5.9).
- [x] Advanced installation routing engine, custom destination overrides & strategy selection (v0.6.0).

## Milestone 4: Distributed Community Sync & P2P Cache
- [ ] BitTorrent / IPFS content-addressed distribution mirrors for large open-source plugin collections.
- [ ] Community rating, compatibility tags (DAW support, OS versions), and user preset sharing.
