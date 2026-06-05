<script>
	import { onMount } from 'svelte';
	import { getBonuses, putBonuses } from '$lib/api.js';
	import { bonusClaims } from '$lib/stores/qso.svelte.js';

	let expanded = $state(false);
	let saved = $state(false);
	let saveTimer;

	// Hardcoded bonus name lookup and counted bonus set from ARRL Field Day rules
	const bonusNames = {
		emergency_power: '100% Emergency Power',
		media_publicity: 'Media Publicity',
		public_location: 'Public Location',
		public_info_table: 'Public Information Table',
		message_to_sm: 'Message to Section Manager',
		message_handling: 'Message Handling',
		satellite_qso: 'Satellite QSO',
		alternate_power: 'Alternate Power',
		w1aw_bulletin: 'W1AW Bulletin',
		educational_activity: 'Educational Activity',
		official_visit: 'Elected Official Visit',
		agency_visit: 'Agency Representative Visit',
		gota_bonus: 'GOTA Station Bonus',
		web_submission: 'Web Submission',
		youth_participation: 'Youth Participation',
		social_media: 'Social Media Promotion',
		safety_officer: 'Safety Officer',
		site_responsibilities: 'Site Responsibilities'
	};

	// Set of bonus IDs that have a count input (IsCounted bonuses)
	const countedBonuses = new Set([
		'emergency_power',
		'message_handling',
		'youth_participation',
		'gota_bonus'
	]);

	function toggle() {
		expanded = !expanded;
	}

	async function loadBonuses() {
		// Try localStorage first for instant display
		try {
			const stored = localStorage.getItem('fdlogger_bonus_claims');
			if (stored) {
				const parsed = JSON.parse(stored);
				for (const [key, val] of Object.entries(parsed)) {
					bonusClaims[key] = val;
				}
			}
		} catch {
			// silently ignore localStorage errors
		}

		// Fetch from server (source of truth per D-04)
		try {
			const data = await getBonuses();
			if (data && typeof data === 'object') {
				for (const [key, val] of Object.entries(data)) {
					bonusClaims[key] = {
						claimed: val.claimed || false,
						count: val.count || 0
					};
				}
			}
		} catch {
			// silently use localStorage/defaults
		}

		// Initialize any missing bonus IDs with defaults
		for (const bonusId of Object.keys(bonusNames)) {
			if (!bonusClaims[bonusId]) {
				bonusClaims[bonusId] = { claimed: false, count: 0 };
			}
		}
	}

	async function handleSubmit(e) {
		e.preventDefault();
		try {
			await putBonuses(bonusClaims);
			saved = true;

			// Backup to localStorage
			try {
				localStorage.setItem('fdlogger_bonus_claims', JSON.stringify(bonusClaims));
			} catch {
				// silently ignore localStorage errors
			}

			if (saveTimer) clearTimeout(saveTimer);
			saveTimer = setTimeout(() => { saved = false; }, 2000);
		} catch {
			// silently handle error
		}
	}

	onMount(() => {
		loadBonuses();
	});
</script>

<div class="bonus-tracker">
	<button class="config-toggle" onclick={toggle} aria-label="Bonuses">
		<span class="toggle-icon">★</span> Bonuses
	</button>

	{#if expanded}
		<div class="config-panel">
			<form onsubmit={handleSubmit}>
				<div class="config-form">
					{#each Object.keys(bonusNames) as bonusId}
						{@const claim = bonusClaims[bonusId] || { claimed: false, count: 0 }}
						{@const name = bonusNames[bonusId] || bonusId}
						<div class="bonus-row">
							<label class="bonus-label">
								<input
									type="checkbox"
									checked={claim.claimed}
									onchange={(e) => {
										if (!bonusClaims[bonusId]) bonusClaims[bonusId] = { claimed: false, count: 0 };
										bonusClaims[bonusId].claimed = e.target.checked;
									}}
								/>
								<span class="bonus-name">{name}</span>
							</label>
							{#if countedBonuses.has(bonusId) && claim.claimed}
								<input
									type="number"
									min="0"
									class="bonus-count"
									value={claim.count || 0}
									onchange={(e) => {
										if (!bonusClaims[bonusId]) bonusClaims[bonusId] = { claimed: false, count: 0 };
										bonusClaims[bonusId].count = parseInt(e.target.value) || 0;
									}}
									aria-label="Count for {name}"
								/>
							{/if}
						</div>
					{/each}
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
	.bonus-tracker {
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
		min-width: 300px;
		max-height: 70vh;
		overflow-y: auto;
	}

	.config-form {
		display: flex;
		flex-direction: column;
		gap: 6px;
	}

	.bonus-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 8px;
		padding: 4px 0;
		border-bottom: 1px solid var(--color-border-light);
	}

	.bonus-row:last-child {
		border-bottom: none;
	}

	.bonus-label {
		display: flex;
		align-items: center;
		gap: 8px;
		flex: 1;
		cursor: pointer;
		font-size: 13px;
		color: var(--color-text-primary);
	}

	.bonus-label input[type="checkbox"] {
		width: 16px;
		height: 16px;
		cursor: pointer;
		accent-color: var(--color-primary);
	}

	.bonus-name {
		font-weight: 500;
	}

	.bonus-count {
		width: 60px;
		padding: 4px 6px;
		font-size: 14px;
		border: 1px solid var(--color-border-light);
		border-radius: 4px;
		background: var(--color-bg);
		color: var(--color-primary);
		text-align: center;
	}

	.bonus-count:focus {
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
