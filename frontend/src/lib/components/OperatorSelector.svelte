<script>
	// OperatorSelector component — per-session operator identity with localStorage persistence
	// Implements: CONF-02 (per-session operator identity)
	let operator = $state(localStorage.getItem('fdlogger_operator') || '');

	function saveOperator() {
		localStorage.setItem('fdlogger_operator', operator);
	}
</script>

<div class="operator-selector">
	<label for="operator-input">Operator:</label>
	<input
		id="operator-input"
		type="text"
		bind:value={operator}
		oninput={saveOperator}
		placeholder="Your callsign or name"
		maxlength="20"
	/>
</div>

<style>
	.operator-selector {
		display: flex;
		align-items: center;
		gap: 6px;
		font-size: 14px;
	}
	.operator-selector input {
		width: 140px;
		padding: 4px 8px;
		border: 1px solid #ccc;
		border-radius: 4px;
		font-size: 14px;
	}
</style>
