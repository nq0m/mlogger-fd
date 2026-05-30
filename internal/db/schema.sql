CREATE TABLE IF NOT EXISTS qsos (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp TEXT NOT NULL,
    callsign TEXT NOT NULL,
    band TEXT NOT NULL,
    mode TEXT NOT NULL,
    sent_exchange TEXT NOT NULL,
    recv_exchange TEXT NOT NULL,
    operator TEXT,
    is_dupe INTEGER NOT NULL DEFAULT 0,
    points INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_qsos_callsign ON qsos(callsign);
CREATE INDEX IF NOT EXISTS idx_qsos_timestamp ON qsos(timestamp);
CREATE INDEX IF NOT EXISTS idx_qsos_band_mode ON qsos(band, mode);
