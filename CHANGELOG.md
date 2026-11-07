# CHANGELOG: VST Monster

## [0.3.7] - Crawler Ingestion Stability, Platform Detection & Curated Seed Expansion
- Enhanced `crawler/db.go` with dynamic platform detection from download URLs (`linux`, `macos`, `windows`) and automated strategy assignment (`silent_install` vs `extract_binaries`).
- Added category column persistence in PostgreSQL `plugins` table during crawler upserts.
- Added detailed ingestion success logging in database pipeline.
- Implemented scraper panic recovery in `crawler/main.go` to prevent thread crashes from halting the pipeline.
- Expanded curated open-source staple seeds with Cardinal Modular Synth and Chow Tape Model.
- Bumped ecosystem version to `v0.3.7`.

## [0.3.6] - Expanded Audio Plugin Format Discovery & Format Filter Chips
- Added Linux LV2 directory scanning (`~/.lv2`, `/usr/lib/lv2`, `/usr/local/lib/lv2`) and Windows CLAP directory scanning (`%CommonProgramFiles%/CLAP`) to `get_default_scan_paths` and `scan_installed_plugins`.
- Added native `.lv2` bundle recognition and clean naming in the local scanning engine.
- Added direct audio format quick-filter pills (`🎚️ VST3`, `⚡ CLAP`) in the dashboard filter bar.
- Enhanced format filter logic to match both installed binary format and catalog inferred download format.
- Added LV2 format support to format dropdown and catalog format inference engine.
- Bumped ecosystem version to `v0.3.6`.

## [0.3.5] - Search Match Indicators & Query Refinements
- Added live search match counter badge in the Find Plugins header showing exact matched package count.
- Added quick "Clear" search shortcut button when an active query is present.
- Added visual focus/glow outline on search input when filtering plugins.
- Synchronized ecosystem version to `v0.3.5`.

## [0.3.4] - Category Discoverability Chips & Fast Navigation
- Added dedicated quick-filter pills for direct category switching (`🎹 Synths`, `🎛️ Effects`) into the primary filter row of `client/src/routes/+page.svelte`.
- Synchronized active toggle states between dropdown controls and interactive quick-filter pills.
- Bumped ecosystem version to `v0.3.4`.

## [0.3.3] - Discoverability Tag Engine, Audio Format Badges & Explorer Quick Actions
- Added dynamic keyword tag extraction (`#synth`, `#wavetable`, `#fm`, `#polyphonic`, `#analog`, `#mpe`, etc.) with interactive click-to-filter pill buttons on catalog cards.
- Added audio format catalog badging (`inferCatalogFormat`) displaying format support (`VST3`, `CLAP`, `AU / VST3`) across all plugins.
- Deep linked directory explorer actions and dynamic filter activations directly from tag pills and inspection cards.

## [0.3.2] - macOS DMG Extraction Engine & Real-time Download Progress Streaming
- Implemented automated macOS DMG volume mounting (`hdiutil attach -nobrowse -readonly`), recursive audio bundle discovery (`.vst3`, `.component`, `.clap`), and clean volume detachment (`hdiutil detach -force`).
- Implemented real-time `install_progress` event payload streaming from Tauri Rust installer engine emitting percentage and downloaded bytes.
- Hooked `listen('install_progress')` event stream to Svelte 5 dashboard updating active install percentages in real-time.

## [0.3.1] - Custom Audio Scan Paths Configuration & Manifest JSON Exporter
- Added `get_default_scan_paths` Tauri command returning operating-system native audio search paths.
- Updated `scan_installed_plugins` to accept arbitrary `extra_paths: Option<Vec<String>>` to inspect user-defined audio plugin directories.
- Implemented "⚙️ Paths" scan path manager modal in Svelte 5 dashboard with persistent `localStorage` support.
- Added "📥 Export JSON" capability to export vetted plugin manifests for offline distribution, backups, and DAW integration.

## [0.3.0] - Batch Plugin Lifecycle Controls, Quick Filter Pills & Native Bundle Validation
- Added `verify_installed_plugin` Tauri command to validate audio bundle binary integrity and file presence.
- Added "Update All" batch operation in Svelte client for updating all installed plugins with newer catalog releases.
- Added visual progress feedback and state disabled triggers for batch updating and rescan operations.
- Added interactive filter count pills (All, Installed, Updates, Not Installed) with dynamic catalog counts.

## [0.2.9] - Advanced Plugin Filtering, Sorting, and Detail Inspection Modal
- Added installed status filtering in dashboard (`all`, `installed`, `update-available`, `not-installed`).
- Added format filtering for detected plugins (`all`, `vst3`, `clap`, `au`, `vst2`).
- Added sort by installed status (`status`) prioritizing locally installed packages.
- Added full plugin detail inspection modal / preview drawer with complete metadata, installation status, path, SHA-256 digest, and action triggers.
- Accessible title buttons and "👁 Details" quick-launch buttons on every catalog card.

## [0.2.8] - Rich Installed Plugin Detection, Version Diffing & Explorer Actions
- Defined `InstalledPluginInfo` in `client/src-tauri/src/installer.rs` containing `name`, `version`, `format`, and `path`.
- Implemented deep bundle version inspection from `VERSION`, `version.txt`, and `package.json` in installed plugin directories.
- Added dynamic format badging (`✓ VST3`, `✓ CLAP`, `✓ AU`, `✓ VST2`) and version status diffing (`v1.0.0 (Latest)` vs `Installed: v1.0.0 (Update to v1.2.0)`) on dashboard cards.
- Integrated direct "📂 Files" quick-launch action button on installed cards to open specific plugin directories.

## [0.2.7] - Dynamic Registry Versioning & Cryptographic Integrity Verification
- Updated `/health` endpoint in `registry/src/index.ts` to dynamically read `VERSION.md` from the project root rather than returning a static string.
- Integrated `sha2` and `hex` into `client/src-tauri/src/installer.rs` to compute stream SHA-256 hashes during plugin binary downloads and enforce cryptographic integrity checks against expected hashes.
- Passed `expectedHash` from `client/src/routes/+page.svelte` when invoking `install_plugin`.

## [0.2.6] - Plugin Directory Explorer & Rescan Integration
- Implemented `open_plugin_directory` Tauri command in `client/src-tauri/src/installer.rs` to open system audio plugin folders in the native file manager (`xdg-open`, `open`, `explorer`).
- Registered `open_plugin_directory` in `client/src-tauri/src/lib.rs`.
- Added "VST3 Folder" opener and "Rescan" controls in `client/src/routes/+page.svelte` for instant directory navigation and catalog refresh.

## [0.2.5] - Local Installed Plugin Scanner & Dynamic Badging
- Added `scan_installed_plugins` Tauri command in `client/src-tauri/src/installer.rs` to detect installed plugins across OS audio paths (`~/.vst3`, `/Library/Audio/Plug-Ins/VST3`, etc.).
- Registered `scan_installed_plugins` in `client/src-tauri/src/lib.rs`.
- Added dynamic "INSTALLED" badges on library cards and an "Installed Only" filter toggle on the single-page dashboard (`client/src/routes/+page.svelte`).

## [0.2.4] - Clean Uninstall Engine & Dashboard Integration
- Implemented `remove_plugin` native command in `client/src-tauri/src/installer.rs` to safely remove plugin directories across Linux (`~/.vst3/`), macOS (`~/Library/Audio/Plug-Ins/VST3/`), and Windows (`Documents/VST3/`).
- Added interactive "Uninstall" button alongside "Install Plugin" on every plugin card in `client/src/routes/+page.svelte` with status tracking and tooltips.
- Enabled `"tray-icon"` feature for `tauri` dependency in `client/src-tauri/Cargo.toml` and verified clean compilation.

## [0.2.3] - Category Tagging, BPB Link Extraction, and Dashboard Filters
- Updated PostgreSQL database schema (`registry/schema.sql`) with a `category` column (synth, effect, dynamics, sampler, utility).
- Enhanced `registry/src/index.ts` to return category in `GET /plugins` and accept it in `POST /plugins`.
- Expanded Single-Page Dashboard in `client/src/routes/+page.svelte` to include Category filtering and sorting options, plus category badges on cards.
- Enhanced Bedroom Producers Blog scraper (`crawler/scrapers/bpb.go`) to extract direct download and anchor links, plus auto-categorize by article title keywords.

## [0.2.2] - Native Zip Extraction & Consolidated Dashboard Registration
- Added native ZIP archive decompression engine in `client/src-tauri/src/installer.rs` targeting Linux standard VST paths (`~/.vst3`).
- Connected crawler with standard modern browser User-Agent headers and introduced seed fallback cataloging in `crawler/main.go` for seamless catalog availability.
- Consolidated Single-Page Dashboard in `client/src/routes/+page.svelte` with live metrics, quick search, crawler execution with live streaming terminal logs, and manual plugin registration modal form.
- Added direct plugin registration API `POST /plugins` in `registry/src/index.ts`.
- Modernized startup orchestration in `start.sh` with automated service checks, non-destructively backing up earlier versions to `scripts/archive/`.

## [0.2.1] - Scraper Refinement & Ecosystem Tray Integration
- Refined GitHub scraper in `crawler/scraper/github.go` to accurately extract author/developer from repository topic links.
- Expanded asset filtering to capture `.exe`, `.msi`, `.dmg`, `.pkg`, `.zip`, and `.tar.gz` releases.
- Connected Bedroom Producers Blog scraper directly into the concurrent crawler pipeline with `sync.WaitGroup` management.
- Wired Tauri system tray quit command to trigger graceful server shutdown before exiting.

## [0.2.0] - Central Command Dashboard
- Redesigned UI into single comprehensive dashboard with tooltips and stats.
- Wired live crawler trigger and logs endpoint into Registry API and UI.

## [0.1.2] - Crawler PostgreSQL Integration
- Wired the Go Colly crawler to directly ingest and populate the PostgreSQL `plugins` and `plugin_releases` schema tables.
- Included struct mapping definitions inside Go wrapper for standard extraction rules.

## [0.1.1] - Registry & UI Wiring
- Built an Express.js backend for the Registry API with PostgreSQL integration.
- Wired the Tauri/SvelteKit frontend to fetch and display plugins from the Registry.

## [0.1.0] - Initial Setup
- Initialized project vision, roadmap, and documentation.
- Defined architecture for Crawler, Registry, and Client.
