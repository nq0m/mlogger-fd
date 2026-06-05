package cabrillo

import (
	"bytes"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jeremy/mlogger-fd/internal/model"
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

	// Read station config, falling back to defaults if not available
	callsign := "N0CALL"
	class := "1D"
	section := "NH"
	power := "LOW"

	err := db.QueryRow("SELECT callsign, class, arrl_section, power_level FROM station_config WHERE id = 1").Scan(&callsign, &class, &section, &power)
	if err != nil && err != sql.ErrNoRows {
		// Config table query error — use defaults silently
		// (table may not exist, permission error, etc.)
		callsign = "N0CALL"
		class = "1D"
		section = "NH"
		power = "LOW"
	}
	// If err == sql.ErrNoRows: defaults already set above
	// If config returned empty class: fall back to default
	if class == "" {
		class = "1D"
	}

	buf.WriteString("START-OF-LOG: 3.0\n")
	buf.WriteString("CREATED-BY: FDLogger v1.0\n")
	buf.WriteString("CONTEST: ARRL-FIELD-DAY\n")
	buf.WriteString(fmt.Sprintf("CALLSIGN: %s\n", callsign))
	buf.WriteString(fmt.Sprintf("ARRL-SECTION: %s\n", section))
	buf.WriteString("CATEGORY-OPERATOR: SINGLE-OP\n")
	buf.WriteString(fmt.Sprintf("CATEGORY-POWER: %s\n", power))
	buf.WriteString("CATEGORY-STATION: PORTABLE\n")
	buf.WriteString(fmt.Sprintf("CATEGORY-CLASS: %s\n", class))

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

	var bonusPoints int
	if err := db.QueryRow(`SELECT COALESCE(SUM(CASE
		WHEN bonus_id = 'emergency_power' THEN claimed * count * 100
		WHEN bonus_id = 'message_handling' THEN claimed * MIN(count, 10) * 10
		WHEN bonus_id = 'youth_participation' THEN claimed * MIN(count, 5) * 20
		WHEN bonus_id = 'gota_bonus' THEN claimed * (count * 5 + CASE WHEN count >= 10 THEN 100 ELSE 0 END)
		WHEN bonus_id = 'web_submission' THEN claimed * 50
		WHEN bonus_id = 'safety_officer' THEN claimed * 100
		WHEN bonus_id = 'site_responsibilities' THEN claimed * 50
		ELSE claimed * 100
	END), 0) FROM bonus_claims`).Scan(&bonusPoints); err != nil {
		bonusPoints = 0
	}

	score := (rawPoints * multiplier) + bonusPoints
	buf.WriteString(fmt.Sprintf("CLAIMED-SCORE: %d\n", score))

	// Add SOAPBOX lines for claimed bonuses
	bonusRows, err := db.Query("SELECT bonus_id, claimed, count FROM bonus_claims WHERE claimed = 1 ORDER BY bonus_id")
	if err == nil {
		defer bonusRows.Close()
		for bonusRows.Next() {
			var bid string
			var claimed, count int
			if err := bonusRows.Scan(&bid, &claimed, &count); err != nil {
				continue
			}
			// Look up bonus name and compute points
			var name string
			var pts int
			switch bid {
			case "emergency_power":
				name = "100% Emergency Power"
				pts = count * 100
			case "media_publicity":
				name = "Media Publicity"
				pts = 100
			case "public_location":
				name = "Public Location"
				pts = 100
			case "public_info_table":
				name = "Public Information Table"
				pts = 100
			case "message_to_sm":
				name = "Message to Section Manager"
				pts = 100
			case "message_handling":
				name = "Message Handling"
				capped := count
				if capped > 10 {
					capped = 10
				}
				pts = capped * 10
			case "satellite_qso":
				name = "Satellite QSO"
				pts = 100
			case "alternate_power":
				name = "Alternate Power"
				pts = 100
			case "w1aw_bulletin":
				name = "W1AW Bulletin"
				pts = 100
			case "educational_activity":
				name = "Educational Activity"
				pts = 100
			case "official_visit":
				name = "Elected Official Visit"
				pts = 100
			case "agency_visit":
				name = "Agency Representative Visit"
				pts = 100
			case "gota_bonus":
				name = "GOTA Station Bonus"
				pts = count*5
				if count >= 10 {
					pts += 100
				}
			case "web_submission":
				name = "Web Submission"
				pts = 50
			case "youth_participation":
				name = "Youth Participation"
				capped := count
				if capped > 5 {
					capped = 5
				}
				pts = capped * 20
			case "social_media":
				name = "Social Media Promotion"
				pts = 100
			case "safety_officer":
				name = "Safety Officer"
				pts = 100
			case "site_responsibilities":
				name = "Site Responsibilities"
				pts = 50
			default:
				// Look up from model.DefaultBonuses
				for i := range model.DefaultBonuses {
					if model.DefaultBonuses[i].ID == bid {
						name = model.DefaultBonuses[i].Name
						pts = model.DefaultBonuses[i].Points
						break
					}
				}
			}
			buf.WriteString(fmt.Sprintf("SOAPBOX: Bonus: %s = %d pts\n", name, pts))
		}
		buf.WriteString(fmt.Sprintf("SOAPBOX: Total Bonus Points = %d\n", bonusPoints))
	}

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
