# Field Test Checklist

Outdoor verification for Field Day Logger before ARRL Field Day — set up in a park or open area with at least 2 operators using separate devices.

## Setup

1. **Set up server** — RPi or laptop, running `fdlogger` binary. Connect to WiFi AP or LAN switch.
2. **Connect 2+ devices** — phones, tablets, or laptops on the same LAN. Open browser to server IP:8080.

## Core Logging

3. **Log QSO from each device** — enter callsign, band, mode, exchange. Verify QSO appears in log table.
4. **Verify real-time display** — log a QSO on one device; confirm it appears on the other within 2 seconds.
5. **Test dupe detection** — enter the same callsign+band+mode twice. Confirm dupe warning and 0 points.
6. **Check stats accuracy** — verify StatsBar shows correct QSO count, rate, and score after logging several QSOs.

## Offline Resilience

7. **Test offline logging** — disable WiFi on one device. Log 3-5 QSOs. Reconnect WiFi. Confirm QSOs sync to server and appear on other device.

## Backup & Export

8. **Download backup** — click ↓ Backup button in header. Confirm browser downloads `fdlogger_backup_*.db` with "Backup downloaded" toast.
9. **Export Cabrillo** — click Export Cabrillo. Confirm browser downloads `.cbr` file. Verify file contains correct QSOs and headers.

## Bonus Tracker

10. **Toggle bonuses** — open BonusTracker. Claim a few bonuses (e.g., Emergency Power, Web Submission). Verify score updates in StatsBar.

## Audio

11. **Test audio alerts** — log a QSO. Confirm audio plays on QSO confirmation. Log a duplicate QSO. Confirm dupe alert plays.
12. **Test mute toggle** — click mute button (🔇). Log another QSO. Confirm no audio plays.

## Visual

13. **Test dark mode** — toggle dark mode (☾). Verify UI is readable in direct sunlight. Log a QSO in dark mode and confirm form is usable.
14. **Test on mobile** — use a phone. Verify touch targets are large enough for outdoor use. Verify no horizontal scrolling.

## Multi-Operator

15. **Simultaneous logging** — both operators log QSOs at the same time for 2-3 minutes. Confirm no conflicts, no data loss, all QSOs visible on both devices.
