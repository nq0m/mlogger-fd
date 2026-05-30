// frontend/src/lib/ws.test.js
// Tests the real ws.js placeholder module (Plan 02-03 will expand it)
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

// Mock the store module that ws.js depends on (created in Phase 1, mocked to avoid pulling in Svelte runes)
vi.mock('$lib/stores/qso.svelte.js', () => {
  const qsos = [];
  const stats = {};
  const fetchStats = vi.fn();
  return { qsos, stats, fetchStats };
});

// Mock WebSocket globally before importing ws.js
const mockWebSocket = {
  readyState: 0,
  send: vi.fn(),
  close: vi.fn(),
  addEventListener: vi.fn(),
  removeEventListener: vi.fn(),
};
const MockWebSocket = vi.fn(function() { return mockWebSocket; });
MockWebSocket.CONNECTING = 0;
MockWebSocket.OPEN = 1;
MockWebSocket.CLOSING = 2;
MockWebSocket.CLOSED = 3;

vi.stubGlobal('WebSocket', MockWebSocket);
vi.stubGlobal('location', { protocol: 'http:', host: 'localhost:5173' });

describe('WebSocket client module (ws.js)', () => {
  it('resolves module imports — ws.js exists and exports expected symbols', async () => {
    const ws = await import('$lib/ws.js');
    expect(ws.connectWebSocket).toBeDefined();
    expect(ws.disconnectWebSocket).toBeDefined();
    expect(ws.wsConnected).toBeDefined();
  });

  it('wsConnected is initially false', async () => {
    const { wsConnected } = await import('$lib/ws.js');
    expect(wsConnected).toBe(false);
  });

  it('connectWebSocket creates a WebSocket to ws://localhost:5173/ws', async () => {
    const { connectWebSocket } = await import('$lib/ws.js');
    connectWebSocket();
    expect(MockWebSocket).toHaveBeenCalledWith('ws://localhost:5173/ws');
  });

  it.skip('onmessage handler processes qso_created events and unshifts into qsos array', async () => {
    // Requires ws.js implementation from Plan 02-03
    expect(true).toBe(true);
  });
});
