import Dexie from 'dexie';

const db = new Dexie('FDLogger');
db.version(1).stores({
	queued_qsos: 'client_id, created_at'
});

export async function enqueueQso(qso) {
	const client_id = crypto.randomUUID();
	await db.queued_qsos.put({ client_id, qso, created_at: new Date().toISOString() });
	return client_id;
}

export function getQueuedQsos() {
	return db.queued_qsos.orderBy('created_at').toArray();
}

export function getQueueCount() {
	return db.queued_qsos.count();
}

export function clearQueued() {
	return db.queued_qsos.clear();
}

export default db;
