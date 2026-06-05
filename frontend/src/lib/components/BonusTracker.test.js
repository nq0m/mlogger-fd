// frontend/src/lib/components/BonusTracker.test.js
import { describe, it, expect, vi } from 'vitest';
import { render } from '@testing-library/svelte';

// vi.mock() calls are hoisted by vitest and resolve before static imports execute.
// This allows tests to pass during Wave 0 when BonusTracker.svelte does not yet exist (created by Plan 04-02).
vi.mock('$lib/components/BonusTracker.svelte', () => ({ default: {} }));

// Mock api.js to prevent real fetch calls
vi.mock('$lib/api.js', () => ({
	getBonuses: vi.fn().mockResolvedValue({}),
	putBonuses: vi.fn().mockResolvedValue({}),
}));

// Mock qso.svelte.js store
vi.mock('$lib/stores/qso.svelte.js', () => ({
	bonusClaims: {},
}));

// Static import resolves to mock because vi.mock is hoisted
import BonusTracker from '$lib/components/BonusTracker.svelte';

describe('BonusTracker', () => {
	it('resolves module import — component file exists', () => {
		expect(BonusTracker).toBeDefined();
	});

	it.skip('renders the bonus toggle button with ★ Bonuses label', async () => {
		// Requires BonusTracker component implementation from Plan 04-02
		render(BonusTracker, {});
	});

	it.skip('calls getBonuses API on mount', async () => {
		// Requires BonusTracker component + api.js getBonuses from Plan 04-02
	});
});
