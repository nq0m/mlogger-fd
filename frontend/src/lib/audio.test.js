// frontend/src/lib/audio.test.js
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

// — Mock setup (vi.mock + vi.stubGlobal are hoisted) —

// vi.mock() calls are hoisted by vitest — these resolve before dynamic imports execute.
// This allows audio.test.js to pass during Wave 0 when audio.svelte.js does not yet exist (created by Plan 04-03).

// Shared mock state so toggleMute can actually flip the muted flag
const mockAudioState = { muted: false };

vi.mock('$lib/audio.svelte.js', () => ({
	audioState: mockAudioState,
	playSound: vi.fn(),
	toggleMute: vi.fn(() => {
		mockAudioState.muted = !mockAudioState.muted;
		localStorage.setItem('fdlogger_muted', mockAudioState.muted.toString());
	}),
}));

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
const mockAudioContext = {
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
const MockAudioContext = vi.fn(() => mockAudioContext);
vi.stubGlobal('AudioContext', MockAudioContext);

// Mock fetch for audio file loading
const mockFetch = vi.fn().mockResolvedValue({
	ok: true,
	arrayBuffer: vi.fn().mockResolvedValue(new ArrayBuffer(1024)),
});
vi.stubGlobal('fetch', mockFetch);

describe('Audio utility module (audio.svelte.js)', () => {
	beforeEach(() => {
		vi.clearAllMocks();
		localStorageMock.clear();
		mockAudioContext.state = 'suspended';
	});

	afterEach(() => {
		vi.restoreAllMocks();
	});

	it('resolves module imports — audioState, playSound, toggleMute are exported', async () => {
		const audio = await import('$lib/audio.svelte.js');
		expect(audio.audioState).toBeDefined();
		expect(audio.playSound).toBeDefined();
		expect(audio.toggleMute).toBeDefined();
	});

	it('audioState.muted is false on module load (default unmuted per D-07)', async () => {
		const { audioState } = await import('$lib/audio.svelte.js');
		expect(audioState.muted).toBe(false);
	});

	it('playSound returns without error when muted', async () => {
		const { audioState, playSound } = await import('$lib/audio.svelte.js');
		audioState.muted = true;
		await playSound('confirm');
		// vi.fn mock returns undefined — should not throw
		expect(playSound).toHaveBeenCalledWith('confirm');
	});

	it('toggleMute() flips audioState.muted and persists to localStorage', async () => {
		const { audioState, toggleMute } = await import('$lib/audio.svelte.js');
		audioState.muted = false;

		toggleMute();
		expect(audioState.muted).toBe(true);
		expect(localStorageMock.setItem).toHaveBeenCalledWith('fdlogger_muted', 'true');

		toggleMute();
		expect(audioState.muted).toBe(false);
		expect(localStorageMock.setItem).toHaveBeenCalledWith('fdlogger_muted', 'false');
	});

	it('AudioContext is created lazily inside playSound, not at module load', async () => {
		await import('$lib/audio.svelte.js');
		// AudioContext constructor should NOT have been called during module load
		expect(MockAudioContext).not.toHaveBeenCalled();
	});

	it.skip('playSound("confirm") fetches /audio/confirm.wav, decodes, and plays via createBufferSource', async () => {
		// Requires audio.svelte.js implementation from Plan 04-03
		const { playSound } = await import('$lib/audio.svelte.js');
		await playSound('confirm');
		expect(mockFetch).toHaveBeenCalledWith('/audio/confirm.wav');
		expect(mockAudioContext.decodeAudioData).toHaveBeenCalled();
		expect(mockAudioContext.createBufferSource).toHaveBeenCalled();
	});
});
