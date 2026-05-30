// StationConfig component tests — Plan 02-01
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, cleanup } from '@testing-library/svelte';
import StationConfig from '$lib/components/StationConfig.svelte';

// Mock api.js
vi.mock('$lib/api.js', () => ({
	getStationConfig: vi.fn().mockResolvedValue({
		callsign: 'N0CALL',
		class: '1D',
		arrl_section: 'EMA',
		transmitter_count: 1,
		power_level: 'LOW',
	}),
	putStationConfig: vi.fn().mockResolvedValue({
		callsign: 'K1ABC',
		class: '2A',
		arrl_section: 'CT',
		transmitter_count: 3,
		power_level: 'HIGH',
	}),
}));

import { getStationConfig, putStationConfig } from '$lib/api.js';

describe('StationConfig - API functions', () => {
	beforeEach(() => {
		vi.clearAllMocks();
	});

	afterEach(() => {
		cleanup();
	});

	it('getStationConfig returns parsed JSON from GET /api/station-config', async () => {
		const result = await getStationConfig();
		expect(getStationConfig).toHaveBeenCalledOnce();
		expect(result.callsign).toBe('N0CALL');
		expect(result.class).toBe('1D');
	});

	it('putStationConfig sends PUT with JSON body', async () => {
		const data = {
			callsign: 'W1AW',
			class: '2A',
			arrl_section: 'CT',
			transmitter_count: 3,
			power_level: 'HIGH',
		};
		const result = await putStationConfig(data);
		expect(putStationConfig).toHaveBeenCalledWith(data);
		expect(result.callsign).toBe('K1ABC');
	});
});

describe('StationConfig - Component', () => {
	beforeEach(() => {
		vi.clearAllMocks();
		getStationConfig.mockResolvedValue({
			callsign: 'N0CALL',
			class: '1D',
			arrl_section: 'EMA',
			transmitter_count: 1,
			power_level: 'LOW',
		});
	});

	afterEach(() => {
		cleanup();
	});

	it('renders toggle button with Config label', () => {
		render(StationConfig, {});
		const toggleBtn = screen.getByRole('button', { name: /config/i });
		expect(toggleBtn).toBeDefined();
	});

	it('on mount, fetches GET /api/station-config', async () => {
		render(StationConfig, {});
		await vi.waitFor(() => {
			expect(getStationConfig).toHaveBeenCalled();
		}, { timeout: 1000 });
	});

	it('renders form with 5 inputs when expanded', async () => {
		render(StationConfig, {});

		const toggleBtn = screen.getByRole('button', { name: /config/i });
		// Use native click to trigger Svelte 5 event handling reliably in jsdom
		toggleBtn.click();

		// Allow Svelte to re-render
		await vi.waitFor(() => {
			const formEl = document.querySelector('.config-panel');
			expect(formEl).not.toBeNull();
		}, { timeout: 500 });

		// Find inputs by their IDs
		const callsignInput = document.querySelector('#cfg-callsign');
		const classInput = document.querySelector('#cfg-class');
		const sectionInput = document.querySelector('#cfg-section');
		const txCountInput = document.querySelector('#cfg-txcount');
		const powerSelect = document.querySelector('#cfg-power');

		expect(callsignInput).not.toBeNull();
		expect(classInput).not.toBeNull();
		expect(sectionInput).not.toBeNull();
		expect(txCountInput).not.toBeNull();
		expect(powerSelect).not.toBeNull();
	});

	it('submitting form calls PUT /api/station-config', async () => {
		render(StationConfig, {});

		// Wait for mount fetch
		await vi.waitFor(() => {
			expect(getStationConfig).toHaveBeenCalled();
		}, { timeout: 1000 });

		const toggleBtn = screen.getByRole('button', { name: /config/i });
		toggleBtn.click();

		await vi.waitFor(() => {
			const formEl = document.querySelector('.config-panel');
			expect(formEl).not.toBeNull();
		}, { timeout: 500 });

		const callsignInput = document.querySelector('#cfg-callsign');
		if (callsignInput) {
			callsignInput.value = 'W1AW';
			callsignInput.dispatchEvent(new Event('input', { bubbles: true }));
		}

		const formEl = document.querySelector('.config-panel form');
		if (formEl) {
			formEl.dispatchEvent(new Event('submit', { bubbles: true, cancelable: true }));
		}

		expect(putStationConfig).toHaveBeenCalled();
	});
});
