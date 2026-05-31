import Dexie from 'dexie';

const db = new Dexie('FDLogger');
db.version(1).stores({
	queued_qsos: 'client_id, created_at'
});
db.version(2).stores({
	queued_qsos: 'client_id, created_at',
	cached_qsos: 'client_id, [callsign+band+mode], timestamp'
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

export function loadCachedQsos() {
	return db.cached_qsos.orderBy('timestamp').reverse().toArray();
}

export async function populateCache(qsos) {
	const entries = qsos.map(q => ({
		client_id: q.client_id || (crypto.randomUUID()),
		id: q.id,
		timestamp: q.timestamp,
		callsign: (q.callsign || '').toUpperCase(),
		band: q.band,
		mode: q.mode,
		recv_exchange: q.recv_exchange,
		operator: q.operator || '',
		is_dupe: q.is_dupe,
		points: q.points
	}));
	await db.cached_qsos.bulkPut(entries);
}

export async function addToCache(qso) {
	await db.cached_qsos.put({
		client_id: qso.client_id || (crypto.randomUUID()),
		id: qso.id,
		timestamp: qso.timestamp,
		callsign: (qso.callsign || '').toUpperCase(),
		band: qso.band,
		mode: qso.mode,
		recv_exchange: qso.recv_exchange,
		operator: qso.operator || '',
		is_dupe: qso.is_dupe,
		points: qso.points
	});
}

export async function offlineDupeCheck(callsign, band, mode) {
	const ucCallsign = callsign.toUpperCase();
	const cachedCount = await db.cached_qsos
		.where({ callsign: ucCallsign, band, mode })
		.count();

	if (cachedCount > 0) return { is_dupe: true };

	const queued = await db.queued_qsos.toArray();
	for (const entry of queued) {
		const q = entry.qso;
		if (q && (q.callsign || '').toUpperCase() === ucCallsign && q.band === band && q.mode === mode) {
			return { is_dupe: true };
		}
	}

	return { is_dupe: false };
}

export default db;
