<script lang="ts">
  import { onMount } from 'svelte';
  import { invoke } from '@tauri-apps/api/core';

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
  let selectedFormat = $state<'all' | 'vst3' | 'clap' | 'au' | 'vst2'>('all');
  let sortBy = $state<'name' | 'developer' | 'license' | 'category' | 'status'>('name');
  let crawlerRunning = $state(false);
  let actionMessage = $state<string | null>(null);
  let isLogsCollapsed = $state(true);
  let showAddPluginModal = $state(false);
  let previewPlugin = $state<Plugin | null>(null);
  let isBatchUpdating = $state(false);
  let isScanningFolders = $state(false);

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

  async function scanLocalPlugins() {
    isScanningFolders = true;
    try {
      const detected: InstalledPluginInfo[] = await invoke('scan_installed_plugins');
      installedPlugins = detected || [];
    } catch (err) {
      // In web browser preview mode, invoke may fail silently
      console.log('Installed plugin scan skipped or not in Tauri:', err);
    } finally {
      isScanningFolders = false;
    }
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
    if (!plugin.download_url) {
      installStatus[plugin.id] = "No direct binary download URL available";
      return;
    }
    installStatus[plugin.id] = "Downloading & Installing...";

    try {
      const response = await invoke('install_plugin', {
        url: plugin.download_url,
        name: plugin.name,
        expectedHash: plugin.sha256_hash || null
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

  // Dynamic counts for status pills
  let countAll = $derived(plugins.length);
  let countInstalled = $derived(plugins.filter(p => isPluginInstalled(p.name)).length);
  let countUpdates = $derived(plugins.filter(p => {
    const inst = getInstalledInfo(p.name);
    return inst ? getVersionStatus(p.version, inst.version).state === 'update-available' : false;
  }).length);
  let countNotInstalled = $derived(plugins.filter(p => !isPluginInstalled(p.name)).length);

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
          matchesFormat = inst ? inst.format.toLowerCase() === selectedFormat : false;
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
    fetchDashboardData();
    const interval = setInterval(fetchDashboardData, 10000);
    return () => clearInterval(interval);
  });
</script>

<div class="dashboard-root">
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
    <div class="modal-backdrop">
      <div class="modal-card preview-card">
        <div class="modal-header">
          <div class="preview-header-info">
            <span class="preview-category">{previewPlugin.category || 'synth'}</span>
            <h2>{previewPlugin.name}</h2>
            <span class="preview-author">by {previewPlugin.developer}</span>
          </div>
          <button class="modal-close" onclick={() => (previewPlugin = null)}>✕</button>
        </div>

        <div class="preview-content">
          <!-- Status Banner -->
          {#if previewInstalled}
            <div class="preview-install-status {previewVersionStatus?.state === 'update-available' ? 'status-update' : 'status-installed'}">
              <span class="icon">✓</span>
              <div>
                <strong>Installed in {previewInstalled.format.toUpperCase()} format</strong>
                <p>Path: <code>{previewInstalled.path}</code></p>
                {#if previewVersionStatus}
                  <small>{previewVersionStatus.label}</small>
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
            <div class="spec-tile">
              <span class="label">Strategy</span>
              <strong class="value">{previewPlugin.strategy || 'extract_binaries'}</strong>
            </div>
          </div>

          <!-- Description -->
          <div class="preview-section">
            <h4>Description & Synthesis Architecture</h4>
            <p class="preview-desc-text">
              {previewPlugin.metadata?.description || 'No extended description provided. Cataloged via VST Monster distributed crawler pipeline.'}
            </p>
          </div>

          <!-- Cryptographic Digest -->
          {#if previewPlugin.sha256_hash}
            <div class="preview-section">
              <h4>Binary SHA-256 Digest</h4>
              <div class="preview-hash-box">
                <code>{previewPlugin.sha256_hash}</code>
              </div>
            </div>
          {/if}

          <!-- Direct Download Link -->
          {#if previewPlugin.download_url}
            <div class="preview-section">
              <h4>Direct Download Source</h4>
              <a href={previewPlugin.download_url} target="_blank" rel="noreferrer" class="preview-download-link">
                🔗 {previewPlugin.download_url}
              </a>
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
          <span class="tooltip-badge" use:tooltip={"Search catalog by plugin name or developer keywords"}>?</span>
        </div>
        <div class="filter-controls">
          <input
            type="text"
            placeholder="Search by name, developer, or synth type..."
            bind:value={searchQuery}
            class="search-input"
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
            class="filter-pill {selectedInstalledStatus === 'all' ? 'pill-active' : ''}"
            onclick={() => (selectedInstalledStatus = 'all')}
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
                  <span class="spec-label">Category</span>
                  <span class="spec-val category-badge" use:tooltip={`Plugin type: ${plugin.category || 'synth'}`}>
                    {plugin.category || 'synth'}
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
              </div>

              <!-- Cryptographic SHA256 Verification -->
              {#if plugin.sha256_hash}
                <div class="hash-container" use:tooltip={`Full SHA-256 Digest: ${plugin.sha256_hash}`}>
                  <span class="hash-label">🔒 SHA-256:</span>
                  <code class="hash-code">{plugin.sha256_hash.substring(0, 16)}...</code>
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

  .card-header h3 {
    font-size: 1rem;
    font-weight: 700;
    color: #f1f5f9;
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
    align-items: center;
    gap: 0.8rem;
    font-size: 0.85rem;
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

  .preview-download-link {
    color: #60a5fa;
    font-size: 0.85rem;
    text-decoration: none;
    word-break: break-all;
  }

  .preview-download-link:hover {
    text-decoration: underline;
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
