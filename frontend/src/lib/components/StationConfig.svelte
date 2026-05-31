<script>
	import { onMount } from 'svelte';
	import { getStationConfig, putStationConfig } from '$lib/api.js';

	let expanded = $state(false);
	let callsign = $state('');
	let cls = $state('');
	let section = $state('');
	let txCount = $state(1);
	let power = $state('LOW');
	let saved = $state(false);
	let saveTimer;

	function toggle() {
		expanded = !expanded;
	}

	async function loadConfig() {
		try {
			const cfg = await getStationConfig();
			callsign = cfg.callsign || '';
			cls = cfg.class || '';
			section = cfg.arrl_section || '';
			txCount = cfg.transmitter_count || 1;
			power = cfg.power_level || 'LOW';
		} catch {
			// silently use defaults
		}
	}

	async function handleSubmit(e) {
		e.preventDefault();
		try {
			await putStationConfig({
				callsign,
				class: cls,
				arrl_section: section,
				transmitter_count: txCount,
				power_level: power,
			});
			saved = true;
			if (saveTimer) clearTimeout(saveTimer);
			saveTimer = setTimeout(() => { saved = false; }, 2000);
		} catch {
			// silently handle error
		}
	}

	onMount(() => {
		loadConfig();
	});
</script>

<div class="station-config">
	<button class="config-toggle" onclick={toggle} aria-label="Config">
		<span class="toggle-icon">⚙</span> Config
	</button>

	{#if expanded}
		<div class="config-panel">
			<form onsubmit={handleSubmit}>
				<div class="config-form">
					<div class="field">
						<label for="cfg-callsign">Callsign</label>
						<input
							id="cfg-callsign"
							type="text"
							bind:value={callsign}
							placeholder="N0CALL"
						/>
					</div>
					<div class="field">
						<label for="cfg-class">Class</label>
						<input
							id="cfg-class"
							type="text"
							bind:value={cls}
							placeholder="1D"
						/>
					</div>
					<div class="field">
						<label for="cfg-section">Section</label>
						<input
							id="cfg-section"
							type="text"
							bind:value={section}
							placeholder="EMA"
						/>
					</div>
					<div class="field">
						<label for="cfg-txcount">Transmitter Count</label>
						<input
							id="cfg-txcount"
							type="number"
							min="1"
							max="20"
							bind:value={txCount}
						/>
					</div>
					<div class="field">
						<label for="cfg-power">Power Level</label>
						<select id="cfg-power" bind:value={power}>
							<option value="LOW">LOW</option>
							<option value="HIGH">HIGH</option>
							<option value="QRP">QRP</option>
						</select>
					</div>
				</div>
				<div class="config-actions">
					<button type="submit" class="save-btn">Save</button>
					{#if saved}
						<span class="saved-msg">Saved!</span>
					{/if}
				</div>
			</form>
		</div>
	{/if}
</div>

<style>
	.station-config {
		display: flex;
		align-items: center;
		gap: 8px;
	}

	.config-toggle {
		padding: 4px 12px;
		font-size: 14px;
		font-weight: 600;
		border: 2px solid var(--color-surface);
		border-radius: 6px;
		background: transparent;
		color: var(--color-surface);
		cursor: pointer;
		white-space: nowrap;
	}

	.config-toggle:hover {
		background: rgba(255,255,255,0.15);
	}

	.toggle-icon {
		margin-right: 4px;
	}

	.config-panel {
		position: absolute;
		top: 44px;
		right: 8px;
		background: var(--color-surface);
		border: 1px solid var(--color-border-light);
		border-radius: 8px;
		padding: 16px;
		box-shadow: 0 4px 16px rgba(0,0,0,0.12);
		z-index: 100;
		min-width: 260px;
	}

	.config-form {
		display: flex;
		flex-direction: column;
		gap: 10px;
	}

	.field {
		display: flex;
		flex-direction: column;
		gap: 3px;
	}

	.field label {
		font-size: 13px;
		font-weight: 600;
		color: var(--color-text-secondary);
	}

	.field input, .field select {
		padding: 6px 10px;
		font-size: 16px;
		border: 1px solid var(--color-border-light);
		border-radius: 6px;
		background: var(--color-bg);
		color: var(--color-primary);
	}

	.field input:focus, .field select:focus {
		outline: none;
		border-color: var(--color-primary);
		box-shadow: 0 0 0 2px var(--color-border-light);
	}

	.config-actions {
		display: flex;
		align-items: center;
		gap: 10px;
		margin-top: 12px;
	}

	.save-btn {
		padding: 6px 20px;
		font-size: 14px;
		font-weight: 600;
		border: none;
		border-radius: 6px;
		background: var(--color-primary);
		color: var(--color-surface);
		cursor: pointer;
	}

	.save-btn:hover {
		background: var(--color-primary);
		filter: brightness(1.15);
	}

	.saved-msg {
		font-size: 14px;
		font-weight: 600;
		color: var(--color-success);
	}
</style>
