<script>
	import { onMount } from 'svelte';
	import QsoEntryForm from '$lib/components/QsoEntryForm.svelte';
	import StatsBar from '$lib/components/StatsBar.svelte';
	import LogTable from '$lib/components/LogTable.svelte';
	import StationConfig from '$lib/components/StationConfig.svelte';
	import OperatorSelector from '$lib/components/OperatorSelector.svelte';
	import { connectWebSocket, wsState } from '$lib/ws.svelte.js';
	import { queueState } from '$lib/sync.svelte.js';
	import { loadCache } from '$lib/stores/qso.svelte.js';

	onMount(() => {
		connectWebSocket();
		loadCache();
	});

	function exportCabrillo() {
		window.location.href = '/api/export/cabrillo';
	}
</script>

<div class="header-bar">
	<div class="header-left">
		<h1 class="title">FD Logger</h1>
		<span class="ws-status" class:online={wsState.connected} class:offline={!wsState.connected}>
			{wsState.connected ? '● Live' : '● Disconnected'}
		</span>
		{#if queueState.queueLength > 0}
			<span class="queue-count">{queueState.syncing ? 'Syncing...' : `${queueState.queueLength} queued`}</span>
		{/if}
	</div>
	<div class="header-center">
		<OperatorSelector />
	</div>
	<div class="header-right">
		<StationConfig />
		<button class="export-btn" onclick={exportCabrillo}>Export Cabrillo</button>
	</div>
</div>
<QsoEntryForm />
<StatsBar />
<LogTable />

<style>
	.header-bar {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 6px 16px;
		background: #1a3a6b;
		color: #fff;
		gap: 12px;
	}

	.header-left {
		display: flex;
		align-items: center;
		gap: 12px;
	}

	.header-center {
		display: flex;
		align-items: center;
	}

	.header-right {
		display: flex;
		align-items: center;
		gap: 8px;
	}

	.title {
		font-size: 18px;
		font-weight: 700;
		margin: 0;
	}

	.ws-status {
		font-size: 11px;
		padding: 2px 8px;
		border-radius: 4px;
		font-weight: 600;
	}

	.ws-status.online {
		color: #1a7a1a;
	}

	.ws-status.offline {
		color: #cc3300;
	}

	.queue-count {
		font-size: 11px;
		padding: 2px 8px;
		border-radius: 4px;
		font-weight: 600;
		background: rgba(255,255,255,0.2);
		color: #fff;
	}

	.export-btn {
		padding: 6px 16px;
		font-size: 14px;
		font-weight: 600;
		border: 2px solid #fff;
		border-radius: 6px;
		background: transparent;
		color: #fff;
		cursor: pointer;
	}

	.export-btn:hover {
		background: rgba(255,255,255,0.15);
	}
</style>
