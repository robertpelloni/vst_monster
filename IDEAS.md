# IDEAS: VST Monster Future & Pivots

- **Audio Fingerprinting & Preset Scanning**: Inspect extracted plugin directories for presets (`.vstpreset`, `.fxp`), extracting metadata for searching by synth sound type (pads, leads, 808s).
- **Yabridge Linux Automation**: Auto-configure Windows `.dll` and `.vst3` plugins on Linux systems using Wine and `yabridge-host` / `yabridgectl sync`.
- **Decentralized Manifest Federation**: Allow developers to host `vst-monster.json` manifests in their own GitHub repositories, indexed automatically via decentralized pull requests.
- **VirusTotal Sandboxing**: Query hash digests against malware databases before triggering silent installation binaries.
- **Client Caching Layer**: Add a local state cache (via Redis or internal Svelte stores) to prevent over-fetching the /plugins endpoint on each UI refresh.
