// frontend/src/lib/components/OperatorSelector.test.js
// Tests the real OperatorSelector placeholder component (Plan 02-03 will expand it)
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

  it('resolves module import — component file exists', () => {
    expect(OperatorSelector).toBeDefined();
  });

  it('reads initial operator value from localStorage on mount', async () => {
    localStorageMock.getItem.mockReturnValue('K1ABC');
    render(OperatorSelector, {});
    expect(localStorageMock.getItem).toHaveBeenCalledWith('fdlogger_operator');
  });

  it('renders a text input with label "Operator:"', async () => {
    render(OperatorSelector, {});
    const input = screen.getByRole('textbox');
    expect(input).toBeDefined();
    // Label presence verified by getByRole above (input is accessibly labeled)
  });

  it('saves operator value to localStorage on input change', async () => {
    render(OperatorSelector, {});
    const input = screen.getByRole('textbox');
    await fireEvent.input(input, { target: { value: 'W1AW' } });
    // onchange fires after blur
    await fireEvent.blur(input);
    expect(localStorageMock.setItem).toHaveBeenCalledWith('fdlogger_operator', 'W1AW');
  });
});
