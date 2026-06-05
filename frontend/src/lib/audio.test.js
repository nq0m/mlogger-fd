// frontend/src/lib/audio.test.js
// Tests for audio.svelte.js — Web Audio API utility module
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

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

// Mock Web Audio API AudioContext
const mockAudioCtxInstance = {
	state: 'suspended',
	resume: vi.fn().mockResolvedValue(undefined),
	decodeAudioData: vi.fn().mockResolvedValue({ duration: 0.5, sampleRate: 44100 }),
	createBufferSource: vi.fn(() => ({
		buffer: null,
		connect: vi.fn(),
		start: vi.fn(),
		disconnect: vi.fn(),
	})),
	destination: {},
	close: vi.fn(),
};
// Use regular function (not arrow) so `new AudioContext()` works in tested code
const MockAudioContext = vi.fn().mockImplementation(function() { return mockAudioCtxInstance; });
vi.stubGlobal('AudioContext', MockAudioContext);

// Mock fetch for audio file loading
const mockFetch = vi.fn().mockResolvedValue({
	ok: true,
	arrayBuffer: vi.fn().mockResolvedValue(new ArrayBuffer(1024)),
});
vi.stubGlobal('fetch', mockFetch);

// Dynamic import to re-import fresh module for each test
async function freshImport() {
	// Clear module cache to get a fresh instance
	return await import('$lib/audio.svelte.js?' + Math.random());
}

describe('Audio utility module (audio.svelte.js)', () => {
	beforeEach(() => {
		vi.clearAllMocks();
		localStorageMock.clear();
		mockAudioCtxInstance.state = 'suspended';
		mockAudioCtxInstance.resume.mockResolvedValue(undefined);
		mockAudioCtxInstance.decodeAudioData.mockResolvedValue({ duration: 0.5, sampleRate: 44100 });
	});

	afterEach(() => {
		vi.restoreAllMocks();
	});

	// Test 1: audioState is an object-based $state export (D-06)
	it('audioState is an object with muted property defaulting to false', async () => {
		const { audioState } = await freshImport();
		expect(audioState).toBeDefined();
		expect(typeof audioState).toBe('object');
		expect(audioState.muted).toBe(false);
	});

	// Test 1 (cont): audioState is reactive $state — mutations are observable
	it('audioState.muted mutations are observable (reactive)', async () => {
		const { audioState } = await freshImport();
		const initial = audioState.muted;
		audioState.muted = true;
		expect(audioState.muted).toBe(true);
		audioState.muted = false;
		expect(audioState.muted).toBe(initial);
	});

	// Test 2: toggleMute() flips audioState.muted and persists to localStorage
	it('toggleMute() flips audioState.muted and persists to localStorage key fdlogger_muted', async () => {
		const { audioState, toggleMute } = await freshImport();
		audioState.muted = false;
		vi.clearAllMocks();

		toggleMute();
		expect(audioState.muted).toBe(true);
		expect(localStorageMock.setItem).toHaveBeenCalledWith('fdlogger_muted', 'true');

		toggleMute();
		expect(audioState.muted).toBe(false);
		expect(localStorageMock.setItem).toHaveBeenCalledWith('fdlogger_muted', 'false');
	});

	// Test 2 (cont): toggleMute persists correct string values
	it('toggleMute persists string "true"/"false" to localStorage', async () => {
		const { audioState, toggleMute } = await freshImport();
		audioState.muted = false;
		vi.clearAllMocks();

		toggleMute();
		expect(localStorageMock.setItem).toHaveBeenCalledWith('fdlogger_muted', 'true');

		vi.clearAllMocks();
		toggleMute();
		expect(localStorageMock.setItem).toHaveBeenCalledWith('fdlogger_muted', 'false');
	});

	// Test 3: playSound creates AudioContext lazily, fetches + decodes + plays via createBufferSource
	it('playSound creates AudioContext lazily (not at module load per D-06)', async () => {
		// Module load should NOT call AudioContext constructor
		await freshImport();
		expect(MockAudioContext).not.toHaveBeenCalled();
	});

	it('playSound("confirm") fetches /audio/confirm.wav, decodes, and plays via createBufferSource', async () => {
		const { playSound } = await freshImport();
		await playSound('confirm');
		expect(mockFetch).toHaveBeenCalledWith('/audio/confirm.wav');
		expect(mockAudioCtxInstance.decodeAudioData).toHaveBeenCalled();
		expect(mockAudioCtxInstance.createBufferSource).toHaveBeenCalled();
	});

	it('playSound caches decoded AudioBuffers — second call does not re-fetch', async () => {
		const { playSound } = await freshImport();
		await playSound('confirm');
		const fetchCount = mockFetch.mock.calls.length;
		const decodeCount = mockAudioCtxInstance.decodeAudioData.mock.calls.length;

		await playSound('confirm');
		// Second call should NOT fetch or decode again (uses cached buffer)
		expect(mockFetch.mock.calls.length).toBe(fetchCount);
		expect(mockAudioCtxInstance.decodeAudioData.mock.calls.length).toBe(decodeCount);
	});

	// Test 4: playSound returns silently when muted
	it('playSound returns silently (no fetch, no AudioContext) when muted', async () => {
		const { audioState, playSound } = await freshImport();
		audioState.muted = true;

		const fetchBefore = mockFetch.mock.calls.length;
		const decodeBefore = mockAudioCtxInstance.decodeAudioData.mock.calls.length;

		await playSound('confirm');

		// Should NOT have fetched or decoded anything
		expect(mockFetch.mock.calls.length).toBe(fetchBefore);
		expect(mockAudioCtxInstance.decodeAudioData.mock.calls.length).toBe(decodeBefore);
	});

	it('playSound does not throw when muted (returns silently)', async () => {
		const { audioState, playSound } = await freshImport();
		audioState.muted = true;
		// Should not throw
		await expect(playSound('confirm')).resolves.toBeUndefined();
	});

	// Test 5: AudioContext.resume() is called if context state is 'suspended' (autoplay policy)
	it('calls AudioContext.resume() when context state is suspended (autoplay policy)', async () => {
		mockAudioCtxInstance.state = 'suspended';
		const { playSound } = await freshImport();
		await playSound('confirm');
		expect(mockAudioCtxInstance.resume).toHaveBeenCalled();
	});

	it('does not call AudioContext.resume() when context is already running', async () => {
		mockAudioCtxInstance.state = 'running';
		const { playSound } = await freshImport();
		await playSound('confirm');
		expect(mockAudioCtxInstance.resume).not.toHaveBeenCalled();
	});

	// Test: Initial localStorage state overrides default
	it('reads initial mute state from localStorage on module load (default unmuted per D-07)', async () => {
		// Pre-set localStorage to muted before import
		localStorageMock.setItem('fdlogger_muted', 'true');
		const { audioState } = await freshImport();
		// Module init reads localStorage and sets muted to true
		expect(audioState.muted).toBe(true);
	});

	it('defaults to unmuted when localStorage has no stored value', async () => {
		// localStorage is empty (cleared in beforeEach)
		const { audioState } = await freshImport();
		expect(audioState.muted).toBe(false);
	});

	// Test: Error handling — playSound never throws
	it('playSound handles fetch errors gracefully (console.warn, no throw)', async () => {
		mockFetch.mockRejectedValueOnce(new Error('Network error'));
		const { playSound } = await freshImport();
		// Must not throw
		await expect(playSound('confirm')).resolves.toBeUndefined();
	});

	it('playSound handles decode errors gracefully (console.warn, no throw)', async () => {
		mockAudioCtxInstance.decodeAudioData.mockRejectedValueOnce(new Error('Decode error'));
		const { playSound } = await freshImport();
		// Must not throw
		await expect(playSound('confirm')).resolves.toBeUndefined();
	});
});
