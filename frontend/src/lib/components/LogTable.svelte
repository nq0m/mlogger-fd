<script>
	import { onMount } from 'svelte';
	import { fetchQsos } from '$lib/api.js';
	import { qsos } from '$lib/stores/qso.svelte.js';

	onMount(async () => {
		try {
			const data = await fetchQsos(50, 0);
			if (data.length) {
				qsos.splice(0, qsos.length, ...data);
			}
		} catch (err) {
			console.error('Failed to load QSOs:', err);
		}
	});

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
				<tr>
					<td class="col-time">{formatTime(q.timestamp)}</td>
					<td class="col-call"><strong>{q.callsign}</strong></td>
					<td>{q.band}</td>
					<td>{q.mode}</td>
					<td>{q.recv_exchange}</td>
					<td class="col-pts">{q.points}</td>
				</tr>
			{/each}
		</tbody>
	</table>
	{#if qsos.length === 0}
		<div class="empty-state">No QSOs logged yet. Start typing above!</div>
	{/if}
</div>

<style>
	.log-table-wrapper {
		flex: 1;
		overflow-y: auto;
	}

	.log-table {
		width: 100%;
		border-collapse: collapse;
		font-size: 15px;
	}

	.log-table th {
		position: sticky;
		top: 0;
		background: #e8e8e8;
		padding: 8px 12px;
		text-align: left;
		font-weight: 600;
		font-size: 13px;
		text-transform: uppercase;
		color: #555;
		border-bottom: 2px solid #ccc;
	}

	.log-table td {
		padding: 8px 12px;
		border-bottom: 1px solid #eee;
	}

	.log-table tbody tr:nth-child(even) {
		background: #fafafa;
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
		color: #888;
		font-style: italic;
	}
</style>
