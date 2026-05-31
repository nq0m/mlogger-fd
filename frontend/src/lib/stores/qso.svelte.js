export const qsos = $state([]);

export const stats = $state({
	total: 0,
	raw_points: 0,
	multiplier: 1,
	score: 0,
	rate_10min: 0,
	rate_1hr: 0,
	breakdown: {}
});

export const stationConfig = $state({
	callsign: 'N0CALL',
	class: '1D',
	arrl_section: 'EMA',
	transmitter_count: 1,
	power_level: 'LOW'
});

export function addQso(qso) {
	qsos.unshift(qso);
}

export async function addQsoOffline(qsoData) {
	const client_id = crypto.randomUUID();
	const { enqueueQso } = await import('$lib/db.js');

	qsoData.client_id = client_id;
	await enqueueQso(qsoData);

	const qso = {
		id: client_id,
		client_id,
		timestamp: new Date().toISOString(),
		callsign: qsoData.callsign,
		band: qsoData.band,
		mode: qsoData.mode,
		recv_exchange: qsoData.recv_exchange,
		operator: qsoData.operator || '',
		is_dupe: false,
		points: 0,
		_offline: true
	};

	addQso(qso);
	return qso;
}

export async function loadCache() {
	const { populateCache } = await import('$lib/db.js');
	try {
		const res = await fetch('/api/qso?limit=9999');
		if (!res.ok) return;
		const data = await res.json();
		await populateCache(data);
	} catch {
		// server unreachable at load time — cache will populate on next load or via WS events
	}
}

loadCache();

export async function fetchStats() {
	try {
		const res = await fetch('/api/stats');
		if (!res.ok) return;
		const data = await res.json();
		Object.assign(stats, data);
	} catch {
		// silently ignore fetch errors
	}
}
