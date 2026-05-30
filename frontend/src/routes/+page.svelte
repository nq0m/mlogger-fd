<script>
	import QsoEntryForm from '$lib/components/QsoEntryForm.svelte';
	import StatsBar from '$lib/components/StatsBar.svelte';
	import LogTable from '$lib/components/LogTable.svelte';

	function exportCabrillo() {
		window.location.href = '/api/export/cabrillo';
	}
</script>

<div class="header-bar">
	<h1 class="title">FD Logger</h1>
	<button class="export-btn" onclick={exportCabrillo}>Export Cabrillo</button>
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
	}

	.title {
		font-size: 18px;
		font-weight: 700;
		margin: 0;
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
