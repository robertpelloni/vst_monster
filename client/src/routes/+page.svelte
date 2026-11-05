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

  // State
  let plugins = $state<Plugin[]>([]);
  let stats = $state<Stats | null>(null);
  let health = $state<Health | null>(null);
  let crawlerLogs = $state<string[]>([]);
  let loading = $state(true);
  let error = $state<string | null>(null);
  let installStatus = $state<Record<string, string>>({});

  // Controls & Filters
  let searchQuery = $state('');
  let selectedLicense = $state('all');
  let selectedCategory = $state('all');
  let selectedPlatform = $state('all');
  let sortBy = $state<'name' | 'developer' | 'license' | 'category'>('name');
  let crawlerRunning = $state(false);
  let actionMessage = $state<string | null>(null);
  let isLogsCollapsed = $state(true);
  let showAddPluginModal = $state(false);

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
      const response = await invoke('install_plugin', { url: plugin.download_url, name: plugin.name });
      installStatus[plugin.id] = `Success: ${response}`;
    } catch (err: any) {
      installStatus[plugin.id] = `Error: ${err}`;
    }
  }

  async function handleRemove(plugin: Plugin) {
    installStatus[plugin.id] = "Removing plugin files...";
    try {
      const response = await invoke('remove_plugin', { name: plugin.name });
      installStatus[plugin.id] = `Removed: ${response}`;
    } catch (err: any) {
      installStatus[plugin.id] = `Error: ${err}`;
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
        return matchesSearch && matchesLicense && matchesCategory && matchesPlatform;
      })
      .sort((a, b) => {
        if (sortBy === 'name') return a.name.localeCompare(b.name);
        if (sortBy === 'developer') return a.developer.localeCompare(b.developer);
        if (sortBy === 'license') return a.license_model.localeCompare(b.license_model);
        if (sortBy === 'category') return (a.category || 'synth').localeCompare(b.category || 'synth');
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

            <select bind:value={sortBy} class="select-input" use:tooltip={"Sort order for the catalog"}>
              <option value="name">Sort: Name</option>
              <option value="developer">Sort: Developer</option>
              <option value="category">Sort: Category</option>
              <option value="license">Sort: License</option>
            </select>
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
        <div>
          <h2>Vetted Plugin Library</h2>
          <span class="section-subtitle">Showing {filteredPlugins.length} of {plugins.length} verified packages</span>
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
            <div class="plugin-card">
              <!-- Header -->
              <div class="card-top">
                <div class="title-group">
                  <h3 class="plugin-title" title={plugin.name}>{plugin.name}</h3>
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
                      {installStatus[plugin.id] === 'Downloading & Installing...' ? 'Installing...' : 'Install Plugin'}
                    </button>
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

  .section-header h2 {
    font-size: 1.3rem;
    font-weight: 700;
  }

  .section-subtitle {
    font-size: 0.8rem;
    color: #94a3b8;
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
