const BASE_URL = '';

export async function createQSO(data) {
	const res = await fetch(`${BASE_URL}/api/qso`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(data)
	});
	const json = await res.json();
	if (!res.ok) {
		throw new Error(json.error || 'Failed to create QSO');
	}
	return json;
}

export async function fetchQsos(limit = 50, offset = 0) {
	const res = await fetch(`${BASE_URL}/api/qso?limit=${limit}&offset=${offset}`);
	if (!res.ok) {
		throw new Error('Failed to fetch QSOs');
	}
	return res.json();
}

export async function checkDupe(callsign, band, mode) {
	const res = await fetch(`${BASE_URL}/api/check-dupe?callsign=${encodeURIComponent(callsign)}&band=${encodeURIComponent(band)}&mode=${encodeURIComponent(mode)}`);
	if (!res.ok) {
		throw new Error('Failed to check dupe');
	}
	return res.json();
}

export async function updateQso(id, data) {
	const res = await fetch(`${BASE_URL}/api/qso/${id}`, {
		method: 'PUT',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(data)
	});
	const json = await res.json();
	if (!res.ok) {
		throw new Error(json.error || 'Failed to update QSO');
	}
	return json;
}

export async function searchQsos(query, limit = 50, offset = 0) {
	const res = await fetch(`${BASE_URL}/api/qso?search=${encodeURIComponent(query)}&limit=${limit}&offset=${offset}`);
	if (!res.ok) {
		throw new Error('Failed to search QSOs');
	}
	return res.json();
}

export async function syncBatch(qsos) {
	const res = await fetch(`${BASE_URL}/api/sync`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ qsos })
	});
	const json = await res.json();
	if (!res.ok) {
		throw new Error(json.error || 'Failed to sync QSOs');
	}
	return json;
}

export async function getStationConfig() {
	const res = await fetch(`${BASE_URL}/api/station-config`);
	if (!res.ok) {
		throw new Error('Failed to fetch station config');
	}
	return res.json();
}

export async function putStationConfig(data) {
	const res = await fetch(`${BASE_URL}/api/station-config`, {
		method: 'PUT',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(data)
	});
	const json = await res.json();
	if (!res.ok) {
		throw new Error(json.error || 'Failed to save station config');
	}
	return json;
}

export async function getBonuses() {
	const res = await fetch(`${BASE_URL}/api/bonuses`);
	if (!res.ok) {
		throw new Error('Failed to fetch bonuses');
	}
	return res.json();
}

export async function putBonuses(data) {
	const res = await fetch(`${BASE_URL}/api/bonuses`, {
		method: 'PUT',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(data)
	});
	const json = await res.json();
	if (!res.ok) {
		throw new Error(json.error || 'Failed to save bonuses');
	}
	return json;
}

export function downloadBackup() {
	window.location.href = `${BASE_URL}/api/backup/db`;
}
