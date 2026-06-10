import { getQueuedQsos, getQueueCount, clearQueued } from '$lib/db.js';
import { syncBatch } from '$lib/api.js';

export const queueState = $state({ queueLength: 0, syncing: false });

export async function refreshQueueCount() {
	queueState.queueLength = await getQueueCount();
}

export async function flushSyncQueue() {
	if (queueState.syncing) return;
	queueState.syncing = true;
	try {
		const queued = await getQueuedQsos();
		if (queued.length === 0) {
			queueState.syncing = false;
			return;
		}
		const qsos = queued.map(q => ({ client_id: q.client_id, ...q.qso }));
		await syncBatch(qsos);
		await clearQueued();
		await refreshQueueCount();
	} catch {
		// queue stays intact for retry
	} finally {
		queueState.syncing = false;
	}
}

let syncTimer = null;

export function startSync() {
	flushSyncQueue();
	if (!syncTimer) {
		syncTimer = setInterval(flushSyncQueue, 30000);
	}
}

export function stopSync() {
	if (syncTimer) {
		clearInterval(syncTimer);
		syncTimer = null;
	}
}

refreshQueueCount();
