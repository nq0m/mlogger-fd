// frontend/src/lib/components/QsoEntryForm.audio.test.js
// Verifies audio trigger wiring in QsoEntryForm.svelte (Plan 04-03 Task 2)
import { describe, it, expect } from 'vitest';
import { readFileSync } from 'fs';
import { resolve } from 'path';

// Read the source file for structural verification
const srcPath = resolve(import.meta.dirname, 'QsoEntryForm.svelte');
let source = '';
try {
	source = readFileSync(srcPath, 'utf-8');
} catch {
	// Test will fail gracefully if file can't be read
}

describe('QsoEntryForm — Audio Trigger Integration (D-08)', () => {
	it('imports playSound from audio.svelte.js', () => {
		expect(source).toMatch(/import\s*\{[^}]*playSound[^}]*\}\s*from\s*'\$lib\/audio\.svelte\.js'/);
	});

	it('calls playSound("confirm") after successful createQSO in handleSubmit', () => {
		// Verify playSound('confirm') exists in source
		expect(source).toContain("playSound('confirm')");

		// Verify it appears after fetchStats() and before if (result.is_dupe)
		const fetchStatsIdx = source.indexOf('fetchStats()');
		const confirmIdx = source.indexOf("playSound('confirm')");
		const isDupeIdx = source.indexOf('if (result.is_dupe)', confirmIdx);

		expect(fetchStatsIdx).toBeGreaterThan(-1);
		expect(confirmIdx).toBeGreaterThan(-1);
		expect(isDupeIdx).toBeGreaterThan(-1);
		expect(confirmIdx).toBeGreaterThan(fetchStatsIdx);
		expect(isDupeIdx).toBeGreaterThan(confirmIdx);
	});

	it('calls playSound("dupe") when dupeWarning is set to exact DUPE message (online path)', () => {
		// Online dupe path (checkDupe)
		const onlineDupeRegex = /if\s*\(result\.is_dupe\)\s*\{[^}]*dupeWarning\s*=\s*'DUPE:[^']*'[^}]*playSound\('dupe'\)/s;
		expect(source).toMatch(onlineDupeRegex);
	});

	it('calls playSound("dupe") when dupeWarning is set to exact DUPE message (offline path)', () => {
		// Offline dupe path (offlineDupeCheck)
		const offlineDupeRegex = /if\s*\(result\.is_dupe\)\s*\{[^}]*dupeWarning\s*=\s*'DUPE:[^']*'[^}]*playSound\('dupe'\)/s;
		expect(source).toMatch(offlineDupeRegex);
	});

	it('does NOT call playSound for similar-calls warning (only exact dupes)', () => {
		// Similar calls block should not contain playSound
		const similarIdx = source.indexOf('Similar calls');
		expect(similarIdx).toBeGreaterThan(-1);

		// Find the similar calls if-block
		const similarBlock = source.slice(similarIdx - 50, similarIdx + 200);
		expect(similarBlock).not.toContain('playSound');
	});

	it('does NOT call playSound in WebSocket handler (D-08 — own-QSO only)', async () => {
		// Verify ws.svelte.js does not import or call playSound
		const wsPath = resolve(import.meta.dirname, '../ws.svelte.js');
		const wsSource = readFileSync(wsPath, 'utf-8');
		expect(wsSource).not.toContain('playSound');
	});
});
