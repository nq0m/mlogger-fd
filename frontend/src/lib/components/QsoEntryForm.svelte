<script>
	import { createQSO, checkDupe } from '$lib/api.js';
	import { addQso, addQsoOffline, fetchStats } from '$lib/stores/qso.svelte.js';
	import { refreshQueueCount } from '$lib/sync.svelte.js';
	import { wsState } from '$lib/ws.svelte.js';
	import { offlineDupeCheck } from '$lib/db.js';

	const bands = ['160M', '80M', '40M', '20M', '15M', '10M', '6M', '2M', '70CM'];
	const modes = ['CW', 'SSB', 'FM', 'RTTY', 'FT8', 'FT4', 'PSK31'];

	let callsign = $state('');
	let band = $state('20M');
	let mode = $state('SSB');
	let recvExchange = $state('');
	let submitting = $state(false);
	let dupeWarning = $state('');

	let callsignInput;
	let lastSubmitTime = 0;

	function validateCallsign(call) {
		if (!call) return 'Callsign required';
		if (call.length < 2) return 'Callsign too short';
		return '';
	}

	async function handleCheckDupe() {
		if (callsign.length < 2) {
			dupeWarning = '';
			return;
		}

		if (!wsState.connected) {
			const result = await offlineDupeCheck(callsign, band, mode);
			if (result.is_dupe) {
				dupeWarning = 'DUPE: Already worked on this band/mode';
			} else {
				dupeWarning = '';
			}
			return;
		}

		try {
			const result = await checkDupe(callsign, band, mode);
			if (result.is_dupe) {
				dupeWarning = 'DUPE: Already worked on this band/mode';
			} else if (result.similar_calls && result.similar_calls.length > 0) {
				dupeWarning = 'Similar calls: ' + result.similar_calls.join(', ');
			} else {
				dupeWarning = '';
			}
		} catch {
			dupeWarning = '';
		}
	}

	async function handleSubmit(e) {
		e.preventDefault();
		if (submitting) return;
		if (Date.now() - lastSubmitTime < 1000) return;
		lastSubmitTime = Date.now();

		const validMsg = validateCallsign(callsign);
		if (validMsg) {
			dupeWarning = validMsg;
			return;
		}

		submitting = true;

		try {
			await handleCheckDupe();

			const result = await createQSO({
				callsign,
				band,
				mode,
				recv_exchange: recvExchange,
				operator: localStorage.getItem('fdlogger_operator') || ''
			});
			addQso(result);
			fetchStats();
			if (result.is_dupe) {
				dupeWarning = 'Logged as duplicate (0 points)';
			} else {
				callsign = '';
				recvExchange = '';
				dupeWarning = '';
				callsignInput?.focus();
			}
		} catch (err) {
			if (err instanceof TypeError || (err.message && err.message.includes('fetch'))) {
				await addQsoOffline({
					callsign,
					band,
					mode,
					recv_exchange: recvExchange,
					operator: localStorage.getItem('fdlogger_operator') || ''
				});
				refreshQueueCount();
				callsign = '';
				recvExchange = '';
				dupeWarning = '';
				callsignInput?.focus();
			} else {
				console.error('Submit failed:', err);
			}
		} finally {
			submitting = false;
		}
	}

	function handleKeydown(e) {
		if (e.ctrlKey && e.key === 'Enter') {
			e.preventDefault();
			handleSubmit(e);
		}
	}
</script>

<form class="qso-form" onsubmit={handleSubmit} onkeydown={handleKeydown}>
	<div class="form-row">
		<div class="callsign-group">
			<input
				bind:this={callsignInput}
				bind:value={callsign}
				type="text"
				placeholder="Callsign"
				tabindex="1"
				autofocus
				class="field-callsign"
				onblur={handleCheckDupe}
			/>
			{#if dupeWarning}
				<span class="dupe-warning">{dupeWarning}</span>
			{/if}
		</div>

		<select bind:value={band} tabindex="2">
			{#each bands as b}
				<option value={b}>{b}</option>
			{/each}
		</select>

		<select bind:value={mode} tabindex="3">
			{#each modes as m}
				<option value={m}>{m}</option>
			{/each}
		</select>

		<input
			bind:value={recvExchange}
			type="text"
			placeholder="Exchange (e.g., 2A NH)"
			tabindex="4"
		/>

		<button type="submit" disabled={submitting}>
			Log QSO <small>(Ctrl+Enter)</small>
		</button>
	</div>
</form>

<style>
	.qso-form {
		padding: 12px 16px;
		background: var(--color-surface);
		border-bottom: 2px solid var(--color-border-strong);
		position: sticky;
		top: 0;
		z-index: 10;
	}

	.form-row {
		display: flex;
		gap: 8px;
		align-items: center;
		flex-wrap: wrap;
	}

	input, select {
		padding: 10px 12px;
		font-size: 16px;
		border: 2px solid var(--color-border);
		border-radius: 6px;
		background: var(--color-surface);
		outline: none;
	}

	input:focus, select:focus {
		border-color: var(--color-accent);
	}

	.field-callsign {
		flex: 0 0 140px;
		min-width: 0;
	}

	.callsign-group {
		display: flex;
		flex-direction: column;
		gap: 2px;
	}

	.dupe-warning {
		font-size: 11px;
		color: var(--color-danger);
		font-weight: 600;
		white-space: nowrap;
	}

	input[type="text"] {
		flex: 0 0 140px;
		min-width: 0;
	}

	select {
		min-width: 80px;
	}

	button {
		padding: 10px 20px;
		font-size: 16px;
		font-weight: 600;
		border: none;
		border-radius: 6px;
		background: var(--color-accent);
		color: var(--color-surface);
		cursor: pointer;
		white-space: nowrap;
	}

	button:hover:not(:disabled) {
		background: var(--color-primary);
	}

	button:disabled {
		opacity: 0.6;
		cursor: default;
	}

	button small {
		font-weight: 400;
		font-size: 12px;
		opacity: 0.8;
	}
</style>
