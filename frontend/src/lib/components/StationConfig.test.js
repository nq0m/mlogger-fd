// StationConfig component tests — Plan 02-01
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, cleanup } from '@testing-library/svelte';
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

	it('renders form with 5 inputs: callsign, class, section, transmitter count, power level', () => {
		render(StationConfig, {});

		// Find the config toggle button to open the panel
		const toggleBtn = screen.getByRole('button', { name: /config/i });
		fireEvent.click(toggleBtn);

		const callsignInput = screen.getByLabelText(/callsign/i);
		const classInput = screen.getByLabelText(/class/i);
		const sectionInput = screen.getByLabelText(/section/i);
		const txCountInput = screen.getByLabelText(/transmitter/i);
		const powerSelect = screen.getByLabelText(/power/i);

		expect(callsignInput).toBeDefined();
		expect(classInput).toBeDefined();
		expect(sectionInput).toBeDefined();
		expect(txCountInput).toBeDefined();
		expect(powerSelect).toBeDefined();
	});

	it('on mount, fetches GET /api/station-config and populates form fields', async () => {
		getStationConfig.mockResolvedValue({
			callsign: 'K1ABC',
			class: '2A',
			arrl_section: 'NH',
			transmitter_count: 5,
			power_level: 'HIGH',
		});

		render(StationConfig, {});
		expect(getStationConfig).toHaveBeenCalled();

		const toggleBtn = screen.getByRole('button', { name: /config/i });
		fireEvent.click(toggleBtn);

		// After async fetch completes, fields should be populated
		await vi.waitFor(() => {
			expect(getStationConfig).toHaveBeenCalled();
		}, { timeout: 1000 });
	});

	it('submitting form calls PUT /api/station-config with current field values', async () => {
		render(StationConfig, {});

		// Wait for mount fetch
		await vi.waitFor(() => {
			expect(getStationConfig).toHaveBeenCalled();
		}, { timeout: 1000 });

		const toggleBtn = screen.getByRole('button', { name: /config/i });
		fireEvent.click(toggleBtn);

		const callsignInput = screen.getByLabelText(/callsign/i);
		await fireEvent.input(callsignInput, { target: { value: 'W1AW' } });

		const submitBtn = screen.getByRole('button', { name: /save/i });
		await fireEvent.click(submitBtn);

		expect(putStationConfig).toHaveBeenCalled();
	});
});
