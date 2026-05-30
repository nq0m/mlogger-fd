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
