// frontend/src/lib/ws.test.js
// Tests the WebSocket client module (ws.svelte.js) — Plan 02-03, Task 1
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

// Mock the store module that ws.svelte.js depends on
let mockQsos = [];
let mockFetchStats = vi.fn();
vi.mock('$lib/stores/qso.svelte.js', () => ({
	get qsos() { return mockQsos; },
	set qsos(v) { mockQsos = v; },
	get fetchStats() { return mockFetchStats; },
}));

// Track the last created WebSocket instance so tests can trigger handlers
let lastWs = null;

// Mock WebSocket as a constructor
class MockWebSocket {
	constructor(url) {
		this.url = url;
		this.onopen = null;
		this.onmessage = null;
		this.onclose = null;
		this.onerror = null;
		lastWs = this;
	}
	close() {
		if (this.onclose) {
			this.onclose({ code: 1000, reason: '' });
		}
	}
}
MockWebSocket.CONNECTING = 0;
MockWebSocket.OPEN = 1;
MockWebSocket.CLOSING = 2;
MockWebSocket.CLOSED = 3;

vi.stubGlobal('WebSocket', MockWebSocket);
vi.stubGlobal('location', { protocol: 'http:', host: 'localhost:5173' });
vi.stubGlobal('console', { error: vi.fn() });

describe('WebSocket client module (ws.js)', () => {
	beforeEach(async () => {
		vi.clearAllMocks();
		vi.useFakeTimers();
		mockQsos = [];
		lastWs = null;
		// Clear module cache so we get a fresh import with clean state
		vi.resetModules();
	});

	afterEach(() => {
		vi.useRealTimers();
	});

	// --- Test 1: connectWebSocket() creates new WebSocket pointing to ws://host/ws ---
	it('connectWebSocket creates a WebSocket to ws://localhost:5173/ws', async () => {
		const { connectWebSocket } = await import('$lib/ws.svelte.js');
		connectWebSocket();
		expect(lastWs).toBeDefined();
		expect(lastWs.url).toBe('ws://localhost:5173/ws');
	});

	// --- Test 2: WebSocket onopen sets wsState.connected to true ---
	it('sets wsState.connected to true on WebSocket onopen', async () => {
		const { connectWebSocket, wsState } = await import('$lib/ws.svelte.js');
		expect(wsState.connected).toBe(false);
		connectWebSocket();
		expect(lastWs.onopen).toBeTypeOf('function');
		lastWs.onopen();
		expect(wsState.connected).toBe(true);
	});

	// --- Test 3: WebSocket onclose sets wsState.connected to false, schedules reconnect after 2000ms ---
	it('sets wsState.connected to false on close and schedules reconnect after 2000ms', async () => {
		const { connectWebSocket, wsState } = await import('$lib/ws.svelte.js');
		connectWebSocket();
		lastWs.onopen();
		expect(wsState.connected).toBe(true);
		lastWs.onclose();
		expect(wsState.connected).toBe(false);
		// Advance 2s — reconnect should create a new WebSocket
		vi.advanceTimersByTime(2000);
		expect(lastWs).toBeDefined();
		expect(lastWs.url).toBe('ws://localhost:5173/ws');
	});

	// --- Test 4: WebSocket onmessage for qso_created unshifts QSO into qsos ---
	it('unshifts a QSO into qsos array on qso_created message', async () => {
		const { connectWebSocket } = await import('$lib/ws.svelte.js');
		connectWebSocket();
		const qsoData = {
			type: 'qso_created',
			id: 'abc123',
			timestamp: '2026-06-27T18:00:00Z',
			callsign: 'W1AW',
			band: '20M',
			mode: 'SSB',
			recv_exchange: '2A CT',
			operator: 'K1ABC',
			is_dupe: false,
			points: 1
		};
		lastWs.onmessage({ data: JSON.stringify(qsoData) });
		expect(mockQsos.length).toBe(1);
		expect(mockQsos[0]).toMatchObject({
			id: 'abc123',
			callsign: 'W1AW',
			band: '20M',
			mode: 'SSB'
		});
	});

	// --- Test 5: Duplicate QSO with same id is skipped ---
	it('skips duplicate QSOs via recentIds deduplication', async () => {
		const { connectWebSocket } = await import('$lib/ws.svelte.js');
		connectWebSocket();
		const qsoData = {
			type: 'qso_created',
			id: 'dup456',
			timestamp: '2026-06-27T18:00:00Z',
			callsign: 'W1AW',
			band: '20M',
			mode: 'SSB',
			recv_exchange: '2A CT',
			operator: 'K1ABC',
			is_dupe: false,
			points: 1
		};
		// First message
		lastWs.onmessage({ data: JSON.stringify(qsoData) });
		expect(mockQsos.length).toBe(1);
		// Second message with same ID should be skipped
		lastWs.onmessage({ data: JSON.stringify(qsoData) });
		expect(mockQsos.length).toBe(1);
	});

	// --- Test 6: recentIds Set is pruned when exceeding MAX_RECENT_IDS (100) ---
	it('prunes recentIds Set when exceeding MAX_RECENT_IDS', async () => {
		const { connectWebSocket } = await import('$lib/ws.svelte.js');
		connectWebSocket();
		// Send 101 unique QSOs — the first one should be pruned
		for (let i = 1; i <= 101; i++) {
			lastWs.onmessage({ data: JSON.stringify({
				type: 'qso_created',
				id: `id_${i}`,
				timestamp: '2026-06-27T18:00:00Z',
				callsign: `W${i}AW`,
				band: '20M',
				mode: 'SSB',
				recv_exchange: '2A CT',
				operator: 'OP',
				is_dupe: false,
				points: 1
			})});
		}
		// 101 QSOs should be in the array (all unique)
		expect(mockQsos.length).toBe(101);
		// Send the first ID again — it should be added since id_1 was pruned from recentIds
		lastWs.onmessage({ data: JSON.stringify({
			type: 'qso_created',
			id: 'id_1',
			timestamp: '2026-06-27T18:00:00Z',
			callsign: 'W1AW',
			band: '20M',
			mode: 'SSB',
			recv_exchange: '2A CT',
			operator: 'OP',
			is_dupe: false,
			points: 1
		})});
		expect(mockQsos.length).toBe(102);
	});

	// --- Test 7: On qso_created, fetchStats() is called ---
	it('calls fetchStats on qso_created message', async () => {
		const { connectWebSocket } = await import('$lib/ws.svelte.js');
		connectWebSocket();
		lastWs.onmessage({ data: JSON.stringify({
			type: 'qso_created',
			id: 'stats1',
			timestamp: '2026-06-27T18:00:00Z',
			callsign: 'W1AW',
			band: '20M',
			mode: 'SSB',
			recv_exchange: '2A CT',
			operator: 'K1ABC',
			is_dupe: false,
			points: 1
		})});
		expect(mockFetchStats).toHaveBeenCalledTimes(1);
	});

	// --- Test 8: disconnectWebSocket() stops reconnection and closes connection ---
	it('disconnectWebSocket stops reconnection and closes connection', async () => {
		const { connectWebSocket, disconnectWebSocket, wsState } = await import('$lib/ws.svelte.js');
		connectWebSocket();
		lastWs.onopen();
		expect(wsState.connected).toBe(true);

		disconnectWebSocket();
		expect(wsState.connected).toBe(false);

		// Advance beyond reconnect timeout — should NOT reconnect
		lastWs = null;
		vi.advanceTimersByTime(3000);
		// No new WebSocket created after disconnect
		expect(lastWs).toBeNull();
	});

	// --- Test 9: Non-qso_created types are gracefully ignored ---
	it('ignores non-qso_created message types', async () => {
		const { connectWebSocket } = await import('$lib/ws.svelte.js');
		connectWebSocket();
		// Send a message with a different type
		lastWs.onmessage({ data: JSON.stringify({
			type: 'stats_update',
			stats: { total: 42 }
		})});
		expect(mockQsos.length).toBe(0);
	});

	// --- Test 10: WebSocket onerror logs to console.error but does not crash ---
	it('logs onerror to console.error without crashing', async () => {
		const { connectWebSocket } = await import('$lib/ws.svelte.js');
		connectWebSocket();
		const err = new Error('connection failed');
		expect(() => lastWs.onerror(err)).not.toThrow();
		expect(console.error).toHaveBeenCalledWith('WebSocket error:', err);
	});
});
