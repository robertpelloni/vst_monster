<script lang="ts">
  import { onMount } from 'svelte';
  import { invoke } from '@tauri-apps/api/core';

  interface Plugin {
    id: string;
    name: string;
    developer: string;
    license_model: string;
    created_at: string;
    updated_at: string;
    version: string | null;
    release_date: string | null;
    platform: string | null;
    architecture: string | null;
    download_url: string | null;
    sha256_hash: string | null;
    strategy: string | null;
    extraction_rules: any | null;
  }

  let plugins = $state<Plugin[]>([]);
  let error = $state('');
  let loading = $state(true);
  let installStatus = $state<Record<string, string>>({});

  onMount(async () => {
    try {
      const res = await fetch('http://localhost:3000/plugins');
      if (!res.ok) throw new Error('Failed to fetch plugins');
      plugins = await res.json();
    } catch (err: any) {
      error = err.message;
    } finally {
      loading = false;
    }
  });

  async function handleInstall(plugin: Plugin) {
    if (!plugin.download_url) {
        installStatus[plugin.id] = "No download URL available";
        return;
    }
    installStatus[plugin.id] = "Installing...";

    try {
      const response = await invoke('install_plugin', { url: plugin.download_url, name: plugin.name });
      installStatus[plugin.id] = `Success: ${response}`;
    } catch (err: any) {
      installStatus[plugin.id] = `Error: ${err}`;
    }
  }

  // Tooltip action
  function tooltip(element: HTMLElement, text: string) {
    let tooltipEl: HTMLDivElement;

    function mouseOver(event: MouseEvent) {
      tooltipEl = document.createElement('div');
      tooltipEl.className = 'tooltip';
      tooltipEl.textContent = text;

      // Position the tooltip
      tooltipEl.style.left = `${event.pageX + 10}px`;
      tooltipEl.style.top = `${event.pageY + 10}px`;

      document.body.appendChild(tooltipEl);
    }

    function mouseMove(event: MouseEvent) {
      if (tooltipEl) {
        tooltipEl.style.left = `${event.pageX + 10}px`;
        tooltipEl.style.top = `${event.pageY + 10}px`;
      }
    }

    function mouseLeave() {
      if (tooltipEl) {
        tooltipEl.remove();
      }
    }

    element.addEventListener('mouseover', mouseOver);
    element.addEventListener('mousemove', mouseMove);
    element.addEventListener('mouseleave', mouseLeave);

    return {
      destroy() {
        element.removeEventListener('mouseover', mouseOver);
        element.removeEventListener('mousemove', mouseMove);
        element.removeEventListener('mouseleave', mouseLeave);
        if (tooltipEl) tooltipEl.remove();
      }
    }
  }
</script>

<main class="container">
  <h1>VST Monster - Registry</h1>
  <p class="subtitle">Explore and install community-vetted VST plugins.</p>

  {#if loading}
    <div class="loader-container">
      <div class="loader"></div>
      <p>Loading plugins...</p>
    </div>
  {:else if error}
    <div class="error-container">
      <p class="error">Error: {error}</p>
      <button onclick={() => window.location.reload()}>Retry</button>
    </div>
  {:else if plugins.length === 0}
    <div class="empty-state">
      <p>No plugins found in the registry. Make sure the crawler is running.</p>
    </div>
  {:else}
    <div class="plugin-grid">
      {#each plugins as plugin}
        <div class="plugin-card">
          <div class="plugin-header">
            <h2 title="{plugin.name}">{plugin.name}</h2>
            <span class="badge {plugin.license_model}" use:tooltip={`License: ${plugin.license_model}`}>{plugin.license_model}</span>
          </div>

          <div class="plugin-details">
            <p><strong>Developer:</strong> <span class="developer-name" use:tooltip={`Created by ${plugin.developer}`}>{plugin.developer}</span></p>
            {#if plugin.version}
              <p><strong>Version:</strong> {plugin.version}</p>
            {/if}
            {#if plugin.platform}
               <p><strong>Platform:</strong> <span class="platform-badge" use:tooltip={`Supported on ${plugin.platform} (${plugin.architecture})`}>{plugin.platform}</span></p>
            {/if}
            {#if plugin.sha256_hash}
               <p class="hash" use:tooltip={`SHA-256 Hash for verification: ${plugin.sha256_hash}`}><strong>Hash:</strong> {plugin.sha256_hash.substring(0, 12)}...</p>
            {/if}
          </div>

          <div class="plugin-actions">
            {#if plugin.download_url}
              <button
                class="install-btn"
                onclick={() => handleInstall(plugin)}
                disabled={installStatus[plugin.id] === 'Installing...'}
                use:tooltip={`Download and install ${plugin.name} natively`}
              >
                {installStatus[plugin.id] === 'Installing...' ? 'Installing...' : 'Install Plugin'}
              </button>
            {:else}
              <button class="install-btn disabled" disabled use:tooltip={"No download URL provided in registry"}>Unavailable</button>
            {/if}
          </div>

          {#if installStatus[plugin.id] && installStatus[plugin.id] !== 'Installing...'}
             <div class="status-message {installStatus[plugin.id].startsWith('Error') ? 'status-error' : 'status-success'}">
               {installStatus[plugin.id]}
             </div>
          {/if}
        </div>
      {/each}
    </div>
  {/if}
</main>

<style>
  :global(.tooltip) {
    position: absolute;
    background: rgba(0, 0, 0, 0.8);
    color: white;
    padding: 6px 10px;
    border-radius: 4px;
    font-size: 0.8rem;
    pointer-events: none;
    z-index: 1000;
    white-space: nowrap;
    box-shadow: 0 2px 4px rgba(0,0,0,0.2);
  }

  :root {
    font-family: Inter, Avenir, Helvetica, Arial, sans-serif;
    color: #0f0f0f;
    background-color: #f6f6f6;
  }

  .container {
    padding: 2rem;
    max-width: 1400px;
    margin: 0 auto;
  }

  h1 {
    text-align: center;
    margin-bottom: 0.5rem;
    font-size: 2.5rem;
    color: #2c3e50;
  }

  .subtitle {
    text-align: center;
    color: #666;
    margin-bottom: 3rem;
  }

  .error-container, .empty-state, .loader-container {
    text-align: center;
    padding: 3rem;
    background: white;
    border-radius: 8px;
    box-shadow: 0 2px 4px rgba(0,0,0,0.1);
  }

  .error {
    color: #e74c3c;
    font-weight: bold;
    margin-bottom: 1rem;
  }

  .loader {
    border: 4px solid #f3f3f3;
    border-top: 4px solid #3498db;
    border-radius: 50%;
    width: 40px;
    height: 40px;
    animation: spin 1s linear infinite;
    margin: 0 auto 1rem auto;
  }

  @keyframes spin {
    0% { transform: rotate(0deg); }
    100% { transform: rotate(360deg); }
  }

  .plugin-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
    gap: 1.5rem;
  }

  .plugin-card {
    background: white;
    padding: 1.5rem;
    border-radius: 12px;
    box-shadow: 0 4px 6px rgba(0,0,0,0.05);
    transition: all 0.3s ease;
    display: flex;
    flex-direction: column;
    border: 1px solid #eee;
  }

  .plugin-card:hover {
    transform: translateY(-4px);
    box-shadow: 0 12px 20px rgba(0,0,0,0.1);
    border-color: #3498db;
  }

  .plugin-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 1rem;
    padding-bottom: 1rem;
    border-bottom: 1px solid #eee;
  }

  .plugin-card h2 {
    margin: 0;
    font-size: 1.25rem;
    color: #2c3e50;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    max-width: 70%;
  }

  .badge {
    padding: 4px 8px;
    border-radius: 12px;
    font-size: 0.75rem;
    font-weight: bold;
    text-transform: uppercase;
  }

  .badge.free { background: #e8f8f5; color: #1abc9c; }
  .badge.commercial { background: #fdf2e9; color: #e67e22; }
  .badge.opensource { background: #ebf5fb; color: #2980b9; }

  .plugin-details {
    flex-grow: 1;
    font-size: 0.9rem;
  }

  .plugin-details p {
    margin: 0.5rem 0;
    color: #555;
    display: flex;
    justify-content: space-between;
  }

  .plugin-details strong {
    color: #333;
  }

  .developer-name {
    color: #3498db;
    cursor: help;
  }

  .platform-badge {
    background: #eee;
    padding: 2px 6px;
    border-radius: 4px;
    font-size: 0.8rem;
    cursor: help;
  }

  .hash {
    font-family: monospace;
    font-size: 0.8rem;
    color: #888 !important;
    cursor: help;
  }

  .plugin-actions {
    margin-top: 1.5rem;
  }

  .install-btn {
    width: 100%;
    padding: 0.75rem;
    background: #3498db;
    color: white;
    border: none;
    border-radius: 6px;
    font-weight: bold;
    cursor: pointer;
    transition: background 0.2s;
  }

  .install-btn:hover:not(:disabled) {
    background: #2980b9;
  }

  .install-btn:disabled {
    background: #95a5a6;
    cursor: not-allowed;
  }

  .status-message {
    margin-top: 1rem;
    padding: 0.5rem;
    border-radius: 4px;
    font-size: 0.85rem;
    text-align: center;
  }

  .status-success {
    background: #e8f8f5;
    color: #1abc9c;
  }

  .status-error {
    background: #fdedec;
    color: #e74c3c;
  }

  @media (prefers-color-scheme: dark) {
    :root {
      color: #f6f6f6;
      background-color: #1a1a1a;
    }

    h1 { color: #f6f6f6; }
    .subtitle { color: #aaa; }

    .plugin-card, .error-container, .empty-state, .loader-container {
      background: #2a2a2a;
      border-color: #333;
    }

    .plugin-header { border-bottom-color: #333; }
    .plugin-card h2 { color: #fff; }
    .plugin-details p { color: #bbb; }
    .plugin-details strong { color: #ddd; }

    .platform-badge {
      background: #444;
      color: #eee;
    }

    .badge.free { background: rgba(26, 188, 156, 0.2); }
    .badge.commercial { background: rgba(230, 126, 34, 0.2); }
    .badge.opensource { background: rgba(41, 128, 185, 0.2); }
  }
</style>
