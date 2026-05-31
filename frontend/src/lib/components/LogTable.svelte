<script>
	import { onMount } from 'svelte';
	import { fetchQsos, searchQsos, updateQso } from '$lib/api.js';
	import { qsos } from '$lib/stores/qso.svelte.js';

	const bands = ['160M', '80M', '40M', '20M', '15M', '10M', '6M', '2M', '70CM'];
	const modes = ['CW', 'SSB', 'FM', 'RTTY', 'FT8', 'FT4', 'PSK31'];

	let searchQuery = $state('');
	let editingId = $state(null);
	let currentOffset = $state(0);
	let hasMore = $state(false);
	let debounceTimer;

	let editCallsign = $state('');
	let editBand = $state('20M');
	let editMode = $state('SSB');
	let editExchange = $state('');

	onMount(() => {
		loadQsos();
	});

	async function loadQsos() {
		try {
			const data = searchQuery
				? await searchQsos(searchQuery, 50, currentOffset)
				: await fetchQsos(50, currentOffset);
			if (currentOffset === 0) {
				qsos.splice(0, qsos.length, ...data);
			} else {
				qsos.push(...data);
			}
			hasMore = data.length === 50;
		} catch (err) {
			console.error('Failed to load QSOs:', err);
		}
	}

	function handleSearch() {
		clearTimeout(debounceTimer);
		debounceTimer = setTimeout(() => {
			currentOffset = 0;
			loadQsos();
		}, 300);
	}

	function loadMore() {
		currentOffset += 50;
		loadQsos();
	}

	function startEdit(qso) {
		editingId = qso.id;
		editCallsign = qso.callsign;
		editBand = qso.band;
		editMode = qso.mode;
		editExchange = qso.recv_exchange || '';
	}

	async function saveEdit() {
		if (!editingId) return;
		try {
			await updateQso(editingId, {
				callsign: editCallsign,
				band: editBand,
				mode: editMode,
				recv_exchange: editExchange
			});

			const idx = qsos.findIndex(q => q.id === editingId);
			if (idx !== -1) {
				qsos[idx].callsign = editCallsign;
				qsos[idx].band = editBand;
				qsos[idx].mode = editMode;
				qsos[idx].recv_exchange = editExchange;
			}
			editingId = null;
		} catch (err) {
			console.error('Edit failed:', err);
		}
	}

	function cancelEdit() {
		editingId = null;
	}

	function formatTime(ts) {
		if (!ts) return '';
		try {
			const d = new Date(ts);
			return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', hour12: false });
		} catch {
			return ts;
		}
	}
</script>

<div class="log-table-wrapper">
	<div class="search-bar">
		<input
			type="text"
			placeholder="Search callsign..."
			bind:value={searchQuery}
			oninput={handleSearch}
			class="search-input"
		/>
	</div>

	<table class="log-table">
		<thead>
			<tr>
				<th>Time</th>
				<th>Callsign</th>
				<th>Band</th>
				<th>Mode</th>
				<th>Exchange</th>
				<th>Pts</th>
			</tr>
		</thead>
		<tbody>
			{#each qsos as q (q.id)}
				{#if editingId === q.id}
					<tr class="edit-row">
						<td class="col-time">{formatTime(q.timestamp)}</td>
						<td><input type="text" bind:value={editCallsign} class="edit-input" /></td>
						<td>
							<select bind:value={editBand} class="edit-select">
								{#each bands as b}
									<option value={b}>{b}</option>
								{/each}
							</select>
						</td>
						<td>
							<select bind:value={editMode} class="edit-select">
								{#each modes as m}
									<option value={m}>{m}</option>
								{/each}
							</select>
						</td>
						<td><input type="text" bind:value={editExchange} class="edit-input" /></td>
						<td class="col-pts">{q.points}</td>
						<td class="edit-actions">
							<button class="btn-save" onclick={saveEdit}>Save</button>
							<button class="btn-cancel" onclick={cancelEdit}>Cancel</button>
						</td>
					</tr>
				{:else}
					<tr onclick={() => startEdit(q)} class="qso-row">
						<td class="col-time">{formatTime(q.timestamp)}</td>
						<td class="col-call"><strong>{q.callsign}</strong></td>
						<td>{q.band}</td>
						<td>{q.mode}</td>
						<td>{q.recv_exchange}</td>
						<td class="col-pts">{q.points}</td>
					</tr>
				{/if}
			{/each}
		</tbody>
	</table>
	{#if qsos.length === 0}
		<div class="empty-state">No QSOs logged yet. Start typing above!</div>
	{/if}

	{#if hasMore}
		<div class="load-more">
			<button onclick={loadMore}>Load More</button>
		</div>
	{/if}
</div>

<style>
	.log-table-wrapper {
		flex: 1;
		overflow-y: auto;
	}

	.search-bar {
		padding: 8px 16px;
		background: var(--color-bg-alt);
		border-bottom: 1px solid var(--color-border-strong);
		position: sticky;
		top: 0;
		z-index: 5;
	}

	.search-input {
		padding: 8px 12px;
		font-size: 14px;
		border: 2px solid var(--color-border);
		border-radius: 6px;
		width: 200px;
		outline: none;
	}

	.search-input:focus {
		border-color: var(--color-accent);
	}

	.log-table {
		width: 100%;
		border-collapse: collapse;
		font-size: 15px;
	}

	.log-table th {
		position: sticky;
		top: 48px;
		background: var(--color-bg-alt);
		padding: 8px 12px;
		text-align: left;
		font-weight: 600;
		font-size: 13px;
		text-transform: uppercase;
		color: var(--color-text-secondary);
		border-bottom: 2px solid var(--color-border);
	}

	.log-table td {
		padding: 8px 12px;
		border-bottom: 1px solid var(--color-border);
	}

	.log-table tbody tr:nth-child(even) {
		background: var(--color-bg);
	}

	.qso-row {
		cursor: pointer;
		transition: background 0.1s;
	}

	.qso-row:hover {
		background: var(--color-highlight) !important;
	}

	.edit-row {
		background: var(--color-highlight) !important;
		border-left: 3px solid #f0c040;
	}

	.edit-row:hover {
		background: var(--color-highlight) !important;
	}

	.edit-input, .edit-select {
		padding: 4px 8px;
		font-size: 14px;
		border: 2px solid var(--color-accent);
		border-radius: 4px;
		width: 100%;
		min-width: 80px;
	}

	.edit-actions {
		white-space: nowrap;
	}

	.btn-save {
		padding: 4px 12px;
		font-size: 13px;
		font-weight: 600;
		border: none;
		border-radius: 4px;
		background: var(--color-success);
		color: var(--color-surface);
		cursor: pointer;
		margin-right: 4px;
	}

	.btn-save:hover {
		background: var(--color-success);
		filter: brightness(0.85);
	}

	.btn-cancel {
		padding: 4px 12px;
		font-size: 13px;
		border: 1px solid var(--color-border);
		border-radius: 4px;
		background: var(--color-surface);
		cursor: pointer;
	}

	.btn-cancel:hover {
		background: var(--color-bg-alt);
	}

	.col-time {
		white-space: nowrap;
		width: 70px;
	}

	.col-call {
		white-space: nowrap;
	}

	.col-pts {
		text-align: center;
		width: 40px;
	}

	.empty-state {
		padding: 40px;
		text-align: center;
		color: var(--color-text-muted);
		font-style: italic;
	}

	.load-more {
		padding: 12px;
		text-align: center;
	}

	.load-more button {
		padding: 8px 24px;
		font-size: 14px;
		border: 2px solid var(--color-accent);
		border-radius: 6px;
		background: var(--color-surface);
		color: var(--color-accent);
		font-weight: 600;
		cursor: pointer;
	}

	.load-more button:hover {
		background: var(--color-highlight);
	}
</style>
