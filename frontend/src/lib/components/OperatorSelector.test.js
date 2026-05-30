// frontend/src/lib/components/OperatorSelector.test.js
// TDD tests for OperatorSelector component — Plan 02-03, Task 2
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, cleanup } from '@testing-library/svelte';

import OperatorSelector from '$lib/components/OperatorSelector.svelte';

// Mock localStorage
const localStorageMock = (() => {
	let store = {};
	return {
		getItem: vi.fn((key) => store[key] || null),
		setItem: vi.fn((key, value) => { store[key] = value; }),
		removeItem: vi.fn((key) => { delete store[key]; }),
		clear: vi.fn(() => { store = {}; }),
	};
})();
vi.stubGlobal('localStorage', localStorageMock);

describe('OperatorSelector', () => {
	beforeEach(() => {
		localStorageMock.clear();
		vi.clearAllMocks();
	});

	afterEach(() => {
		cleanup();
	});

	// Test 1: Renders a text input with label "Operator:" and placeholder
	it('renders a text input with label Operator: and placeholder', () => {
		render(OperatorSelector, {});
		const input = screen.getByRole('textbox', { name: /operator/i });
		expect(input).toBeDefined();
		expect(input.placeholder).toBe('Your callsign or name');
	});

	// Test 2: Typing in the input updates the displayed value reactively
	it('updates displayed value reactively when typing', async () => {
		render(OperatorSelector, {});
		const input = screen.getByRole('textbox', { name: /operator/i });
		await fireEvent.input(input, { target: { value: 'W1AW' } });
		expect(input.value).toBe('W1AW');
	});

	// Test 3: On input change, value is saved to localStorage under key 'fdlogger_operator'
	it('saves value to localStorage on input change', async () => {
		render(OperatorSelector, {});
		const input = screen.getByRole('textbox', { name: /operator/i });
		await fireEvent.input(input, { target: { value: 'K1ABC' } });
		expect(localStorageMock.setItem).toHaveBeenCalledWith('fdlogger_operator', 'K1ABC');
	});

	// Test 4: On mount, reads initial value from localStorage (empty string if not set)
	it('reads initial value from localStorage on mount', async () => {
		localStorageMock.getItem.mockReturnValueOnce('N0CALL');
		render(OperatorSelector, {});
		expect(localStorageMock.getItem).toHaveBeenCalledWith('fdlogger_operator');
	});

	it('shows saved operator value from localStorage on mount', async () => {
		// Prime localStorage with a value
		localStorageMock.setItem('fdlogger_operator', 'W1AW');
		localStorageMock.getItem.mockReturnValueOnce('W1AW');
		render(OperatorSelector, {});
		const input = screen.getByRole('textbox', { name: /operator/i });
		expect(input.value).toBe('W1AW');
	});

	// Test 5: (Verified in QsoEntryForm integration — operator sent in createQSO payload)
	// Test 6: OperatorSelector maxlength is 20 characters
	it('has maxlength of 20 characters on the input', async () => {
		render(OperatorSelector, {});
		const input = screen.getByRole('textbox', { name: /operator/i });
		expect(input.maxLength).toBe(20);
	});
});
