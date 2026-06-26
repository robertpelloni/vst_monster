<script lang="ts">
  import { onMount } from 'svelte';
  import { invoke } from '@tauri-apps/api/core';

  interface Plugin {
    id: string;
    name: string;
    developer: string;
    license_model: string;
    created_at: string;
  }

  let plugins = $state<Plugin[]>([]);
  let error = $state('');
  let loading = $state(true);
  let installStatus = $state<{ [key: string]: string }>({});

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

  async function installPlugin(plugin: Plugin) {
    try {
      installStatus[plugin.id] = 'Installing...';
      const dummyUrl = 'https://example.com/dummy.vst'; // In a real scenario, this would come from the registry DB plugin_distributions
      const response = await invoke('install_plugin', { url: dummyUrl, name: plugin.name });
      installStatus[plugin.id] = response as string;
    } catch (err: any) {
      installStatus[plugin.id] = `Error: ${err}`;
    }
  }
</script>

<main class="container">
  <h1>VST Monster - Registry</h1>

  {#if loading}
    <p>Loading plugins...</p>
  {:else if error}
    <p class="error">Error: {error}</p>
  {:else if plugins.length === 0}
    <p>No plugins found in the registry.</p>
  {:else}
    <div class="plugin-grid">
      {#each plugins as plugin}
        <div class="plugin-card">
          <h2>{plugin.name}</h2>
          <p><strong>Developer:</strong> {plugin.developer}</p>
          <p><strong>License:</strong> {plugin.license_model}</p>
          <button onclick={() => installPlugin(plugin)}>Install</button>
          {#if installStatus[plugin.id]}
            <p class="status">{installStatus[plugin.id]}</p>
          {/if}
        </div>
      {/each}
    </div>
  {/if}
</main>

<style>
  :root {
    font-family: Inter, Avenir, Helvetica, Arial, sans-serif;
    color: #0f0f0f;
    background-color: #f6f6f6;
  }

  .container {
    padding: 2rem;
    max-width: 1200px;
    margin: 0 auto;
  }

  h1 {
    text-align: center;
    margin-bottom: 2rem;
  }

  .error {
    color: red;
    text-align: center;
  }

  .plugin-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
    gap: 1.5rem;
  }

  .plugin-card {
    background: white;
    padding: 1.5rem;
    border-radius: 8px;
    box-shadow: 0 2px 4px rgba(0,0,0,0.1);
    transition: transform 0.2s;
  }

  .plugin-card:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 8px rgba(0,0,0,0.15);
  }

  .plugin-card h2 {
    margin: 0 0 1rem 0;
    font-size: 1.25rem;
    color: #2c3e50;
  }

  .plugin-card p {
    margin: 0.5rem 0;
    color: #666;
  }

  .plugin-card button {
    margin-top: 1rem;
    padding: 0.5rem 1rem;
    background-color: #007bff;
    color: white;
    border: none;
    border-radius: 4px;
    cursor: pointer;
    font-weight: bold;
  }

  .plugin-card button:hover {
    background-color: #0056b3;
  }

  .status {
    font-size: 0.8rem;
    margin-top: 0.5rem;
    color: #0056b3;
  }

  @media (prefers-color-scheme: dark) {
    :root {
      color: #f6f6f6;
      background-color: #2f2f2f;
    }
    .plugin-card {
      background: #3f3f3f;
    }
    .plugin-card h2 {
      color: #fff;
    }
    .plugin-card p {
      color: #ccc;
    }
    .status {
        color: #4da3ff;
    }
  }
</style>
