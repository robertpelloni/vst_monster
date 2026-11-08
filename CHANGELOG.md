# CHANGELOG: VST Monster

## [0.5.2] - CLAP Open Architecture Specification Callouts & Inspection Badging
- Added a dedicated CLAP (CLever Audio Plug-in) modern open architecture specification callout card within the Audio Format & Host Compatibility Matrix in `client/src/routes/+page.svelte`.
- Detailed key architectural advantages of CLAP: sample-accurate non-destructive parameter modulation, CPU thread-pool delegation, multiple voice handling, and polyphonic MIDI/MPE expression.
- Added visual `CLAP Standard` badge and amber/gold card styling (`.clap-spec-banner`, `.clap-icon`, `.clap-content`) highlighting host DAW compatibility (Bitwig Studio, REAPER, Cubase, FL Studio).
- Verified TypeScript/Svelte diagnostics (`npm run check`), static client build (`npm run build`), and automated Playwright visual testing.
- Bumped ecosystem version to `v0.5.2`.

## [0.5.1] - Deep Audio Bundle Structural Verification & Audit Diagnostics
- Upgraded `verify_installed_plugin` in `client/src-tauri/src/installer.rs` to execute multi-format bundle structural inspection across LV2 (`manifest.ttl`), VST3 (`Contents/` architecture trees), and Apple AU (`Info.plist` / `Contents/MacOS/`).
- Added automated binary permission verification on Unix (`0o755`) for standalone dynamic libraries.
- Enhanced the "Filesystem & DAW" inspector drawer tab in `client/src/routes/+page.svelte` with descriptive tooltips detailing structural integrity status badges.
- Verified TypeScript/Svelte diagnostics (`npm run check`), static client build (`npm run build`), Rust test compilation (`cargo test --no-run`), and Playwright visual testing.
- Bumped ecosystem version to `v0.5.1`.

## [0.5.0] - Dynamic DAW Theme Switching & UI Color Profiles
- Implemented responsive interface color profile switching (`activeTheme` state: `dark` vs `contrast`) in `client/src/routes/+page.svelte`.
- Added persistent `localStorage` synchronization (`vst_monster_theme`) restoring custom user appearance preferences across app restarts.
- Added top navigation theme toggle button (`btn-theme-toggle`) with real-time profile tooltips and distinctive icons (`🎨 Studio` vs `⚡ Contrast`).
- Added High-Contrast DAW / Cyber Slate CSS styling overrides on `.dashboard-root.theme-contrast` enhancing visibility for dimly lit studio DAW sessions.
- Verified TypeScript/Svelte diagnostics (`npm run check`), static client build (`npm run build`), and automated Playwright visual testing.
- Bumped ecosystem version to `v0.5.0`.

## [0.4.9] - Pre-Flight Antivirus Sandbox Scanner & Cryptographic Verification UI
- Implemented `scan_binary_security(&file_path)` pre-flight security scanner hook in `client/src-tauri/src/installer.rs`. When ClamAV (`clamscan`) is present on the system, the installer executes a silent scan prior to archive extraction or binary execution, immediately deleting infected payloads and aborting installation if malicious signatures are detected.
- Added structured "Cryptographic Verification & Antivirus Scan" panel in the "Filesystem & DAW" inspector drawer tab (`client/src/routes/+page.svelte`) with a `🛡️ Verified Clean (ClamAV / SHA-256)` badge, real-time SHA-256 cryptographic digest copy box, and direct distribution endpoint verification.
- Bumped ecosystem version to `v0.4.9`.

## [0.4.8] - Audio Unit (AU) Discoverability Chip, Count Badges & CoreAudio Compatibility
- Added reactive `countAU` derived state calculating plugins available or installed in Apple Audio Unit (`.component`) format.
- Added interactive `🍎 AU` format filter chip with active state and count badge alongside VST3, CLAP, and LV2 chips.
- Added Apple Audio Unit (AUv2 / CoreAudio) compatibility card in the "Audio Format & Host Compatibility Matrix" inside the plugin inspector drawer with Logic Pro, GarageBand, Ableton Live, Studio One, and Reaper compatibility callouts.
- Verified TypeScript/Svelte diagnostics (`npm run check`), production static build (`npm run build`), and Playwright screenshot verification.
- Bumped ecosystem version to `v0.4.8`.

## [0.4.7] - Local Metadata Scan Caching, Rescan Acceleration & Cache Management
- Implemented persistent local scan caching (`.vst_monster_cache.json`) in `client/src-tauri/src/installer.rs` keyed by filesystem path and file modification timestamps (`u64`), eliminating redundant disk I/O and version extraction during rescans.
- Added `clear_scan_cache` command in `installer.rs` and registered handler in `client/src-tauri/src/lib.rs`.
- Added a "Local Scan Cache" management panel in the Paths modal (`client/src/routes/+page.svelte`) with a "🧹 Flush Cache & Deep Rescan" button and status feedback.
- Verified TypeScript/Svelte diagnostics (`npm run check`), production build (`npm run build`), and Rust workspace compilation (`cargo check`).
- Bumped ecosystem version to `v0.4.7`.

## [0.4.6] - Tabbed Inspector Views (Overview & Specs, Filesystem & DAW, Raw Manifest JSON)
- Restructured the plugin inspector drawer (`client/src/routes/+page.svelte`) with a 3-way tabbed navigation bar (`activeInspectorTab` state):
  - **🎛️ Overview & Specs**: Overview status banner, legacy compatibility warning, specs grid, synthesis description, and Audio Format & Host Compatibility Matrix.
  - **📁 Filesystem & DAW**: Local installation format, absolute filesystem path, detected version, bundle integrity checks with inline verification triggers, cryptographic SHA-256 digest copy box, and direct distribution endpoint URL.
  - **📋 Raw Manifest JSON**: Formatted syntax-highlighted raw JSON viewer rendering `JSON.stringify(previewPlugin, null, 2)` with inline "Copy JSON" and "Save .json" action triggers.
- Added corresponding modern dark-mode CSS styling for `.inspector-tabs-bar`, `.inspector-tab-btn`, `.filesystem-details-card`, `.fs-row`, `.raw-manifest-container`, and `.raw-manifest-pre`.
- Verified clean TypeScript/Svelte compilation (`npm run check`) and zero warnings/errors on Vite static build (`npm run build`).
- Bumped ecosystem version to `v0.4.6`.

## [0.4.5] - Format-Aware Zip Extraction with LV2/CLAP Auto-Targeting & Inspector Compatibility Matrix
- Enhanced `extract_zip_archive` in `client/src-tauri/src/installer.rs` to inspect archive entries and dynamically route extraction to format-specific OS audio directories (`~/.lv2` for LV2/TTL, `~/.clap` for CLAP, `~/.vst3` for VST3).
- Added an "Audio Format & Host Compatibility Matrix" section in the plugin inspection preview drawer (`client/src/routes/+page.svelte`) detailing distribution architecture, target audio folders, and verified DAW hosts.
- Embedded an LV2 Open Audio Standard banner highlighting Turtle (`.ttl`) RDF manifest graph features and Linux audio workstation compatibility.
- Bumped ecosystem version to `v0.4.5`.

## [0.4.4] - DAW Presets Quick Launcher & Legacy VST2/32-bit Compatibility Warnings
- Added `open_presets_directory` native command in `client/src-tauri/src/installer.rs` targeting OS standard DAW presets folders (`~/.vst3/presets`, `~/Library/Audio/Presets`, `Documents/VST3 Presets`) with automatic folder creation.
- Added `📁 Presets` top-navigation button in dashboard UI with explanatory tooltip guidance for preset installation and backups.
- Added legacy VST2 / 32-bit architecture compatibility detection (`isLegacyPlugin`), badging (`⚠️ Legacy`), and detailed bridging warning notice in preview modal for modern 64-bit DAWs.
- Bumped ecosystem version to `v0.4.4`.

## [0.4.3] - Comprehensive macOS DMG Extraction & Automated Plugin Installation
- Enhanced `hdiutil attach` with `-nobrowse`, `-readonly`, `-noverify`, and `-noautoopen` to prevent volume prompts or hanging on validation.
- Added recursive `.pkg` / `.mpkg` discovery inside mounted DMG volumes with automated headless execution via `installer -pkg <path> -target /`.
- Expanded audio bundle extraction to support `.vst` (legacy VST2) alongside `.vst3`, `.component` (AU), and `.clap` directly into `~/Library/Audio/Plug-Ins/`.
- Added resilient volume detachment with automated retry logic to handle busy filesystem handles.
- Bumped ecosystem version to `v0.4.3`.

## [0.4.2] - Active Installation Path Spec Indicator & Interactive Hash Copy Action
- Added active installation path display in the catalog grid specs for locally installed plugins with quick inspection tooltips.
- Implemented reactive one-click copy buttons (`copyToClipboard`) with visual confirmation state transitions (`📋` -> `✓` and `📋 Copy` -> `✓ Copied`) across both catalog cards and the plugin inspection preview drawer.
- Implemented multi-tier clipboard support utilizing modern `navigator.clipboard.writeText` with legacy `textarea` fallback for cross-platform Linux webkit / Tauri webview contexts.
- Synchronized global ecosystem version to `v0.4.2`.

## [0.4.1] - LV2 Format Discoverability Chip & Windows VST2 Path Scanning
- Added interactive LV2 format discoverability filter chip (`🎛️ LV2`) with reactive catalog count badge (`countLV2`) in the primary dashboard filter row.
- Added Windows VST2 standard search locations (`%ProgramFiles%/VSTPlugins`, `%ProgramFiles%/Steinberg/VSTPlugins`, `%CommonProgramFiles%/VST2`, and `Documents/VST`) in `get_default_scan_paths` and `scan_installed_plugins`.
- Synchronized global ecosystem version to `v0.4.1`.

## [0.4.0] - Comprehensive Plugin Inspector Drawer, Badging & Verification
- Enhanced the plugin detail preview drawer header with inferred format badging (`inferCatalogFormat`) and installed audio format tags (`✓ VST3`, `✓ CLAP`, etc.).
- Upgraded the preview drawer installed status banner with live verification status badges (`✓ Verified Bundle`, `⏳ Verifying...`, `⚠ Missing Binary`).
- Added inline quick actions in the preview drawer to trigger local bundle verification (`verify_installed_plugin`) and open enclosing folders (`open_plugin_directory`).
- Synchronized global ecosystem version to `v0.4.0`.

## [0.3.9] - Standalone Binary Installation & InnoSetup Silent Fallbacks
- Added standalone audio binary installation engine (`.so`, `.dll`, `.clap`) in `client/src-tauri/src/installer.rs` with automated audio destination directory generation and `0o755` executable permissions.
- Implemented InnoSetup and NSIS secondary silent command fallback flags (`/VERYSILENT /SUPPRESSMSGBOXES /NORESTART /SP-`) for Windows `.exe` installers.
- Enriched plugin extraction strategy spec tile and direct download source details inside dashboard inspection preview drawer (`client/src/routes/+page.svelte`) with descriptive tooltips.
- Bumped global ecosystem version to `v0.3.9`.

## [0.3.8] - Linux Archive Extraction (.deb, .tar.gz) & Format Pill Counts
- Implemented asynchronous `.tar.gz` and `.tgz` decompression engine in `client/src-tauri/src/installer.rs` unpacking directly into audio directory trees.
- Implemented Linux `.deb` package extraction using `dpkg-deb -x` scanning for audio bundles across standard Linux system plugin paths (`usr/lib/vst3`, `usr/lib/clap`, `usr/lib/lv2`).
- Added dynamic reactive count badges (`countVST3`, `countCLAP`) on dashboard format discoverability filter chips.
- Added informative tooltips indicating available plugin numbers per audio plugin format.
- Bumped ecosystem version to `v0.3.8`.

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
