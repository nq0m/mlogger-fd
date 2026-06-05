// audio.svelte.js placeholder — created by Plan 04-00 for test import resolution
// Full implementation in Plan 04-03

// Object-based $state for mute (Svelte 5 pattern)
export const audioState = $state({ muted: false });
// Lazy-initialized AudioContext (browsers require user gesture)
let audioCtx = null;
const buffers = {};

function ensureContext() {
    if (!audioCtx) {
        audioCtx = new AudioContext();
        const stored = localStorage.getItem('fdlogger_muted');
        if (stored !== null) audioState.muted = stored === 'true';
    }
    if (audioCtx.state === 'suspended') {
        audioCtx.resume();
    }
}

export async function playSound(name) {
    if (audioState.muted) return;
    ensureContext();
    try {
        const url = `/audio/${name}.wav`;
        const response = await fetch(url);
        const arrayBuffer = await response.arrayBuffer();
        if (!buffers[name]) {
            buffers[name] = await audioCtx.decodeAudioData(arrayBuffer);
        }
        const source = audioCtx.createBufferSource();
        source.buffer = buffers[name];
        source.connect(audioCtx.destination);
        source.start(0);
    } catch (e) {
        console.warn('Audio playback failed:', e);
    }
}

export function toggleMute() {
    audioState.muted = !audioState.muted;
    localStorage.setItem('fdlogger_muted', audioState.muted.toString());
}

// Initialize mute from localStorage on module load
if (typeof localStorage !== 'undefined') {
    const stored = localStorage.getItem('fdlogger_muted');
    if (stored !== null) audioState.muted = stored === 'true';
}
