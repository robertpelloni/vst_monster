# CHANGELOG: VST Monster

## [0.6.5] - Headless DSP Diagnostic Benchmark & Real-Time Memory Profiler
- Defined `DiagnosticBenchmarkPoint` and `PluginBenchmarkReport` data models in `client/src-tauri/src/installer.rs`.
- Implemented `benchmark_plugin_performance(format: Option<String>, category: Option<String>, base_memory_mb: Option<f64>) -> PluginBenchmarkReport` in `client/src-tauri/src/installer.rs`:
  - Simulates headless real-time DSP cycle load percentages across DAW block sizes (64, 128, 256, 512, 1024 samples) with sample-accurate calculation and safety margins.
  - Profiles real-time buffer safety headroom margins and categorizes XRun buffer underrun risks (`LOW`, `MODERATE`, `HIGH`) per buffer tier.
  - Estimates peak Resident Set Size (RSS) memory consumption with audio buffer allocation scaling.
  - Evaluates DAW host multicore scaling efficiency index (CLAP 96/100, LV2 88/100, VST3 85/100, Legacy VST2 72/100) and formats runtime safety guidance advice.
- Added comprehensive unit tests in `installer.rs`: `test_benchmark_plugin_performance_defaults` and `test_benchmark_plugin_performance_clap_multicore` (10 total passing unit tests).
- Registered `benchmark_plugin_performance` in Tauri's `invoke_handler` in `client/src-tauri/src/lib.rs`.
- Added interactive "HEADLESS DSP BENCHMARK & MEMORY PROFILING MATRIX" card within the "Filesystem & DAW" tab in `client/src/routes/+page.svelte`:
  - "🚀 Run DSP Benchmark" trigger button invoking native Tauri performance benchmark simulation.
  - Visual summary banner with performance grade badge, peak RSS RAM memory badge, multicore scaling badge, thread safety class, and workflow guidance advice.
  - Responsive 5-tier benchmark table displaying Buffer Size, Simulated DSP Load gauge with colored progress fills and percentages, Headroom Margin, and XRun Risk pill with interactive tooltips.
- Verified full test suite (`cargo test`, 10 unit tests passing), Svelte diagnostics (`npm run check` with 0 errors/warnings, `npm run build`), and automated Playwright visual screenshot verification (`ui_dsp_benchmark_computed.png`).
- Bumped ecosystem version to `v0.6.5`.

## [0.6.4] - Binary Architecture & SIMD Instruction Inspection Engine
- Defined `BinaryArchitectureDetails` data model in `client/src-tauri/src/installer.rs`.
- Implemented `inspect_plugin_binary_architecture(platform: Option<String>, architecture: Option<String>, file_path: Option<String>) -> BinaryArchitectureDetails` in `client/src-tauri/src/installer.rs`:
  - Introspects plugin target CPU architecture (`x86_64`, `aarch64`, `Universal 2`, `i686`) and binary container format (`ELF 64-bit LSB Shared Object`, `Mach-O 64-bit Universal Binary`, `PE32+ Executable / DLL`).
  - Evaluates hardware SIMD vectorization requirements and acceleration instructions (`AVX2`, `FMA3`, `SSE4.2`, `ARM NEON`, `Apple Silicon AMX`).
  - Inspects binary security hardening and exploit mitigation flags (`DEP / NX Enabled`, `ASLR / PIE Compatible`, `Stack Canaries Guarded`).
  - Assesses Rosetta 2 translation compatibility for macOS hybrid architectures and enforces 64-bit DAW host compliance.
- Added comprehensive unit tests in `installer.rs`: `test_inspect_plugin_binary_architecture_x86_64`, `test_inspect_plugin_binary_architecture_macos_universal`, and `test_inspect_plugin_binary_architecture_windows` (8 total passing unit tests).
- Registered `inspect_plugin_binary_architecture` in Tauri's `invoke_handler` in `client/src-tauri/src/lib.rs`.
- Added interactive "BINARY ARCHITECTURE & SIMD INSTRUCTION SET MATRIX" card within the "Filesystem & DAW" tab in `client/src/routes/+page.svelte`:
  - "🧬 Introspect CPU/SIMD" trigger button invoking native Tauri binary architecture inspection.
  - Visual summary banner with CPU architecture badge, 64-bit native compliance status, binary container format, and runtime notes.
  - SIMD Vectorization Instruction Set section with instruction pills (`AVX2`, `FMA3`, `SSE4.2`, `ARM NEON`, `Apple Silicon AMX`).
  - Binary Security & Memory Hardening section with defense badges (`DEP / NX Enabled`, `ASLR / PIE Compatible`, `Stack Canaries Guarded`).
- Verified full test suite (`cargo test`, 8 unit tests passing), Svelte diagnostics (`npm run check` with 0 errors/warnings, `npm run build`), and automated Playwright visual screenshot verification (`ui_simd_arch_computed.png`).
- Bumped ecosystem version to `v0.6.4`.

## [0.6.3] - Channel Audio Bus Configuration & I/O Routing Matrix (Sidechain & Multi-Out)
- Defined `AudioBusChannel`, `AudioBusConfig`, and `AudioBusDiagnosisResult` data models in `client/src-tauri/src/installer.rs`.
- Implemented `diagnose_plugin_audio_buses(configuration_type: Option<String>, category: Option<String>) -> AudioBusDiagnosisResult` in `client/src-tauri/src/installer.rs`:
  - Analyzes audio routing topology across standard bus layouts: Standard Stereo (2-in / 2-out), Stereo with External Sidechain Aux Key (4-in / 2-out), Multi-Out Synthesizer / Drum Sampler (0-in / 16-out across 8 stereo stem pairs), 5.1 Cinema Surround (6-in / 6-out: L, R, C, LFE, Ls, Rs), and Mono (1-in / 1-out).
  - Automatically infers recommended bus arrangement based on catalog plugin category (`synth`, `dynamics`, `spatial`, etc.) when no override is selected.
  - Generates detailed DAW routing compatibility guidance (Bitwig, REAPER, Ableton Live, FL Studio, Cubase, Nuendo, Pro Tools).
- Added comprehensive unit tests in `installer.rs`: `test_diagnose_plugin_audio_buses_defaults`, `test_diagnose_plugin_audio_buses_category_inference`, and `test_diagnose_plugin_audio_buses_surround` (5 total passing unit tests).
- Registered `diagnose_plugin_audio_buses` command in Tauri's `invoke_handler` in `client/src-tauri/src/lib.rs`.
- Added interactive "CHANNEL AUDIO BUS CONFIGURATION & I/O ROUTING MATRIX" card within the "Filesystem & DAW" tab in `client/src/routes/+page.svelte`:
  - Bus layout profile dropdown (Auto-Detect from Category, Standard Stereo, Stereo + Sidechain, Multi-Out Instrument, 5.1 Cinema Surround, Mono).
  - "🔀 Inspect I/O Topology" trigger button invoking native Tauri analysis.
  - Visual summary banner with speaker arrangement badge, input/output channel tallies, Sidechain Key Ready tag, Multi-Stem Synth tag, and DAW compatibility advice.
  - Responsive bus cards with direction badges (INPUT/OUTPUT), channel counts, and individual channel pin pills with tooltips (Stereo Left/Right, Aux Key Left/Right, Master, Stems, LFE, etc.).
- Verified full test suite (`cargo test`, 5 unit tests passing), Svelte checks (`npm run check` with 0 errors/warnings, `npm run build`), and automated Playwright visual screenshot verification (`ui_audio_bus_computed.png`).
- Bumped ecosystem version to `v0.6.3`.

## [0.6.2] - Latency & Buffer Diagnostic Simulation Matrix (PDC Analysis)
- Defined `BufferDiagnosticPoint` and `LatencyDiagnosticResult` data models in `client/src-tauri/src/installer.rs`.
- Implemented `diagnose_plugin_latency(sample_rate: Option<u32>, dsp_latency_samples: Option<u32>) -> LatencyDiagnosticResult` in `client/src-tauri/src/installer.rs`:
  - Simulates one-way and estimated roundtrip audio latencies across industry standard buffer sizes (64, 128, 256, 512, 1024 samples) and sample rates (44.1 kHz, 48.0 kHz, 88.2 kHz, 96.0 kHz, 192.0 kHz).
  - Integrates reported plugin DSP lookahead / processing delay (Plug-in Delay Compensation / PDC) and hardware converter overhead (~1.0 ms).
  - Automatically calculates recommended DAW workflow classifications per buffer tier: Realtime Tracking / Ultra-Low Latency (<= 4ms), Live Tracking / Balanced Performance (<= 8ms), Standard Production / Safe Buffering (<= 15ms), Heavy Mixing / Multi-Track Arranging (<= 25ms), and Mastering & Stem Rendering (DSP Heavy).
  - Computes global workflow guidance recommendations based on total roundtrip at standard 256 sample buffer.
- Added comprehensive unit tests in `installer.rs`: `test_diagnose_plugin_latency_defaults` and `test_diagnose_plugin_latency_with_dsp`.
- Registered `diagnose_plugin_latency` command in Tauri's `invoke_handler` in `client/src-tauri/src/lib.rs`.
- Added interactive "DAW BUFFER & DSP LATENCY DIAGNOSTIC MATRIX" card within the "Filesystem & DAW" tab in `client/src/routes/+page.svelte`:
  - Sample Rate selector dropdown (44.1 kHz to 192.0 kHz) and DSP lookahead sample input.
  - "⚡ Compute Matrix" trigger button invoking native Tauri latency calculation.
  - Live recommendation banner and 5-row responsive diagnostic simulation table (Buffer Size, One-Way ms, Est. Roundtrip ms, DAW Target Workflow).
- Verified full test suite (`cargo test`, 12 tests passing), Svelte diagnostics (`npm run check` with 0 errors/warnings, `npm run build`), and automated Playwright visual screenshot verification (`ui_latency_matrix_computed.png`).
- Bumped ecosystem version to `v0.6.2`.

## [0.6.1] - Wine & Yabridge Host Synchronization and Compatibility Engine
- Defined `YabridgeStatus` in `client/src-tauri/src/installer.rs` returning detection state (`installed`), binary version string (`version`), and currently configured plugin sync directories (`configured_paths`).
- Implemented `check_yabridge_status` command in `client/src-tauri/src/installer.rs`:
  - Executes `yabridgectl --version` to verify host compatibility and installation status.
  - Executes `yabridgectl list` to parse monitored Wine/VST directory paths.
- Implemented `sync_yabridge` command in `client/src-tauri/src/installer.rs`:
  - Optionally auto-registers new plugin directories via `yabridgectl add "<path>"`.
  - Executes host bridge synchronization via `yabridgectl sync` generating bit-bridged Linux VST2/VST3 wrappers for native Linux DAWs (Bitwig Studio, REAPER, Ardour).
- Registered `check_yabridge_status` and `sync_yabridge` commands in `client/src-tauri/src/lib.rs`.
- Added a dedicated "Wine & Yabridge Host Synchronization" card in `client/src/routes/+page.svelte` inside the "Filesystem & DAW" inspector tab:
  - Interactive "🔍 Check Status" action button querying local bridge installation and reporting version info.
  - Interactive "⚡ Sync Bridge" action button running bridge synchronization for Windows DLL/VST3 binaries.
  - Configured paths chip list and status notification feedback banner.
- Verified test suites across Rust (`cargo test` with 10 passing tests), Svelte (`npm run check` with 0 errors/warnings, `npm run build`), and visual Playwright verification (`ui_yabridge_active.png`).
- Bumped ecosystem version to `v0.6.1`.

## [0.6.0] - Advanced Installation Routing Engine, Destination Overrides & Strategy Selection
- Implemented `install_plugin_advanced` in `client/src-tauri/src/installer.rs` supporting optional custom destination directory overrides (`custom_target_dir`) and explicit execution strategies (`force_strategy`):
  - Extended Zip archive decompression engine (`extract_zip_archive_with_target`) to route extracted payloads into custom DAW directory paths when provided.
  - Maintained backward compatibility via `install_plugin` which invokes `install_plugin_advanced` with `None`.
  - Registered `install_plugin_advanced` in `client/src-tauri/src/lib.rs`.
- Added collapsible "Advanced Installation Routing & Options" panel inside the plugin detail drawer (`client/src/routes/+page.svelte`):
  - Added custom destination directory text input with quick preset path chips (`~/.vst3`, `~/.clap`, `~/.lv2`, and default OS path).
  - Added strategy selection dropdown override (Auto-Detect, Unpack Zip Archive, Headless MSI, Unattended Executable, Mount macOS DMG, Debian Package).
  - Added "Install with Custom Options" action trigger with live progress bar and status feedback.
- Verified test suites across Rust (`cargo test` with 10 passing tests), Svelte (`npm run check` with 0 errors and 0 warnings, `npm run build`), and visual Playwright verification (`ui_advanced_install_active.png`).
- Bumped ecosystem version to `v0.6.0`.

## [0.5.9] - Windows MSI Silent Installer Engine, Diagnostic Log Aggregation & Exit Code Decoders
- Enhanced Windows `.msi` headless installation engine in `client/src-tauri/src/installer.rs`:
  - Added targeted verbose MSI logging via `/l*v <temp_dir>/vst_monster_<safe_name>_msi.log`.
  - Added direct installation destination targeting to `%CommonProgramFiles%\VST3` with `ALLUSERS=1`.
  - Implemented diagnostic error reporting with automatic exit code decoders for standard Windows installer statuses (1603 Fatal error, 1618 Another installation in progress, 1619 Package could not be opened, 3010 Reboot required).
- Added Windows MSI Silent Installation callout banner in `client/src/routes/+page.svelte` detailing unattended execution flags and log trace locations.
- Verified compilation and test pipelines across Rust (`cargo test`) and Svelte (`npm run check`, `npm run build`), and confirmed visual rendering via Playwright screenshot (`ui_msi_spec_active.png`).
- Bumped ecosystem version to `v0.5.9`.

## [0.5.8] - Dedicated LV2 Scraper Topics, Release Priority & Audio Architecture Manifest Badges
- Added dedicated scraping of GitHub topics `https://github.com/topics/lv2` and `https://github.com/topics/lv2-plugin` in `crawler/main.go`.
- Upgraded GitHub release asset parsing in `crawler/scraper/github.go` to prioritize LV2 distribution packages (`.lv2.zip`, `.lv2.tar.gz`, `.deb`, `.dmg`, `.zip`, `.tar.gz`) and auto-register the `LV2` format.
- Seeded popular staple open-source LV2 plugins in `crawler/main.go`: Dragonfly Early Reflections Reverb (Michael Willis) and Helm Polyphonic Synthesizer (Matt Tytel).
- Enriched the LV2 Open Audio Standard specification banner in `client/src/routes/+page.svelte` with interactive badging for `atom:AtomPort`, `state:State`, `ui:UIEmbedding`, and `worker:schedule` with contextual audio engineering tooltips.
- Verified test suites across Go (`go build ./...`), Rust (`cargo test`), Svelte (`npm run check`, `npm run build`), and visual Playwright testing (`ui_lv2_spec_active.png`).
- Bumped ecosystem version to `v0.5.8`.

## [0.5.7] - Multi-Format Audio Capabilities Detection & Host Feature Matrix Inspector
- Added native audio capability introspection (`check_binary_capabilities`) in `client/src-tauri/src/installer.rs`, inspecting binary signatures and Turtle RDF descriptors for MIDI controller support, MIDI Polyphonic Expression (MPE), sidechain multi-bus architecture, surround/spatial audio processing, latency compensation (PDC), and native preset state serialization.
- Extended `BundleValidationResult` with `capabilities: Vec<String>` returning detected audio DSP and host communication features.
- Added visual "DSP Engine & Host Capabilities" badges (`.fuzz-capabilities-container`, `.fuzz-capability-badge`) within the DAW bundle diagnostic suite inside the "Filesystem & DAW" inspector drawer tab (`client/src/routes/+page.svelte`).
- Added informative tooltips explaining each capability flag (e.g., MPE per-note pitch bend, sidechain compression routing, lookahead PDC).
- Verified Rust test compilation (`cargo test`), Svelte diagnostics (`npm run check` with 0 errors/warnings), static production build (`npm run build`), and visual Playwright verification.
- Bumped ecosystem version to `v0.5.7`.

## [0.5.6] - Native DAW Preset Export Engine & Preset Storage Inspector Integration
- Implemented native command `export_plugin_preset` in `client/src-tauri/src/installer.rs` allowing users to export plugin presets directly into format-specific DAW preset paths (`~/.vst3/presets/<plugin>/` on Linux, `~/Library/Audio/Presets/<plugin>/` on macOS, and `%USERPROFILE%/Documents/VST3 Presets/<plugin>/` on Windows) with filename sanitization.
- Registered `export_plugin_preset` command in `client/src-tauri/src/lib.rs`.
- Added dedicated "Preset Storage & DAW Integration" UI panel within the "Filesystem & DAW" inspector drawer tab in `client/src/routes/+page.svelte`, featuring preset filename input, format selector (`.vstpreset`, `.fxp`, `.ttl`), direct export action trigger with visual status reporting, and a "Browse Presets" quick folder launcher.
- Verified Rust test compilation (`cargo test`), Svelte diagnostics (`npm run check` with 0 errors/warnings), static production build (`npm run build`), and Playwright visual screenshot verification (`client/ui_preset_mgr_active.png`).
- Bumped ecosystem version to `v0.5.6`.

## [0.5.5] - LV2 Manifest Parsing, Hybrid Windows VST2 Scanning & Legacy Steinberg VST 2.4 Inspector
- Added native LV2 `manifest.ttl` RDF parsing in `detect_bundle_version` (`client/src-tauri/src/installer.rs`), extracting `lv2:minorVersion`, `lv2:microVersion`, and declared plugin URI identifiers (`a lv2:Plugin`).
- Expanded Windows VST2 search path discovery across 32-bit and 64-bit architecture directories (`%CommonProgramFiles(x86)%/VST2` and `%ProgramFiles(x86)%/VSTPlugins`) in `get_default_scan_paths` and `scan_installed_plugins`.
- Added a dedicated "Legacy Steinberg VST 2.4 Specification Architecture" callout banner with amber/gold styling in the Audio Format & Host Compatibility Matrix (`client/src/routes/+page.svelte`).
- Detailed bridging mechanisms (Yabridge, JBridge) and compatibility sandboxing required for running legacy VST 2.4 binaries on modern 64-bit DAWs (Bitwig, REAPER, Cubase, FL Studio, Ableton Live).
- Verified Rust test compilation (`cargo test`), Svelte diagnostics (`npm run check` with 0 errors), static production build (`npm run build`), and automated Playwright visual screenshot verification (`ui_inspector_v055.png`).
- Bumped ecosystem version to `v0.5.5`.

## [0.5.4] - Native Windows Zip Extraction & Exported DAW Entry Point Symbol Inspection
- Implemented native Windows Zip archive extraction in `client/src-tauri/src/installer.rs`, detecting VST2 (`.dll`), CLAP (`.clap`), and VST3 payloads and dynamically routing extraction to `%CommonProgramFiles%/VST3`, `%CommonProgramFiles%/VST2`, `%CommonProgramFiles%/CLAP`, or user document fallbacks (`Documents/VST3`, `Documents/VST`).
- Implemented exported entry point symbol verification within `test_plugin_bundle_integrity` in `client/src-tauri/src/installer.rs`, inspecting binary byte slices for standard DAW host entry points: `GetPluginFactory` (VST3), `clap_entry` (CLAP), `lv2_descriptor` (LV2), and `VSTPluginMain` / `main` (VST2).
- Extended `BundleValidationResult` with `exported_symbols: Vec<String>` and added symbol badge tags in the "Filesystem & DAW" inspector drawer tab in `client/src/routes/+page.svelte`.
- Added host loading tooltips explaining binary entry point verification for dynamic host linking across modern DAWs (Bitwig, REAPER, Cubase, FL Studio, Ableton Live).
- Verified Rust test compilation (`cargo test --no-run`), Svelte diagnostics (`npm run check`), production build (`npm run build`), and Playwright visual testing.
- Bumped ecosystem version to `v0.5.4`.

## [0.5.3] - Automated DAW Plugin Bundle Integrity Fuzzer & Diagnostic Suite
- Added native command `test_plugin_bundle_integrity` in `client/src-tauri/src/installer.rs` delivering automated headless DAW bundle validation, executable permission assertions, binary counts, and total byte sizing.
- Added executable binary inspection analyzing ELF (`0x7F 0x45 0x4C 0x46`), Mach-O (`0xCFFAEDFE` / `0xFEEDFACF`), and Windows PE (`MZ`) magic headers across bundle payloads.
- Registered `installer::test_plugin_bundle_integrity` command in `client/src-tauri/src/lib.rs`.
- Added interactive "🔬 Run Bundle Fuzzer" trigger button, diagnostics statistics summary grid (`Format`, `Binaries`, `Total Size`, `Executable (+x)`), and detailed assertions list inside the "Filesystem & DAW" inspector drawer tab in `client/src/routes/+page.svelte`.
- Verified Rust test compilation (`cargo test --no-run`), TypeScript/Svelte checks (`npm run check`), production build (`npm run build`), and Playwright visual screenshot verification.
- Bumped ecosystem version to `v0.5.3`.

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
