<script>
	import { onMount } from 'svelte';
	import QsoEntryForm from '$lib/components/QsoEntryForm.svelte';
	import StatsBar from '$lib/components/StatsBar.svelte';
	import LogTable from '$lib/components/LogTable.svelte';
	import StationConfig from '$lib/components/StationConfig.svelte';
	import BonusTracker from '$lib/components/BonusTracker.svelte';
	import OperatorSelector from '$lib/components/OperatorSelector.svelte';
	import { connectWebSocket, wsState } from '$lib/ws.svelte.js';
	import { audioState, toggleMute } from '$lib/audio.svelte.js';
	import { queueState } from '$lib/sync.svelte.js';
	import { loadCache } from '$lib/stores/qso.svelte.js';
	import { downloadBackup } from '$lib/api.js';

	let theme = $state('light');
	let backupToast = $state(false);
	let backupTimer;

	function initTheme() {
		const stored = localStorage.getItem('fdlogger_theme');
		if (stored === 'light' || stored === 'dark') {
			theme = stored;
		} else if (window.matchMedia('(prefers-color-scheme: dark)').matches) {
			theme = 'dark';
		}
		applyTheme();
	}

	function toggleTheme() {
		theme = theme === 'light' ? 'dark' : 'light';
		localStorage.setItem('fdlogger_theme', theme);
		applyTheme();
	}

	function applyTheme() {
		document.body.setAttribute('data-theme', theme);
	}

	onMount(() => {
		initTheme();
		connectWebSocket();
		loadCache();
	});

	function exportCabrillo() {
		window.location.href = '/api/export/cabrillo';
	}

	function handleBackup() {
		downloadBackup();
		backupToast = true;
		if (backupTimer) clearTimeout(backupTimer);
		backupTimer = setTimeout(() => { backupToast = false; }, 2000);
	}
</script>

<div class="header-bar">
	<div class="header-left">
		<button class="theme-toggle" onclick={toggleTheme} aria-label="Toggle dark mode">
			{theme === 'light' ? '☀' : '☾'}
		</button>
		<button class="theme-toggle" onclick={toggleMute} aria-label="Toggle audio">
			{audioState.muted ? '🔇' : '🔊'}
		</button>
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
		<BonusTracker />
		<button class="export-btn" onclick={exportCabrillo}>Export Cabrillo</button>
		<button class="export-btn" onclick={handleBackup}>↓ Backup</button>
	</div>
	{#if backupToast}
		<span class="saved-msg" style="position:absolute; top:44px; right: 8px; z-index: 101; background: var(--color-surface); padding: 4px 12px; border-radius: 6px; font-size: 14px; font-weight: 600; color: var(--color-success);">Backup downloaded</span>
	{/if}
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
		background: var(--color-primary);
		color: var(--color-surface);
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
		color: var(--color-success);
	}

	.ws-status.offline {
		color: var(--color-danger);
	}

	.queue-count {
		font-size: 11px;
		padding: 2px 8px;
		border-radius: 4px;
		font-weight: 600;
		background: rgba(255,255,255,0.2);
		color: var(--color-surface);
	}

	.export-btn {
		padding: 6px 16px;
		font-size: 14px;
		font-weight: 600;
		border: 2px solid var(--color-surface);
		border-radius: 6px;
		background: transparent;
		color: var(--color-surface);
		cursor: pointer;
		min-height: 48px;
	}

	.export-btn:hover {
		background: rgba(255,255,255,0.15);
	}

	.theme-toggle {
		background: none;
		border: 1px solid rgba(255,255,255,0.3);
		border-radius: 4px;
		color: var(--color-surface);
		font-size: 16px;
		cursor: pointer;
		padding: 2px 6px;
		min-width: 48px;
		min-height: 48px;
	}

	.theme-toggle:hover {
		background: rgba(255,255,255,0.15);
	}

	@media (max-width: 768px) {
		.header-bar {
			flex-wrap: wrap;
			padding: 6px 10px;
			gap: 6px;
		}
		.header-left, .header-center, .header-right {
			flex: 1 1 auto;
		}
		.title {
			font-size: 16px;
		}
	}

	@media (max-width: 500px) {
		.header-bar {
			padding: 4px 8px;
		}
		.header-right {
			flex-basis: 100%;
			justify-content: flex-end;
		}
	}
</style>
