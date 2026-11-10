# MEMORY: VST Monster Architectural Observations

## Codebase Traits
- **Crawler**: Built in Go for high concurrency and performance. Scrapers: KVR Audio, GitHub topics, Bedroom Producers Blog. Employs modern browser user agents and fallback seed mechanisms for resilience.
- **Registry**: Express Node.js/TypeScript backend with PostgreSQL (pg) and Redis. Exposes plugin search, health checks, real-time stats, crawler triggering, log streaming, and manual plugin entry.
- **Client**: Tauri v2 / Rust for native OS integration and minimal footprint, paired with a SvelteKit UI in Single-Page Application (SPA) mode using Svelte 5 runes.

## Design Preferences
- Single-page consolidated dashboard layout without subpage fragmentation.
- High-value features displayed first (Quick Search, Crawler Controls, Live Metrics).
- Strategy-based installation engine handling MSI, EXE, PKG, and direct ZIP archive decompression to audio directories (`~/.vst3`).
- Contextual tooltips providing detailed explanation across all interactive controls.
- Interactive audio engineering diagnostic matrix cards in the "Filesystem & DAW" inspector tab for PDC latency, sidechain/surround bus routing, CPU/SIMD binary architecture, headless DSP cycle benchmarks, plugin parameter enumeration with MIDI automation mapping, real-time audio thread stress testing with crash recovery profiling, headless audio loopback rendering with dynamic frequency response spectrum analysis, automated preset state diffing & snapshot management, real-time audio loopback metering with True Peak (dBTP) LUFS EBU R128 analysis, multi-channel phase correlation with Lissajous goniometer vectorscope projection, deep LV2 RDF manifest graph inspection with port topology matrices, CLAP factory descriptors, Apple Audio Unit (AUv2/AUv3) CoreAudio introspection, Linux Audio PipeWire/JACK graph routing diagnostics, Avid AAX Native / HDX DSP inspection, and Windows MSI unattended silent installation engine parameters and exit code remediation.
- Go crawler normalizer automatically inspects release archives for `.lv2` bundles and tags `has_lv2_manifest: true`, surfaced visually in the UI through Turtle validation status chips.
- Go crawler monitors 14 dedicated audio engineering and DSP topics on GitHub (including `audio-unit`, `juce`, `vst-instruments`, and `faust`), extracting Apple Audio Unit (`AU`) formats from repository names, descriptions, or `.component` assets and tagging Faust DSP language metadata.
