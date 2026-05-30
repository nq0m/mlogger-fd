<script>
	import { createQSO } from '$lib/api.js';
	import { addQso } from '$lib/stores/qso.svelte.js';

	const bands = ['160M', '80M', '40M', '20M', '15M', '10M', '6M', '2M', '70CM'];
	const modes = ['CW', 'SSB', 'FM', 'RTTY', 'FT8', 'FT4', 'PSK31'];

	let callsign = $state('');
	let band = $state('20M');
	let mode = $state('SSB');
	let recvExchange = $state('');
	let submitting = $state(false);

	let callsignInput;

	async function handleSubmit(e) {
		e.preventDefault();
		if (submitting) return;
		submitting = true;

		try {
			const result = await createQSO({
				callsign,
				band,
				mode,
				recv_exchange: recvExchange
			});
			addQso(result);
			callsign = '';
			recvExchange = '';
			callsignInput?.focus();
		} catch (err) {
			console.error('Submit failed:', err);
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
		<input
			bind:this={callsignInput}
			bind:value={callsign}
			type="text"
			placeholder="Callsign"
			tabindex="1"
			autofocus
			class="field-callsign"
		/>

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
		background: #fff;
		border-bottom: 2px solid #ddd;
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
		border: 2px solid #ccc;
		border-radius: 6px;
		background: #fff;
		outline: none;
	}

	input:focus, select:focus {
		border-color: #2266cc;
	}

	.field-callsign {
		flex: 0 0 140px;
		min-width: 0;
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
		background: #2266cc;
		color: #fff;
		cursor: pointer;
		white-space: nowrap;
	}

	button:hover:not(:disabled) {
		background: #1a52a3;
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
