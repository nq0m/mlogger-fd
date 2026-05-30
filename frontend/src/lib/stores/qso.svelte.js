export const qsos = $state([]);

export function addQso(qso) {
	qsos.unshift(qso);
}
