// WebSocket client module for real-time multi-user QSO sync
// Implements: SYNC-02 (client-side WebSocket listener for real-time QSO updates)
import { qsos, fetchStats } from '$lib/stores/qso.svelte.js';

// Use object-based $state since Svelte 5 forbids reassigning exported $state variables
export const wsState = $state({ connected: false });

let ws = null;
let reconnectTimer = null;
let shouldReconnect = true;

// Deduplication: track recent QSO IDs to avoid double-display
const recentIds = new Set();
const MAX_RECENT_IDS = 100;

function pruneRecentIds() {
	if (recentIds.size > MAX_RECENT_IDS) {
		const entries = [...recentIds];
		entries.slice(0, entries.length - MAX_RECENT_IDS).forEach(id => recentIds.delete(id));
	}
}

export function connectWebSocket() {
	const protocol = location.protocol === 'https:' ? 'wss:' : 'ws:';
	const url = `${protocol}//${location.host}/ws`;
	shouldReconnect = true;

	function connect() {
		if (!shouldReconnect) return;

		ws = new WebSocket(url);

		ws.onopen = () => {
			wsState.connected = true;
			if (reconnectTimer) {
				clearTimeout(reconnectTimer);
				reconnectTimer = null;
			}
		};

		ws.onmessage = (event) => {
			try {
				const data = JSON.parse(event.data);
				if (data.type === 'qso_created') {
					// Deduplicate by QSO ID
					if (recentIds.has(data.id)) return;
					recentIds.add(data.id);
					pruneRecentIds();

					qsos.unshift({
						id: data.id,
						timestamp: data.timestamp,
						callsign: data.callsign,
						band: data.band,
						mode: data.mode,
						recv_exchange: data.recv_exchange,
						operator: data.operator,
						is_dupe: data.is_dupe,
						points: data.points,
					});
					// Refresh stats to keep scoreboard current
					fetchStats();
				}
			} catch (e) {
				console.error('WebSocket message parse error:', e);
			}
		};

		ws.onclose = () => {
			wsState.connected = false;
			// Reconnect after 2 seconds (LAN-appropriate)
			reconnectTimer = setTimeout(connect, 2000);
		};

		ws.onerror = (err) => {
			console.error('WebSocket error:', err);
			// onclose will fire after onerror
		};
	}

	connect();
}

export function disconnectWebSocket() {
	shouldReconnect = false;
	if (reconnectTimer) {
		clearTimeout(reconnectTimer);
		reconnectTimer = null;
	}
	if (ws) {
		ws.close();
		ws = null;
	}
	wsState.connected = false;
}
