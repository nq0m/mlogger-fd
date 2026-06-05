// Audio utility module — Web Audio API wrapper with mute state
// Implements: UX-03 (audio alerts for operator feedback), D-06 (audio file loading), D-07 (mute toggle), D-08 (own-QSO-only sounds)

// Object-based $state for mute (Svelte 5 pattern — primitive $state would break on reassignment)
// Pattern follows ws.svelte.js line 7: export const wsState = $state({ connected: false })
export const audioState = $state({ muted: false });

// Lazy-initialized AudioContext (browsers require user gesture per Chrome autoplay policy)
let audioCtx = null;
const buffers = {};  // Cache decoded AudioBuffers (D-06: decode-once, play-many)

function ensureContext() {
	if (!audioCtx) {
		audioCtx = new AudioContext();
	}
	// Resume if suspended (autoplay policy compliance — Chrome 66+)
	if (audioCtx.state === 'suspended') {
		audioCtx.resume();
	}
}

async function loadSound(name) {
	// Return cached buffer if already decoded (D-06: decode once, play many)
	if (buffers[name]) return buffers[name];
	const response = await fetch(`/audio/${name}.wav`);
	const arrayBuffer = await response.arrayBuffer();
	buffers[name] = await audioCtx.decodeAudioData(arrayBuffer);
	return buffers[name];
}

export async function playSound(name) {
	// Mute guard — return silently when muted (D-07)
	if (audioState.muted) return;
	ensureContext();
	try {
		const buffer = await loadSound(name);
		const source = audioCtx.createBufferSource();
		source.buffer = buffer;
		source.connect(audioCtx.destination);
		source.start(0);
	} catch (e) {
		// Silent fallback — never throw to caller (missing audio files, decode errors, etc.)
		console.warn('Audio playback failed:', e);
	}
}

export function toggleMute() {
	audioState.muted = !audioState.muted;
	localStorage.setItem('fdlogger_muted', audioState.muted.toString());
}

// Initialize mute state from localStorage on module load (D-07: default unmuted)
// SSR guard: typeof localStorage !== 'undefined' check for server-side rendering safety
if (typeof localStorage !== 'undefined') {
	const stored = localStorage.getItem('fdlogger_muted');
	if (stored !== null) audioState.muted = stored === 'true';
}
