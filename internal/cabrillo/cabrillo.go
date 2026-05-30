package cabrillo

import (
	"bytes"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

var bandToFreq = map[string]string{
	"160M": "  1800",
	"80M":  "  3500",
	"40M":  "  7000",
	"20M":  " 14000",
	"15M":  " 21000",
	"10M":  " 28000",
	"6M":   " 50000",
	"2M":   "144000",
	"70CM": "432000",
}

var modeToCabrillo = map[string]string{
	"CW":    "CW  ",
	"SSB":   "PH  ",
	"FM":    "FM  ",
	"RTTY":  "RY  ",
	"FT8":   "DG  ",
	"FT4":   "DG  ",
	"PSK31": "DG  ",
}

func Generate(db *sql.DB) (string, error) {
	var buf bytes.Buffer

	buf.WriteString("START-OF-LOG: 3.0\n")
	buf.WriteString("CREATED-BY: FDLogger v1.0\n")
	buf.WriteString("CONTEST: ARRL-FIELD-DAY\n")
	buf.WriteString("CALLSIGN: N0CALL\n")
	buf.WriteString("ARRL-SECTION: NH\n")
	buf.WriteString("CATEGORY-OPERATOR: SINGLE-OP\n")
	buf.WriteString("CATEGORY-POWER: LOW\n")
	buf.WriteString("CATEGORY-STATION: PORTABLE\n")

	var rawPoints int
	var multiplier int
	if err := db.QueryRow("SELECT COALESCE(SUM(points), 0) FROM qsos WHERE is_dupe = 0").Scan(&rawPoints); err != nil {
		rawPoints = 0
	}
	if err := db.QueryRow("SELECT COUNT(DISTINCT band || '_' || mode) FROM qsos WHERE is_dupe = 0").Scan(&multiplier); err != nil {
		multiplier = 1
	}
	if multiplier < 1 {
		multiplier = 1
	}
	score := rawPoints * multiplier
	buf.WriteString(fmt.Sprintf("CLAIMED-SCORE: %d\n", score))

	rows, err := db.Query(`SELECT timestamp, callsign, band, mode, sent_exchange, recv_exchange, is_dupe
		FROM qsos ORDER BY timestamp ASC`)
	if err != nil {
		rows = nil
	}

	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var ts, callsign, band, mode, sentEx, recvEx string
			var isDupe int
			if err := rows.Scan(&ts, &callsign, &band, &mode, &sentEx, &recvEx, &isDupe); err != nil {
				continue
			}

			t, err := time.Parse(time.RFC3339, ts)
			if err != nil {
				continue
			}
			date := t.Format("2006-01-02")
			timeStr := t.Format("1504")

			freq := bandToFreq[band]
			if freq == "" {
				freq = "     0"
			}

			modeCode := modeToCabrillo[strings.ToUpper(mode)]
			if modeCode == "" {
				modeCode = "PH  "
			}

			line := fmt.Sprintf("QSO: %6s %-4s %s %s %-10s %-10s %-10s %-10s\n",
				freq, modeCode, date, timeStr,
				callsign, padRight(sentEx, 10), callsign, padRight(recvEx, 10))

			if isDupe == 1 {
				line = strings.Replace(line, "QSO:", "QSO: ---dupe---", 1)
			}

			buf.WriteString(line)
		}
	}

	buf.WriteString("END-OF-LOG:\n")

	return buf.String(), nil
}

func padRight(s string, length int) string {
	if len(s) >= length {
		return s[:length]
	}
	return s + strings.Repeat(" ", length-len(s))
}
