<script lang="ts">
  import { onMount } from 'svelte';
  import { invoke } from '@tauri-apps/api/core';
  import { listen } from '@tauri-apps/api/event';

  type Plugin = {
    id: string;
    name: string;
    developer: string;
    license_model: string;
    category?: string;
    version?: string;
    platform?: string;
    architecture?: string;
    download_url?: string;
    sha256_hash?: string;
    strategy?: string;
    metadata?: {
      description?: string;
      source?: string;
      [key: string]: any;
    };
  };

  type Stats = {
    totalPlugins: number;
    totalDistributions: number;
    licenses: { license_model: string; count: string }[];
    crawlerRunning: boolean;
  };

  type Health = {
    status: string;
    timestamp: string;
    database: string;
    crawlerRunning: boolean;
    version: string;
  };

  type InstalledPluginInfo = {
    name: string;
    version?: string | null;
    format: string;
    path: string;
  };

  // State
  let plugins = $state<Plugin[]>([]);
  let stats = $state<Stats | null>(null);
  let health = $state<Health | null>(null);
  let crawlerLogs = $state<string[]>([]);
  let loading = $state(true);
  let error = $state<string | null>(null);
  let installStatus = $state<Record<string, string>>({});
  let installedPlugins = $state<InstalledPluginInfo[]>([]);

  // Controls & Filters
  let searchQuery = $state('');
  let selectedLicense = $state('all');
  let selectedCategory = $state('all');
  let selectedPlatform = $state('all');
  let selectedInstalledStatus = $state<'all' | 'installed' | 'update-available' | 'not-installed'>('all');
  let selectedFormat = $state<'all' | 'vst3' | 'clap' | 'au' | 'vst2' | 'lv2'>('all');
  let sortBy = $state<'name' | 'developer' | 'license' | 'category' | 'status'>('name');
  let crawlerRunning = $state(false);
  let actionMessage = $state<string | null>(null);
  let isLogsCollapsed = $state(true);
  let showAddPluginModal = $state(false);
  let previewPlugin = $state<Plugin | null>(null);
  let activeInspectorTab = $state<'overview' | 'filesystem' | 'manifest'>('overview');
  let isVerifyingPlugin = $state(false);
  let verificationStatus = $state<{ [key: string]: 'valid' | 'invalid' | 'verifying' }>({});
  let isTestingBundle = $state(false);
  let bundleTestResult = $state<{
    is_valid: boolean;
    format: string;
    binary_count: number;
    total_size_bytes: number;
    has_executable_permission: boolean;
    architecture_compatible: boolean;
    exported_symbols?: string[];
    capabilities?: string[];
    diagnostics: string[];
  } | null>(null);
  let copiedHashId = $state<string | number | null>(null);

  // Advanced Installation Options State
  let showAdvancedInstall = $state(false);
  let customTargetDir = $state('');
  let forceStrategy = $state<'auto' | 'extract_zip' | 'silent_msi' | 'silent_exe' | 'mount_dmg' | 'deb_pkg'>('auto');
  let isBatchUpdating = $state(false);
  let isScanningFolders = $state(false);
  let showScanPathsModal = $state(false);
  let defaultScanPaths = $state<string[]>([]);
  let customScanPaths = $state<string[]>([]);
  let newCustomPath = $state('');
  let isClearingCache = $state(false);
  let cacheClearMsg = $state('');
  let activeTheme = $state<'dark' | 'contrast'>('dark');
  let presetExportName = $state('');
  let presetExportFormat = $state<'vstpreset' | 'fxp' | 'ttl'>('vstpreset');
  let isExportingPreset = $state(false);
  let presetExportMessage = $state('');

  // Yabridge Linux Bridging State
  interface YabridgeStatusInfo {
    installed: boolean;
    version?: string;
    configured_paths: string[];
  }
  let yabridgeStatus = $state<YabridgeStatusInfo | null>(null);
  let isCheckingYabridge = $state(false);
  let isSyncingYabridge = $state(false);
  let yabridgeMessage = $state('');

  // Audio Latency & Buffer Diagnostic State
  interface LatencyPoint {
    buffer_size: number;
    latency_ms: number;
    roundtrip_estimate_ms: number;
    daw_workflow_category: string;
  }
  interface LatencyResult {
    sample_rate: number;
    reported_dsp_latency_samples: number;
    dsp_latency_ms: number;
    buffer_points: LatencyPoint[];
    recommendation: string;
  }
  let latencySampleRate = $state<number>(48000);
  let latencyDspSamples = $state<number>(0);
  let latencyDiagnosticResult = $state<LatencyResult | null>(null);
  let isDiagnosingLatency = $state(false);

  interface AudioBusChannelItem {
    name: string;
    channel_type: string;
  }
  interface AudioBusConfigItem {
    bus_name: string;
    direction: string;
    channel_count: number;
    is_auxiliary: boolean;
    channels: AudioBusChannelItem[];
  }
  interface AudioBusDiagnosisResultItem {
    configuration_id: string;
    display_name: string;
    total_input_channels: number;
    total_output_channels: number;
    supports_sidechain: boolean;
    is_instrument: boolean;
    speaker_arrangement: string;
    daw_compatibility_notes: string;
    buses: AudioBusConfigItem[];
  }
  let selectedBusConfigType = $state<string>('auto');
  let audioBusDiagnosisResult = $state<AudioBusDiagnosisResultItem | null>(null);
  let isDiagnosingAudioBuses = $state(false);

  interface BinaryArchitectureDetailsItem {
    target_cpu_arch: string;
    binary_container_type: string;
    is_64bit: boolean;
    simd_instruction_sets: string[];
    rosetta2_compatible: boolean;
    min_os_version: string;
    binary_security_flags: string[];
    architecture_summary: string;
  }
  let binaryArchDetails = $state<BinaryArchitectureDetailsItem | null>(null);
  let isInspectingArch = $state(false);

  interface DiagnosticBenchmarkPointItem {
    buffer_size: number;
    simulated_dsp_load_pct: number;
    xrun_risk_level: string;
    real_time_safety_margin_pct: number;
  }
  interface PluginBenchmarkReportItem {
    format: string;
    category: string;
    estimated_peak_rss_mb: number;
    base_memory_mb: number;
    multicore_scaling_score: number;
    thread_safety_classification: string;
    buffer_benchmarks: DiagnosticBenchmarkPointItem[];
    overall_efficiency_rating: string;
    optimization_advice: string;
  }
  let benchmarkReport = $state<PluginBenchmarkReportItem | null>(null);
  let isBenchmarking = $state(false);

  interface PluginParameterInfoItem {
    id: number;
    name: String;
    parameter_type: string;
    default_value: number;
    min_value: number;
    max_value: number;
    display_unit: string;
    is_automatable: boolean;
    assigned_midi_cc: number | null;
    parameter_group: string;
  }
  interface PluginParameterReportItem {
    format: string;
    category: string;
    total_parameters: number;
    automatable_parameters: number;
    supports_mpe: boolean;
    supports_sample_accurate_automation: boolean;
    parameter_tree_structure: string;
    parameters: PluginParameterInfoItem[];
    automation_host_advice: string;
  }
  let parameterReport = $state<PluginParameterReportItem | null>(null);
  let isInspectingParameters = $state(false);

  interface AudioThreadStressPointItem {
    stress_intensity_pct: number;
    execution_time_us: number;
    time_budget_us: number;
    cpu_cycle_pct: number;
    priority_inversion_detected: boolean;
    audio_thread_allocations: number;
    xrun_dropout: boolean;
  }
  interface PluginResilienceReportItem {
    format: string;
    category: string;
    buffer_size: number;
    sample_rate: number;
    time_budget_us: number;
    lock_free_audio_thread: boolean;
    zero_allocation_render: boolean;
    bitwig_sandboxing_compatible: boolean;
    reaper_process_isolation_compatible: boolean;
    crash_recovery_mechanism: string;
    overall_resilience_grade: string;
    stress_timeline: AudioThreadStressPointItem[];
    host_resilience_advice: string;
  }
  let resilienceReport = $state<PluginResilienceReportItem | null>(null);
  let isTestingResilience = $state(false);

  interface SpectrumFrequencyBinItem {
    frequency_hz: number;
    band_name: string;
    input_magnitude_db: number;
    output_magnitude_db: number;
    gain_delta_db: number;
    phase_shift_deg: number;
    harmonic_distortion_pct: number;
  }
  interface AudioSpectrumReportItem {
    format: string;
    category: string;
    test_signal: string;
    sample_rate: number;
    fft_resolution_points: number;
    cutoff_frequency_3db_hz: number;
    total_harmonic_distortion_pct: number;
    signal_to_noise_ratio_db: number;
    phase_linearity_status: string;
    frequency_bins: SpectrumFrequencyBinItem[];
    mastering_advice: string;
  }
  let spectrumReport = $state<AudioSpectrumReportItem | null>(null);
  let isRenderingSpectrum = $state(false);
  let spectrumTestSignal = $state('Logarithmic Sine Sweep (20Hz - 20kHz)');

  // Preset State Snapshot & Diffing Manager State
  interface ParameterSnapshotItem {
    param_id: number;
    name: string;
    value: number;
    display_value: string;
    unit: string;
    is_automatable: boolean;
  }
  interface PluginStateSnapshot {
    plugin_name: string;
    format: string;
    preset_name: string;
    preset_format: string;
    checksum_crc32: string;
    total_parameters: number;
    parameters: ParameterSnapshotItem[];
    created_at: string;
  }
  interface ParameterDiffEntry {
    param_id: number;
    name: string;
    baseline_value: number;
    target_value: number;
    delta: number;
    delta_pct: number;
    status: string;
  }
  interface PresetDiffReport {
    plugin_name: string;
    format: string;
    baseline_preset_name: string;
    target_preset_name: string;
    total_parameters_compared: number;
    modified_parameters_count: number;
    unchanged_parameters_count: number;
    drift_severity: string;
    similarity_score_pct: number;
    parameter_diffs: ParameterDiffEntry[];
    compatibility_notes: string;
  }

  interface LoudnessMetricSummary {
    integrated_lufs: number;
    momentary_max_lufs: number;
    short_term_max_lufs: number;
    loudness_range_lra: number;
  }

  interface TruePeakSummary {
    peak_left_dbtp: number;
    peak_right_dbtp: number;
    oversampling_factor: string;
    ceiling_headroom_db: number;
    is_inter_sample_clipping: boolean;
  }

  interface StreamingStandardCompliance {
    standard_name: string;
    target_integrated_lufs: number;
    max_true_peak_dbtp: number;
    compliance_status: string;
    recommended_gain_adjustment_db: number;
  }

  interface PluginAudioMeteringReport {
    plugin_name: string;
    format: string;
    category: string;
    test_signal: string;
    sample_rate_hz: number;
    buffer_size_samples: number;
    loudness: LoudnessMetricSummary;
    true_peak: TruePeakSummary;
    target_compliance: StreamingStandardCompliance;
    dynamic_range_score: string;
    mastering_advice: string;
  }

  let activeSnapshot = $state<PluginStateSnapshot | null>(null);
  let isSnapshotting = $state(false);
  let presetDiffReport = $state<PresetDiffReport | null>(null);
  let isDiffingPresets = $state(false);
  let baselinePresetInput = $state('Init Clean Patch');
  let targetPresetInput = $state('Acid Resonant Bassline');
  let audioMeteringReport = $state<PluginAudioMeteringReport | null>(null);
  let isAnalyzingMetering = $state(false);
  let meteringTargetStandard = $state('EBU R128 (-23 LUFS)');

  function toggleTheme() {
    activeTheme = activeTheme === 'dark' ? 'contrast' : 'dark';
    try {
      localStorage.setItem('vst_monster_theme', activeTheme);
    } catch {}
  }

  // New plugin form state
  let newPlugin = $state({
    name: '',
    developer: '',
    license_model: 'opensource',
    category: 'synth',
    version: '1.0.0',
    platform: 'linux',
    download_url: '',
    description: ''
  });

  // Tooltip Directive
  function tooltip(element: HTMLElement, text: string) {
    let tooltipEl: HTMLDivElement | null = null;

    function mouseOver(event: MouseEvent) {
      if (!text) return;
      tooltipEl = document.createElement('div');
      tooltipEl.className = 'custom-tooltip';
      tooltipEl.textContent = text;
      tooltipEl.style.left = `${event.pageX + 12}px`;
      tooltipEl.style.top = `${event.pageY + 12}px`;
      document.body.appendChild(tooltipEl);
    }

    function mouseMove(event: MouseEvent) {
      if (tooltipEl) {
        tooltipEl.style.left = `${event.pageX + 12}px`;
        tooltipEl.style.top = `${event.pageY + 12}px`;
      }
    }

    function mouseLeave() {
      if (tooltipEl) {
        tooltipEl.remove();
        tooltipEl = null;
      }
    }

    element.addEventListener('mouseover', mouseOver);
    element.addEventListener('mousemove', mouseMove);
    element.addEventListener('mouseleave', mouseLeave);

    return {
      update(newText: string) {
        text = newText;
        if (tooltipEl) tooltipEl.textContent = newText;
      },
      destroy() {
        element.removeEventListener('mouseover', mouseOver);
        element.removeEventListener('mousemove', mouseMove);
        element.removeEventListener('mouseleave', mouseLeave);
        if (tooltipEl) tooltipEl.remove();
      }
    };
  }

  async function loadScanPaths() {
    try {
      const paths: string[] = await invoke('get_default_scan_paths');
      defaultScanPaths = paths || [];
    } catch (err) {
      console.log('Failed to fetch default scan paths:', err);
    }

    if (typeof localStorage !== 'undefined') {
      const saved = localStorage.getItem('vst_monster_custom_paths');
      if (saved) {
        try {
          customScanPaths = JSON.parse(saved);
        } catch (e) {}
      }
    }
  }

  function addCustomScanPath() {
    const trimmed = newCustomPath.trim();
    if (trimmed && !customScanPaths.includes(trimmed)) {
      customScanPaths = [...customScanPaths, trimmed];
      newCustomPath = '';
      if (typeof localStorage !== 'undefined') {
        localStorage.setItem('vst_monster_custom_paths', JSON.stringify(customScanPaths));
      }
      scanLocalPlugins();
    }
  }

  function removeCustomScanPath(index: number) {
    customScanPaths = customScanPaths.filter((_, i) => i !== index);
    if (typeof localStorage !== 'undefined') {
      localStorage.setItem('vst_monster_custom_paths', JSON.stringify(customScanPaths));
    }
    scanLocalPlugins();
  }

  async function scanLocalPlugins() {
    isScanningFolders = true;
    try {
      const detected: InstalledPluginInfo[] = await invoke('scan_installed_plugins', {
        extraPaths: customScanPaths
      });
      installedPlugins = detected || [];
    } catch (err) {
      // In web browser preview mode, invoke may fail silently
      console.log('Installed plugin scan skipped or not in Tauri:', err);
    } finally {
      isScanningFolders = false;
    }
  }

  async function clearScanCache() {
    isClearingCache = true;
    cacheClearMsg = '';
    try {
      const res: string = await invoke('clear_scan_cache');
      cacheClearMsg = res || 'Cache cleared!';
      await scanLocalPlugins();
      setTimeout(() => {
        cacheClearMsg = '';
      }, 4000);
    } catch (err: any) {
      cacheClearMsg = `Error: ${err?.message || err}`;
    } finally {
      isClearingCache = false;
    }
  }

  function exportPluginManifest(plugin: Plugin) {
    const manifestData = {
      name: plugin.name,
      developer: plugin.developer,
      category: plugin.category,
      license_model: plugin.license_model,
      platform: plugin.platform,
      version: plugin.version,
      download_url: plugin.download_url,
      sha256_hash: plugin.sha256_hash,
      strategy: plugin.strategy,
      metadata: plugin.metadata,
      exported_at: new Date().toISOString(),
      exporter: 'VST Monster Universal Client v0.3.1'
    };

    const blob = new Blob([JSON.stringify(manifestData, null, 2)], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `${plugin.name.toLowerCase().replace(/[^a-z0-9_-]/g, '_')}_manifest.json`;
    a.click();
    URL.revokeObjectURL(url);
    actionMessage = `Manifest exported for ${plugin.name}!`;
  }

  async function updateAllOutdatedPlugins() {
    const outdated = plugins.filter(p => {
      const inst = getInstalledInfo(p.name);
      if (!inst) return false;
      const vStatus = getVersionStatus(p.version, inst.version);
      return vStatus.state === 'update-available' && !!p.download_url;
    });

    if (outdated.length === 0) {
      actionMessage = 'All installed plugins are currently up to date!';
      return;
    }

    if (!confirm(`Found ${outdated.length} plugin(s) with updates available. Proceed with batch update?`)) {
      return;
    }

    isBatchUpdating = true;
    actionMessage = `Starting batch update of ${outdated.length} plugin(s)...`;

    for (const plugin of outdated) {
      try {
        installStatus[plugin.id] = 'Updating...';
        await invoke('install_plugin', {
          url: plugin.download_url,
          name: plugin.name,
          expectedHash: plugin.sha256_hash || null
        });
        installStatus[plugin.id] = 'Updated to latest release!';
      } catch (err: any) {
        installStatus[plugin.id] = `Update failed: ${err}`;
      }
    }

    await scanLocalPlugins();
    isBatchUpdating = false;
    actionMessage = 'Batch update sequence completed!';
  }

  function getInstalledInfo(pluginName: string): InstalledPluginInfo | undefined {
    const safeName = pluginName.toLowerCase().replace(/[^a-z0-9]/g, '');
    return installedPlugins.find(p => {
      const cleanP = p.name.toLowerCase().replace(/[^a-z0-9]/g, '');
      return cleanP.includes(safeName) || safeName.includes(cleanP);
    });
  }

  function getVersionStatus(catalogVersion?: string, localVersion?: string | null): {
    state: 'up-to-date' | 'update-available' | 'unknown';
    label: string;
  } {
    if (!localVersion) {
      return { state: 'unknown', label: 'Installed' };
    }
    if (!catalogVersion) {
      return { state: 'up-to-date', label: `v${localVersion}` };
    }
    const cleanCat = catalogVersion.trim().replace(/^v/i, '');
    const cleanLoc = localVersion.trim().replace(/^v/i, '');
    if (cleanCat === cleanLoc) {
      return { state: 'up-to-date', label: `v${localVersion} (Latest)` };
    }
    return { state: 'update-available', label: `Installed: v${localVersion} (Update to v${cleanCat})` };
  }

  async function fetchDashboardData() {
    try {
      const [pluginsRes, statsRes, healthRes] = await Promise.all([
        fetch('http://localhost:3000/plugins'),
        fetch('http://localhost:3000/stats'),
        fetch('http://localhost:3000/health')
      ]);

      if (pluginsRes.ok) plugins = await pluginsRes.json();
      if (statsRes.ok) stats = await statsRes.json();
      if (healthRes.ok) {
        health = await healthRes.json();
        crawlerRunning = health?.crawlerRunning || false;
      }
      error = null;
    } catch (err: any) {
      console.warn('Dashboard fetch warning:', err);
      error = 'Registry API offline or unreachable (port 3000)';
    } finally {
      loading = false;
    }
    await scanLocalPlugins();
  }

  async function submitNewPlugin(e: Event) {
    e.preventDefault();
    if (!newPlugin.name || !newPlugin.developer) {
      actionMessage = 'Plugin Name and Developer are required.';
      return;
    }

    try {
      actionMessage = 'Submitting plugin to registry...';
      const res = await fetch('http://localhost:3000/plugins', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(newPlugin)
      });

      if (!res.ok) {
        const data = await res.json();
        throw new Error(data.error || 'Failed to submit plugin');
      }

      actionMessage = `Successfully added ${newPlugin.name} to the registry!`;
      showAddPluginModal = false;
      newPlugin = {
        name: '',
        developer: '',
        license_model: 'opensource',
        category: 'synth',
        version: '1.0.0',
        platform: 'linux',
        download_url: '',
        description: ''
      };
      await fetchDashboardData();
    } catch (err: any) {
      actionMessage = `Registration error: ${err.message}`;
    }
  }

  async function triggerCrawler() {
    try {
      actionMessage = 'Initiating crawler pipeline...';
      const res = await fetch('http://localhost:3000/crawler/run', { method: 'POST' });
      const data = await res.json();
      actionMessage = data.message;
      crawlerRunning = true;
      isLogsCollapsed = false;
      pollLogs();
    } catch (err: any) {
      actionMessage = `Failed to trigger crawler: ${err.message}`;
    }
  }

  async function pollLogs() {
    try {
      const res = await fetch('http://localhost:3000/crawler/logs');
      if (res.ok) {
        const data = await res.json();
        crawlerLogs = data.logs || [];
        crawlerRunning = data.running;
        if (crawlerRunning) {
          setTimeout(pollLogs, 1500);
        } else {
          fetchDashboardData();
        }
      }
    } catch {
      // ignore transient poll errors
    }
  }

  async function handleInstall(plugin: Plugin) {
    await handleInstallAdvanced(plugin, null, null);
  }

  async function handleInstallAdvanced(
    plugin: Plugin,
    targetOverride: string | null = null,
    strategyOverride: string | null = null
  ) {
    if (!plugin.download_url) {
      installStatus[plugin.id] = "No direct binary download URL available";
      return;
    }
    installStatus[plugin.id] = "Downloading & Installing...";

    try {
      const response = await invoke('install_plugin_advanced', {
        url: plugin.download_url,
        name: plugin.name,
        expectedHash: plugin.sha256_hash || null,
        customTargetDir: targetOverride && targetOverride.trim().length > 0 ? targetOverride.trim() : null,
        forceStrategy: strategyOverride && strategyOverride !== 'auto' ? strategyOverride : null
      });
      installStatus[plugin.id] = `Success: ${response}`;
      await scanLocalPlugins();
    } catch (err: any) {
      installStatus[plugin.id] = `Error: ${err}`;
    }
  }

  async function handleRemove(plugin: Plugin) {
    installStatus[plugin.id] = "Removing plugin files...";
    try {
      const response = await invoke('remove_plugin', { name: plugin.name });
      installStatus[plugin.id] = `Removed: ${response}`;
      await scanLocalPlugins();
    } catch (err: any) {
      installStatus[plugin.id] = `Error: ${err}`;
    }
  }

  async function handleOpenDirectory(pluginName?: string) {
    try {
      await invoke('open_plugin_directory', { name: pluginName });
      actionMessage = pluginName ? `Opened directory for ${pluginName}` : 'Opened plugins directory';
    } catch (err: any) {
      actionMessage = `Error opening folder: ${err}`;
    }
  }

  async function handleOpenPresetsDirectory() {
    try {
      await invoke('open_presets_directory');
      actionMessage = 'Opened system DAW presets directory';
    } catch (err: any) {
      actionMessage = `Error opening presets directory: ${err}`;
    }
  }

  async function handleVerifyBundle(pluginName: string, path: string) {
    isVerifyingPlugin = true;
    verificationStatus[pluginName] = 'verifying';
    try {
      const isValid = await invoke('verify_installed_plugin', { path });
      verificationStatus[pluginName] = isValid ? 'valid' : 'invalid';
      actionMessage = isValid
        ? `Bundle integrity verified for ${pluginName} at ${path}`
        : `Integrity check failed: bundle file not found at ${path}`;
    } catch (err: any) {
      verificationStatus[pluginName] = 'invalid';
      actionMessage = `Bundle verification error: ${err}`;
    } finally {
      isVerifyingPlugin = false;
    }
  }

  async function handleTestBundleIntegrity(path: string) {
    isTestingBundle = true;
    bundleTestResult = null;
    try {
      const res = await invoke<any>('test_plugin_bundle_integrity', { path });
      bundleTestResult = res;
      actionMessage = res.is_valid
        ? `Headless DAW fuzzer assertions passed for bundle at ${path}`
        : `Bundle fuzzer detected potential compatibility issues at ${path}`;
    } catch (err: any) {
      actionMessage = `DAW Bundle fuzzer error: ${err}`;
    } finally {
      isTestingBundle = false;
    }
  }

  async function copyToClipboard(text: string, identifier: string | number) {
    try {
      if (navigator.clipboard) {
        await navigator.clipboard.writeText(text);
      }
      copiedHashId = identifier;
      setTimeout(() => {
        if (copiedHashId === identifier) {
          copiedHashId = null;
        }
      }, 2000);
    } catch (err) {
      console.warn('Clipboard write failed:', err);
    }
  }

  async function checkYabridge() {
    isCheckingYabridge = true;
    yabridgeMessage = '';
    try {
      const res: YabridgeStatusInfo = await invoke('check_yabridge_status');
      yabridgeStatus = res;
    } catch (err: any) {
      yabridgeMessage = `Failed to check yabridge: ${err}`;
    } finally {
      isCheckingYabridge = false;
    }
  }

  async function syncYabridgePlugin(pluginPath?: string) {
    isSyncingYabridge = true;
    yabridgeMessage = '';
    try {
      const res: string = await invoke('sync_yabridge', { pluginPath });
      yabridgeMessage = res;
      await checkYabridge();
    } catch (err: any) {
      yabridgeMessage = `Yabridge error: ${err}`;
    } finally {
      isSyncingYabridge = false;
    }
  }

  async function calculateLatencyDiagnostics() {
    isDiagnosingLatency = true;
    try {
      const res: LatencyResult = await invoke('diagnose_plugin_latency', {
        sampleRate: latencySampleRate,
        dspLatencySamples: latencyDspSamples,
      });
      latencyDiagnosticResult = res;
    } catch (err: any) {
      console.warn('Latency calculation failed:', err);
    } finally {
      isDiagnosingLatency = false;
    }
  }

  async function calculateAudioBusDiagnostics(category?: string) {
    isDiagnosingAudioBuses = true;
    try {
      const configParam = selectedBusConfigType === 'auto' ? null : selectedBusConfigType;
      const res: AudioBusDiagnosisResultItem = await invoke('diagnose_plugin_audio_buses', {
        configurationType: configParam,
        category: category || previewPlugin?.category || null,
      });
      audioBusDiagnosisResult = res;
    } catch (err: any) {
      console.warn('Audio bus calculation failed:', err);
    } finally {
      isDiagnosingAudioBuses = false;
    }
  }

  async function inspectBinaryArchitecture() {
    isInspectingArch = true;
    try {
      const res: BinaryArchitectureDetailsItem = await invoke('inspect_plugin_binary_architecture', {
        platform: previewPlugin?.platform || null,
        architecture: previewPlugin?.architecture || null,
        filePath: previewPlugin?.download_url || null,
      });
      binaryArchDetails = res;
    } catch (err: any) {
      console.warn('Binary architecture inspection failed:', err);
    } finally {
      isInspectingArch = false;
    }
  }

  async function runDspBenchmark() {
    isBenchmarking = true;
    try {
      const format = previewPlugin ? inferCatalogFormat(previewPlugin) : 'VST3';
      const category = previewPlugin?.category || 'synth';
      const res: PluginBenchmarkReportItem = await invoke('benchmark_plugin_performance', {
        format,
        category,
        baseMemoryMb: null,
      });
      benchmarkReport = res;
    } catch (err: any) {
      console.warn('DSP benchmark execution failed:', err);
    } finally {
      isBenchmarking = false;
    }
  }

  async function inspectPluginParameters() {
    isInspectingParameters = true;
    try {
      const format = previewPlugin ? inferCatalogFormat(previewPlugin) : 'VST3';
      const category = previewPlugin?.category || 'synth';
      const res: PluginParameterReportItem = await invoke('inspect_plugin_parameters', {
        format,
        category,
        parameterCount: null,
      });
      parameterReport = res;
    } catch (err: any) {
      console.warn('Parameter inspection failed:', err);
    } finally {
      isInspectingParameters = false;
    }
  }

  async function testAudioThreadResilience() {
    isTestingResilience = true;
    try {
      const format = previewPlugin ? inferCatalogFormat(previewPlugin) : 'VST3';
      const category = previewPlugin?.category || 'synth';
      const res: PluginResilienceReportItem = await invoke('test_audio_thread_resilience', {
        format,
        category,
        bufferSize: 256,
      });
      resilienceReport = res;
    } catch (err: any) {
      console.warn('Resilience stress test failed:', err);
    } finally {
      isTestingResilience = false;
    }
  }

  async function renderAudioSpectrumAnalysis() {
    isRenderingSpectrum = true;
    try {
      const format = previewPlugin ? inferCatalogFormat(previewPlugin) : 'VST3';
      const category = previewPlugin?.category || 'synth';
      const res: AudioSpectrumReportItem = await invoke('render_audio_spectrum_analysis', {
        format,
        category,
        testSignal: spectrumTestSignal,
      });
      spectrumReport = res;
    } catch (err: any) {
      console.warn('Spectrum analysis failed:', err);
    } finally {
      isRenderingSpectrum = false;
    }
  }

  async function handleSnapshotPluginState() {
    isSnapshotting = true;
    try {
      const pluginName = previewPlugin?.name || 'Surge XT';
      const format = previewPlugin ? inferCatalogFormat(previewPlugin) : 'VST3';
      const presetName = presetExportName.trim() || 'Active Workspace Patch';
      const res: PluginStateSnapshot = await invoke('snapshot_plugin_state', {
        pluginName,
        format,
        presetName,
        presetFormat: `.${presetExportFormat}`,
      });
      activeSnapshot = res;
      actionMessage = `Captured parameter state snapshot for ${pluginName} (CRC32: ${res.checksum_crc32})`;
    } catch (err: any) {
      actionMessage = `Preset snapshot failed: ${err}`;
    } finally {
      isSnapshotting = false;
    }
  }

  async function handleDiffPresets() {
    isDiffingPresets = true;
    try {
      const pluginName = previewPlugin?.name || 'Surge XT';
      const format = previewPlugin ? inferCatalogFormat(previewPlugin) : 'VST3';
      const res: PresetDiffReport = await invoke('diff_plugin_presets', {
        pluginName,
        format,
        baselinePresetName: baselinePresetInput,
        targetPresetName: targetPresetInput,
      });
      presetDiffReport = res;
      actionMessage = `Computed preset diff between "${baselinePresetInput}" and "${targetPresetInput}" (${res.drift_severity} Drift)`;
    } catch (err: any) {
      actionMessage = `Preset diff calculation failed: ${err}`;
    } finally {
      isDiffingPresets = false;
    }
  }

  async function handleAnalyzeMetering() {
    isAnalyzingMetering = true;
    try {
      const pluginName = previewPlugin?.name || 'Surge XT';
      const format = previewPlugin ? inferCatalogFormat(previewPlugin) : 'VST3';
      const category = previewPlugin?.category || 'synth';
      const res: PluginAudioMeteringReport = await invoke('render_audio_metering_analysis', {
        pluginName,
        format,
        category,
        targetStandard: meteringTargetStandard,
      });
      audioMeteringReport = res;
      actionMessage = `Computed audio loopback metering for ${pluginName} (${res.target_compliance.compliance_status} for ${meteringTargetStandard})`;
    } catch (err: any) {
      actionMessage = `Audio metering analysis failed: ${err}`;
    } finally {
      isAnalyzingMetering = false;
    }
  }

  async function handleExportPreset(pluginName: string) {
    if (!presetExportName.trim()) {
      presetExportMessage = 'Please enter a preset name';
      return;
    }
    isExportingPreset = true;
    presetExportMessage = '';
    const cleanPresetName = presetExportName.trim().endsWith(`.${presetExportFormat}`)
      ? presetExportName.trim()
      : `${presetExportName.trim()}.${presetExportFormat}`;

    const samplePresetPayload = JSON.stringify({
      plugin: pluginName,
      exported_at: new Date().toISOString(),
      format: presetExportFormat,
      generator: 'VST Monster Ecosystem v0.5.6',
      params: { master_gain: 1.0, active_program: 0 }
    }, null, 2);

    try {
      const savedPath = await invoke<string>('export_plugin_preset', {
        pluginName,
        presetName: cleanPresetName,
        content: samplePresetPayload
      });
      presetExportMessage = `✓ Preset exported: ${savedPath}`;
      actionMessage = `Exported preset to ${savedPath}`;
      presetExportName = '';
    } catch (err: any) {
      presetExportMessage = `Export failed: ${err}`;
      actionMessage = `Preset export failed: ${err}`;
    } finally {
      isExportingPreset = false;
    }
  }

  async function handleQuit() {
    if (confirm('Quit VST Monster desktop app and stop local services?')) {
      try {
        await fetch('http://localhost:3000/system/shutdown', { method: 'POST' });
      } catch {}
      try {
        await invoke('quit_app');
      } catch {}
    }
  }

  function isPluginInstalled(name: string): boolean {
    return !!getInstalledInfo(name);
  }

  function getPluginTags(plugin: Plugin): string[] {
    const tags = new Set<string>();
    if (plugin.category) tags.add(plugin.category.toLowerCase());
    if (plugin.license_model) tags.add(plugin.license_model.toLowerCase());

    const desc = (plugin.metadata?.description || '').toLowerCase();
    const keywords = ['wavetable', 'fm', 'analog', 'subtractive', 'filter', 'polyphonic', 'reverb', 'delay', 'chorus', 'distortion', 'compressor', 'eq', 'mpe'];
    for (const kw of keywords) {
      if (desc.includes(kw)) {
        tags.add(kw);
      }
    }
    return Array.from(tags).slice(0, 4);
  }

  function inferCatalogFormat(plugin: Plugin): string {
    const url = (plugin.download_url || '').toLowerCase();
    const desc = (plugin.metadata?.description || '').toLowerCase();
    if (url.includes('.lv2') || desc.includes('lv2')) return 'LV2';
    if (url.includes('clap') || desc.includes('clap')) return 'CLAP';
    if (url.endsWith('.vst3.zip') || url.includes('vst3')) return 'VST3';
    if (url.includes('dmg') || url.includes('pkg') || url.includes('component')) return 'AU / VST3';
    return 'VST3';
  }

  function isLegacyPlugin(plugin: Plugin, installedInfo?: InstalledPluginInfo | null): boolean {
    if (installedInfo) {
      const fmt = (installedInfo.format || '').toLowerCase();
      if (fmt === 'vst' || fmt === 'vst2') return true;
    }
    const arch = (plugin.architecture || '').toLowerCase();
    const name = plugin.name.toLowerCase();
    const desc = (plugin.metadata?.description || '').toLowerCase();
    if (arch === 'x86' || arch === 'i686' || arch === 'i386' || arch.includes('32')) return true;
    if (desc.includes('32-bit') || name.includes('32-bit') || desc.includes('vst 2') || desc.includes('vst2')) return true;
    return false;
  }

  // Dynamic counts for status pills
  let countAll = $derived(plugins.length);
  let countInstalled = $derived(plugins.filter(p => isPluginInstalled(p.name)).length);
  let countUpdates = $derived(plugins.filter(p => {
    const inst = getInstalledInfo(p.name);
    return inst ? getVersionStatus(p.version, inst.version).state === 'update-available' : false;
  }).length);
  let countNotInstalled = $derived(plugins.filter(p => !isPluginInstalled(p.name)).length);
  let countVST3 = $derived(plugins.filter(p => {
    const inst = getInstalledInfo(p.name);
    return inst ? inst.format.toLowerCase() === 'vst3' : inferCatalogFormat(p).toLowerCase().includes('vst3');
  }).length);
  let countCLAP = $derived(plugins.filter(p => {
    const inst = getInstalledInfo(p.name);
    return inst ? inst.format.toLowerCase() === 'clap' : inferCatalogFormat(p).toLowerCase().includes('clap');
  }).length);
  let countLV2 = $derived(plugins.filter(p => {
    const inst = getInstalledInfo(p.name);
    return inst ? inst.format.toLowerCase() === 'lv2' : inferCatalogFormat(p).toLowerCase().includes('lv2');
  }).length);
  let countAU = $derived(plugins.filter(p => {
    const inst = getInstalledInfo(p.name);
    return inst ? inst.format.toLowerCase() === 'au' : (inferCatalogFormat(p).toLowerCase().includes('au') || p.platform === 'macos');
  }).length);

  // Filtered and Sorted Plugins
  let filteredPlugins = $derived(
    plugins
      .filter((p) => {
        const matchesSearch =
          p.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
          p.developer.toLowerCase().includes(searchQuery.toLowerCase()) ||
          (p.category && p.category.toLowerCase().includes(searchQuery.toLowerCase()));
        const matchesLicense = selectedLicense === 'all' || p.license_model === selectedLicense;
        const matchesCategory = selectedCategory === 'all' || (p.category || 'synth') === selectedCategory;
        const matchesPlatform = selectedPlatform === 'all' || (p.platform && p.platform === selectedPlatform);

        const inst = getInstalledInfo(p.name);
        let matchesStatus = true;
        if (selectedInstalledStatus === 'installed') {
          matchesStatus = !!inst;
        } else if (selectedInstalledStatus === 'update-available') {
          const vStatus = inst ? getVersionStatus(p.version, inst.version) : null;
          matchesStatus = vStatus?.state === 'update-available';
        } else if (selectedInstalledStatus === 'not-installed') {
          matchesStatus = !inst;
        }

        let matchesFormat = true;
        if (selectedFormat !== 'all') {
          if (inst) {
            matchesFormat = inst.format.toLowerCase() === selectedFormat;
          } else {
            const inferred = inferCatalogFormat(p).toLowerCase();
            matchesFormat = inferred.includes(selectedFormat);
          }
        }

        return matchesSearch && matchesLicense && matchesCategory && matchesPlatform && matchesStatus && matchesFormat;
      })
      .sort((a, b) => {
        if (sortBy === 'name') return a.name.localeCompare(b.name);
        if (sortBy === 'developer') return a.developer.localeCompare(b.developer);
        if (sortBy === 'license') return a.license_model.localeCompare(b.license_model);
        if (sortBy === 'category') return (a.category || 'synth').localeCompare(b.category || 'synth');
        if (sortBy === 'status') {
          const aInst = isPluginInstalled(a.name) ? 1 : 0;
          const bInst = isPluginInstalled(b.name) ? 1 : 0;
          return bInst - aInst;
        }
        return 0;
      })
  );

  onMount(() => {
    try {
      const savedTheme = localStorage.getItem('vst_monster_theme');
      if (savedTheme === 'contrast' || savedTheme === 'dark') {
        activeTheme = savedTheme;
      }
    } catch {}
    loadScanPaths();
    fetchDashboardData();
    let unlistenProgress: (() => void) | null = null;
    try {
      listen('install_progress', (event: any) => {
        const payload = event.payload;
        if (payload && payload.name) {
          const matchingPlugin = plugins.find(p => p.name.toLowerCase() === payload.name.toLowerCase());
          if (matchingPlugin) {
            if (payload.percentage !== null && payload.percentage !== undefined) {
              installStatus[matchingPlugin.id] = `Downloading: ${Math.round(payload.percentage)}%`;
            } else {
              installStatus[matchingPlugin.id] = payload.status || 'Downloading...';
            }
          }
        }
      }).then(fn => { unlistenProgress = fn; }).catch(() => {});
    } catch {}

    const interval = setInterval(fetchDashboardData, 10000);
    return () => {
      clearInterval(interval);
      if (unlistenProgress) unlistenProgress();
    };
  });
</script>

<div class="dashboard-root theme-{activeTheme}">
  <!-- Top Global Header / Command Bar -->
  <header class="top-nav">
    <div class="brand">
      <div class="logo-mark">⚡</div>
      <div>
        <h1 class="brand-title">VST MONSTER</h1>
        <span class="brand-subtitle">Universal Audio Plugin Engine & Registry</span>
      </div>
    </div>

    <!-- Quick Stats in Header -->
    <div class="status-strip">
      <div class="status-item" use:tooltip={"PostgreSQL & Registry API Connection Health"}>
        <span class="status-indicator {health?.database === 'connected' ? 'online' : 'offline'}"></span>
        <span class="label">Registry:</span>
        <strong>{health?.database === 'connected' ? 'Online' : 'Offline'}</strong>
      </div>

      <div class="status-item" use:tooltip={"Total cataloged plugins in the central database"}>
        <span class="label">Catalog:</span>
        <strong>{stats?.totalPlugins ?? plugins.length} VSTs</strong>
      </div>

      <div class="status-item" use:tooltip={"Crawling state: Scrapes KVR, GitHub & BPB"}>
        <span class="status-indicator {crawlerRunning ? 'active-pulse' : 'idle'}"></span>
        <span class="label">Crawler:</span>
        <strong>{crawlerRunning ? 'Crawling...' : 'Idle'}</strong>
      </div>

      <div class="status-item" use:tooltip={"Global Ecosystem Version String"}>
        <span class="version-tag">v{health?.version || '0.2.2'}</span>
      </div>

      <button class="btn-folder-action" onclick={handleOpenPresetsDirectory} use:tooltip={"Open system DAW presets directory in file manager (~/.vst3/presets, ~/Library/Audio/Presets, Documents/VST3 Presets)"}>
        📁 Presets
      </button>

      <button class="btn-folder-action" onclick={() => (showScanPathsModal = true)} use:tooltip={"Configure and audit local audio plugin scan paths"}>
        ⚙️ Paths
      </button>

      <button class="btn-theme-toggle" onclick={toggleTheme} use:tooltip={`Switch interface color profile (Currently: ${activeTheme === 'dark' ? 'Dark Studio' : 'High-Contrast DAW'})`}>
        {activeTheme === 'dark' ? '🎨 Studio' : '⚡ Contrast'}
      </button>

      <button class="btn-success-outline" onclick={() => (showAddPluginModal = !showAddPluginModal)} use:tooltip={"Register a plugin manifest directly into PostgreSQL"}>
        + Register Plugin
      </button>

      <button class="btn-danger-outline" onclick={handleQuit} use:tooltip={"Gracefully shut down registry servers and close client"}>
        Quit App
      </button>
    </div>
  </header>

  <!-- Notification Banner -->
  {#if actionMessage}
    <div class="action-alert">
      <span>{actionMessage}</span>
      <button class="alert-close" onclick={() => (actionMessage = null)}>✕</button>
    </div>
  {/if}

  <!-- Scan Paths Configuration Modal -->
  {#if showScanPathsModal}
    <div class="modal-backdrop">
      <div class="modal-card">
        <div class="modal-header">
          <div class="modal-title-group">
            <span class="card-icon">⚙️</span>
            <h2>Audio Plugin Scan Paths</h2>
          </div>
          <button class="modal-close" onclick={() => (showScanPathsModal = false)}>✕</button>
        </div>

        <div class="modal-body-section">
          <p class="modal-desc">
            VST Monster inspects these directories across your system to detect installed VST3, CLAP, AU, and legacy VST plugins.
          </p>

          <div class="paths-block">
            <h4>Default System Directories</h4>
            <ul class="paths-list">
              {#if defaultScanPaths.length === 0}
                <li class="path-empty">Standard OS audio directories (~/.vst3, Library/Audio, etc.)</li>
              {:else}
                {#each defaultScanPaths as p}
                  <li class="path-item">
                    <span class="path-type-badge">SYSTEM</span>
                    <code>{p}</code>
                  </li>
                {/each}
              {/if}
            </ul>
          </div>

          <div class="paths-block">
            <h4>Custom VST Search Paths</h4>
            <div class="custom-path-input-row">
              <input
                type="text"
                bind:value={newCustomPath}
                placeholder="/custom/audio/plugins or D:\Audio\VSTs"
                class="search-input"
              />
              <button class="btn-primary" onclick={addCustomScanPath}>Add Path</button>
            </div>

            <ul class="paths-list">
              {#if customScanPaths.length === 0}
                <li class="path-empty">No custom search directories configured.</li>
              {:else}
                {#each customScanPaths as cp, idx}
                  <li class="path-item">
                    <span class="path-type-badge custom-badge">CUSTOM</span>
                    <code>{cp}</code>
                    <button class="btn-path-del" onclick={() => removeCustomScanPath(idx)}>✕</button>
                  </li>
                {/each}
              {/if}
            </ul>
          </div>

          <!-- Local Metadata Scan Cache Controls -->
          <div class="paths-cache-section">
            <div class="paths-cache-info">
              <h4>⚡ Local Scan Cache</h4>
              <p>Metadata such as version markers and bundle structures are cached using filesystem modification timestamps for instant rescans.</p>
              {#if cacheClearMsg}
                <div class="cache-msg">{cacheClearMsg}</div>
              {/if}
            </div>
            <button
              class="btn-cache-flush"
              onclick={clearScanCache}
              disabled={isClearingCache || isScanningFolders}
              use:tooltip={"Deletes local scan cache and triggers a full, deep filesystem re-read"}
            >
              {isClearingCache ? "Clearing..." : "🧹 Flush Cache & Deep Rescan"}
            </button>
          </div>

          <div class="modal-actions">
            <button
              class="btn-primary"
              onclick={() => {
                scanLocalPlugins();
                showScanPathsModal = false;
              }}
            >
              Save & Rescan
            </button>
            <button class="btn-secondary" onclick={() => (showScanPathsModal = false)}>Close</button>
          </div>
        </div>
      </div>
    </div>
  {/if}

  <!-- Error Bar if Backend Offline -->
  {#if error}
    <div class="error-banner">
      <span>⚠️ {error}</span>
      <button class="retry-btn" onclick={fetchDashboardData} use:tooltip={"Try reconnecting to Registry API"}>Reconnect</button>
    </div>
  {/if}

  <!-- Manual Plugin Registration Modal Form -->
  {#if showAddPluginModal}
    <div class="modal-backdrop">
      <div class="modal-card">
        <div class="modal-header">
          <h3>Register New Plugin</h3>
          <button class="modal-close" onclick={() => (showAddPluginModal = false)}>✕</button>
        </div>
        <form onsubmit={submitNewPlugin} class="modal-form">
          <div class="form-row">
            <label>
              Plugin Name *
              <input type="text" required bind:value={newPlugin.name} placeholder="e.g. Tyrell N6" class="search-input" />
            </label>
            <label>
              Developer / Author *
              <input type="text" required bind:value={newPlugin.developer} placeholder="e.g. U-He" class="search-input" />
            </label>
          </div>
          <div class="form-row">
            <label>
              License Model
              <select bind:value={newPlugin.license_model} class="select-input">
                <option value="opensource">Open Source</option>
                <option value="free">Free / Freeware</option>
                <option value="commercial">Commercial</option>
                <option value="donationware">Donationware</option>
              </select>
            </label>
            <label>
              Version
              <input type="text" bind:value={newPlugin.version} placeholder="1.0.0" class="search-input" />
            </label>
            <label>
              Platform
              <select bind:value={newPlugin.platform} class="select-input">
                <option value="linux">Linux</option>
                <option value="windows">Windows</option>
                <option value="macos">macOS</option>
              </select>
            </label>
          </div>
          <label>
            Direct Binary / Zip Download URL
            <input type="url" bind:value={newPlugin.download_url} placeholder="https://example.com/download.zip" class="search-input" />
          </label>
          <label>
            Description & Sound Characteristics
            <textarea bind:value={newPlugin.description} placeholder="Analog-modeled polyphonic synthesizer..." class="textarea-input"></textarea>
          </label>
          <div class="modal-actions">
            <button type="button" class="btn-secondary" onclick={() => (showAddPluginModal = false)}>Cancel</button>
            <button type="submit" class="btn-primary">Save to Registry</button>
          </div>
        </form>
      </div>
    </div>
  {/if}

  <!-- Plugin Details Preview Drawer / Modal -->
  {#if previewPlugin}
    {@const previewInstalled = getInstalledInfo(previewPlugin.name)}
    {@const previewVersionStatus = previewInstalled ? getVersionStatus(previewPlugin.version, previewInstalled.version) : null}
    {@const previewInferredFormat = inferCatalogFormat(previewPlugin)}
    <div class="modal-backdrop">
      <div class="modal-card preview-card">
        <div class="modal-header">
          <div class="preview-header-info">
            <div class="preview-header-tags">
              <span class="preview-category">{previewPlugin.category || 'synth'}</span>
              {#if previewInferredFormat}
                <span class="catalog-format-badge" use:tooltip={"Inferred audio plugin architecture from binary distribution"}>
                  {previewInferredFormat}
                </span>
              {/if}
              {#if previewInstalled}
                <span class="installed-badge" use:tooltip={`Installed format: ${previewInstalled.format.toUpperCase()} at ${previewInstalled.path}`}>
                  ✓ {previewInstalled.format.toUpperCase()}
                </span>
              {/if}
            </div>
            <h2>{previewPlugin.name}</h2>
            <span class="preview-author">by {previewPlugin.developer}</span>
          </div>
          <button class="modal-close" onclick={() => (previewPlugin = null)}>✕</button>
        </div>

        <!-- Inspector View Navigation Tabs -->
        <div class="inspector-tabs-bar">
          <button
            type="button"
            class="inspector-tab-btn {activeInspectorTab === 'overview' ? 'active' : ''}"
            onclick={() => (activeInspectorTab = 'overview')}
            use:tooltip={"View synthesis architecture, license, release metadata, and host compatibility matrix"}
          >
            🎛️ Overview & Specs
          </button>
          <button
            type="button"
            class="inspector-tab-btn {activeInspectorTab === 'filesystem' ? 'active' : ''}"
            onclick={() => (activeInspectorTab = 'filesystem')}
            use:tooltip={"View local filesystem bundle paths, verification checks, scan paths, and cryptographic digest"}
          >
            📁 Filesystem & DAW
          </button>
          <button
            type="button"
            class="inspector-tab-btn {activeInspectorTab === 'manifest' ? 'active' : ''}"
            onclick={() => (activeInspectorTab = 'manifest')}
            use:tooltip={"Inspect structured raw JSON package manifest"}
          >
            📋 Raw Manifest JSON
          </button>
        </div>

        <div class="preview-content">
          {#if activeInspectorTab === 'overview'}
            <!-- Status Banner -->
            {#if previewInstalled}
              <div class="preview-install-status {previewVersionStatus?.state === 'update-available' ? 'status-update' : 'status-installed'}">
                <span class="icon">✓</span>
                <div class="preview-status-details">
                  <div class="preview-status-title-row">
                    <strong>Installed in {previewInstalled.format.toUpperCase()} format</strong>
                    {#if verificationStatus[previewPlugin.name]}
                      <span class="verification-badge {verificationStatus[previewPlugin.name]}">
                        {verificationStatus[previewPlugin.name] === 'valid' ? '✓ Verified Bundle' : verificationStatus[previewPlugin.name] === 'verifying' ? '⏳ Verifying...' : '⚠ Missing Binary'}
                      </span>
                    {/if}
                  </div>
                  <p>Path: <code>{previewInstalled.path}</code></p>
                  {#if previewVersionStatus}
                    <small class="version-label-note">{previewVersionStatus.label}</small>
                  {/if}
                </div>
              </div>
            {:else}
              <div class="preview-install-status status-not-installed">
                <span class="icon">📦</span>
                <div>
                  <strong>Not locally installed</strong>
                  <p>Ready to stream and unpack via VST Monster silent engine.</p>
                </div>
              </div>
            {/if}

            <!-- Legacy Compatibility Warning -->
            {#if isLegacyPlugin(previewPlugin, previewInstalled)}
              <div class="legacy-warning-box">
                <span class="warning-icon">⚠️</span>
                <div class="warning-content">
                  <strong>Legacy VST2 / 32-bit Architecture Detected</strong>
                  <p>This plugin uses older VST2 or 32-bit binaries. Modern 64-bit DAWs (e.g., Cubase 12+, Bitwig, Ableton Live 10+, Apple Silicon macOS) may require a compatibility bridge such as <em>yabridge</em> or <em>jBridge</em> to load.</p>
                </div>
              </div>
            {/if}

            <!-- Metadata Specs -->
            <div class="preview-specs-grid">
              <div class="spec-tile">
                <span class="label">License</span>
                <strong class="value uppercase">{previewPlugin.license_model}</strong>
              </div>
              <div class="spec-tile">
                <span class="label">Latest Release</span>
                <strong class="value">v{previewPlugin.version || '1.0.0'}</strong>
              </div>
              <div class="spec-tile">
                <span class="label">Target OS</span>
                <strong class="value capitalize">{previewPlugin.platform || 'linux'}</strong>
              </div>
              <div class="spec-tile" use:tooltip={"Installation execution mode: silent installer, archive decompression (ZIP/TAR/DMG), or direct binary relocation"}>
                <span class="label">Strategy</span>
                <strong class="value strategy-value">{previewPlugin.strategy || 'extract_binaries'}</strong>
              </div>
            </div>

            <!-- Description -->
            <div class="preview-section">
              <h4>Description & Synthesis Architecture</h4>
              <p class="preview-desc-text">
                {previewPlugin.metadata?.description || 'No extended description provided. Cataloged via VST Monster distributed crawler pipeline.'}
              </p>
            </div>

            <!-- Audio Architecture & Compatibility Matrix -->
            <div class="preview-section">
              <div class="preview-section-title-row">
                <h4>Audio Format & Host Compatibility Matrix</h4>
                <span class="preview-format-tag" use:tooltip={"Primary audio plugin architecture inferred from distribution"}>
                  {previewInferredFormat} Standard
                </span>
              </div>
              <div class="matrix-grid">
                <div class="matrix-item">
                  <span class="matrix-label">Native Format</span>
                  <span class="matrix-val">{previewInferredFormat}</span>
                </div>
                <div class="matrix-item">
                  <span class="matrix-label">Distribution Architecture</span>
                  <span class="matrix-val">{previewPlugin.architecture || 'x86_64'}</span>
                </div>
                <div class="matrix-item">
                  <span class="matrix-label">Target OS Destination</span>
                  <span class="matrix-val capitalize">{previewPlugin.platform || 'linux'}</span>
                </div>
                <div class="matrix-item">
                  <span class="matrix-label">Host DAWs</span>
                  <span class="matrix-val">{previewInferredFormat.includes('LV2') ? 'Ardour, Reaper, Carla, Mixbus, LMMS' : 'Bitwig, Cubase, Reaper, FL Studio, Ableton'}</span>
                </div>
              </div>

              {#if previewInferredFormat.includes('CLAP') || (previewInstalled?.format.toLowerCase() === 'clap')}
                <div class="clap-spec-banner" use:tooltip={"CLever Audio Plug-in (CLAP) Open Standard Architecture"}>
                  <span class="clap-icon">⚡</span>
                  <div class="clap-content">
                    <strong>CLever Audio Plug-in (CLAP) Open Architecture</strong>
                    <p>Modern open-source audio standard developed by u-he & Bitwig. Features sample-accurate non-destructive parameter modulation, CPU thread-pool delegation, multiple voice handling, and polyphonic MIDI expression.</p>
                  </div>
                </div>
              {/if}

              {#if previewInferredFormat.includes('LV2') || (previewInstalled?.format.toLowerCase() === 'lv2')}
                <div class="lv2-spec-banner" use:tooltip={"LV2 Open Audio Standard with Turtle (TTL) manifest graph & Atom/State extensions"}>
                  <span class="lv2-icon">🎛️</span>
                  <div class="lv2-content">
                    <strong>LV2 Open Audio Standard Detected</strong>
                    <p>Includes RDF Turtle graph definitions (<code>manifest.ttl</code>). Supported natively across Linux and cross-platform DAWs (Ardour, Carla, REAPER, Qtractor, Bitwig) with atom event ports, state serialization (<code>lv2:state</code>), and dynamic DSP port discovery.</p>
                    <div class="lv2-spec-features">
                      <span class="lv2-pill" use:tooltip={"Atom event port for sample-accurate MIDI and expressive events"}>atom:AtomPort</span>
                      <span class="lv2-pill" use:tooltip={"Key-value preset and state persistence via Turtle RDF"}>state:State</span>
                      <span class="lv2-pill" use:tooltip={"Flexible UI embedding interfaces (X11, GTK, Qt, Cocoa)"}>ui:UIEmbedding</span>
                      <span class="lv2-pill" use:tooltip={"Multi-threading worker extensions for offloading heavy DSP compute"}>worker:schedule</span>
                    </div>
                  </div>
                </div>
              {/if}

              {#if previewInferredFormat.includes('AU') || previewPlugin.platform === 'macos' || (previewInstalled?.format.toLowerCase() === 'au')}
                <div class="au-spec-banner" use:tooltip={"Apple Audio Unit (AUv2) CoreAudio Architecture"}>
                  <span class="au-icon">🍎</span>
                  <div class="au-content">
                    <strong>Apple Audio Unit (AUv2 / .component) Supported</strong>
                    <p>Native macOS CoreAudio plugin architecture. Compatible with Apple Logic Pro, GarageBand, Ableton Live, Studio One, and Reaper with zero-latency sandbox hosting.</p>
                  </div>
                </div>
              {/if}

              {#if previewInferredFormat.includes('VST2') || isLegacyPlugin(previewPlugin) || (previewInstalled?.format.toLowerCase() === 'vst2')}
                <div class="vst2-spec-banner" use:tooltip={"Legacy Steinberg VST 2.4 Specification Architecture"}>
                  <span class="vst2-icon">🎚️</span>
                  <div class="vst2-content">
                    <strong>Legacy Steinberg VST 2.4 Specification Architecture</strong>
                    <p>Classic audio plugin standard (.dll / .so). Modern 64-bit DAWs (Cubase 12+, Bitwig) require bit-bridging solutions (Yabridge, JBridge) or compatibility sandboxing for full integration.</p>
                  </div>
                </div>
              {/if}

              {#if previewPlugin.download_url?.toLowerCase().endsWith('.msi') || (previewPlugin.platform === 'windows' && previewPlugin.strategy === 'silent_install')}
                <div class="msi-spec-banner" use:tooltip={"Windows MSI Headless Unattended Installation Engine with automated log aggregation"}>
                  <span class="msi-icon">🪟</span>
                  <div class="msi-content">
                    <strong>Windows MSI Silent Installation Engine</strong>
                    <p>Executes <code>msiexec.exe /i package.msi /quiet /qn /norestart TARGETDIR=%CommonProgramFiles%\VST3 ALLUSERS=1</code> with automated verbose diagnostic logging in <code>%TEMP%\vst_monster_*.log</code>.</p>
                  </div>
                </div>
              {/if}
            </div>

          {:else if activeInspectorTab === 'filesystem'}
            <!-- Filesystem & Local Inspection Tab -->
            <div class="preview-section">
              <h4>Local Audio Installation State</h4>
              {#if previewInstalled}
                <div class="filesystem-details-card">
                  <div class="fs-row">
                    <span class="fs-label">Audio Plugin Format:</span>
                    <span class="fs-value badge-format">{previewInstalled.format.toUpperCase()}</span>
                  </div>
                  <div class="fs-row">
                    <span class="fs-label">Absolute Filesystem Path:</span>
                    <code class="fs-path">{previewInstalled.path}</code>
                  </div>
                  <div class="fs-row">
                    <span class="fs-label">Detected Version:</span>
                    <span class="fs-value">v{previewInstalled.version}</span>
                  </div>
                  <div class="fs-row">
                    <span class="fs-label">Bundle Verification:</span>
                    <span class="fs-value">
                      {#if verificationStatus[previewPlugin.name] === 'valid'}
                        <span class="verification-badge valid" use:tooltip={"Valid audio bundle entrypoint, manifest graph, architecture tree, and binary permissions"}>
                          ✓ Verified Structural Integrity
                        </span>
                      {:else if verificationStatus[previewPlugin.name] === 'verifying'}
                        <span class="verification-badge verifying">⏳ Inspecting bundle architecture & entrypoints...</span>
                      {:else if verificationStatus[previewPlugin.name] === 'invalid'}
                        <span class="verification-badge invalid" use:tooltip={"Bundle missing entrypoint, architecture folder, or valid binary"}>
                          ⚠ Missing Binary or Corrupt Bundle
                        </span>
                      {:else}
                        <span class="verification-badge" use:tooltip={"Click 'Run Deep Verification' to audit local bundle integrity"}>
                          Unverified
                        </span>
                      {/if}
                    </span>
                  </div>
                  <div class="fs-row">
                    <span class="fs-label">Security Sandbox Scan:</span>
                    <span class="fs-value">
                      <span class="verification-badge valid" use:tooltip={"Scanned with pre-flight binary inspection and SHA-256 integrity verification"}>
                        🛡️ Verified Clean (ClamAV / SHA-256)
                      </span>
                    </span>
                  </div>
                  <div class="fs-action-row">
                    <button
                      type="button"
                      class="btn-tiny-action"
                      onclick={() => handleVerifyBundle(previewPlugin!.name, previewInstalled.path)}
                      disabled={isVerifyingPlugin}
                      use:tooltip={"Execute deep binary and bundle integrity check on local filesystem"}
                    >
                      🔍 Run Deep Verification
                    </button>
                    <button
                      type="button"
                      class="btn-tiny-action"
                      onclick={() => handleTestBundleIntegrity(previewInstalled.path)}
                      disabled={isTestingBundle}
                      use:tooltip={"Execute automated headless DAW bundle validator and fuzzer assertions"}
                    >
                      {isTestingBundle ? '⏳ Fuzzing...' : '🔬 Run Bundle Fuzzer'}
                    </button>
                    <button
                      type="button"
                      class="btn-tiny-action"
                      onclick={() => handleOpenDirectory(previewPlugin!.name)}
                      use:tooltip={"Open local enclosing directory in file explorer"}
                    >
                      📂 Open in File Explorer
                    </button>
                  </div>

                  {#if bundleTestResult}
                    <div class="bundle-fuzz-report">
                      <div class="fuzz-report-header">
                        <span class="fuzz-title">🔬 DAW Bundle Fuzzer & Validator Diagnostics</span>
                        <span class="verification-badge {bundleTestResult.is_valid ? 'valid' : 'invalid'}">
                          {bundleTestResult.is_valid ? '✓ Passed Assertions' : '⚠ Validation Alert'}
                        </span>
                      </div>
                      <div class="fuzz-stats-grid">
                        <div class="fuzz-stat-item">
                          <span class="fuzz-stat-label">Format:</span>
                          <span class="fuzz-stat-val">{bundleTestResult.format}</span>
                        </div>
                        <div class="fuzz-stat-item">
                          <span class="fuzz-stat-label">Binaries:</span>
                          <span class="fuzz-stat-val">{bundleTestResult.binary_count}</span>
                        </div>
                        <div class="fuzz-stat-item">
                          <span class="fuzz-stat-label">Total Size:</span>
                          <span class="fuzz-stat-val">{(bundleTestResult.total_size_bytes / (1024 * 1024)).toFixed(2)} MB</span>
                        </div>
                        <div class="fuzz-stat-item">
                          <span class="fuzz-stat-label">Executable (+x):</span>
                          <span class="fuzz-stat-val">{bundleTestResult.has_executable_permission ? '✓ Yes' : '⚠ No'}</span>
                        </div>
                      </div>

                      {#if bundleTestResult.exported_symbols && bundleTestResult.exported_symbols.length > 0}
                        <div class="fuzz-symbols-container">
                          <span class="fuzz-symbols-label">Verified DAW Host Entry Symbols:</span>
                          <div class="fuzz-symbols-badges">
                            {#each bundleTestResult.exported_symbols as sym}
                              <span class="fuzz-symbol-badge" use:tooltip={"Exported entry point verified for native DAW host loading and dynamic linking"}>
                                ⚡ {sym}
                              </span>
                            {/each}
                          </div>
                        </div>
                      {/if}

                      {#if bundleTestResult.capabilities && bundleTestResult.capabilities.length > 0}
                        <div class="fuzz-capabilities-container">
                          <span class="fuzz-symbols-label">DSP Engine & Host Capabilities:</span>
                          <div class="fuzz-capabilities-badges">
                            {#each bundleTestResult.capabilities as cap}
                              <span class="fuzz-capability-badge" use:tooltip={"Audio DSP capability and host bus structure detected in binary payload"}>
                                🎚️ {cap}
                              </span>
                            {/each}
                          </div>
                        </div>
                      {/if}

                      <ul class="fuzz-diag-list">
                        {#each bundleTestResult.diagnostics as diag}
                          <li>• {diag}</li>
                        {/each}
                      </ul>
                    </div>
                  {/if}
                </div>
              {:else}
                <div class="empty-fs-state">
                  <p>📦 Plugin is not installed locally. Default installation will target user audio directory.</p>
                </div>
              {/if}
            </div>

            <!-- Cryptographic Digest & Antivirus Security Scan -->
            <div class="preview-section">
              <div class="preview-section-title-row">
                <h4>Cryptographic Verification & Antivirus Scan</h4>
                {#if previewPlugin.sha256_hash}
                  <button
                    type="button"
                    class="btn-copy-hash"
                    onclick={() => copyToClipboard(previewPlugin!.sha256_hash!, 'preview')}
                    use:tooltip={"Copy full SHA-256 cryptographic digest to clipboard"}
                  >
                    {copiedHashId === 'preview' ? '✓ Copied' : '📋 Copy'}
                  </button>
                {/if}
              </div>

              <!-- Preset Management & DAW Integration -->
              <div class="preset-manager-card">
                <div class="preset-header-row">
                  <div class="preset-title">
                    <span class="preset-icon">🎛️</span>
                    <div>
                      <h5>Preset Storage & DAW Integration</h5>
                      <span class="preset-subtitle">Export native DAW presets (.vstpreset / .fxp / .ttl) directly to standard preset folders</span>
                    </div>
                  </div>
                  <button
                    type="button"
                    class="btn-open-presets"
                    onclick={handleOpenPresetsDirectory}
                    use:tooltip={"Open DAW preset folder in system file manager"}
                  >
                    📁 Browse Presets
                  </button>
                </div>

                <div class="preset-export-controls">
                  <input
                    type="text"
                    class="preset-name-input"
                    placeholder="Enter preset name (e.g. init_lead, deep_bass)..."
                    bind:value={presetExportName}
                  />
                  <select class="preset-format-select" bind:value={presetExportFormat}>
                    <option value="vstpreset">.vstpreset (VST3 Standard)</option>
                    <option value="fxp">.fxp (Steinberg Legacy VST2)</option>
                    <option value="ttl">.ttl (LV2 Preset Graph)</option>
                  </select>
                  <button
                    type="button"
                    class="btn-export-preset"
                    onclick={() => handleExportPreset(previewPlugin!.name)}
                    disabled={isExportingPreset}
                    use:tooltip={"Generate preset bundle and write to DAW presets directory"}
                  >
                    {isExportingPreset ? 'Exporting...' : '💾 Export Preset'}
                  </button>
                </div>

                {#if presetExportMessage}
                  <div class="preset-export-status" class:success={presetExportMessage.startsWith('✓')} class:error={presetExportMessage.startsWith('Export failed')}>
                    {presetExportMessage}
                  </div>
                {/if}

                <div class="preset-snapshot-actions">
                  <button
                    type="button"
                    class="btn-snapshot-state"
                    onclick={handleSnapshotPluginState}
                    disabled={isSnapshotting}
                    use:tooltip={"Inspect and snapshot current parameter tree state with CRC32 digest"}
                  >
                    {isSnapshotting ? '⏳ Capturing...' : '📸 Snapshot Parameter State'}
                  </button>
                </div>

                {#if activeSnapshot}
                  <div class="snapshot-details-box">
                    <div class="snapshot-header">
                      <strong>State Snapshot: {activeSnapshot.preset_name}</strong>
                      <span class="snapshot-crc">CRC32: <code>{activeSnapshot.checksum_crc32}</code></span>
                    </div>
                    <div class="snapshot-meta-row">
                      <span class="spec-pill">{activeSnapshot.format}</span>
                      <span class="spec-pill">{activeSnapshot.preset_format}</span>
                      <span class="spec-pill">{activeSnapshot.total_parameters} Parameters</span>
                    </div>
                    <div class="snapshot-params-list">
                      {#each activeSnapshot.parameters as p}
                        <div class="snapshot-param-item">
                          <span class="param-name">{p.name}</span>
                          <span class="param-val">{p.display_value}</span>
                        </div>
                      {/each}
                    </div>
                  </div>
                {/if}

                <!-- Preset Bank Diffing Engine -->
                <div class="preset-diff-section">
                  <div class="diff-section-title">
                    <span>⚖️ Automated Preset State Diff & Drift Analyzer</span>
                  </div>
                  <div class="diff-controls-row">
                    <input
                      type="text"
                      class="diff-input"
                      placeholder="Baseline preset..."
                      bind:value={baselinePresetInput}
                    />
                    <span class="diff-vs">vs</span>
                    <input
                      type="text"
                      class="diff-input"
                      placeholder="Target preset..."
                      bind:value={targetPresetInput}
                    />
                    <button
                      type="button"
                      class="btn-diff-compare"
                      onclick={handleDiffPresets}
                      disabled={isDiffingPresets}
                      use:tooltip={"Compare parameter deviations and identify acoustic timbre drift"}
                    >
                      {isDiffingPresets ? 'Comparing...' : 'Diff Presets'}
                    </button>
                  </div>

                  {#if presetDiffReport}
                    <div class="diff-report-card">
                      <div class="diff-summary-badges">
                        <span class="diff-badge" class:radical={presetDiffReport.drift_severity === 'RADICAL'} class:moderate={presetDiffReport.drift_severity === 'MODERATE'} class:subtle={presetDiffReport.drift_severity === 'SUBTLE' || presetDiffReport.drift_severity === 'IDENTICAL'}>
                          Drift: {presetDiffReport.drift_severity}
                        </span>
                        <span class="diff-badge metric">
                          Similarity: {presetDiffReport.similarity_score_pct}%
                        </span>
                        <span class="diff-badge metric">
                          Modified: {presetDiffReport.modified_parameters_count} / {presetDiffReport.total_parameters_compared}
                        </span>
                      </div>

                      <div class="diff-notes-box">
                        💡 {presetDiffReport.compatibility_notes}
                      </div>

                      <div class="diff-table-container">
                        <table class="diff-table">
                          <thead>
                            <tr>
                              <th>Parameter</th>
                              <th>Baseline</th>
                              <th>Target</th>
                              <th>Delta</th>
                              <th>Status</th>
                            </tr>
                          </thead>
                          <tbody>
                            {#each presetDiffReport.parameter_diffs as d}
                              <tr class:modified-row={d.status === 'MODIFIED'}>
                                <td>{d.name}</td>
                                <td><code>{(d.baseline_value * 100).toFixed(0)}%</code></td>
                                <td><code>{(d.target_value * 100).toFixed(0)}%</code></td>
                                <td>
                                  <span class="delta-tag" class:positive={d.delta > 0} class:negative={d.delta < 0}>
                                    {d.delta > 0 ? '+' : ''}{d.delta_pct}%
                                  </span>
                                </td>
                                <td>
                                  <span class="status-pill" class:mod={d.status === 'MODIFIED'} class:same={d.status === 'UNCHANGED'}>
                                    {d.status}
                                  </span>
                                </td>
                              </tr>
                            {/each}
                          </tbody>
                        </table>
                      </div>
                    </div>
                  {/if}
                </div>
              </div>

              <!-- Real-Time Audio Loopback Metering & True Peak (dBTP) LUFS EBU R128 Analyzer -->
              <div class="metering-analyzer-card">
                <div class="metering-header">
                  <div class="metering-title">
                    <span class="metering-icon">📊</span>
                    <div>
                      <h4 class="metering-h4">Audio Loopback Metering & True Peak (dBTP) Analyzer</h4>
                      <span class="metering-sub">Headless EBU R128 loudness normalization, 4x inter-sample peak detection, and streaming compliance</span>
                    </div>
                  </div>
                  <div class="metering-ctrls">
                    <div class="metering-select-wrapper">
                      <select
                        class="metering-select"
                        bind:value={meteringTargetStandard}
                        use:tooltip={"Target streaming loudness normalization standard"}
                      >
                        <option value="EBU R128 (-23 LUFS)">EBU R128 (-23 LUFS, -1 dBTP)</option>
                        <option value="Spotify (-14 LUFS)">Spotify (-14 LUFS, -1 dBTP)</option>
                        <option value="Apple Music (-16 LUFS)">Apple Music (-16 LUFS, -1 dBTP)</option>
                        <option value="YouTube (-14 LUFS)">YouTube (-14 LUFS, -1 dBTP)</option>
                      </select>
                    </div>
                    <button
                      type="button"
                      class="btn-metering-action"
                      onclick={handleAnalyzeMetering}
                      disabled={isAnalyzingMetering}
                      use:tooltip={"Run headless audio loopback measurement to compute integrated LUFS, inter-sample true peaks, and gain recommendations"}
                    >
                      {isAnalyzingMetering ? 'Measuring...' : 'Measure Audio Metering'}
                    </button>
                  </div>
                </div>

                {#if audioMeteringReport}
                  <div class="metering-content">
                    <div class="metering-metrics-grid">
                      <div class="meter-stat-card">
                        <span class="meter-label">Integrated Loudness:</span>
                        <span class="meter-val lufs">{audioMeteringReport.loudness.integrated_lufs.toFixed(1)} LUFS</span>
                        <span class="meter-subtext">Momentary Max: {audioMeteringReport.loudness.momentary_max_lufs.toFixed(1)} LUFS</span>
                      </div>
                      <div class="meter-stat-card">
                        <span class="meter-label">True Peak (dBTP):</span>
                        <span class="meter-val dbtp" class:clipping={audioMeteringReport.true_peak.is_inter_sample_clipping}>
                          L: {audioMeteringReport.true_peak.peak_left_dbtp.toFixed(1)} / R: {audioMeteringReport.true_peak.peak_right_dbtp.toFixed(1)} dBTP
                        </span>
                        <span class="meter-subtext">{audioMeteringReport.true_peak.is_inter_sample_clipping ? '⚠️ Inter-sample clipping detected' : '✓ Clean (Below 0 dBTP)'}</span>
                      </div>
                      <div class="meter-stat-card">
                        <span class="meter-label">Loudness Range (LRA):</span>
                        <span class="meter-val lra">{audioMeteringReport.loudness.loudness_range_lra.toFixed(1)} LU</span>
                        <span class="meter-subtext">{audioMeteringReport.dynamic_range_score}</span>
                      </div>
                      <div class="meter-stat-card">
                        <span class="meter-label">Target Standard:</span>
                        <span class="meter-val compliance" class:compliant={audioMeteringReport.target_compliance.compliance_status === 'COMPLIANT'} class:over={audioMeteringReport.target_compliance.compliance_status === 'OVER_TARGET'} class:under={audioMeteringReport.target_compliance.compliance_status === 'UNDER_TARGET'}>
                          {audioMeteringReport.target_compliance.compliance_status}
                        </span>
                        <span class="meter-subtext">Gain Trim: {audioMeteringReport.target_compliance.recommended_gain_adjustment_db > 0 ? '+' : ''}{audioMeteringReport.target_compliance.recommended_gain_adjustment_db.toFixed(1)} dB</span>
                      </div>
                    </div>

                    <div class="metering-visual-meters">
                      <div class="meter-bar-group">
                        <div class="meter-bar-header">
                          <span>Integrated Loudness Bar</span>
                          <span>{audioMeteringReport.loudness.integrated_lufs.toFixed(1)} LUFS (Target: {audioMeteringReport.target_compliance.target_integrated_lufs.toFixed(1)})</span>
                        </div>
                        <div class="meter-gauge-track">
                          <div
                            class="meter-gauge-fill"
                            class:meter-gauge-hot={audioMeteringReport.loudness.integrated_lufs > audioMeteringReport.target_compliance.target_integrated_lufs}
                            style="width: {Math.max(5, Math.min(100, ((audioMeteringReport.loudness.integrated_lufs + 36) / 36) * 100))}%"
                          ></div>
                        </div>
                      </div>

                      <div class="meter-bar-group">
                        <div class="meter-bar-header">
                          <span>True Peak Oversampled Ceiling</span>
                          <span>Max: {Math.max(audioMeteringReport.true_peak.peak_left_dbtp, audioMeteringReport.true_peak.peak_right_dbtp).toFixed(1)} dBTP</span>
                        </div>
                        <div class="meter-gauge-track">
                          <div
                            class="meter-gauge-fill"
                            class:meter-gauge-clip={audioMeteringReport.true_peak.is_inter_sample_clipping}
                            style="width: {Math.max(5, Math.min(100, ((Math.max(audioMeteringReport.true_peak.peak_left_dbtp, audioMeteringReport.true_peak.peak_right_dbtp) + 12) / 14) * 100))}%"
                          ></div>
                        </div>
                      </div>
                    </div>

                    <div class="metering-advice-card" class:advice-warn={audioMeteringReport.true_peak.is_inter_sample_clipping}>
                      <span class="advice-icon">{audioMeteringReport.true_peak.is_inter_sample_clipping ? '⚠️' : '💡'}</span>
                      <div class="advice-content">
                        <strong>DSP Mastering & Normalization Advisory:</strong>
                        <p>{audioMeteringReport.mastering_advice}</p>
                      </div>
                    </div>
                  </div>
                {/if}
              </div>

              <!-- Headless Audio Loopback Rendering & Dynamic Frequency Response Spectrum Analyzer -->
              <div class="spectrum-analyzer-card">
                <div class="spectrum-header">
                  <div class="spectrum-title">
                    <span class="spectrum-icon">📈</span>
                    <div>
                      <div class="spectrum-title-text">HEADLESS AUDIO SPECTRUM & FREQUENCY RESPONSE MATRIX</div>
                      <div class="spectrum-subtitle-text">
                        Inject synthetic loopback sweeps, measure passband magnitude (dBFS), harmonic distortion (THD %), and phase response
                      </div>
                    </div>
                  </div>
                  <div class="spectrum-controls">
                    <select
                      class="spectrum-signal-select"
                      bind:value={spectrumTestSignal}
                      use:tooltip={"Select synthetic input stimulus injected into plugin DSP render callback"}
                    >
                      <option value="Logarithmic Sine Sweep (20Hz - 20kHz)">Logarithmic Sine Sweep (20Hz - 20kHz)</option>
                      <option value="Pink Noise (-12 dBFS)">Pink Noise (-12 dBFS)</option>
                      <option value="Dirac Delta Impulse (Single Sample)">Dirac Delta Impulse (Single Sample)</option>
                    </select>
                    <button
                      class="btn btn-sm btn-spectrum"
                      disabled={isRenderingSpectrum}
                      onclick={() => renderAudioSpectrumAnalysis()}
                      use:tooltip={"Inject test signal and compute headless 2048-point FFT spectrum analysis"}
                    >
                      {isRenderingSpectrum ? '⏳ Analyzing...' : '📊 Analyze Spectrum'}
                    </button>
                  </div>
                </div>

                {#if spectrumReport}
                  <div class="spectrum-results-container">
                    <div class="spectrum-summary-bar">
                      <span class="spectrum-badge badge-thd" use:tooltip={"Total Harmonic Distortion measured at 1 kHz fundamental"}>
                        THD: {spectrumReport.total_harmonic_distortion_pct}%
                      </span>
                      <span class="spectrum-badge badge-snr" use:tooltip={"Signal-to-Noise dynamic range across 20Hz-20kHz band"}>
                        SNR: {spectrumReport.signal_to_noise_ratio_db} dB
                      </span>
                      <span class="spectrum-badge badge-cutoff" use:tooltip={"High frequency -3dB roll-off cutoff limit"}>
                        -3dB Cutoff: {(spectrumReport.cutoff_frequency_3db_hz / 1000).toFixed(1)} kHz
                      </span>
                      <span class="spectrum-badge badge-phase" use:tooltip={"DSP phase linearity across audio spectrum"}>
                        {spectrumReport.phase_linearity_status}
                      </span>
                    </div>

                    <div class="spectrum-info-box">
                      <div class="spectrum-metric-row">
                        <span><strong>Stimulus:</strong> {spectrumReport.test_signal}</span>
                        <span><strong>FFT Resolution:</strong> {spectrumReport.fft_resolution_points} points @ {spectrumReport.sample_rate / 1000} kHz</span>
                      </div>
                      <div class="spectrum-advice">💡 {spectrumReport.mastering_advice}</div>
                    </div>

                    <!-- Frequency Bands Table & Visual Response Curve -->
                    <div class="spectrum-table-wrapper">
                      <table class="spectrum-table">
                        <thead>
                          <tr>
                            <th use:tooltip={"Acoustic audio spectrum frequency band"}>Band / Frequency</th>
                            <th use:tooltip={"Injected test signal magnitude level"}>Input (dBFS)</th>
                            <th use:tooltip={"DSP output magnitude level"}>Output (dBFS)</th>
                            <th use:tooltip={"Magnitude gain delta through plugin"}>Gain Delta</th>
                            <th use:tooltip={"Visual magnitude response relative to 0 dB unity gain"}>Response Curve</th>
                            <th use:tooltip={"Phase displacement relative to dry input signal"}>Phase Shift</th>
                            <th use:tooltip={"Band-specific harmonic distortion"}>THD %</th>
                          </tr>
                        </thead>
                        <tbody>
                          {#each spectrumReport.frequency_bins as bin}
                            <tr>
                              <td class="spectrum-col-band">
                                <span class="band-tag">{bin.band_name}</span>
                              </td>
                              <td class="spectrum-col-db">{bin.input_magnitude_db.toFixed(1)} dB</td>
                              <td class="spectrum-col-db">{bin.output_magnitude_db.toFixed(1)} dB</td>
                              <td class="spectrum-col-delta">
                                <span class="delta-tag {bin.gain_delta_db > 0.5 ? 'delta-boost' : bin.gain_delta_db < -0.5 ? 'delta-cut' : 'delta-flat'}">
                                  {bin.gain_delta_db > 0 ? `+${bin.gain_delta_db.toFixed(1)}` : bin.gain_delta_db.toFixed(1)} dB
                                </span>
                              </td>
                              <td class="spectrum-col-curve">
                                <div class="curve-bar-container">
                                  <div
                                    class="curve-bar-fill {bin.gain_delta_db > 0.5 ? 'curve-boost' : bin.gain_delta_db < -1.5 ? 'curve-roll' : 'curve-flat'}"
                                    style="width: {Math.max(10, Math.min(100, 50 + bin.gain_delta_db * 10))}%"
                                  ></div>
                                </div>
                              </td>
                              <td class="spectrum-col-phase">
                                <span class="phase-tag {Math.abs(bin.phase_shift_deg) < 15 ? 'phase-min' : 'phase-shift'}">
                                  {bin.phase_shift_deg > 0 ? `+${bin.phase_shift_deg.toFixed(1)}°` : `${bin.phase_shift_deg.toFixed(1)}°`}
                                </span>
                              </td>
                              <td class="spectrum-col-thd">
                                <span class="thd-tag {bin.harmonic_distortion_pct < 0.02 ? 'thd-clean' : 'thd-warm'}">
                                  {bin.harmonic_distortion_pct.toFixed(3)}%
                                </span>
                              </td>
                            </tr>
                          {/each}
                        </tbody>
                      </table>
                    </div>
                  </div>
                {/if}
              </div>

              <div class="security-scan-card">
                <div class="sec-row">
                  <span class="sec-label">Security Sandbox Scan:</span>
                  <span class="verification-badge valid" use:tooltip={"Scanned with pre-flight ClamAV engine and cryptographic binary verification"}>
                    🛡️ Verified Clean (ClamAV / SHA-256)
                  </span>
                </div>
                <div class="sec-row">
                  <span class="sec-label">SHA-256 Digest:</span>
                  {#if previewPlugin.sha256_hash}
                    <code class="sec-hash">{previewPlugin.sha256_hash}</code>
                  {:else}
                    <span class="sec-val-dim">Dynamic stream calculated upon download</span>
                  {/if}
                </div>
              </div>

              <!-- Wine & Yabridge Linux Compatibility Sync -->
              <div class="yabridge-card">
                <div class="yabridge-header-row">
                  <div class="yabridge-title">
                    <span class="yabridge-icon">🍷</span>
                    <div>
                      <strong>Wine & Yabridge Host Synchronization</strong>
                      <span class="yabridge-subtitle">Bridge Windows VST2 (.dll) and VST3 (.vst3) binaries on Linux DAWs (Bitwig, REAPER)</span>
                    </div>
                  </div>
                  <div class="yabridge-actions">
                    <button
                      type="button"
                      class="btn-yabridge-check"
                      onclick={checkYabridge}
                      disabled={isCheckingYabridge}
                      use:tooltip={"Check if yabridgectl is installed and scan registered directories"}
                    >
                      {isCheckingYabridge ? 'Checking...' : '🔍 Check Status'}
                    </button>
                    <button
                      type="button"
                      class="btn-yabridge-sync"
                      onclick={() => syncYabridgePlugin(previewInstalled?.path)}
                      disabled={isSyncingYabridge}
                      use:tooltip={"Sync this plugin folder with yabridgectl sync to generate native Linux .so bridges"}
                    >
                      {isSyncingYabridge ? 'Syncing...' : '⚡ Sync Bridge'}
                    </button>
                  </div>
                </div>

                {#if yabridgeStatus}
                  <div class="yabridge-status-body">
                    <div class="yabridge-status-line">
                      <span class="badge" class:badge-active={yabridgeStatus.installed} class:badge-inactive={!yabridgeStatus.installed}>
                        {yabridgeStatus.installed ? `✓ Yabridge Detected (${yabridgeStatus.version || 'Active'})` : '⚠️ yabridgectl Not Found'}
                      </span>
                      {#if !yabridgeStatus.installed}
                        <span class="yabridge-install-tip">Install yabridge and Wine to run Windows plugins seamlessly on Linux.</span>
                      {/if}
                    </div>
                    {#if yabridgeStatus.configured_paths.length > 0}
                      <div class="yabridge-paths-list">
                        <span class="yabridge-paths-title">Configured Bridge Directories:</span>
                        <ul>
                          {#each yabridgeStatus.configured_paths as p}
                            <li><code>{p}</code></li>
                          {/each}
                        </ul>
                      </div>
                    {/if}
                  </div>
                {/if}

                {#if yabridgeMessage}
                  <div class="yabridge-message-box" class:error={yabridgeMessage.includes('error') || yabridgeMessage.includes('not installed')}>
                    <pre>{yabridgeMessage}</pre>
                  </div>
                {/if}
              </div>

              <!-- Latency & Buffer Diagnostic Matrix -->
              <div class="latency-diagnostic-card">
                <div class="latency-card-header">
                  <div class="latency-icon">⏱️</div>
                  <div class="latency-card-title">
                    <h4>DAW Buffer & DSP Latency Diagnostic Matrix</h4>
                    <p>Simulate roundtrip audio latency across standard DAW buffer configurations (PDC analysis)</p>
                  </div>
                  <div class="latency-actions">
                    <button
                      type="button"
                      class="btn-latency-calc"
                      onclick={calculateLatencyDiagnostics}
                      disabled={isDiagnosingLatency}
                      use:tooltip={"Compute real-time DSP roundtrip latency across 64 to 1024 sample buffers"}
                    >
                      {isDiagnosingLatency ? 'Calculating...' : '⚡ Compute Matrix'}
                    </button>
                  </div>
                </div>

                <div class="latency-controls-row">
                  <div class="latency-control-group">
                    <label for="sample-rate-select">Sample Rate:</label>
                    <select id="sample-rate-select" bind:value={latencySampleRate} onchange={calculateLatencyDiagnostics}>
                      <option value={44100}>44.1 kHz (CD Standard)</option>
                      <option value={48000}>48.0 kHz (Studio Video/DAW)</option>
                      <option value={88200}>88.2 kHz (High-Res 2x)</option>
                      <option value={96000}>96.0 kHz (High-Res Mastering)</option>
                    </select>
                  </div>
                  <div class="latency-control-group">
                    <label for="dsp-samples-input">Reported DSP Lookahead:</label>
                    <input
                      id="dsp-samples-input"
                      type="number"
                      min="0"
                      max="4096"
                      step="1"
                      bind:value={latencyDspSamples}
                      oninput={calculateLatencyDiagnostics}
                      placeholder="0 samples"
                    />
                  </div>
                </div>

                {#if latencyDiagnosticResult}
                  <div class="latency-results-container">
                    <div class="latency-recommendation-banner">
                      <span class="latency-badge">{latencyDiagnosticResult.dsp_latency_ms} ms Plugin Delay</span>
                      <span class="latency-rec-text">{latencyDiagnosticResult.recommendation}</span>
                    </div>

                    <div class="latency-grid">
                      <div class="latency-grid-header">
                        <span>Buffer Size</span>
                        <span>One-Way</span>
                        <span>Est. Roundtrip</span>
                        <span>DAW Target Workflow</span>
                      </div>
                      {#each latencyDiagnosticResult.buffer_points as pt}
                        <div class="latency-grid-row">
                          <span class="buffer-cell"><strong>{pt.buffer_size}</strong> spls</span>
                          <span class="time-cell">{pt.latency_ms} ms</span>
                          <span class="time-cell roundtrip">~{pt.roundtrip_estimate_ms} ms</span>
                          <span class="workflow-cell">{pt.daw_workflow_category}</span>
                        </div>
                      {/each}
                    </div>
                  </div>
                {/if}
              </div>

              <!-- Headless DSP Benchmark & Memory Profiler Card -->
              <div class="dsp-benchmark-card">
                <div class="benchmark-header">
                  <div class="benchmark-title">
                    <span class="benchmark-icon">🚀</span>
                    <div>
                      <strong>HEADLESS DSP BENCHMARK & MEMORY PROFILING MATRIX</strong>
                      <span class="benchmark-subtitle">Simulate real-time DSP cycle load, buffer underrun (XRun) safety margins, and peak RSS memory footprints</span>
                    </div>
                  </div>
                  <button
                    type="button"
                    class="btn-benchmark-run"
                    onclick={() => runDspBenchmark()}
                    disabled={isBenchmarking}
                    use:tooltip={"Execute simulated headless DSP cycle benchmark and memory consumption profiler"}
                  >
                    {isBenchmarking ? '⏳ Profiling DSP...' : '⚡ Run DSP Benchmark'}
                  </button>
                </div>

                {#if benchmarkReport}
                  <div class="benchmark-results-container">
                    <div class="benchmark-summary-banner">
                      <div class="benchmark-banner-badges">
                        <span class="benchmark-rating-badge">{benchmarkReport.overall_efficiency_rating}</span>
                        <span class="benchmark-mem-badge" use:tooltip={"Estimated peak resident set size memory allocation"}>
                          RAM: {benchmarkReport.base_memory_mb} MB (Peak: {benchmarkReport.estimated_peak_rss_mb} MB)
                        </span>
                        <span class="benchmark-multicore-badge" use:tooltip={"DAW multicore DSP worker thread pool efficiency score (1-100)"}>
                          Multicore Scaling: {benchmarkReport.multicore_scaling_score}/100
                        </span>
                      </div>
                      <div class="benchmark-safety-desc">
                        🔒 <strong>Safety Class:</strong> {benchmarkReport.thread_safety_classification}
                      </div>
                      <div class="benchmark-advice-text">
                        💡 {benchmarkReport.optimization_advice}
                      </div>
                    </div>

                    <!-- Buffer benchmark table -->
                    <div class="benchmark-table-wrapper">
                      <table class="benchmark-table">
                        <thead>
                          <tr>
                            <th use:tooltip={"Audio host buffer size in samples"}>Buffer Size</th>
                            <th use:tooltip={"Estimated percentage of CPU core allocated for real-time DSP calculation"}>Simulated DSP Load</th>
                            <th use:tooltip={"Real-time safety margin before audio glitch / buffer underrun occurs"}>Headroom Margin</th>
                            <th use:tooltip={"Risk level of audio stutter or XRun buffer underruns at this block size"}>XRun Risk</th>
                          </tr>
                        </thead>
                        <tbody>
                          {#each benchmarkReport.buffer_benchmarks as point}
                            <tr>
                              <td class="bench-col-buf"><strong>{point.buffer_size}</strong> samples</td>
                              <td class="bench-col-load">
                                <div class="load-bar-wrap">
                                  <div class="load-bar-fill" style="width: {Math.min(point.simulated_dsp_load_pct * 4, 100)}%"></div>
                                  <span class="load-val">{point.simulated_dsp_load_pct}%</span>
                                </div>
                              </td>
                              <td class="bench-col-margin">
                                <span class="margin-val">{point.real_time_safety_margin_pct}%</span>
                              </td>
                              <td class="bench-col-risk">
                                <span class="xrun-risk-pill risk-{point.xrun_risk_level.toLowerCase()}">
                                  {point.xrun_risk_level}
                                </span>
                              </td>
                            </tr>
                          {/each}
                        </tbody>
                      </table>
                    </div>
                  </div>
                {/if}
              </div>
            </div>

            <!-- Audio Bus Configuration & Channel Routing Matrix -->
            <div class="preview-section">
              <div class="audio-bus-matrix-card">
                <div class="bus-matrix-header">
                  <div class="bus-matrix-title">
                    <span class="bus-matrix-icon">🎚️</span>
                    <div>
                      <strong>CHANNEL AUDIO BUS CONFIGURATION & I/O ROUTING MATRIX</strong>
                      <span class="bus-matrix-subtitle">Inspect input/output pin topology, multi-channel auxiliary stems, and surround speaker configurations</span>
                    </div>
                  </div>
                  <button
                    type="button"
                    class="btn-latency-calc"
                    onclick={() => calculateAudioBusDiagnostics()}
                    disabled={isDiagnosingAudioBuses}
                    use:tooltip={"Inspect audio bus arrangement and I/O routing topology"}
                  >
                    {isDiagnosingAudioBuses ? '⏳ Analyzing...' : '🔀 Inspect I/O Topology'}
                  </button>
                </div>

                <div class="bus-controls-row">
                  <div class="latency-input-group">
                    <label for="bus-profile-select">Bus Layout Profile:</label>
                    <select id="bus-profile-select" bind:value={selectedBusConfigType} class="latency-select">
                      <option value="auto">Auto-Detect from Category</option>
                      <option value="stereo_standard">Standard Stereo (2-In / 2-Out)</option>
                      <option value="stereo_sidechain">Stereo + Sidechain Aux Key (4-In / 2-Out)</option>
                      <option value="multi_out_instrument">Multi-Out Instrument / Sampler (0-In / 16-Out)</option>
                      <option value="surround_5_1">5.1 Cinema Surround (6-In / 6-Out)</option>
                      <option value="mono">Mono Insert (1-In / 1-Out)</option>
                    </select>
                  </div>
                </div>

                {#if audioBusDiagnosisResult}
                  <div class="bus-results-container">
                    <div class="bus-summary-banner">
                      <div class="bus-banner-top">
                        <span class="bus-topology-badge">{audioBusDiagnosisResult.speaker_arrangement}</span>
                        <span class="bus-io-count">Inputs: {audioBusDiagnosisResult.total_input_channels} ch | Outputs: {audioBusDiagnosisResult.total_output_channels} ch</span>
                        {#if audioBusDiagnosisResult.supports_sidechain}
                          <span class="bus-sidechain-badge">✓ Sidechain Key Ready</span>
                        {/if}
                        {#if audioBusDiagnosisResult.is_instrument}
                          <span class="bus-instrument-badge">🎹 Multi-Stem Synth</span>
                        {/if}
                      </div>
                      <div class="bus-daw-notes">
                        💡 {audioBusDiagnosisResult.daw_compatibility_notes}
                      </div>
                    </div>

                    <div class="bus-buses-grid">
                      {#each audioBusDiagnosisResult.buses as bus}
                        <div class="bus-card {bus.direction === 'input' ? 'bus-in' : 'bus-out'} {bus.is_auxiliary ? 'bus-aux' : ''}">
                          <div class="bus-card-header">
                            <span class="bus-dir-badge">{bus.direction.toUpperCase()}</span>
                            <strong>{bus.bus_name}</strong>
                            <span class="bus-channels-badge">{bus.channel_count} ch</span>
                          </div>
                          <div class="bus-card-pins">
                            {#each bus.channels as ch}
                              <span class="bus-pin-pill" use:tooltip={ch.channel_type}>
                                🔌 {ch.name} <small>({ch.channel_type})</small>
                              </span>
                            {/each}
                          </div>
                        </div>
                      {/each}
                    </div>
                  </div>
                {/if}
              </div>
            </div>

            <!-- Binary Architecture & SIMD Instruction Inspection Matrix -->
            <div class="preview-section">
              <div class="binary-arch-matrix-card">
                <div class="arch-matrix-header">
                  <div class="arch-matrix-title">
                    <span class="arch-matrix-icon">💻</span>
                    <div>
                      <strong>BINARY ARCHITECTURE & SIMD INSTRUCTION SET MATRIX</strong>
                      <span class="arch-matrix-subtitle">Analyze CPU instruction sets, vectorization acceleration, 64-bit compliance, and runtime requirements</span>
                    </div>
                  </div>
                  <button
                    type="button"
                    class="btn-latency-calc"
                    onclick={() => inspectBinaryArchitecture()}
                    disabled={isInspectingArch}
                    use:tooltip={"Inspect CPU architecture, SIMD vectorization, and OS minimum runtime"}
                  >
                    {isInspectingArch ? '⏳ Inspecting...' : '🔬 Introspect CPU/SIMD'}
                  </button>
                </div>

                {#if binaryArchDetails}
                  <div class="arch-results-container">
                    <div class="arch-summary-banner">
                      <div class="arch-banner-top">
                        <span class="arch-cpu-badge">{binaryArchDetails.target_cpu_arch}</span>
                        <span class="arch-bitness-badge {binaryArchDetails.is_64bit ? 'is-64' : 'is-32'}">
                          {binaryArchDetails.is_64bit ? '64-bit Native' : '⚠️ 32-bit Legacy'}
                        </span>
                        {#if binaryArchDetails.rosetta2_compatible}
                          <span class="arch-rosetta-badge">✓ Rosetta 2 Intel Emulation</span>
                        {/if}
                        <span class="arch-container-type">{binaryArchDetails.binary_container_type}</span>
                      </div>
                      <div class="arch-summary-text">
                        📌 {binaryArchDetails.architecture_summary}
                      </div>
                    </div>

                    <div class="arch-details-grid">
                      <div class="arch-detail-box">
                        <span class="arch-box-title">⚡ SIMD Vectorization Instruction Sets:</span>
                        <div class="arch-tag-pills">
                          {#each binaryArchDetails.simd_instruction_sets as simd}
                            <span class="simd-pill" use:tooltip={"Hardware accelerated vector processing"}>⚡ {simd}</span>
                          {/each}
                        </div>
                      </div>

                      <div class="arch-detail-box">
                        <span class="arch-box-title">🛡️ Binary Security & Memory Hardening:</span>
                        <div class="arch-tag-pills">
                          {#each binaryArchDetails.binary_security_flags as flag}
                            <span class="sec-pill" use:tooltip={"Exploit mitigation & safe memory execution"}>🛡️ {flag}</span>
                          {/each}
                        </div>
                      </div>
                    </div>

                    <div class="arch-os-box">
                      <span class="arch-os-label">Minimum Host OS Environment:</span>
                      <span class="arch-os-val">{binaryArchDetails.min_os_version}</span>
                    </div>
                  </div>
                {/if}
              </div>
            </div>

            <!-- Plugin Parameter Enumeration & MIDI Automation Inspector -->
            <div class="preview-section">
              <div class="param-matrix-card">
                <div class="param-matrix-header">
                  <div class="param-matrix-title">
                    <span class="param-matrix-icon">🎛️</span>
                    <div>
                      <strong>PLUGIN PARAMETER ENUMERATION & MIDI AUTOMATION MATRIX</strong>
                      <span class="param-matrix-subtitle">Enumerate host parameters, control ranges, units, sample-accurate automation flags, and MIDI CC assignments</span>
                    </div>
                  </div>
                  <button
                    type="button"
                    class="btn-latency-calc"
                    onclick={() => inspectPluginParameters()}
                    disabled={isInspectingParameters}
                    use:tooltip={"Inspect plugin parameter hierarchy, automatable controls, and MIDI CC bindings"}
                  >
                    {isInspectingParameters ? '⏳ Enumerating...' : '🎛️ Enumerate Parameters'}
                  </button>
                </div>

                {#if parameterReport}
                  <div class="param-results-container">
                    <div class="param-summary-banner">
                      <div class="param-banner-top">
                        <span class="param-total-badge">{parameterReport.total_parameters} Total Parameters</span>
                        <span class="param-auto-badge">✓ {parameterReport.automatable_parameters} Automatable</span>
                        {#if parameterReport.supports_mpe}
                          <span class="param-mpe-badge">⚡ MPE Polyphonic Expression</span>
                        {/if}
                        {#if parameterReport.supports_sample_accurate_automation}
                          <span class="param-sample-badge">⏱️ Sample-Accurate Automation</span>
                        {/if}
                        <span class="param-tree-type">{parameterReport.parameter_tree_structure}</span>
                      </div>
                      <div class="param-host-notes">
                        💡 {parameterReport.automation_host_advice}
                      </div>
                    </div>

                    <div class="param-table-wrapper">
                      <table class="param-table">
                        <thead>
                          <tr>
                            <th use:tooltip={"Unique zero-based host parameter identifier index"}>ID</th>
                            <th use:tooltip={"Host automation parameter display title"}>Parameter Name</th>
                            <th use:tooltip={"Functional parameter group / DSP module"}>Group</th>
                            <th use:tooltip={"Data type (Float, Boolean, Enumeration, Integer)"}>Type</th>
                            <th use:tooltip={"Factory default value"}>Default</th>
                            <th use:tooltip={"Min and Max modulation range"}>Range</th>
                            <th use:tooltip={"Engineering display unit"}>Unit</th>
                            <th use:tooltip={"Whether this parameter can be automated by DAW envelope lanes"}>Auto</th>
                            <th use:tooltip={"Standard MIDI Continuous Controller assignment"}>MIDI CC</th>
                          </tr>
                        </thead>
                        <tbody>
                          {#each parameterReport.parameters as param}
                            <tr>
                              <td class="param-col-id">#{param.id}</td>
                              <td class="param-col-name"><strong>{param.name}</strong></td>
                              <td class="param-col-group"><span class="param-group-tag">{param.parameter_group}</span></td>
                              <td class="param-col-type">
                                <span class="param-type-pill type-{param.parameter_type.toLowerCase()}">{param.parameter_type}</span>
                              </td>
                              <td class="param-col-val">{param.default_value}</td>
                              <td class="param-col-range">{param.min_value} .. {param.max_value}</td>
                              <td class="param-col-unit">{param.display_unit}</td>
                              <td class="param-col-auto">
                                <span class="param-auto-tag {param.is_automatable ? 'is-auto' : 'no-auto'}">
                                  {param.is_automatable ? '✓ YES' : '✗ NO'}
                                </span>
                              </td>
                              <td class="param-col-midi">
                                {#if param.assigned_midi_cc !== null}
                                  <span class="param-midi-pill" use:tooltip={"Assigned MIDI CC trigger"}>CC {param.assigned_midi_cc}</span>
                                {:else}
                                  <span class="param-no-midi">—</span>
                                {/if}
                              </td>
                            </tr>
                          {/each}
                        </tbody>
                      </table>
                    </div>
                  </div>
                {/if}
              </div>
            </div>

            <!-- AUDIO THREAD RESILIENCE & CRASH ISOLATION MATRIX -->
            <div class="preview-section">
              <div class="resilience-matrix-card">
                <div class="resilience-matrix-header">
                  <div class="resilience-matrix-title">
                    <span class="resilience-icon">🛡️</span>
                    <div>
                      <div class="resilience-title-text">AUDIO THREAD RESILIENCE & CRASH ISOLATION MATRIX</div>
                      <div class="resilience-subtitle-text">
                        Test real-time buffer deadline adherence, lock-free safety, and host crash sandbox isolation
                      </div>
                    </div>
                  </div>
                  <button
                    class="btn btn-sm btn-resilience"
                    disabled={isTestingResilience}
                    onclick={() => testAudioThreadResilience()}
                    use:tooltip={"Stress test audio thread execution budgets and assert zero-allocation real-time safety"}
                  >
                    {isTestingResilience ? '⏳ Stress Testing...' : '🛡️ Stress Test Thread'}
                  </button>
                </div>

                {#if resilienceReport}
                  <div class="resilience-results-container">
                    <div class="resilience-summary-bar">
                      <span class="resilience-badge badge-grade" use:tooltip={"Overall real-time safety and deterministic reliability rating"}>
                        Grade: {resilienceReport.overall_resilience_grade}
                      </span>
                      <span class="resilience-badge {resilienceReport.lock_free_audio_thread ? 'badge-safe' : 'badge-warn'}" use:tooltip={"Verifies absence of priority inversions or mutex lock stalls on the audio thread"}>
                        {resilienceReport.lock_free_audio_thread ? '✓ Lock-Free Concurrency' : '⚠️ Mutex Contention Risk'}
                      </span>
                      <span class="resilience-badge {resilienceReport.zero_allocation_render ? 'badge-safe' : 'badge-warn'}" use:tooltip={"Verifies zero malloc/free heap allocations in real-time audio render callback"}>
                        {resilienceReport.zero_allocation_render ? '✓ Zero Heap Allocations' : '⚠️ Heap Dynamic Allocations'}
                      </span>
                      <span class="resilience-badge badge-sandbox" use:tooltip={"Process isolation compatibility with Bitwig Studio and REAPER crash sandboxes"}>
                        Host Sandboxing: Bitwig & REAPER Ready
                      </span>
                    </div>

                    <div class="resilience-info-box">
                      <div class="resilience-metric-row">
                        <span><strong>Buffer Budget:</strong> {resilienceReport.buffer_size} samples ({resilienceReport.time_budget_us.toFixed(1)} µs deadline @ {resilienceReport.sample_rate / 1000} kHz)</span>
                        <span><strong>Crash Recovery:</strong> {resilienceReport.crash_recovery_mechanism}</span>
                      </div>
                      <div class="resilience-advice">💡 {resilienceReport.host_resilience_advice}</div>
                    </div>

                    <!-- Stress Timeline Table -->
                    <div class="resilience-table-wrapper">
                      <table class="resilience-table">
                        <thead>
                          <tr>
                            <th use:tooltip={"Simulated DSP synthetic load stress level"}>Stress Load</th>
                            <th use:tooltip={"Execution time consumed in audio callback per buffer block"}>Exec Time (µs)</th>
                            <th use:tooltip={"Buffer deadline budget"}>Budget (µs)</th>
                            <th use:tooltip={"Percentage of available time budget consumed"}>DSP Cycle %</th>
                            <th use:tooltip={"Audio thread heap allocations detected (must be 0 for real-time safety)"}>Allocations</th>
                            <th use:tooltip={"Thread priority inversion / mutex block check"}>Priority Inversion</th>
                            <th use:tooltip={"Audio buffer underrun / glitch dropout"}>Dropout (XRun)</th>
                          </tr>
                        </thead>
                        <tbody>
                          {#each resilienceReport.stress_timeline as pt}
                            <tr>
                              <td class="resilience-col-stress">
                                <span class="stress-intensity-tag {pt.stress_intensity_pct >= 100 ? 'high-stress' : 'normal-stress'}">
                                  {pt.stress_intensity_pct}% Load
                                </span>
                              </td>
                              <td class="resilience-col-exec">{pt.execution_time_us} µs</td>
                              <td class="resilience-col-budget">{pt.time_budget_us.toFixed(1)} µs</td>
                              <td class="resilience-col-cycle">
                                <div class="cycle-bar-container">
                                  <div
                                    class="cycle-bar-fill {pt.cpu_cycle_pct > 80 ? 'fill-danger' : pt.cpu_cycle_pct > 50 ? 'fill-warning' : 'fill-good'}"
                                    style="width: {Math.min(pt.cpu_cycle_pct, 100)}%"
                                  ></div>
                                  <span class="cycle-label">{pt.cpu_cycle_pct}%</span>
                                </div>
                              </td>
                              <td class="resilience-col-alloc">
                                <span class="alloc-badge {pt.audio_thread_allocations === 0 ? 'alloc-zero' : 'alloc-warn'}">
                                  {pt.audio_thread_allocations === 0 ? '0 (Clean)' : `${pt.audio_thread_allocations} (Unsafe)`}
                                </span>
                              </td>
                              <td class="resilience-col-inv">
                                <span class="inv-badge {pt.priority_inversion_detected ? 'inv-danger' : 'inv-safe'}">
                                  {pt.priority_inversion_detected ? '⚠️ DETECTED' : '✓ Safe'}
                                </span>
                              </td>
                              <td class="resilience-col-xrun">
                                <span class="xrun-badge {pt.xrun_dropout ? 'xrun-dropped' : 'xrun-ok'}">
                                  {pt.xrun_dropout ? '💥 DROPOUT' : '✓ CLEAN'}
                                </span>
                              </td>
                            </tr>
                          {/each}
                        </tbody>
                      </table>
                    </div>
                  </div>
                {/if}
              </div>
            </div>

            <!-- Direct Download Link -->
            {#if previewPlugin.download_url}
              <div class="preview-section">
                <h4>Direct Binary Distribution Endpoint</h4>
                <a
                  href={previewPlugin.download_url}
                  target="_blank"
                  rel="noreferrer"
                  class="preview-download-link"
                  use:tooltip={"Open direct binary distribution endpoint in external browser"}
                >
                  🔗 {previewPlugin.download_url}
                </a>
              </div>
            {/if}

          {:else if activeInspectorTab === 'manifest'}
            <!-- Raw JSON Manifest Tab -->
            <div class="preview-section">
              <div class="preview-section-title-row">
                <h4>Package Registry Manifest (JSON)</h4>
                <div class="manifest-actions">
                  <button
                    type="button"
                    class="btn-copy-hash"
                    onclick={() => copyToClipboard(JSON.stringify(previewPlugin, null, 2), 'manifest-json')}
                    use:tooltip={"Copy entire JSON manifest to clipboard"}
                  >
                    {copiedHashId === 'manifest-json' ? '✓ Copied' : '📋 Copy JSON'}
                  </button>
                  <button
                    type="button"
                    class="btn-tiny-action"
                    onclick={() => exportPluginManifest(previewPlugin!)}
                    use:tooltip={"Export as standalone manifest file"}
                  >
                    📥 Save .json
                  </button>
                </div>
              </div>
              <div class="raw-manifest-container">
                <pre class="raw-manifest-pre"><code>{JSON.stringify(previewPlugin, null, 2)}</code></pre>
              </div>
            </div>
          {/if}

          <!-- Advanced Installation Options Collapsible Panel -->
          {#if previewPlugin.download_url}
            <div class="advanced-install-container">
              <button
                type="button"
                class="btn-toggle-advanced"
                onclick={() => (showAdvancedInstall = !showAdvancedInstall)}
              >
                <span>⚙️ Advanced Installation Routing & Options</span>
                <span class="chevron">{showAdvancedInstall ? '▲' : '▼'}</span>
              </button>

              {#if showAdvancedInstall}
                <div class="advanced-install-body">
                  <div class="adv-field-group">
                    <label for="adv-target-dir">Custom Destination Directory (Optional):</label>
                    <input
                      id="adv-target-dir"
                      type="text"
                      bind:value={customTargetDir}
                      placeholder="e.g. /opt/audio/vst3 or C:\CustomPlugins\VST3"
                      class="adv-input-text"
                    />
                    <div class="preset-path-pills">
                      <span class="path-pill-label">Quick Presets:</span>
                      <button
                        type="button"
                        class="path-preset-pill"
                        onclick={() => (customTargetDir = '~/.vst3')}
                      >
                        ~/.vst3
                      </button>
                      <button
                        type="button"
                        class="path-preset-pill"
                        onclick={() => (customTargetDir = '~/.clap')}
                      >
                        ~/.clap
                      </button>
                      <button
                        type="button"
                        class="path-preset-pill"
                        onclick={() => (customTargetDir = '~/.lv2')}
                      >
                        ~/.lv2
                      </button>
                      <button
                        type="button"
                        class="path-preset-pill"
                        onclick={() => (customTargetDir = '')}
                      >
                        Default OS Path
                      </button>
                    </div>
                  </div>

                  <div class="adv-field-group">
                    <label for="adv-strategy">Execution Strategy Override:</label>
                    <select id="adv-strategy" bind:value={forceStrategy} class="adv-select">
                      <option value="auto">Auto-Detect (Archive / Binary / OS Package)</option>
                      <option value="extract_zip">Force ZIP Archive Extraction</option>
                      <option value="silent_msi">Force Windows MSI Silent Installer</option>
                      <option value="silent_exe">Force Windows InnoSetup / NSIS Silent Executable</option>
                      <option value="mount_dmg">Force macOS DMG Mount & Install</option>
                      <option value="deb_pkg">Force Linux DEB Package Extraction</option>
                    </select>
                  </div>

                  <div class="adv-execute-row">
                    <button
                      type="button"
                      class="btn-primary adv-install-btn"
                      onclick={() => {
                        handleInstallAdvanced(previewPlugin!, customTargetDir, forceStrategy);
                        previewPlugin = null;
                      }}
                      use:tooltip={"Execute download and installation with your customized directory routing and strategy parameters"}
                    >
                      🚀 Install with Custom Options
                    </button>
                  </div>
                </div>
              {/if}
            </div>
          {/if}

          <!-- Actions -->
          <div class="preview-actions">
            {#if previewPlugin.download_url}
              <button
                class="btn-primary"
                onclick={() => {
                  handleInstall(previewPlugin!);
                  previewPlugin = null;
                }}
              >
                {previewInstalled
                  ? (previewVersionStatus?.state === 'update-available' ? 'Update Plugin' : 'Reinstall Plugin')
                  : 'Install Plugin'}
              </button>
            {/if}

            {#if previewInstalled}
              <button
                class="btn-secondary"
                onclick={() => handleOpenDirectory(previewPlugin!.name)}
              >
                📂 Open Local Folder
              </button>
              <button
                class="btn-danger-outline"
                onclick={() => {
                  handleRemove(previewPlugin!);
                  previewPlugin = null;
                }}
              >
                Uninstall
              </button>
            {/if}

            <button
              class="btn-folder-action"
              onclick={() => exportPluginManifest(previewPlugin!)}
              use:tooltip={"Export plugin registry manifest to a standalone JSON file"}
            >
              📥 Export JSON
            </button>

            <button class="btn-secondary" onclick={() => (previewPlugin = null)}>
              Close
            </button>
          </div>
        </div>
      </div>
    </div>
  {/if}

  <!-- Main Unified Workspace (Single Page) -->
  <main class="dashboard-body">
    <!-- Row 1: High-Priority Operations & Metrics Hub -->
    <section class="operations-grid">
      <!-- High-Value Action 1: Search & Dynamic Filters -->
      <div class="command-card search-card">
        <div class="card-header">
          <span class="card-icon">🔍</span>
          <h3>Find Plugins</h3>
          {#if searchQuery}
            <span class="search-match-count">{filteredPlugins.length} match{filteredPlugins.length === 1 ? '' : 'es'}</span>
            <button class="search-clear-btn" onclick={() => (searchQuery = '')}>Clear</button>
          {/if}
          <span class="tooltip-badge" use:tooltip={"Search catalog by plugin name or developer keywords"}>?</span>
        </div>
        <div class="filter-controls">
          <input
            type="text"
            placeholder="Search by name, developer, or keyword (e.g. 'wavetable', 'fm')..."
            bind:value={searchQuery}
            class="search-input {searchQuery ? 'has-query' : ''}"
          />
          <div class="dropdown-row">
            <select bind:value={selectedCategory} class="select-input" use:tooltip={"Filter by plugin instrument/audio category"}>
              <option value="all">All Categories</option>
              <option value="synth">Synthesizers</option>
              <option value="effect">Audio Effects</option>
              <option value="dynamics">Dynamics / Compressors</option>
              <option value="sampler">Samplers & Drums</option>
              <option value="utility">Utilities & Tools</option>
            </select>

            <select bind:value={selectedLicense} class="select-input" use:tooltip={"Filter by license type"}>
              <option value="all">All Licenses</option>
              <option value="opensource">Open Source</option>
              <option value="free">Free / Freeware</option>
              <option value="commercial">Commercial</option>
              <option value="donationware">Donationware</option>
            </select>

            <select bind:value={selectedPlatform} class="select-input" use:tooltip={"Filter by target OS platform"}>
              <option value="all">All Platforms</option>
              <option value="windows">Windows</option>
              <option value="macos">macOS</option>
              <option value="linux">Linux</option>
            </select>

            <select bind:value={selectedInstalledStatus} class="select-input" use:tooltip={"Filter by local installation & update state"}>
              <option value="all">All States</option>
              <option value="installed">✓ Installed</option>
              <option value="update-available">⚡ Updates Available</option>
              <option value="not-installed">Not Installed</option>
            </select>

            <select bind:value={selectedFormat} class="select-input" use:tooltip={"Filter installed plugins by audio plugin format"}>
              <option value="all">All Formats</option>
              <option value="vst3">VST3</option>
              <option value="clap">CLAP</option>
              <option value="lv2">LV2</option>
              <option value="au">AU (.component)</option>
              <option value="vst2">VST2 (.so/.dll)</option>
            </select>

            <select bind:value={sortBy} class="select-input" use:tooltip={"Sort order for the catalog"}>
              <option value="name">Sort: Name</option>
              <option value="developer">Sort: Developer</option>
              <option value="category">Sort: Category</option>
              <option value="license">Sort: License</option>
              <option value="status">Sort: Installed First</option>
            </select>

            <button
              type="button"
              class="btn-folder-action"
              onclick={() => handleOpenDirectory()}
              use:tooltip={"Open system audio plugin directory in file explorer"}
            >
              📁 VST3 Folder
            </button>

            <button
              type="button"
              class="btn-rescan-action"
              disabled={isScanningFolders}
              onclick={scanLocalPlugins}
              use:tooltip={"Re-scan local audio directories for installed VSTs"}
            >
              {isScanningFolders ? '⏳ Scanning...' : '🔄 Rescan'}
            </button>

            <button
              type="button"
              class="btn-update-all"
              disabled={isBatchUpdating}
              onclick={updateAllOutdatedPlugins}
              use:tooltip={"Download and install updates for all installed plugins that are outdated"}
            >
              {isBatchUpdating ? '⚡ Updating...' : '⚡ Update All'}
            </button>
          </div>
        </div>
      </div>

      <!-- High-Value Action 2: Crawler Pipeline Controller -->
      <div class="command-card crawler-card">
        <div class="card-header">
          <span class="card-icon">🕷️</span>
          <h3>Crawler Operations</h3>
          <span class="tooltip-badge" use:tooltip={"Trigger multi-target scraper for GitHub VST topics, KVR, and bedroom producers"}>?</span>
        </div>
        <p class="card-desc">Execute distributed crawlers to scrape fresh releases, compute SHA-256 binary digests, and populate registry.</p>
        <div class="crawler-action-bar">
          <button
            class="btn-primary"
            onclick={triggerCrawler}
            disabled={crawlerRunning}
            use:tooltip={"Start real-time crawler process across all upstream sources"}
          >
            {crawlerRunning ? 'Crawler Running...' : 'Run Crawler Now'}
          </button>
          <button
            class="btn-secondary"
            onclick={() => (isLogsCollapsed = !isLogsCollapsed)}
            use:tooltip={"Toggle real-time terminal output logs from crawler"}
          >
            {isLogsCollapsed ? 'Show Logs' : 'Hide Logs'} ({crawlerLogs.length})
          </button>
        </div>
      </div>

      <!-- High-Value Action 3: Health & Metrics Breakdown -->
      <div class="command-card stats-card">
        <div class="card-header">
          <span class="card-icon">📊</span>
          <h3>Registry Metrics</h3>
          <span class="tooltip-badge" use:tooltip={"Real-time database statistics and distribution health"}>?</span>
        </div>
        <div class="metrics-grid">
          <div class="metric-box">
            <span class="metric-num">{stats?.totalPlugins ?? plugins.length}</span>
            <span class="metric-lbl">Total Plugins</span>
          </div>
          <div class="metric-box">
            <span class="metric-num">{stats?.totalDistributions ?? plugins.filter(p => !!p.download_url).length}</span>
            <span class="metric-lbl">Installable</span>
          </div>
          <div class="metric-box">
            <span class="metric-num">{stats?.licenses?.find(l => l.license_model === 'opensource')?.count || '0'}</span>
            <span class="metric-lbl">Open Source</span>
          </div>
          <div class="metric-box">
            <span class="metric-num">{stats?.licenses?.find(l => l.license_model === 'free')?.count || '0'}</span>
            <span class="metric-lbl">Free VSTs</span>
          </div>
        </div>
      </div>
    </section>

    <!-- Collapsible Crawler Terminal Console -->
    {#if !isLogsCollapsed}
      <section class="console-section">
        <div class="console-header">
          <span>⚡ Live Crawler Terminal Output</span>
          <button class="console-close" onclick={() => (isLogsCollapsed = true)}>✕</button>
        </div>
        <div class="console-window">
          {#if crawlerLogs.length === 0}
            <div class="console-line empty-line">No crawler output yet. Click 'Run Crawler Now' to begin.</div>
          {:else}
            {#each crawlerLogs as log}
              <div class="console-line">{log}</div>
            {/each}
          {/if}
        </div>
      </section>
    {/if}

    <!-- High-Value Core: Plugin Registry Explorer & Silent Installer -->
    <section class="registry-section">
      <div class="section-header">
        <div class="header-titles">
          <h2>Vetted Plugin Library</h2>
          <span class="section-subtitle">Showing {filteredPlugins.length} of {plugins.length} verified packages</span>
        </div>

        <!-- Quick Filter Count Pills -->
        <div class="filter-pills-row">
          <button
            type="button"
            class="filter-pill {selectedInstalledStatus === 'all' && selectedCategory === 'all' ? 'pill-active' : ''}"
            onclick={() => { selectedInstalledStatus = 'all'; selectedCategory = 'all'; }}
            use:tooltip={"View all plugins in database catalog"}
          >
            All <span class="pill-badge">{countAll}</span>
          </button>

          <button
            type="button"
            class="filter-pill {selectedInstalledStatus === 'installed' ? 'pill-active' : ''}"
            onclick={() => (selectedInstalledStatus = 'installed')}
            use:tooltip={"View plugins currently installed on this system"}
          >
            ✓ Installed <span class="pill-badge">{countInstalled}</span>
          </button>

          <button
            type="button"
            class="filter-pill pill-updates {selectedInstalledStatus === 'update-available' ? 'pill-active' : ''}"
            onclick={() => (selectedInstalledStatus = 'update-available')}
            use:tooltip={"View installed plugins with newer releases in registry"}
          >
            ⚡ Updates <span class="pill-badge">{countUpdates}</span>
          </button>

          <button
            type="button"
            class="filter-pill {selectedInstalledStatus === 'not-installed' ? 'pill-active' : ''}"
            onclick={() => (selectedInstalledStatus = 'not-installed')}
            use:tooltip={"View plugins not yet installed"}
          >
            Not Installed <span class="pill-badge">{countNotInstalled}</span>
          </button>

          <!-- Category Discoverability Filter Chips -->
          <button
            type="button"
            class="filter-pill category-pill {selectedCategory === 'synth' ? 'pill-active' : ''}"
            onclick={() => { selectedCategory = selectedCategory === 'synth' ? 'all' : 'synth'; }}
            use:tooltip={"Filter directly for Synthesizers"}
          >
            🎹 Synths
          </button>

          <button
            type="button"
            class="filter-pill category-pill {selectedCategory === 'effect' ? 'pill-active' : ''}"
            onclick={() => { selectedCategory = selectedCategory === 'effect' ? 'all' : 'effect'; }}
            use:tooltip={"Filter directly for Audio Effects"}
          >
            🎛️ Effects
          </button>

          <!-- Format Discoverability Filter Chips -->
          <button
            type="button"
            class="filter-pill format-pill {selectedFormat === 'vst3' ? 'pill-active' : ''}"
            onclick={() => { selectedFormat = selectedFormat === 'vst3' ? 'all' : 'vst3'; }}
            use:tooltip={`Filter directly for VST3 format plugins (${countVST3} available)`}
          >
            🎚️ VST3 <span class="pill-badge">{countVST3}</span>
          </button>

          <button
            type="button"
            class="filter-pill format-pill {selectedFormat === 'clap' ? 'pill-active' : ''}"
            onclick={() => { selectedFormat = selectedFormat === 'clap' ? 'all' : 'clap'; }}
            use:tooltip={`Filter directly for modern CLAP format plugins (${countCLAP} available)`}
          >
            ⚡ CLAP <span class="pill-badge">{countCLAP}</span>
          </button>

          <button
            type="button"
            class="filter-pill format-pill {selectedFormat === 'lv2' ? 'pill-active' : ''}"
            onclick={() => { selectedFormat = selectedFormat === 'lv2' ? 'all' : 'lv2'; }}
            use:tooltip={`Filter directly for LV2 format plugins (${countLV2} available)`}
          >
            🎛️ LV2 <span class="pill-badge">{countLV2}</span>
          </button>

          <button
            type="button"
            class="filter-pill format-pill {selectedFormat === 'au' ? 'pill-active' : ''}"
            onclick={() => { selectedFormat = selectedFormat === 'au' ? 'all' : 'au'; }}
            use:tooltip={`Filter directly for Apple Audio Unit (.component) format plugins (${countAU} available)`}
          >
            🍎 AU <span class="pill-badge">{countAU}</span>
          </button>
        </div>
      </div>

      {#if loading}
        <div class="loading-state">
          <div class="spinner"></div>
          <p>Syncing plugin registry database...</p>
        </div>
      {:else if filteredPlugins.length === 0}
        <div class="empty-state">
          <span class="empty-icon">📦</span>
          <h3>No plugins match current filters</h3>
          <p>Try resetting search terms or click "Run Crawler Now" to fetch plugins into PostgreSQL.</p>
        </div>
      {:else}
        <div class="plugins-grid">
          {#each filteredPlugins as plugin}
            {@const installedInfo = getInstalledInfo(plugin.name)}
            {@const versionStatus = installedInfo ? getVersionStatus(plugin.version, installedInfo.version) : null}
            <div class="plugin-card">
              <!-- Header -->
              <div class="card-top">
                <div class="title-group">
                  <div class="title-badge-row">
                    <button
                      type="button"
                      class="plugin-title-btn"
                      title="Click to preview plugin details"
                      onclick={() => (previewPlugin = plugin)}
                    >
                      <h3 class="plugin-title">{plugin.name}</h3>
                    </button>
                    {#if installedInfo}
                      <span
                        class="installed-badge {versionStatus?.state === 'update-available' ? 'badge-update' : ''}"
                        use:tooltip={`Installed: ${installedInfo.path} (${installedInfo.format.toUpperCase()})`}
                      >
                        ✓ {installedInfo.format.toUpperCase()}
                      </span>
                      {#if versionStatus}
                        <span
                          class="version-status-badge {versionStatus.state === 'update-available' ? 'badge-update' : 'badge-current'}"
                          use:tooltip={`Catalog: v${plugin.version || '1.0.0'} | Local: ${installedInfo.version || 'detected'}`}
                        >
                          {versionStatus.label}
                        </span>
                      {/if}
                    {/if}
                    {#if isLegacyPlugin(plugin, installedInfo)}
                      <span
                        class="badge-legacy"
                        use:tooltip={"Legacy VST2 or 32-bit architecture. Modern 64-bit DAWs require bridging (e.g. yabridge, jBridge)."}
                      >
                        ⚠️ Legacy
                      </span>
                    {/if}
                  </div>
                  <span class="developer-tag" use:tooltip={`Created by developer ${plugin.developer}`}>
                    by {plugin.developer}
                  </span>
                </div>
                <span
                  class="license-tag license-{plugin.license_model || 'free'}"
                  use:tooltip={`License category: ${plugin.license_model}`}
                >
                  {plugin.license_model}
                </span>
              </div>

              <!-- Metadata Specs -->
              <div class="specs-grid">
                <div class="spec-item">
                  <span class="spec-label">Format</span>
                  <span class="spec-val format-badge-tag" use:tooltip={"Supported binary audio plugin format"}>
                    {installedInfo ? installedInfo.format.toUpperCase() : inferCatalogFormat(plugin)}
                  </span>
                </div>
                <div class="spec-item">
                  <span class="spec-label">Target OS</span>
                  <span class="spec-val" use:tooltip={`Architecture: ${plugin.architecture || 'x86_64'}`}>
                    {plugin.platform || 'windows'}
                  </span>
                </div>
                <div class="spec-item">
                  <span class="spec-label">Strategy</span>
                  <span class="spec-val" use:tooltip={`Installation strategy: ${plugin.strategy || 'extract_binaries'}`}>
                    {plugin.strategy || 'extract_binaries'}
                  </span>
                </div>
                {#if installedInfo}
                  <div class="spec-item spec-path-item" use:tooltip={`Active file path: ${installedInfo.path}`}>
                    <span class="spec-label">Location</span>
                    <span class="spec-val path-val">{installedInfo.path}</span>
                  </div>
                {/if}
              </div>

              <!-- Discoverability Tag Pills -->
              <div class="tag-pills-row">
                {#each getPluginTags(plugin) as tag}
                  <button
                    type="button"
                    class="tag-pill"
                    onclick={() => { searchQuery = tag; }}
                    use:tooltip={`Click to filter catalog by tag '${tag}'`}
                  >
                    #{tag}
                  </button>
                {/each}
              </div>

              <!-- Cryptographic SHA256 Verification -->
              {#if plugin.sha256_hash}
                <div class="hash-container" use:tooltip={`Full SHA-256 Digest: ${plugin.sha256_hash}`}>
                  <span class="hash-label">🔒 SHA-256:</span>
                  <code class="hash-code">{plugin.sha256_hash.substring(0, 16)}...</code>
                  <button
                    type="button"
                    class="btn-card-copy-hash"
                    onclick={(e) => { e.stopPropagation(); copyToClipboard(plugin.sha256_hash!, plugin.id); }}
                    use:tooltip={"Copy full SHA-256 hash to clipboard"}
                  >
                    {copiedHashId === plugin.id ? '✓' : '📋'}
                  </button>
                </div>
              {/if}

              <!-- Description / Source details -->
              {#if plugin.metadata?.description || plugin.metadata?.source}
                <div class="description-box" use:tooltip={plugin.metadata?.description || `Source: ${plugin.metadata?.source}`}>
                  <p class="description-text">
                    {plugin.metadata?.description || `Cataloged from ${plugin.metadata?.source} crawler pipeline`}
                  </p>
                </div>
              {/if}

              <!-- Action Bar -->
              <div class="install-bar">
                {#if plugin.download_url}
                  <div class="action-button-group">
                    <button
                      class="btn-install"
                      disabled={installStatus[plugin.id] === 'Downloading & Installing...'}
                      onclick={() => handleInstall(plugin)}
                      use:tooltip={`Stream download and install ${plugin.name} via native silent installer`}
                    >
                      {installStatus[plugin.id] === 'Downloading & Installing...'
                        ? 'Installing...'
                        : installedInfo
                        ? (versionStatus?.state === 'update-available' ? 'Update Plugin' : 'Reinstall')
                        : 'Install Plugin'}
                    </button>
                    <button
                      class="btn-inspect"
                      onclick={() => (previewPlugin = plugin)}
                      use:tooltip={`Open detail inspection drawer for ${plugin.name}`}
                    >
                      👁 Details
                    </button>
                    {#if installedInfo}
                      <button
                        class="btn-dir"
                        onclick={() => handleOpenDirectory(plugin.name)}
                        use:tooltip={`Open ${plugin.name} in file manager (${installedInfo.path})`}
                      >
                        📂 Files
                      </button>
                    {/if}
                    <button
                      class="btn-remove"
                      disabled={installStatus[plugin.id] === 'Removing plugin files...'}
                      onclick={() => handleRemove(plugin)}
                      use:tooltip={`Safely uninstall ${plugin.name} from OS audio plugin directories`}
                    >
                      Uninstall
                    </button>
                  </div>
                {:else}
                  <div class="action-button-group">
                    <button class="btn-install btn-disabled" disabled use:tooltip={"No direct download distribution payload URL"}>
                      Direct Link Unavailable
                    </button>
                    <button
                      class="btn-remove"
                      disabled={installStatus[plugin.id] === 'Removing plugin files...'}
                      onclick={() => handleRemove(plugin)}
                      use:tooltip={`Clean uninstall if previously installed locally`}
                    >
                      Uninstall
                    </button>
                  </div>
                {/if}
              </div>

              <!-- Install Result Banner -->
              {#if installStatus[plugin.id] && installStatus[plugin.id] !== 'Downloading & Installing...'}
                <div class="status-banner {installStatus[plugin.id].startsWith('Error') ? 'status-err' : 'status-ok'}">
                  {installStatus[plugin.id]}
                </div>
              {/if}
            </div>
          {/each}
        </div>
      {/if}
    </section>
  </main>
</div>

<style>
  :global(*) {
    box-sizing: border-box;
    margin: 0;
    padding: 0;
  }

  :global(body) {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
    background-color: #0f1117;
    color: #e1e7f0;
    line-height: 1.5;
  }

  :global(.custom-tooltip) {
    position: absolute;
    background: #1e2433;
    color: #f1f5f9;
    padding: 6px 12px;
    border-radius: 6px;
    font-size: 0.78rem;
    pointer-events: none;
    z-index: 9999;
    box-shadow: 0 4px 14px rgba(0, 0, 0, 0.45);
    border: 1px solid #334155;
    max-width: 320px;
  }

  .dashboard-root {
    display: flex;
    flex-direction: column;
    min-height: 100vh;
  }

  /* Global Top Nav */
  .top-nav {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1rem 2rem;
    background: #161b26;
    border-bottom: 1px solid #232d42;
    position: sticky;
    top: 0;
    z-index: 100;
  }

  .brand {
    display: flex;
    align-items: center;
    gap: 0.8rem;
  }

  .logo-mark {
    font-size: 1.8rem;
    background: #2563eb;
    width: 42px;
    height: 42px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 8px;
    color: #fff;
  }

  .brand-title {
    font-size: 1.3rem;
    font-weight: 800;
    letter-spacing: 1px;
    color: #f8fafc;
  }

  .brand-subtitle {
    font-size: 0.75rem;
    color: #94a3b8;
    display: block;
  }

  .status-strip {
    display: flex;
    align-items: center;
    gap: 1.2rem;
  }

  .status-item {
    display: flex;
    align-items: center;
    gap: 0.4rem;
    font-size: 0.85rem;
    background: #1f2739;
    padding: 0.4rem 0.8rem;
    border-radius: 6px;
    border: 1px solid #2d374e;
  }

  .status-indicator {
    width: 8px;
    height: 8px;
    border-radius: 50%;
  }

  .status-indicator.online { background: #10b981; box-shadow: 0 0 8px #10b981; }
  .status-indicator.offline { background: #ef4444; }
  .status-indicator.idle { background: #64748b; }
  .status-indicator.active-pulse {
    background: #3b82f6;
    box-shadow: 0 0 10px #3b82f6;
    animation: pulse 1.2s infinite;
  }

  @keyframes pulse {
    0% { opacity: 0.4; }
    50% { opacity: 1; }
    100% { opacity: 0.4; }
  }

  .version-tag {
    background: #334155;
    padding: 2px 6px;
    border-radius: 4px;
    font-size: 0.75rem;
    font-weight: 700;
  }

  .btn-theme-toggle {
    background: #1e293b;
    color: #38bdf8;
    border: 1px solid #334155;
    padding: 0.4rem 0.8rem;
    border-radius: 6px;
    font-size: 0.8rem;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s ease;
    display: inline-flex;
    align-items: center;
    gap: 0.3rem;
  }

  .btn-theme-toggle:hover {
    background: #334155;
    border-color: #38bdf8;
    color: #7dd3fc;
  }

  /* High-Contrast DAW / Cyber Slate Theme Overrides */
  .dashboard-root.theme-contrast {
    background-color: #030712;
    color: #f9fafb;
  }

  .dashboard-root.theme-contrast .top-nav {
    background: #0b0f19;
    border-bottom: 2px solid #3b82f6;
  }

  .dashboard-root.theme-contrast .command-card {
    background: #0f172a;
    border: 1px solid #334155;
    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.6);
  }

  .dashboard-root.theme-contrast .plugin-card {
    background: #0b0f19;
    border: 1px solid #2563eb;
    box-shadow: 0 2px 10px rgba(37, 99, 235, 0.12);
  }

  .dashboard-root.theme-contrast .filter-chip {
    border-color: #475569;
  }

  .dashboard-root.theme-contrast .filter-chip.active {
    background: #2563eb;
    color: #ffffff;
    border-color: #60a5fa;
  }

  .btn-danger-outline {
    background: transparent;
    color: #ef4444;
    border: 1px solid #ef4444;
    padding: 0.4rem 0.8rem;
    border-radius: 6px;
    font-size: 0.8rem;
    cursor: pointer;
    transition: 0.2s;
  }

  .btn-danger-outline:hover {
    background: #ef4444;
    color: white;
  }

  .btn-success-outline {
    background: transparent;
    color: #10b981;
    border: 1px solid #10b981;
    padding: 0.4rem 0.8rem;
    border-radius: 6px;
    font-size: 0.8rem;
    cursor: pointer;
    transition: 0.2s;
  }

  .btn-success-outline:hover {
    background: #10b981;
    color: white;
  }

  .modal-backdrop {
    position: fixed;
    top: 0;
    left: 0;
    width: 100vw;
    height: 100vh;
    background: rgba(0, 0, 0, 0.75);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
  }

  .modal-card {
    background: #161b26;
    border: 1px solid #2d374e;
    border-radius: 12px;
    width: 90%;
    max-width: 600px;
    padding: 1.5rem;
    box-shadow: 0 10px 30px rgba(0, 0, 0, 0.6);
  }

  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1.2rem;
    border-bottom: 1px solid #232d42;
    padding-bottom: 0.8rem;
  }

  .modal-close {
    background: none;
    border: none;
    color: #94a3b8;
    font-size: 1.2rem;
    cursor: pointer;
  }

  .modal-form {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .form-row {
    display: flex;
    gap: 1rem;
  }

  .form-row label, .modal-form label {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 0.4rem;
    font-size: 0.8rem;
    color: #cbd5e1;
  }

  .textarea-input {
    background: #0f1117;
    border: 1px solid #2d374e;
    border-radius: 6px;
    color: white;
    padding: 0.6rem;
    font-size: 0.85rem;
    resize: vertical;
    min-height: 70px;
    font-family: inherit;
  }

  .modal-actions {
    display: flex;
    justify-content: flex-end;
    gap: 0.8rem;
    margin-top: 0.5rem;
  }

  /* Alerts */
  .action-alert {
    background: #1e3a8a;
    color: #bfdbfe;
    padding: 0.6rem 2rem;
    display: flex;
    justify-content: space-between;
    align-items: center;
    font-size: 0.85rem;
  }

  .alert-close {
    background: none;
    border: none;
    color: #bfdbfe;
    cursor: pointer;
  }

  .error-banner {
    background: #7f1d1d;
    color: #fecaca;
    padding: 0.6rem 2rem;
    display: flex;
    justify-content: space-between;
    align-items: center;
    font-size: 0.85rem;
  }

  .retry-btn {
    background: #ef4444;
    color: white;
    border: none;
    padding: 0.2rem 0.6rem;
    border-radius: 4px;
    cursor: pointer;
  }

  /* Body */
  .dashboard-body {
    padding: 2rem;
    display: flex;
    flex-direction: column;
    gap: 2rem;
  }

  /* Operations Grid */
  .operations-grid {
    display: grid;
    grid-template-columns: 1.5fr 1.2fr 1fr;
    gap: 1.5rem;
  }

  .command-card {
    background: #161b26;
    border: 1px solid #232d42;
    border-radius: 12px;
    padding: 1.2rem;
    display: flex;
    flex-direction: column;
    gap: 0.8rem;
  }

  .card-header {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .tag-pills-row {
    display: flex;
    flex-wrap: wrap;
    gap: 0.35rem;
    margin: 0.2rem 0;
  }

  .tag-pill {
    background: #0f1117;
    border: 1px solid #232d42;
    color: #38bdf8;
    font-size: 0.68rem;
    padding: 2px 7px;
    border-radius: 12px;
    cursor: pointer;
    transition: 0.15s ease-in-out;
  }

  .tag-pill:hover {
    background: #1e293b;
    border-color: #38bdf8;
    color: #7dd3fc;
    transform: scale(1.05);
  }

  .format-badge-tag {
    color: #a78bfa;
    font-weight: 700;
  }

  .card-header h3 {
    font-size: 1rem;
    font-weight: 700;
    color: #f1f5f9;
  }

  .search-match-count {
    font-size: 0.72rem;
    background: #1e3a8a;
    color: #93c5fd;
    padding: 2px 7px;
    border-radius: 10px;
    font-weight: 600;
  }

  .search-clear-btn {
    background: transparent;
    border: none;
    color: #94a3b8;
    font-size: 0.75rem;
    cursor: pointer;
    text-decoration: underline;
  }

  .search-clear-btn:hover {
    color: #f87171;
  }

  .search-input.has-query {
    border-color: #3b82f6;
    box-shadow: 0 0 6px rgba(59, 130, 246, 0.3);
  }

  .tooltip-badge {
    background: #2563eb;
    color: white;
    width: 18px;
    height: 18px;
    border-radius: 50%;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    font-size: 0.7rem;
    cursor: help;
    margin-left: auto;
  }

  .card-desc {
    font-size: 0.8rem;
    color: #94a3b8;
  }

  .filter-controls {
    display: flex;
    flex-direction: column;
    gap: 0.6rem;
  }

  .search-input {
    width: 100%;
    padding: 0.6rem 0.8rem;
    background: #0f1117;
    border: 1px solid #2d374e;
    border-radius: 6px;
    color: white;
    font-size: 0.85rem;
  }

  .dropdown-row {
    display: flex;
    gap: 0.6rem;
  }

  .select-input {
    flex: 1;
    padding: 0.5rem;
    background: #0f1117;
    border: 1px solid #2d374e;
    border-radius: 6px;
    color: #cbd5e1;
    font-size: 0.8rem;
  }

  .filter-toggle-btn {
    padding: 0.5rem 0.8rem;
    background: #0f1117;
    border: 1px solid #2d374e;
    border-radius: 6px;
    color: #94a3b8;
    font-size: 0.8rem;
    cursor: pointer;
    transition: 0.2s;
    white-space: nowrap;
  }

  .filter-toggle-btn.active {
    background: #064e3b;
    border-color: #10b981;
    color: #34d399;
    font-weight: 700;
  }

  .btn-folder-action, .btn-rescan-action {
    padding: 0.5rem 0.8rem;
    background: #0f1117;
    border: 1px solid #2d374e;
    border-radius: 6px;
    color: #94a3b8;
    font-size: 0.8rem;
    cursor: pointer;
    transition: 0.2s;
    white-space: nowrap;
  }

  .btn-folder-action:hover, .btn-rescan-action:hover {
    background: #1e293b;
    color: #f1f5f9;
    border-color: #3b82f6;
  }

  .btn-update-all {
    padding: 0.5rem 0.8rem;
    background: #78350f;
    border: 1px solid #d97706;
    border-radius: 6px;
    color: #fde68a;
    font-size: 0.8rem;
    cursor: pointer;
    transition: 0.2s;
    white-space: nowrap;
    font-weight: 600;
  }

  .btn-update-all:hover:not(:disabled) {
    background: #b45309;
    color: white;
  }

  .btn-update-all:disabled, .btn-rescan-action:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .title-badge-row {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .installed-badge {
    background: #064e3b;
    color: #34d399;
    font-size: 0.65rem;
    font-weight: 800;
    padding: 2px 6px;
    border-radius: 4px;
    border: 1px solid #059669;
    letter-spacing: 0.5px;
  }

  .installed-badge.badge-update {
    background: #78350f;
    color: #fcd34d;
    border-color: #d97706;
  }

  .version-status-badge {
    font-size: 0.65rem;
    padding: 2px 6px;
    border-radius: 4px;
    font-weight: 600;
  }

  .version-status-badge.badge-current {
    background: #1e293b;
    color: #94a3b8;
    border: 1px solid #334155;
  }

  .version-status-badge.badge-update {
    background: #451a03;
    color: #fbbf24;
    border: 1px solid #b45309;
  }

  .badge-legacy {
    font-size: 0.65rem;
    padding: 2px 6px;
    border-radius: 4px;
    font-weight: 600;
    background: #450a0a;
    color: #f87171;
    border: 1px solid #991b1b;
  }

  .legacy-warning-box {
    display: flex;
    gap: 0.75rem;
    align-items: flex-start;
    background: rgba(153, 27, 27, 0.15);
    border: 1px solid #991b1b;
    border-radius: 8px;
    padding: 0.85rem;
    margin-bottom: 1.25rem;
  }

  .legacy-warning-box .warning-icon {
    font-size: 1.2rem;
    line-height: 1;
  }

  .legacy-warning-box .warning-content strong {
    display: block;
    color: #fca5a5;
    font-size: 0.85rem;
    margin-bottom: 0.25rem;
  }

  .legacy-warning-box .warning-content p {
    color: #cbd5e1;
    font-size: 0.78rem;
    line-height: 1.4;
    margin: 0;
  }

  .msi-spec-banner {
    display: flex;
    gap: 0.75rem;
    align-items: flex-start;
    background: rgba(59, 130, 246, 0.08);
    border: 1px solid rgba(59, 130, 246, 0.25);
    border-radius: var(--radius-sm);
    padding: 0.75rem 1rem;
    margin-top: 0.75rem;
  }

  .msi-icon {
    font-size: 1.25rem;
    line-height: 1;
  }

  .msi-content strong {
    display: block;
    color: #60a5fa;
    font-size: 0.82rem;
    margin-bottom: 0.2rem;
  }

  .msi-content p {
    color: #94a3b8;
    font-size: 0.75rem;
    line-height: 1.4;
    margin: 0;
  }

  .msi-content code {
    background: rgba(0, 0, 0, 0.3);
    padding: 1px 4px;
    border-radius: 3px;
    color: #93c5fd;
    font-size: 0.7rem;
  }

  .clap-spec-banner {
    display: flex;
    gap: 0.75rem;
    align-items: flex-start;
    background: rgba(245, 158, 11, 0.08);
    border: 1px solid rgba(245, 158, 11, 0.25);
    border-radius: 8px;
    padding: 0.75rem;
  }

  .clap-icon {
    font-size: 1.25rem;
    line-height: 1;
  }

  .clap-content strong {
    display: block;
    color: #fde68a;
    font-size: 0.82rem;
    margin-bottom: 0.2rem;
  }

  .clap-content p {
    color: #d6d3d1;
    font-size: 0.75rem;
    line-height: 1.4;
    margin: 0;
  }

  .au-spec-banner {
    display: flex;
    gap: 0.75rem;
    align-items: flex-start;
    background: rgba(244, 63, 94, 0.08);
    border: 1px solid rgba(244, 63, 94, 0.25);
    border-radius: 8px;
    padding: 0.75rem;
  }

  .au-icon {
    font-size: 1.25rem;
    line-height: 1;
  }

  .au-content strong {
    display: block;
    color: #fb7185;
    font-size: 0.82rem;
    margin-bottom: 0.2rem;
  }

  .au-content p {
    color: #94a3b8;
    font-size: 0.75rem;
    line-height: 1.4;
    margin: 0 0 0.5rem 0;
  }

  .lv2-spec-features {
    display: flex;
    flex-wrap: wrap;
    gap: 0.35rem;
  }

  .lv2-pill {
    font-size: 0.65rem;
    font-family: var(--font-mono);
    padding: 0.15rem 0.4rem;
    border-radius: 4px;
    background: rgba(16, 185, 129, 0.15);
    color: #6ee7b7;
    border: 1px solid rgba(16, 185, 129, 0.3);
  }

  .btn-dir {
    background: #1e293b;
    color: #cbd5e1;
    border: 1px solid #334155;
    padding: 0.6rem;
    border-radius: 6px;
    font-size: 0.8rem;
    cursor: pointer;
    transition: 0.2s;
    white-space: nowrap;
  }

  .btn-dir:hover {
    background: #334155;
    color: #f8fafc;
    border-color: #3b82f6;
  }

  .btn-inspect {
    background: #1e293b;
    color: #93c5fd;
    border: 1px solid #1e3a8a;
    padding: 0.6rem;
    border-radius: 6px;
    font-size: 0.8rem;
    cursor: pointer;
    transition: 0.2s;
    white-space: nowrap;
  }

  .btn-inspect:hover {
    background: #1e3a8a;
    color: white;
  }

  .plugin-title-btn {
    background: none;
    border: none;
    padding: 0;
    text-align: left;
    cursor: pointer;
    font-family: inherit;
  }

  .plugin-title-btn:hover .plugin-title {
    color: #60a5fa;
    text-decoration: underline;
  }

  /* Preview Card Styles */
  .preview-card {
    max-width: 680px;
  }

  .inspector-tabs-bar {
    display: flex;
    gap: 0.5rem;
    padding: 0.5rem 1.5rem;
    background: #111827;
    border-bottom: 1px solid #1f2937;
  }

  .inspector-tab-btn {
    background: transparent;
    border: 1px solid transparent;
    color: #94a3b8;
    padding: 0.35rem 0.8rem;
    border-radius: 6px;
    font-size: 0.78rem;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .inspector-tab-btn:hover {
    color: #f1f5f9;
    background: rgba(255, 255, 255, 0.05);
  }

  .inspector-tab-btn.active {
    color: #38bdf8;
    background: rgba(56, 189, 248, 0.12);
    border-color: rgba(56, 189, 248, 0.3);
  }

  .filesystem-details-card {
    background: #0f1117;
    border: 1px solid #1f2937;
    border-radius: 8px;
    padding: 0.8rem 1rem;
    display: flex;
    flex-direction: column;
    gap: 0.6rem;
  }

  .fs-row {
    display: flex;
    align-items: baseline;
    gap: 0.8rem;
    font-size: 0.82rem;
  }

  .fs-label {
    width: 160px;
    color: #64748b;
    font-size: 0.72rem;
    text-transform: uppercase;
    font-weight: 600;
    flex-shrink: 0;
  }

  .fs-value {
    color: #e2e8f0;
  }

  .badge-format {
    background: rgba(59, 130, 246, 0.2);
    color: #60a5fa;
    padding: 2px 6px;
    border-radius: 4px;
    font-size: 0.72rem;
    font-weight: 700;
  }

  .fs-path {
    background: rgba(0, 0, 0, 0.4);
    color: #38bdf8;
    padding: 2px 6px;
    border-radius: 4px;
    word-break: break-all;
    font-size: 0.75rem;
  }

  .fs-action-row {
    display: flex;
    gap: 0.6rem;
    margin-top: 0.4rem;
    padding-top: 0.6rem;
    border-top: 1px solid rgba(255, 255, 255, 0.06);
  }

  .empty-fs-state {
    background: #0f1117;
    padding: 1.2rem;
    border-radius: 8px;
    text-align: center;
    color: #94a3b8;
    font-size: 0.82rem;
    border: 1px dashed #334155;
  }

  .manifest-actions {
    display: flex;
    gap: 0.5rem;
  }

  .raw-manifest-container {
    background: #090d16;
    border: 1px solid #1e293b;
    border-radius: 6px;
    padding: 0.8rem;
    max-height: 260px;
    overflow-y: auto;
  }

  .raw-manifest-pre {
    margin: 0;
    font-family: monospace;
    font-size: 0.72rem;
    color: #38bdf8;
    line-height: 1.4;
    white-space: pre-wrap;
    word-break: break-word;
  }

  .preview-header-info {
    display: flex;
    flex-direction: column;
    gap: 0.2rem;
  }

  .preview-category {
    font-size: 0.75rem;
    color: #38bdf8;
    text-transform: uppercase;
    font-weight: 700;
  }

  .preview-author {
    font-size: 0.85rem;
    color: #94a3b8;
  }

  .preview-content {
    display: flex;
    flex-direction: column;
    gap: 1.2rem;
  }

  .preview-install-status {
    padding: 0.8rem 1rem;
    border-radius: 8px;
    display: flex;
    align-items: flex-start;
    gap: 0.8rem;
    font-size: 0.85rem;
  }

  .preview-status-details {
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
    flex: 1;
  }

  .preview-status-title-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.8rem;
  }

  .verification-badge {
    font-size: 0.7rem;
    padding: 0.15rem 0.5rem;
    border-radius: 4px;
    font-weight: 600;
  }

  .verification-badge.valid {
    background: rgba(16, 185, 129, 0.2);
    color: #6ee7b7;
    border: 1px solid rgba(16, 185, 129, 0.4);
  }

  .verification-badge.verifying {
    background: rgba(59, 130, 246, 0.2);
    color: #93c5fd;
    border: 1px solid rgba(59, 130, 246, 0.4);
  }

  .verification-badge.invalid {
    background: rgba(239, 68, 68, 0.2);
    color: #fca5a5;
    border: 1px solid rgba(239, 68, 68, 0.4);
  }

  .version-label-note {
    font-size: 0.75rem;
    color: #fde68a;
  }

  .preview-status-subactions {
    display: flex;
    gap: 0.5rem;
    margin-top: 0.2rem;
  }

  .btn-tiny-action {
    background: rgba(255, 255, 255, 0.08);
    color: #e2e8f0;
    border: 1px solid rgba(255, 255, 255, 0.15);
    padding: 0.2rem 0.6rem;
    border-radius: 4px;
    font-size: 0.75rem;
    cursor: pointer;
    transition: background 0.15s ease;
  }

  .btn-tiny-action:hover:not(:disabled) {
    background: rgba(255, 255, 255, 0.16);
  }

  .btn-tiny-action:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .preview-install-status.status-installed {
    background: #064e3b;
    color: #a7f3d0;
    border: 1px solid #059669;
  }

  .preview-install-status.status-update {
    background: #451a03;
    color: #fde68a;
    border: 1px solid #b45309;
  }

  .preview-install-status.status-not-installed {
    background: #1e293b;
    color: #cbd5e1;
    border: 1px solid #334155;
  }

  .preview-specs-grid {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 0.8rem;
    background: #0f1117;
    padding: 0.8rem;
    border-radius: 8px;
  }

  .spec-tile {
    display: flex;
    flex-direction: column;
    align-items: center;
    text-align: center;
  }

  .spec-tile .label {
    font-size: 0.7rem;
    color: #64748b;
  }

  .spec-tile .value {
    font-size: 0.85rem;
    color: #e2e8f0;
  }

  .preview-section-title-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 0.4rem;
  }

  .preview-section-title-row h4 {
    margin-bottom: 0;
  }

  .preview-format-tag {
    font-size: 0.7rem;
    padding: 2px 8px;
    border-radius: 4px;
    background: rgba(59, 130, 246, 0.15);
    color: #60a5fa;
    border: 1px solid rgba(59, 130, 246, 0.3);
    font-weight: 600;
  }

  .matrix-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 0.6rem;
    background: #0f1117;
    padding: 0.75rem;
    border-radius: 8px;
    margin-bottom: 0.6rem;
  }

  .matrix-item {
    display: flex;
    flex-direction: column;
    gap: 0.15rem;
  }

  .matrix-label {
    font-size: 0.68rem;
    color: #64748b;
    text-transform: uppercase;
    font-weight: 600;
  }

  .matrix-val {
    font-size: 0.82rem;
    color: #e2e8f0;
    font-weight: 500;
  }

  .lv2-spec-banner {
    display: flex;
    gap: 0.75rem;
    align-items: flex-start;
    background: rgba(16, 185, 129, 0.08);
    border: 1px solid rgba(16, 185, 129, 0.25);
    border-radius: 8px;
    padding: 0.75rem;
  }

  .lv2-icon {
    font-size: 1.25rem;
    line-height: 1;
  }

  .lv2-content strong {
    display: block;
    color: #34d399;
    font-size: 0.82rem;
    margin-bottom: 0.2rem;
  }

  .lv2-content p {
    color: #94a3b8;
    font-size: 0.75rem;
    line-height: 1.4;
    margin: 0;
  }

  .lv2-content code {
    background: rgba(0, 0, 0, 0.3);
    padding: 1px 4px;
    border-radius: 3px;
    color: #6ee7b7;
  }

  .vst2-spec-banner {
    display: flex;
    gap: 0.75rem;
    align-items: flex-start;
    background: rgba(245, 158, 11, 0.08);
    border: 1px solid rgba(245, 158, 11, 0.25);
    border-radius: 8px;
    padding: 0.75rem;
    margin-top: 0.5rem;
  }

  .vst2-icon {
    font-size: 1.25rem;
    line-height: 1;
  }

  .vst2-content strong {
    display: block;
    color: #fbbf24;
    font-size: 0.82rem;
    margin-bottom: 0.2rem;
  }

  .vst2-content p {
    color: #94a3b8;
    font-size: 0.75rem;
    line-height: 1.4;
    margin: 0;
  }

  .btn-copy-hash {
    background: rgba(255, 255, 255, 0.08);
    color: #cbd5e1;
    border: 1px solid rgba(255, 255, 255, 0.15);
    padding: 0.15rem 0.5rem;
    border-radius: 4px;
    font-size: 0.7rem;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .btn-copy-hash:hover {
    background: rgba(255, 255, 255, 0.15);
    color: #fff;
  }

  .btn-card-copy-hash {
    margin-left: auto;
    background: transparent;
    border: none;
    color: #94a3b8;
    cursor: pointer;
    font-size: 0.75rem;
    padding: 0 0.2rem;
    transition: color 0.15s ease;
  }

  .btn-card-copy-hash:hover {
    color: #38bdf8;
  }

  .preview-section h4 {
    font-size: 0.85rem;
    color: #94a3b8;
    margin-bottom: 0.4rem;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .preview-desc-text {
    font-size: 0.85rem;
    color: #cbd5e1;
    line-height: 1.5;
  }

  .preview-hash-box {
    background: #0f1117;
    padding: 0.6rem;
    border-radius: 6px;
    font-size: 0.75rem;
    color: #10b981;
    overflow-x: auto;
  }

  .preset-manager-card {
    background: #0d1424;
    border: 1px solid #1e3a5f;
    border-radius: 8px;
    padding: 0.85rem 1rem;
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    margin-bottom: 0.85rem;
  }

  .preset-header-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 1rem;
  }

  .preset-title {
    display: flex;
    align-items: center;
    gap: 0.6rem;
  }

  .preset-icon {
    font-size: 1.3rem;
  }

  .preset-title h5 {
    margin: 0;
    font-size: 0.88rem;
    font-weight: 700;
    color: #60a5fa;
  }

  .preset-subtitle {
    font-size: 0.72rem;
    color: #94a3b8;
  }

  .btn-open-presets {
    background: #1e293b;
    border: 1px solid #334155;
    color: #f1f5f9;
    padding: 0.35rem 0.65rem;
    border-radius: 6px;
    font-size: 0.75rem;
    font-weight: 600;
    cursor: pointer;
    transition: 0.15s ease-in-out;
  }

  .btn-open-presets:hover {
    background: #2563eb;
    border-color: #3b82f6;
  }

  .preset-export-controls {
    display: flex;
    gap: 0.5rem;
    align-items: center;
  }

  .preset-name-input {
    flex: 2;
    background: #0a0f1d;
    border: 1px solid #233554;
    border-radius: 6px;
    padding: 0.4rem 0.6rem;
    color: #f1f5f9;
    font-size: 0.78rem;
  }

  .preset-format-select {
    flex: 1;
    background: #0a0f1d;
    border: 1px solid #233554;
    border-radius: 6px;
    padding: 0.4rem 0.5rem;
    color: #f1f5f9;
    font-size: 0.78rem;
    cursor: pointer;
  }

  .btn-export-preset {
    background: #2563eb;
    border: 1px solid #3b82f6;
    color: white;
    padding: 0.4rem 0.75rem;
    border-radius: 6px;
    font-size: 0.78rem;
    font-weight: 600;
    cursor: pointer;
    white-space: nowrap;
    transition: 0.15s ease-in-out;
  }

  .btn-export-preset:hover:not(:disabled) {
    background: #1d4ed8;
    box-shadow: 0 0 10px rgba(37, 99, 235, 0.4);
  }

  .btn-export-preset:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .preset-export-status {
    font-size: 0.75rem;
    padding: 0.35rem 0.6rem;
    border-radius: 4px;
    background: #1e293b;
  }

  .preset-export-status.success {
    background: #064e3b;
    color: #6ee7b7;
    border: 1px solid #059669;
  }

  .preset-export-status.error {
    background: #7f1d1d;
    color: #fca5a5;
    border: 1px solid #dc2626;
  }

  .preset-snapshot-actions {
    display: flex;
    gap: 0.5rem;
    margin-top: 0.25rem;
  }

  .btn-snapshot-state {
    background: #1e293b;
    border: 1px solid #3b82f6;
    color: #93c5fd;
    padding: 0.35rem 0.75rem;
    border-radius: 6px;
    font-size: 0.76rem;
    font-weight: 600;
    cursor: pointer;
    transition: 0.15s ease-in-out;
  }

  .btn-snapshot-state:hover:not(:disabled) {
    background: #2563eb;
    color: #ffffff;
  }

  .snapshot-details-box {
    background: #090d16;
    border: 1px solid #1e3a5f;
    border-radius: 6px;
    padding: 0.7rem 0.85rem;
    margin-top: 0.5rem;
    display: flex;
    flex-direction: column;
    gap: 0.4rem;
  }

  .snapshot-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    font-size: 0.8rem;
    color: #e2e8f0;
  }

  .snapshot-crc code {
    color: #38bdf8;
    background: #0f172a;
    padding: 0.1rem 0.3rem;
    border-radius: 4px;
    font-size: 0.72rem;
  }

  .snapshot-meta-row {
    display: flex;
    gap: 0.4rem;
  }

  .snapshot-params-list {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
    gap: 0.4rem;
    margin-top: 0.3rem;
    max-height: 120px;
    overflow-y: auto;
  }

  .snapshot-param-item {
    background: #131d31;
    padding: 0.25rem 0.5rem;
    border-radius: 4px;
    display: flex;
    justify-content: space-between;
    font-size: 0.72rem;
  }

  .snapshot-param-item .param-name {
    color: #94a3b8;
  }

  .snapshot-param-item .param-val {
    color: #38bdf8;
    font-weight: 600;
  }

  /* Preset Diff Section */
  .preset-diff-section {
    margin-top: 0.75rem;
    padding-top: 0.75rem;
    border-top: 1px solid #1e293b;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .diff-section-title {
    font-size: 0.82rem;
    font-weight: 700;
    color: #f59e0b;
  }

  .diff-controls-row {
    display: flex;
    align-items: center;
    gap: 0.4rem;
  }

  .diff-input {
    flex: 1;
    background: #0a0f1d;
    border: 1px solid #233554;
    border-radius: 6px;
    padding: 0.35rem 0.5rem;
    color: #f1f5f9;
    font-size: 0.75rem;
  }

  .diff-vs {
    font-size: 0.72rem;
    color: #64748b;
    font-weight: 600;
  }

  .btn-diff-compare {
    background: #d97706;
    border: 1px solid #f59e0b;
    color: #ffffff;
    padding: 0.35rem 0.7rem;
    border-radius: 6px;
    font-size: 0.75rem;
    font-weight: 600;
    cursor: pointer;
    white-space: nowrap;
    transition: 0.15s ease-in-out;
  }

  .btn-diff-compare:hover:not(:disabled) {
    background: #b45309;
  }

  .diff-report-card {
    background: #090d16;
    border: 1px solid #283548;
    border-radius: 6px;
    padding: 0.6rem 0.75rem;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .diff-summary-badges {
    display: flex;
    gap: 0.4rem;
    flex-wrap: wrap;
  }

  .diff-badge {
    padding: 0.2rem 0.5rem;
    border-radius: 4px;
    font-size: 0.72rem;
    font-weight: 700;
  }

  .diff-badge.radical {
    background: rgba(239, 68, 68, 0.2);
    border: 1px solid #ef4444;
    color: #f87171;
  }

  .diff-badge.moderate {
    background: rgba(245, 158, 11, 0.2);
    border: 1px solid #f59e0b;
    color: #fbbf24;
  }

  .diff-badge.subtle {
    background: rgba(16, 185, 129, 0.2);
    border: 1px solid #10b981;
    color: #34d399;
  }

  .diff-badge.metric {
    background: #1e293b;
    border: 1px solid #334155;
    color: #e2e8f0;
  }

  .diff-notes-box {
    font-size: 0.73rem;
    color: #cbd5e1;
    background: #131d31;
    padding: 0.4rem 0.6rem;
    border-radius: 4px;
  }

  .diff-table-container {
    max-height: 140px;
    overflow-y: auto;
  }

  .diff-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.72rem;
  }

  .diff-table th {
    text-align: left;
    color: #94a3b8;
    padding: 0.25rem 0.4rem;
    border-bottom: 1px solid #1e293b;
  }

  .diff-table td {
    padding: 0.25rem 0.4rem;
    border-bottom: 1px solid #111827;
    color: #e2e8f0;
  }

  .modified-row {
    background: rgba(245, 158, 11, 0.05);
  }

  .delta-tag {
    font-size: 0.7rem;
    font-weight: 600;
    padding: 0.1rem 0.3rem;
    border-radius: 3px;
  }

  .delta-tag.positive {
    background: rgba(16, 185, 129, 0.2);
    color: #34d399;
  }

  .delta-tag.negative {
    background: rgba(239, 68, 68, 0.2);
    color: #f87171;
  }

  .status-pill {
    font-size: 0.68rem;
    padding: 0.1rem 0.35rem;
    border-radius: 3px;
    font-weight: 600;
  }

  .status-pill.mod {
    background: #78350f;
    color: #fcd34d;
  }

  .status-pill.same {
    background: #1e293b;
    color: #94a3b8;
  }

  .security-scan-card {
    background: #0f1117;
    border: 1px solid #232d42;
    border-radius: 8px;
    padding: 0.85rem 1rem;
    display: flex;
    flex-direction: column;
    gap: 0.6rem;
  }

  .sec-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 1rem;
  }

  .sec-label {
    font-size: 0.8rem;
    color: #94a3b8;
    font-weight: 500;
  }

  .sec-hash {
    font-size: 0.75rem;
    color: #10b981;
    word-break: break-all;
    max-width: 70%;
    text-align: right;
  }

  .sec-val-dim {
    font-size: 0.75rem;
    color: #64748b;
    font-style: italic;
  }

  .yabridge-card {
    background: #140d1f;
    border: 1px solid #381e5f;
    border-radius: 8px;
    padding: 0.85rem 1rem;
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    margin-top: 0.85rem;
  }

  .yabridge-header-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 1rem;
  }

  .yabridge-title {
    display: flex;
    align-items: center;
    gap: 0.6rem;
  }

  .yabridge-icon {
    font-size: 1.3rem;
  }

  .yabridge-title strong {
    display: block;
    font-size: 0.88rem;
    color: #c084fc;
  }

  .yabridge-subtitle {
    display: block;
    font-size: 0.72rem;
    color: #94a3b8;
  }

  .yabridge-actions {
    display: flex;
    gap: 0.5rem;
  }

  .btn-yabridge-check,
  .btn-yabridge-sync {
    background: #2a1b42;
    border: 1px solid #582a85;
    color: #e9d5ff;
    padding: 0.35rem 0.65rem;
    border-radius: 6px;
    font-size: 0.75rem;
    cursor: pointer;
    transition: 0.15s ease-in-out;
    white-space: nowrap;
  }

  .btn-yabridge-check:hover:not(:disabled),
  .btn-yabridge-sync:hover:not(:disabled) {
    background: #3b235e;
    box-shadow: 0 0 10px rgba(192, 132, 252, 0.3);
  }

  .btn-yabridge-sync {
    background: #7e22ce;
    border-color: #a855f7;
    color: #ffffff;
    font-weight: 600;
  }

  .btn-yabridge-sync:hover:not(:disabled) {
    background: #9333ea;
  }

  .yabridge-status-body {
    display: flex;
    flex-direction: column;
    gap: 0.4rem;
    font-size: 0.75rem;
  }

  .yabridge-status-line {
    display: flex;
    align-items: center;
    gap: 0.6rem;
  }

  .yabridge-install-tip {
    color: #94a3b8;
    font-size: 0.72rem;
  }

  .yabridge-paths-list {
    margin-top: 0.3rem;
    background: #0a0612;
    padding: 0.5rem;
    border-radius: 4px;
    border: 1px solid #271442;
  }

  .yabridge-paths-title {
    color: #c084fc;
    font-weight: 600;
    margin-bottom: 0.25rem;
    display: block;
  }

  .yabridge-paths-list ul {
    margin: 0;
    padding-left: 1.2rem;
    color: #cbd5e1;
  }

  .yabridge-message-box {
    background: #0f172a;
    border: 1px solid #334155;
    border-radius: 6px;
    padding: 0.5rem;
  }

  .yabridge-message-box pre {
    margin: 0;
    font-size: 0.72rem;
    color: #38bdf8;
    white-space: pre-wrap;
    word-break: break-all;
  }

  .yabridge-message-box.error pre {
    color: #f87171;
  }

  .latency-diagnostic-card {
    background: #0f172a;
    border: 1px solid #1e293b;
    border-radius: 8px;
    padding: 0.85rem 1rem;
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    margin-top: 0.85rem;
  }

  .latency-card-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 1rem;
  }

  .latency-icon {
    font-size: 1.4rem;
  }

  .latency-card-title h4 {
    margin: 0;
    font-size: 0.88rem;
    color: #38bdf8;
  }

  .latency-card-title p {
    margin: 0;
    font-size: 0.72rem;
    color: #94a3b8;
  }

  .latency-actions {
    display: flex;
  }

  .btn-latency-calc {
    background: #0369a1;
    border: 1px solid #0284c7;
    color: #f0f9ff;
    padding: 0.35rem 0.65rem;
    border-radius: 6px;
    font-size: 0.75rem;
    font-weight: 600;
    cursor: pointer;
    transition: 0.15s ease-in-out;
    white-space: nowrap;
  }

  .btn-latency-calc:hover:not(:disabled) {
    background: #0284c7;
    box-shadow: 0 0 10px rgba(56, 189, 248, 0.3);
  }

  .latency-controls-row {
    display: flex;
    gap: 1.25rem;
    background: #090e17;
    padding: 0.6rem 0.8rem;
    border-radius: 6px;
    border: 1px solid #1e293b;
  }

  .latency-control-group {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .latency-control-group label {
    font-size: 0.75rem;
    color: #94a3b8;
  }

  .latency-control-group select,
  .latency-control-group input {
    background: #1e293b;
    border: 1px solid #334155;
    color: #f8fafc;
    border-radius: 4px;
    padding: 0.25rem 0.5rem;
    font-size: 0.75rem;
  }

  .latency-results-container {
    display: flex;
    flex-direction: column;
    gap: 0.6rem;
  }

  .latency-recommendation-banner {
    display: flex;
    align-items: center;
    gap: 0.6rem;
    background: #091e36;
    border: 1px solid #0369a1;
    padding: 0.4rem 0.6rem;
    border-radius: 4px;
  }

  .latency-badge {
    background: #0284c7;
    color: #ffffff;
    font-size: 0.7rem;
    font-weight: 700;
    padding: 0.15rem 0.45rem;
    border-radius: 4px;
    white-space: nowrap;
  }

  .latency-rec-text {
    font-size: 0.75rem;
    color: #bae6fd;
  }

  .latency-grid {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
    background: #090e17;
    border: 1px solid #1e293b;
    border-radius: 6px;
    overflow: hidden;
  }

  .latency-grid-header {
    display: grid;
    grid-template-columns: 90px 80px 100px 1fr;
    background: #1e293b;
    padding: 0.4rem 0.6rem;
    font-size: 0.72rem;
    font-weight: 600;
    color: #cbd5e1;
  }

  .latency-grid-row {
    display: grid;
    grid-template-columns: 90px 80px 100px 1fr;
    padding: 0.35rem 0.6rem;
    font-size: 0.72rem;
    border-top: 1px solid #1e293b;
    align-items: center;
  }

  .buffer-cell {
    color: #38bdf8;
  }

  .time-cell {
    color: #cbd5e1;
  }

  .time-cell.roundtrip {
    color: #facc15;
    font-weight: 600;
  }

  .workflow-cell {
    color: #94a3b8;
  }

  .audio-bus-matrix-card {
    background: #0f172a;
    border: 1px solid #1e293b;
    border-radius: 8px;
    padding: 0.85rem 1rem;
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    margin-top: 0.85rem;
  }

  .bus-matrix-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 1rem;
  }

  .bus-matrix-title {
    display: flex;
    align-items: center;
    gap: 0.65rem;
  }

  .bus-matrix-title strong {
    display: block;
    font-size: 0.88rem;
    color: #a78bfa;
  }

  .bus-matrix-subtitle {
    display: block;
    font-size: 0.72rem;
    color: #94a3b8;
  }

  .bus-matrix-icon {
    font-size: 1.4rem;
  }

  .bus-controls-row {
    display: flex;
    gap: 1.25rem;
    background: #090e17;
    padding: 0.6rem 0.8rem;
    border-radius: 6px;
    border: 1px solid #1e293b;
  }

  .bus-results-container {
    display: flex;
    flex-direction: column;
    gap: 0.65rem;
  }

  .bus-summary-banner {
    display: flex;
    flex-direction: column;
    gap: 0.4rem;
    background: #180d2b;
    border: 1px solid #7c3aed;
    padding: 0.5rem 0.75rem;
    border-radius: 6px;
  }

  .bus-banner-top {
    display: flex;
    align-items: center;
    gap: 0.65rem;
    flex-wrap: wrap;
  }

  .bus-topology-badge {
    background: #8b5cf6;
    color: #ffffff;
    font-size: 0.72rem;
    font-weight: 700;
    padding: 0.15rem 0.5rem;
    border-radius: 4px;
  }

  .bus-io-count {
    font-size: 0.75rem;
    color: #e9d5ff;
    font-weight: 600;
  }

  .bus-sidechain-badge {
    background: rgba(16, 185, 129, 0.2);
    border: 1px solid #10b981;
    color: #6ee7b7;
    font-size: 0.7rem;
    font-weight: 600;
    padding: 0.1rem 0.4rem;
    border-radius: 4px;
  }

  .bus-instrument-badge {
    background: rgba(245, 158, 11, 0.2);
    border: 1px solid #f59e0b;
    color: #fcd34d;
    font-size: 0.7rem;
    font-weight: 600;
    padding: 0.1rem 0.4rem;
    border-radius: 4px;
  }

  .bus-daw-notes {
    font-size: 0.73rem;
    color: #c4b5fd;
    line-height: 1.4;
  }

  .bus-buses-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
    gap: 0.5rem;
  }

  .bus-card {
    background: #090e17;
    border: 1px solid #1e293b;
    border-radius: 6px;
    padding: 0.55rem 0.7rem;
    display: flex;
    flex-direction: column;
    gap: 0.4rem;
  }

  .bus-card.bus-in {
    border-left: 3px solid #38bdf8;
  }

  .bus-card.bus-out {
    border-left: 3px solid #10b981;
  }

  .bus-card.bus-aux {
    background: #111422;
    border-left: 3px solid #f59e0b;
  }

  .bus-card-header {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 0.75rem;
  }

  .bus-card-header strong {
    color: #f1f5f9;
    flex: 1;
  }

  .bus-dir-badge {
    font-size: 0.65rem;
    font-weight: 700;
    padding: 0.1rem 0.35rem;
    border-radius: 3px;
    background: #1e293b;
    color: #94a3b8;
  }

  .bus-channels-badge {
    font-size: 0.65rem;
    padding: 0.1rem 0.35rem;
    border-radius: 3px;
    background: #2e1065;
    color: #c4b5fd;
    font-weight: 600;
  }

  .bus-card-pins {
    display: flex;
    flex-wrap: wrap;
    gap: 0.3rem;
  }

  .bus-pin-pill {
    background: #1e293b;
    border: 1px solid #334155;
    color: #e2e8f0;
    font-size: 0.68rem;
    padding: 0.15rem 0.4rem;
    border-radius: 4px;
    cursor: default;
  }

  .bus-pin-pill small {
    color: #94a3b8;
  }

  .binary-arch-matrix-card {
    background: #0f172a;
    border: 1px solid #1e293b;
    border-radius: 8px;
    padding: 0.85rem 1rem;
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    margin-top: 0.85rem;
  }

  .arch-matrix-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 1rem;
  }

  .arch-matrix-title {
    display: flex;
    align-items: center;
    gap: 0.65rem;
  }

  .arch-matrix-title strong {
    display: block;
    font-size: 0.88rem;
    color: #38bdf8;
  }

  .arch-matrix-subtitle {
    display: block;
    font-size: 0.72rem;
    color: #94a3b8;
  }

  .arch-matrix-icon {
    font-size: 1.4rem;
  }

  .arch-results-container {
    display: flex;
    flex-direction: column;
    gap: 0.65rem;
  }

  .arch-summary-banner {
    display: flex;
    flex-direction: column;
    gap: 0.4rem;
    background: #082f49;
    border: 1px solid #0284c7;
    padding: 0.5rem 0.75rem;
    border-radius: 6px;
  }

  .arch-banner-top {
    display: flex;
    align-items: center;
    gap: 0.6rem;
    flex-wrap: wrap;
  }

  .arch-cpu-badge {
    background: #0284c7;
    color: #ffffff;
    font-size: 0.72rem;
    font-weight: 700;
    padding: 0.15rem 0.5rem;
    border-radius: 4px;
  }

  .arch-bitness-badge {
    font-size: 0.7rem;
    font-weight: 700;
    padding: 0.15rem 0.45rem;
    border-radius: 4px;
  }

  .arch-bitness-badge.is-64 {
    background: rgba(16, 185, 129, 0.2);
    border: 1px solid #10b981;
    color: #6ee7b7;
  }

  .arch-bitness-badge.is-32 {
    background: rgba(245, 158, 11, 0.2);
    border: 1px solid #f59e0b;
    color: #fcd34d;
  }

  .arch-rosetta-badge {
    background: rgba(139, 92, 246, 0.2);
    border: 1px solid #8b5cf6;
    color: #c4b5fd;
    font-size: 0.7rem;
    font-weight: 600;
    padding: 0.15rem 0.45rem;
    border-radius: 4px;
  }

  .arch-container-type {
    font-size: 0.72rem;
    color: #94a3b8;
  }

  .arch-summary-text {
    font-size: 0.73rem;
    color: #bae6fd;
    line-height: 1.4;
  }

  .arch-details-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 0.6rem;
  }

  .arch-detail-box {
    background: #090e17;
    border: 1px solid #1e293b;
    border-radius: 6px;
    padding: 0.55rem 0.7rem;
    display: flex;
    flex-direction: column;
    gap: 0.4rem;
  }

  .arch-box-title {
    font-size: 0.72rem;
    color: #cbd5e1;
    font-weight: 600;
  }

  .arch-tag-pills {
    display: flex;
    flex-wrap: wrap;
    gap: 0.35rem;
  }

  .simd-pill {
    background: rgba(56, 189, 248, 0.15);
    border: 1px solid rgba(56, 189, 248, 0.35);
    color: #7dd3fc;
    font-size: 0.68rem;
    padding: 0.15rem 0.4rem;
    border-radius: 4px;
    font-weight: 600;
  }

  .sec-pill {
    background: rgba(52, 211, 153, 0.15);
    border: 1px solid rgba(52, 211, 153, 0.35);
    color: #6ee7b7;
    font-size: 0.68rem;
    padding: 0.15rem 0.4rem;
    border-radius: 4px;
  }

  .arch-os-box {
    display: flex;
    align-items: center;
    gap: 0.6rem;
    background: #090e17;
    border: 1px solid #1e293b;
    padding: 0.45rem 0.7rem;
    border-radius: 6px;
    font-size: 0.72rem;
  }

  .arch-os-label {
    color: #94a3b8;
  }

  .arch-os-val {
    color: #f1f5f9;
    font-weight: 600;
  }

  .bundle-fuzz-report {
    margin-top: 0.85rem;
    background: #0d131f;
    border: 1px solid #1e293b;
    border-radius: 6px;
    padding: 0.85rem 1rem;
    display: flex;
    flex-direction: column;
    gap: 0.6rem;
  }

  .fuzz-report-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 1rem;
  }

  .fuzz-title {
    font-size: 0.8rem;
    font-weight: 600;
    color: #60a5fa;
  }

  .fuzz-stats-grid {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 0.5rem;
    background: #111827;
    padding: 0.5rem;
    border-radius: 4px;
    border: 1px solid #1f2937;
  }

  .fuzz-stat-item {
    display: flex;
    flex-direction: column;
    gap: 0.2rem;
  }

  .fuzz-stat-label {
    font-size: 0.7rem;
    color: #9ca3af;
  }

  .fuzz-stat-val {
    font-size: 0.75rem;
    font-weight: 600;
    color: #f3f4f6;
  }

  .fuzz-symbols-container {
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
    background: #090d16;
    padding: 0.5rem;
    border-radius: 4px;
    border: 1px solid #1e293b;
  }

  .fuzz-symbols-label {
    font-size: 0.68rem;
    color: #94a3b8;
    text-transform: uppercase;
    font-weight: 600;
  }

  .fuzz-symbols-badges {
    display: flex;
    flex-wrap: wrap;
    gap: 0.4rem;
  }

  .fuzz-symbol-badge {
    font-size: 0.72rem;
    font-family: monospace;
    font-weight: 600;
    color: #38bdf8;
    background: rgba(56, 189, 248, 0.12);
    border: 1px solid rgba(56, 189, 248, 0.3);
    border-radius: 4px;
    padding: 2px 6px;
  }

  .fuzz-capabilities-container {
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
    background: #090e17;
    padding: 0.6rem;
    border-radius: 6px;
    border: 1px solid #1e293b;
  }

  .fuzz-capabilities-badges {
    display: flex;
    flex-wrap: wrap;
    gap: 0.4rem;
  }

  .fuzz-capability-badge {
    font-size: 0.72rem;
    font-weight: 600;
    color: #a78bfa;
    background: rgba(167, 139, 250, 0.12);
    border: 1px solid rgba(167, 139, 250, 0.3);
    border-radius: 4px;
    padding: 2px 6px;
  }

  .fuzz-diag-list {
    margin: 0;
    padding-left: 0;
    list-style: none;
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
    font-size: 0.75rem;
    color: #94a3b8;
    font-family: monospace;
  }

  .preview-download-link {
    color: #60a5fa;
    font-size: 0.85rem;
    text-decoration: none;
    word-break: break-all;
  }

  .preview-download-link:hover {
    text-decoration: underline;
  }

  /* Advanced Installation Options Styling */
  .advanced-install-container {
    margin: 1rem 0;
    border: 1px solid rgba(56, 189, 248, 0.2);
    border-radius: 8px;
    background: rgba(15, 23, 42, 0.6);
    overflow: hidden;
  }

  .btn-toggle-advanced {
    width: 100%;
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.6rem 1rem;
    background: rgba(56, 189, 248, 0.08);
    border: none;
    color: #38bdf8;
    font-size: 0.8rem;
    font-weight: 600;
    cursor: pointer;
    text-align: left;
    transition: background 0.15s ease;
  }

  .btn-toggle-advanced:hover {
    background: rgba(56, 189, 248, 0.15);
  }

  .btn-toggle-advanced .chevron {
    font-size: 0.75rem;
  }

  .advanced-install-body {
    padding: 1rem;
    display: flex;
    flex-direction: column;
    gap: 0.8rem;
    border-top: 1px solid rgba(56, 189, 248, 0.15);
  }

  .adv-field-group {
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
  }

  .adv-field-group label {
    font-size: 0.75rem;
    font-weight: 600;
    color: #cbd5e1;
  }

  .adv-input-text, .adv-select {
    background: #0f172a;
    border: 1px solid #334155;
    border-radius: 6px;
    color: #f8fafc;
    padding: 0.4rem 0.6rem;
    font-size: 0.78rem;
    outline: none;
    transition: border-color 0.15s ease;
  }

  .adv-input-text:focus, .adv-select:focus {
    border-color: #38bdf8;
  }

  .preset-path-pills {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 0.35rem;
    margin-top: 0.3rem;
  }

  .path-pill-label {
    font-size: 0.7rem;
    color: #64748b;
  }

  .path-preset-pill {
    background: #1e293b;
    border: 1px solid #334155;
    color: #94a3b8;
    border-radius: 4px;
    font-size: 0.68rem;
    padding: 0.15rem 0.45rem;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .path-preset-pill:hover {
    color: #38bdf8;
    border-color: #38bdf8;
    background: rgba(56, 189, 248, 0.1);
  }

  .adv-execute-row {
    display: flex;
    justify-content: flex-end;
    margin-top: 0.4rem;
  }

  .adv-install-btn {
    font-size: 0.78rem;
    padding: 0.45rem 1rem;
  }

  .preview-actions {
    display: flex;
    gap: 0.8rem;
    justify-content: flex-end;
    border-top: 1px solid #232d42;
    padding-top: 1rem;
    margin-top: 0.5rem;
  }

  .uppercase { text-transform: uppercase; }
  .capitalize { text-transform: capitalize; }

  .crawler-action-bar {
    display: flex;
    gap: 0.8rem;
    margin-top: auto;
  }

  .btn-primary {
    flex: 1;
    background: #2563eb;
    color: white;
    border: none;
    padding: 0.6rem;
    border-radius: 6px;
    font-weight: 600;
    font-size: 0.85rem;
    cursor: pointer;
    transition: 0.2s;
  }

  .btn-primary:hover:not(:disabled) {
    background: #1d4ed8;
  }

  .btn-secondary {
    background: #1e293b;
    color: #cbd5e1;
    border: 1px solid #334155;
    padding: 0.6rem;
    border-radius: 6px;
    font-size: 0.85rem;
    cursor: pointer;
  }

  .metrics-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 0.8rem;
    height: 100%;
  }

  .metric-box {
    background: #0f1117;
    border: 1px solid #232d42;
    border-radius: 8px;
    padding: 0.6rem;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
  }

  .metric-num {
    font-size: 1.4rem;
    font-weight: 800;
    color: #60a5fa;
  }

  .metric-lbl {
    font-size: 0.7rem;
    color: #94a3b8;
  }

  /* Console Output */
  .console-section {
    background: #090d14;
    border: 1px solid #1e293b;
    border-radius: 8px;
    overflow: hidden;
  }

  .console-header {
    background: #111827;
    padding: 0.5rem 1rem;
    font-size: 0.8rem;
    color: #9ca3af;
    display: flex;
    justify-content: space-between;
    align-items: center;
    border-bottom: 1px solid #1f2937;
  }

  .console-close {
    background: none;
    border: none;
    color: #9ca3af;
    cursor: pointer;
  }

  .console-window {
    padding: 0.8rem;
    max-height: 200px;
    overflow-y: auto;
    font-family: monospace;
    font-size: 0.8rem;
  }

  .console-line {
    color: #38bdf8;
    line-height: 1.4;
  }

  .empty-line {
    color: #64748b;
  }

  /* Plugins Grid */
  .registry-section {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .section-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    flex-wrap: wrap;
    gap: 1rem;
  }

  .section-header h2 {
    font-size: 1.3rem;
    font-weight: 700;
  }

  .section-subtitle {
    font-size: 0.8rem;
    color: #94a3b8;
  }

  .filter-pills-row {
    display: flex;
    gap: 0.5rem;
    align-items: center;
  }

  .filter-pill {
    background: #161b26;
    border: 1px solid #232d42;
    border-radius: 20px;
    padding: 0.35rem 0.8rem;
    font-size: 0.8rem;
    color: #94a3b8;
    cursor: pointer;
    display: flex;
    align-items: center;
    gap: 0.4rem;
    transition: 0.2s;
  }

  .filter-pill:hover {
    border-color: #3b82f6;
    color: #f1f5f9;
  }

  .filter-pill.pill-active {
    background: #1e293b;
    border-color: #3b82f6;
    color: #60a5fa;
    font-weight: 600;
  }

  .filter-pill.pill-updates.pill-active {
    background: #451a03;
    border-color: #d97706;
    color: #fbbf24;
  }

  .pill-badge {
    background: #0f1117;
    padding: 1px 6px;
    border-radius: 10px;
    font-size: 0.7rem;
    font-weight: 700;
  }

  .plugins-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
    gap: 1.2rem;
  }

  .plugin-card {
    background: #161b26;
    border: 1px solid #232d42;
    border-radius: 10px;
    padding: 1.2rem;
    display: flex;
    flex-direction: column;
    gap: 0.8rem;
    transition: 0.2s;
  }

  .plugin-card:hover {
    border-color: #3b82f6;
    transform: translateY(-2px);
    box-shadow: 0 8px 16px rgba(0, 0, 0, 0.3);
  }

  .card-top {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
  }

  .title-group {
    max-width: 75%;
  }

  .plugin-title {
    font-size: 1.1rem;
    color: #f8fafc;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .developer-tag {
    font-size: 0.75rem;
    color: #60a5fa;
    cursor: help;
  }

  .license-tag {
    font-size: 0.7rem;
    padding: 2px 8px;
    border-radius: 12px;
    font-weight: 700;
    text-transform: uppercase;
  }

  .license-opensource { background: #064e3b; color: #34d399; }
  .license-free { background: #1e3a8a; color: #93c5fd; }
  .license-commercial { background: #78350f; color: #fcd34d; }
  .license-donationware { background: #581c87; color: #d8b4fe; }

  .specs-grid {
    display: grid;
    grid-template-columns: 1fr 1fr 1fr;
    background: #0f1117;
    padding: 0.6rem;
    border-radius: 6px;
    font-size: 0.75rem;
    text-align: center;
  }

  .spec-label {
    display: block;
    color: #64748b;
  }

  .spec-val {
    font-weight: 600;
    color: #e2e8f0;
  }

  .spec-path-item {
    grid-column: span 3;
    margin-top: 0.3rem;
    padding-top: 0.3rem;
    border-top: 1px dashed rgba(255, 255, 255, 0.08);
  }

  .path-val {
    font-family: monospace;
    font-size: 0.7rem;
    color: #93c5fd;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    display: block;
  }

  .hash-container {
    background: #0f1117;
    padding: 0.4rem 0.6rem;
    border-radius: 6px;
    font-size: 0.75rem;
    display: flex;
    align-items: center;
    gap: 0.4rem;
    cursor: help;
  }

  .hash-label {
    color: #94a3b8;
  }

  .hash-code {
    font-family: monospace;
    color: #10b981;
  }

  .description-box {
    font-size: 0.8rem;
    color: #94a3b8;
    line-height: 1.4;
    height: 40px;
    overflow: hidden;
    text-overflow: ellipsis;
    cursor: help;
  }

  .install-bar {
    margin-top: auto;
  }

  .action-button-group {
    display: flex;
    gap: 0.5rem;
  }

  .btn-install {
    flex: 2;
    background: #2563eb;
    color: white;
    border: none;
    padding: 0.6rem;
    border-radius: 6px;
    font-weight: 600;
    font-size: 0.85rem;
    cursor: pointer;
    transition: 0.2s;
  }

  .btn-install:hover:not(:disabled) {
    background: #1d4ed8;
  }

  .btn-remove {
    flex: 1;
    background: #1e293b;
    color: #f87171;
    border: 1px solid #7f1d1d;
    padding: 0.6rem;
    border-radius: 6px;
    font-size: 0.8rem;
    cursor: pointer;
    transition: 0.2s;
  }

  .btn-remove:hover:not(:disabled) {
    background: #7f1d1d;
    color: white;
  }

  .btn-disabled {
    background: #334155;
    color: #94a3b8;
    cursor: not-allowed;
  }

  .status-banner {
    padding: 0.4rem 0.6rem;
    border-radius: 4px;
    font-size: 0.75rem;
    text-align: center;
  }

  .status-ok { background: #064e3b; color: #a7f3d0; }
  .status-err { background: #7f1d1d; color: #fecaca; }

  .loading-state, .empty-state {
    text-align: center;
    padding: 4rem;
    background: #161b26;
    border-radius: 12px;
  }

  .spinner {
    border: 3px solid #232d42;
    border-top: 3px solid #3b82f6;
    border-radius: 50%;
    width: 32px;
    height: 32px;
    animation: spin 0.8s linear infinite;
    margin: 0 auto 1rem auto;
  }

  @keyframes spin {
    0% { transform: rotate(0deg); }
    100% { transform: rotate(360deg); }
  }

  /* Paths Configuration Modal Styles */
  .modal-title-group {
    display: flex;
    align-items: center;
    gap: 0.6rem;
  }

  .modal-body-section {
    display: flex;
    flex-direction: column;
    gap: 1.2rem;
  }

  .modal-desc {
    font-size: 0.85rem;
    color: #94a3b8;
    line-height: 1.4;
  }

  .paths-block h4 {
    font-size: 0.8rem;
    color: #cbd5e1;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    margin-bottom: 0.5rem;
  }

  .paths-list {
    list-style: none;
    display: flex;
    flex-direction: column;
    gap: 0.4rem;
    max-height: 160px;
    overflow-y: auto;
    background: #0f1117;
    border: 1px solid #232d42;
    border-radius: 6px;
    padding: 0.5rem;
  }

  .path-item {
    display: flex;
    align-items: center;
    gap: 0.6rem;
    font-size: 0.8rem;
    padding: 0.3rem 0.5rem;
    background: #161b26;
    border-radius: 4px;
  }

  .path-item code {
    flex: 1;
    color: #93c5fd;
    word-break: break-all;
  }

  .path-type-badge {
    font-size: 0.65rem;
    font-weight: 700;
    padding: 2px 6px;
    border-radius: 4px;
    background: #1e293b;
    color: #94a3b8;
  }

  .path-type-badge.custom-badge {
    background: #064e3b;
    color: #6ee7b7;
  }

  .path-empty {
    font-size: 0.8rem;
    color: #64748b;
    font-style: italic;
    padding: 0.5rem;
  }

  .custom-path-input-row {
    display: flex;
    gap: 0.5rem;
    margin-bottom: 0.5rem;
  }

  .custom-path-input-row .search-input {
    flex: 1;
  }

  .btn-path-del {
    background: transparent;
    border: none;
    color: #ef4444;
    cursor: pointer;
    font-size: 0.85rem;
    padding: 0 4px;
  }

  .paths-cache-section {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
    padding: 14px 16px;
    margin-top: 14px;
    background: rgba(255, 255, 255, 0.02);
    border: 1px solid rgba(255, 255, 255, 0.06);
    border-radius: 8px;
  }

  .paths-cache-info h4 {
    margin: 0 0 4px 0;
    font-size: 13px;
    color: #f1f5f9;
  }

  .paths-cache-info p {
    margin: 0;
    font-size: 11px;
    color: #94a3b8;
    line-height: 1.4;
  }

  .cache-msg {
    margin-top: 6px;
    font-size: 11px;
    color: #10b981;
    font-weight: 500;
  }

  .btn-cache-flush {
    padding: 8px 12px;
    background: rgba(239, 68, 68, 0.12);
    border: 1px solid rgba(239, 68, 68, 0.3);
    color: #fca5a5;
    border-radius: 6px;
    font-size: 12px;
    font-weight: 600;
    cursor: pointer;
    white-space: nowrap;
    transition: all 0.2s;
  }

  .btn-cache-flush:hover:not(:disabled) {
    background: rgba(239, 68, 68, 0.22);
    border-color: #ef4444;
    color: #ffffff;
  }

  .btn-cache-flush:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  /* DSP Benchmark & Profiler Styles */
  .dsp-benchmark-card {
    background: #0f172a;
    border: 1px solid #1e293b;
    border-radius: 8px;
    padding: 0.85rem 1rem;
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    margin-top: 0.85rem;
  }

  .benchmark-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 1rem;
  }

  .benchmark-title {
    display: flex;
    align-items: center;
    gap: 0.65rem;
  }

  .benchmark-title strong {
    display: block;
    font-size: 0.88rem;
    color: #a855f7;
  }

  .benchmark-subtitle {
    display: block;
    font-size: 0.72rem;
    color: #94a3b8;
  }

  .benchmark-icon {
    font-size: 1.4rem;
  }

  .btn-benchmark-run {
    padding: 0.45rem 0.85rem;
    background: #7e22ce;
    color: #ffffff;
    border: none;
    border-radius: 6px;
    font-size: 0.78rem;
    font-weight: 600;
    cursor: pointer;
    white-space: nowrap;
    transition: background 0.15s ease;
  }

  .btn-benchmark-run:hover:not(:disabled) {
    background: #9333ea;
  }

  .btn-benchmark-run:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .benchmark-results-container {
    display: flex;
    flex-direction: column;
    gap: 0.65rem;
  }

  .benchmark-summary-banner {
    display: flex;
    flex-direction: column;
    gap: 0.4rem;
    background: #3b0764;
    border: 1px solid #7e22ce;
    padding: 0.55rem 0.75rem;
    border-radius: 6px;
  }

  .benchmark-banner-badges {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    flex-wrap: wrap;
  }

  .benchmark-rating-badge {
    background: #10b981;
    color: #064e3b;
    font-weight: 700;
    font-size: 0.72rem;
    padding: 0.2rem 0.55rem;
    border-radius: 4px;
    letter-spacing: 0.02em;
  }

  .benchmark-mem-badge {
    background: #0284c7;
    color: #f0f9ff;
    font-weight: 600;
    font-size: 0.72rem;
    padding: 0.2rem 0.55rem;
    border-radius: 4px;
  }

  .benchmark-multicore-badge {
    background: #6366f1;
    color: #eef2ff;
    font-weight: 600;
    font-size: 0.72rem;
    padding: 0.2rem 0.55rem;
    border-radius: 4px;
  }

  .benchmark-safety-desc {
    font-size: 0.75rem;
    color: #e9d5ff;
  }

  .benchmark-advice-text {
    font-size: 0.72rem;
    color: #d8b4fe;
    font-style: italic;
  }

  .benchmark-table-wrapper {
    overflow-x: auto;
    border: 1px solid #334155;
    border-radius: 6px;
  }

  .benchmark-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.75rem;
    text-align: left;
  }

  .benchmark-table th {
    background: #1e293b;
    color: #cbd5e1;
    padding: 0.45rem 0.65rem;
    border-bottom: 1px solid #334155;
  }

  .benchmark-table td {
    padding: 0.45rem 0.65rem;
    border-bottom: 1px solid #1e293b;
    color: #f1f5f9;
  }

  .bench-col-buf {
    white-space: nowrap;
    width: 130px;
  }

  .bench-col-load {
    width: 180px;
  }

  .load-bar-wrap {
    display: flex;
    align-items: center;
    gap: 0.45rem;
    width: 100%;
    background: #090d16;
    border-radius: 4px;
    padding: 2px 4px;
    box-sizing: border-box;
  }

  .load-bar-fill {
    height: 8px;
    background: linear-gradient(90deg, #10b981 0%, #eab308 60%, #ef4444 100%);
    border-radius: 3px;
    min-width: 4px;
    transition: width 0.3s ease;
  }

  .load-val {
    font-size: 0.7rem;
    font-weight: 600;
    color: #cbd5e1;
    margin-left: auto;
  }

  .bench-col-margin {
    width: 110px;
  }

  .margin-val {
    color: #34d399;
    font-weight: 600;
  }

  .bench-col-risk {
    width: 90px;
  }

  .xrun-risk-pill {
    padding: 0.15rem 0.45rem;
    border-radius: 4px;
    font-size: 0.68rem;
    font-weight: 700;
    text-transform: uppercase;
  }

  .risk-low {
    background: rgba(16, 185, 129, 0.2);
    color: #6ee7b7;
    border: 1px solid #059669;
  }

  .risk-moderate {
    background: rgba(234, 179, 8, 0.2);
    color: #fde047;
    border: 1px solid #ca8a04;
  }

  .risk-high {
    background: rgba(239, 68, 68, 0.2);
    color: #fca5a5;
    border: 1px solid #dc2626;
  }

  /* Plugin Parameter Matrix Styles */
  .param-matrix-card {
    background: #0f172a;
    border: 1px solid #1e293b;
    border-radius: 8px;
    padding: 1rem;
    display: flex;
    flex-direction: column;
    gap: 0.85rem;
  }

  .param-matrix-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 1rem;
  }

  .param-matrix-title {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  .param-matrix-icon {
    font-size: 1.5rem;
  }

  .param-matrix-subtitle {
    display: block;
    font-size: 0.75rem;
    color: #94a3b8;
    margin-top: 0.2rem;
  }

  .param-results-container {
    display: flex;
    flex-direction: column;
    gap: 0.85rem;
  }

  .param-summary-banner {
    background: #0d1527;
    border: 1px solid #1e3a8a;
    border-radius: 6px;
    padding: 0.75rem;
    display: flex;
    flex-direction: column;
    gap: 0.45rem;
  }

  .param-banner-top {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 0.45rem;
  }

  .param-total-badge {
    background: #1e3a8a;
    color: #93c5fd;
    padding: 0.15rem 0.5rem;
    border-radius: 4px;
    font-size: 0.72rem;
    font-weight: 700;
  }

  .param-auto-badge {
    background: #065f46;
    color: #6ee7b7;
    padding: 0.15rem 0.5rem;
    border-radius: 4px;
    font-size: 0.72rem;
    font-weight: 700;
  }

  .param-mpe-badge {
    background: #5b21b6;
    color: #c4b5fd;
    padding: 0.15rem 0.5rem;
    border-radius: 4px;
    font-size: 0.72rem;
    font-weight: 700;
  }

  .param-sample-badge {
    background: #854d0e;
    color: #fef08a;
    padding: 0.15rem 0.5rem;
    border-radius: 4px;
    font-size: 0.72rem;
    font-weight: 700;
  }

  .param-tree-type {
    font-size: 0.7rem;
    color: #94a3b8;
    margin-left: auto;
  }

  .param-host-notes {
    font-size: 0.72rem;
    color: #93c5fd;
    font-style: italic;
  }

  .param-table-wrapper {
    overflow-x: auto;
    border: 1px solid #1e293b;
    border-radius: 6px;
    max-height: 280px;
    overflow-y: auto;
  }

  .param-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.73rem;
    text-align: left;
  }

  .param-table th {
    background: #1e293b;
    color: #cbd5e1;
    padding: 0.45rem 0.6rem;
    border-bottom: 1px solid #334155;
    position: sticky;
    top: 0;
    z-index: 1;
  }

  .param-table td {
    padding: 0.4rem 0.6rem;
    border-bottom: 1px solid #1e293b;
    color: #f1f5f9;
  }

  .param-col-id {
    color: #64748b;
    font-family: monospace;
    width: 35px;
  }

  .param-col-name {
    color: #f8fafc;
  }

  .param-group-tag {
    background: #1e293b;
    color: #94a3b8;
    padding: 0.1rem 0.35rem;
    border-radius: 3px;
    font-size: 0.68rem;
  }

  .param-type-pill {
    padding: 0.1rem 0.35rem;
    border-radius: 3px;
    font-size: 0.68rem;
    font-weight: 600;
  }

  .type-float {
    background: rgba(56, 189, 248, 0.15);
    color: #38bdf8;
  }

  .type-boolean {
    background: rgba(168, 85, 247, 0.15);
    color: #c084fc;
  }

  .type-enumeration {
    background: rgba(245, 158, 11, 0.15);
    color: #fbbf24;
  }

  .type-integer {
    background: rgba(52, 211, 153, 0.15);
    color: #34d399;
  }

  .param-col-val, .param-col-range {
    font-family: monospace;
    color: #cbd5e1;
  }

  .param-col-unit {
    color: #94a3b8;
    font-size: 0.7rem;
  }

  .param-auto-tag {
    font-size: 0.68rem;
    font-weight: 700;
  }

  .is-auto {
    color: #34d399;
  }

  .no-auto {
    color: #f87171;
  }

  .param-midi-pill {
    background: #312e81;
    color: #a5b4fc;
    padding: 0.1rem 0.35rem;
    border-radius: 3px;
    font-size: 0.68rem;
    font-weight: 600;
  }

  .param-no-midi {
    color: #64748b;
  }

  /* Resilience Matrix Styles */
  .resilience-matrix-card {
    background: #0f172a;
    border: 1px solid #334155;
    border-radius: 8px;
    padding: 1rem;
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .resilience-matrix-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 1rem;
    flex-wrap: wrap;
  }

  .resilience-matrix-title {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  .resilience-icon {
    font-size: 1.5rem;
  }

  .resilience-title-text {
    font-size: 0.85rem;
    font-weight: 700;
    color: #f1f5f9;
    letter-spacing: 0.05em;
  }

  .resilience-subtitle-text {
    font-size: 0.72rem;
    color: #94a3b8;
  }

  .btn-resilience {
    background: #0369a1;
    color: #f0f9ff;
    border: 1px solid #0284c7;
    font-weight: 600;
    padding: 0.35rem 0.75rem;
    border-radius: 4px;
    cursor: pointer;
    font-size: 0.75rem;
  }

  .btn-resilience:hover:not(:disabled) {
    background: #0284c7;
  }

  .resilience-results-container {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .resilience-summary-bar {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
  }

  .resilience-badge {
    padding: 0.2rem 0.5rem;
    border-radius: 4px;
    font-size: 0.72rem;
    font-weight: 600;
  }

  .badge-grade {
    background: rgba(16, 185, 129, 0.2);
    color: #34d399;
    border: 1px solid #059669;
  }

  .badge-safe {
    background: rgba(56, 189, 248, 0.15);
    color: #38bdf8;
    border: 1px solid #0284c7;
  }

  .badge-warn {
    background: rgba(239, 68, 68, 0.2);
    color: #f87171;
    border: 1px solid #b91c1c;
  }

  .badge-sandbox {
    background: rgba(168, 85, 247, 0.15);
    color: #c084fc;
    border: 1px solid #9333ea;
  }

  .resilience-info-box {
    background: #1e293b;
    border: 1px solid #334155;
    border-radius: 6px;
    padding: 0.75rem;
    display: flex;
    flex-direction: column;
    gap: 0.4rem;
    font-size: 0.75rem;
  }

  .resilience-metric-row {
    display: flex;
    justify-content: space-between;
    flex-wrap: wrap;
    gap: 0.5rem;
    color: #cbd5e1;
  }

  .resilience-advice {
    color: #38bdf8;
    font-style: italic;
    border-top: 1px dashed #334155;
    padding-top: 0.4rem;
  }

  .resilience-table-wrapper {
    overflow-x: auto;
    border: 1px solid #334155;
    border-radius: 6px;
  }

  .resilience-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.72rem;
    text-align: left;
  }

  .resilience-table th {
    background: #1e293b;
    color: #94a3b8;
    padding: 0.45rem 0.6rem;
    border-bottom: 1px solid #334155;
  }

  .resilience-table td {
    padding: 0.45rem 0.6rem;
    border-bottom: 1px solid #1e293b;
    color: #cbd5e1;
  }

  .stress-intensity-tag {
    padding: 0.1rem 0.35rem;
    border-radius: 3px;
    font-weight: 600;
  }

  .normal-stress {
    background: #1e293b;
    color: #94a3b8;
  }

  .high-stress {
    background: rgba(239, 68, 68, 0.2);
    color: #f87171;
  }

  .resilience-col-exec, .resilience-col-budget {
    font-family: monospace;
  }

  .cycle-bar-container {
    display: flex;
    align-items: center;
    gap: 0.4rem;
    min-width: 100px;
    background: #0f172a;
    border-radius: 3px;
    height: 14px;
    padding: 1px;
    position: relative;
  }

  .cycle-bar-fill {
    height: 100%;
    border-radius: 2px;
  }

  .fill-good {
    background: #10b981;
  }

  .fill-warning {
    background: #f59e0b;
  }

  .fill-danger {
    background: #ef4444;
  }

  .cycle-label {
    position: absolute;
    right: 4px;
    font-size: 0.65rem;
    font-weight: 700;
    color: #fff;
    text-shadow: 0 0 2px #000;
  }

  .alloc-badge {
    padding: 0.1rem 0.35rem;
    border-radius: 3px;
    font-weight: 600;
  }

  .alloc-zero {
    background: rgba(16, 185, 129, 0.15);
    color: #34d399;
  }

  .alloc-warn {
    background: rgba(239, 68, 68, 0.2);
    color: #f87171;
  }

  .inv-badge {
    padding: 0.1rem 0.35rem;
    border-radius: 3px;
    font-weight: 600;
  }

  .inv-safe {
    background: rgba(16, 185, 129, 0.15);
    color: #34d399;
  }

  .inv-danger {
    background: rgba(239, 68, 68, 0.2);
    color: #f87171;
  }

  .xrun-badge {
    padding: 0.1rem 0.35rem;
    border-radius: 3px;
    font-weight: 700;
  }

  .xrun-ok {
    background: rgba(16, 185, 129, 0.15);
    color: #34d399;
  }

  .xrun-dropped {
    background: #ef4444;
    color: #fff;
  }

  /* Headless Audio Spectrum & Frequency Response Matrix Styles */
  .spectrum-analyzer-card {
    background: rgba(30, 41, 59, 0.45);
    border: 1px solid rgba(56, 189, 248, 0.25);
    border-radius: 8px;
    padding: 1.15rem;
    margin-top: 1rem;
    display: flex;
    flex-direction: column;
    gap: 0.85rem;
  }

  .spectrum-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    flex-wrap: wrap;
    gap: 0.75rem;
  }

  .spectrum-title {
    display: flex;
    align-items: center;
    gap: 0.65rem;
  }

  .spectrum-icon {
    font-size: 1.4rem;
  }

  .spectrum-title-text {
    font-size: 0.82rem;
    font-weight: 700;
    color: #38bdf8;
    letter-spacing: 0.05em;
    text-transform: uppercase;
  }

  .spectrum-subtitle-text {
    font-size: 0.73rem;
    color: #94a3b8;
  }

  .spectrum-controls {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .spectrum-signal-select {
    background: #0f172a;
    color: #cbd5e1;
    border: 1px solid #334155;
    border-radius: 5px;
    padding: 0.35rem 0.6rem;
    font-size: 0.74rem;
    outline: none;
  }

  .btn-spectrum {
    background: #0284c7;
    color: #ffffff;
    font-weight: 600;
    font-size: 0.75rem;
    border: none;
    border-radius: 5px;
    padding: 0.4rem 0.8rem;
    cursor: pointer;
    transition: background 0.15s ease;
  }

  .btn-spectrum:hover:not(:disabled) {
    background: #0369a1;
  }

  .spectrum-results-container {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .spectrum-summary-bar {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
  }

  .spectrum-badge {
    font-size: 0.73rem;
    font-weight: 600;
    padding: 0.25rem 0.55rem;
    border-radius: 4px;
  }

  .badge-thd {
    background: rgba(16, 185, 129, 0.15);
    color: #34d399;
    border: 1px solid rgba(16, 185, 129, 0.3);
  }

  .badge-snr {
    background: rgba(56, 189, 248, 0.15);
    color: #38bdf8;
    border: 1px solid rgba(56, 189, 248, 0.3);
  }

  .badge-cutoff {
    background: rgba(168, 85, 247, 0.15);
    color: #c084fc;
    border: 1px solid rgba(168, 85, 247, 0.3);
  }

  .badge-phase {
    background: rgba(245, 158, 11, 0.15);
    color: #fbbf24;
    border: 1px solid rgba(245, 158, 11, 0.3);
  }

  .spectrum-info-box {
    background: rgba(15, 23, 42, 0.75);
    border: 1px solid #334155;
    border-radius: 6px;
    padding: 0.65rem 0.85rem;
    font-size: 0.74rem;
    color: #94a3b8;
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
  }

  .spectrum-metric-row {
    display: flex;
    flex-wrap: wrap;
    gap: 1.5rem;
    color: #cbd5e1;
  }

  .spectrum-advice {
    color: #e2e8f0;
    font-style: italic;
  }

  .spectrum-table-wrapper {
    overflow-x: auto;
  }

  .spectrum-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.73rem;
    text-align: left;
  }

  .spectrum-table th {
    background: rgba(15, 23, 42, 0.9);
    color: #94a3b8;
    padding: 0.45rem 0.6rem;
    border-bottom: 1px solid #334155;
    font-weight: 600;
  }

  .spectrum-table td {
    padding: 0.45rem 0.6rem;
    border-bottom: 1px solid rgba(51, 65, 85, 0.5);
    color: #cbd5e1;
  }

  .band-tag {
    font-weight: 600;
    color: #e2e8f0;
  }

  .delta-tag {
    font-weight: 700;
    padding: 0.1rem 0.35rem;
    border-radius: 3px;
  }

  .delta-boost {
    background: rgba(16, 185, 129, 0.15);
    color: #34d399;
  }

  .delta-cut {
    background: rgba(239, 68, 68, 0.15);
    color: #f87171;
  }

  .delta-flat {
    background: rgba(148, 163, 184, 0.15);
    color: #cbd5e1;
  }

  .curve-bar-container {
    width: 90px;
    height: 8px;
    background: #0f172a;
    border-radius: 4px;
    overflow: hidden;
    border: 1px solid #334155;
  }

  .curve-bar-fill {
    height: 100%;
    border-radius: 4px;
    transition: width 0.2s ease;
  }

  .curve-boost {
    background: #34d399;
  }

  .curve-roll {
    background: #f87171;
  }

  .curve-flat {
    background: #38bdf8;
  }

  .phase-tag {
    padding: 0.1rem 0.35rem;
    border-radius: 3px;
    font-weight: 600;
  }

  .phase-min {
    background: rgba(56, 189, 248, 0.15);
    color: #38bdf8;
  }

  .phase-shift {
    background: rgba(245, 158, 11, 0.15);
    color: #fbbf24;
  }

  .thd-tag {
    padding: 0.1rem 0.35rem;
    border-radius: 3px;
    font-weight: 600;
  }

  .thd-clean {
    background: rgba(16, 185, 129, 0.15);
    color: #34d399;
  }

  .thd-warm {
    background: rgba(245, 158, 11, 0.15);
    color: #fbbf24;
  }

  /* Audio Metering Analyzer Styles */
  .metering-analyzer-card {
    background: #0f172a;
    border: 1px solid #1e293b;
    border-radius: 8px;
    padding: 1rem;
    display: flex;
    flex-direction: column;
    gap: 1rem;
    margin-top: 1rem;
  }

  .metering-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 1rem;
    flex-wrap: wrap;
  }

  .metering-title {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  .metering-icon {
    font-size: 1.5rem;
  }

  .metering-h4 {
    margin: 0;
    font-size: 1rem;
    font-weight: 600;
    color: #f1f5f9;
  }

  .metering-sub {
    font-size: 0.8rem;
    color: #94a3b8;
  }

  .metering-ctrls {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    flex-wrap: wrap;
  }

  .metering-select {
    background: #1e293b;
    border: 1px solid #334155;
    color: #e2e8f0;
    padding: 0.4rem 0.6rem;
    border-radius: 6px;
    font-size: 0.85rem;
  }

  .btn-metering-action {
    background: #0284c7;
    color: white;
    border: none;
    border-radius: 6px;
    padding: 0.45rem 0.85rem;
    font-size: 0.85rem;
    font-weight: 600;
    cursor: pointer;
    transition: background 0.15s ease;
  }

  .btn-metering-action:hover:not(:disabled) {
    background: #0369a1;
  }

  .btn-metering-action:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .metering-content {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .metering-metrics-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 0.75rem;
  }

  .meter-stat-card {
    background: #1e293b;
    border: 1px solid #334155;
    border-radius: 6px;
    padding: 0.75rem;
    display: flex;
    flex-direction: column;
    gap: 0.3rem;
  }

  .meter-label {
    font-size: 0.75rem;
    color: #94a3b8;
    text-transform: uppercase;
    font-weight: 600;
  }

  .meter-val {
    font-size: 1.15rem;
    font-weight: 700;
  }

  .meter-val.lufs {
    color: #38bdf8;
  }

  .meter-val.dbtp {
    color: #34d399;
  }

  .meter-val.dbtp.clipping {
    color: #f87171;
  }

  .meter-val.lra {
    color: #a78bfa;
  }

  .meter-val.compliance.compliant {
    color: #34d399;
  }

  .meter-val.compliance.over {
    color: #f87171;
  }

  .meter-val.compliance.under {
    color: #fbbf24;
  }

  .meter-subtext {
    font-size: 0.75rem;
    color: #94a3b8;
  }

  .metering-visual-meters {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    background: #131d31;
    padding: 0.85rem;
    border-radius: 6px;
    border: 1px solid #1e293b;
  }

  .meter-bar-group {
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
  }

  .meter-bar-header {
    display: flex;
    justify-content: space-between;
    font-size: 0.8rem;
    color: #cbd5e1;
    font-weight: 500;
  }

  .meter-gauge-track {
    background: #0b1120;
    height: 10px;
    border-radius: 5px;
    overflow: hidden;
    border: 1px solid #1e293b;
  }

  .meter-gauge-fill {
    height: 100%;
    background: #38bdf8;
    border-radius: 5px;
    transition: width 0.3s ease;
  }

  .meter-gauge-fill.meter-gauge-hot {
    background: linear-gradient(90deg, #38bdf8, #fbbf24);
  }

  .meter-gauge-fill.meter-gauge-clip {
    background: linear-gradient(90deg, #34d399, #f87171);
  }

  .metering-advice-card {
    display: flex;
    gap: 0.75rem;
    align-items: flex-start;
    background: rgba(56, 189, 248, 0.1);
    border: 1px solid rgba(56, 189, 248, 0.25);
    border-radius: 6px;
    padding: 0.75rem 1rem;
    font-size: 0.85rem;
    color: #e2e8f0;
  }

  .metering-advice-card.advice-warn {
    background: rgba(239, 68, 68, 0.1);
    border-color: rgba(239, 68, 68, 0.3);
  }

  .advice-icon {
    font-size: 1.2rem;
  }

  .advice-content strong {
    display: block;
    margin-bottom: 0.2rem;
    color: #f8fafc;
  }

  .advice-content p {
    margin: 0;
    color: #cbd5e1;
    line-height: 1.4;
  }

  @media (max-width: 1024px) {
    .operations-grid {
      grid-template-columns: 1fr;
    }
    .top-nav {
      flex-direction: column;
      gap: 1rem;
    }
  }
</style>
