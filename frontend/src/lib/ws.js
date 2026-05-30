// WebSocket client module placeholder — will be implemented by Plan 02-03
export let wsConnected = false;

export function connectWebSocket() {
  const protocol = location.protocol === 'https:' ? 'wss:' : 'ws:';
  const ws = new WebSocket(`${protocol}//${location.host}/ws`);
  return ws;
}

export function disconnectWebSocket() {
  wsConnected = false;
}
