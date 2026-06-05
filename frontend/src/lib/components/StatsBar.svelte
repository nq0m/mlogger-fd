<script>
	import { stats, fetchStats } from '$lib/stores/qso.svelte.js';
	import { onMount } from 'svelte';

	onMount(() => {
		fetchStats();
	});

	let showBreakdown = $state(false);

	function toggleBreakdown() {
		showBreakdown = !showBreakdown;
	}

	let breakdownEntries = $derived(
		stats.breakdown ? Object.entries(stats.breakdown).sort() : []
	);
</script>

<div class="stats-bar">
	<div class="stat rate">
		<span class="stat-label">Rate</span>
		<span class="stat-value">{stats.rate_10min}</span>
		<span class="stat-unit">/hr</span>
	</div>
	<div class="stat">
		<span class="stat-label">QSOs</span>
		<span class="stat-value">{stats.total}</span>
	</div>
	<div class="stat">
		<span class="stat-label">Pts</span>
		<span class="stat-value">{stats.raw_points}</span>
	</div>
	<div class="stat bonus">
		<span class="stat-label">Bonus</span>
		<span class="stat-value">{stats.bonus_points || 0}</span>
	</div>
	<div class="stat">
		<span class="stat-label">Mult</span>
		<span class="stat-value">{stats.multiplier}</span>
	</div>
	<div class="stat score">
		<span class="stat-label">Score</span>
		<span class="stat-value">{stats.score}</span>
	</div>
	<button class="breakdown-toggle" onclick={toggleBreakdown}>
		{showBreakdown ? '▼' : '▶'} Breakdown
	</button>
</div>

{#if showBreakdown}
	<div class="breakdown-panel">
		<table class="breakdown-table">
			<thead>
				<tr>
					<th>Band</th>
					<th>Mode</th>
					<th>QSOs</th>
				</tr>
			</thead>
			<tbody>
				{#each breakdownEntries as [key, count]}
					{@const parts = key.split('_')}
					<tr>
						<td>{parts[0]}</td>
						<td>{parts[1] || ''}</td>
						<td>{count}</td>
					</tr>
				{/each}
			</tbody>
		</table>
		{#if breakdownEntries.length === 0}
			<div class="empty-breakdown">No QSOs yet</div>
		{/if}
	</div>
{/if}

<style>
	.stats-bar {
		display: flex;
		gap: 16px;
		align-items: center;
		padding: 8px 16px;
		background: var(--color-highlight);
		border-bottom: 1px solid var(--color-border-light);
		font-size: 14px;
		flex-wrap: wrap;
	}

	.stat {
		display: flex;
		gap: 4px;
		align-items: baseline;
	}

	.stat-label {
		color: var(--color-text-secondary);
		font-size: 11px;
		text-transform: uppercase;
		font-weight: 600;
	}

	.stat-value {
		font-size: 20px;
		font-weight: 700;
		color: var(--color-primary);
	}

	.stat-unit {
		font-size: 11px;
		color: var(--color-text-secondary);
	}

	.rate .stat-value {
		color: var(--color-danger);
	}

	.score .stat-value {
		color: var(--color-success);
	}

	.bonus .stat-value {
		color: #cc8800;
	}

	:global([data-theme='dark']) .bonus .stat-value {
		color: #ffaa00;
	}

	.breakdown-toggle {
		margin-left: auto;
		padding: 4px 10px;
		font-size: 12px;
		border: 1px solid var(--color-text-muted);
		border-radius: 4px;
		background: var(--color-surface);
		cursor: pointer;
	}

	.breakdown-toggle:hover {
		background: var(--color-bg-alt);
	}

	.breakdown-panel {
		padding: 8px 16px;
		background: var(--color-bg);
		border-bottom: 1px solid var(--color-border-strong);
	}

	.breakdown-table {
		width: 100%;
		font-size: 13px;
		border-collapse: collapse;
	}

	.breakdown-table th {
		text-align: left;
		padding: 4px 8px;
		color: var(--color-text-secondary);
		font-size: 11px;
		text-transform: uppercase;
	}

	.breakdown-table td {
		padding: 4px 8px;
	}

	.empty-breakdown {
		padding: 8px;
		color: var(--color-text-muted);
		font-style: italic;
		font-size: 13px;
	}
</style>
