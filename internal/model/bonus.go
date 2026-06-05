package model

// BonusItem represents immutable metadata for a single ARRL Field Day bonus type.
type BonusItem struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	RuleRef      string `json:"rule_ref"`
	Points       int    `json:"points"`
	IsCounted    bool   `json:"is_counted"`
	MaxCount     int    `json:"max_count,omitempty"`
	DefaultCount int    `json:"default_count,omitempty"`
}

// BonusClaim represents per-bonus mutable claim state, persisted to the database.
type BonusClaim struct {
	BonusID   string `json:"bonus_id"`
	Claimed   bool   `json:"claimed"`
	Count     int    `json:"count"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

// DefaultBonuses is the hardcoded 2026 ARRL Field Day bonus list per section 7.3 of the rules.
// 18 bonus items total.
var DefaultBonuses = []BonusItem{
	{
		ID: "emergency_power", Name: "100% Emergency Power", RuleRef: "7.3.1",
		Points: 100, IsCounted: true, MaxCount: 20,
	},
	{
		ID: "media_publicity", Name: "Media Publicity", RuleRef: "7.3.2",
		Points: 100, IsCounted: false,
	},
	{
		ID: "public_location", Name: "Public Location", RuleRef: "7.3.3",
		Points: 100, IsCounted: false,
	},
	{
		ID: "public_info_table", Name: "Public Information Table", RuleRef: "7.3.4",
		Points: 100, IsCounted: false,
	},
	{
		ID: "message_to_sm", Name: "Message to Section Manager", RuleRef: "7.3.5",
		Points: 100, IsCounted: false,
	},
	{
		ID: "message_handling", Name: "Message Handling", RuleRef: "7.3.6",
		Points: 10, IsCounted: true, MaxCount: 10,
	},
	{
		ID: "satellite_qso", Name: "Satellite QSO", RuleRef: "7.3.7",
		Points: 100, IsCounted: false,
	},
	{
		ID: "alternate_power", Name: "Alternate Power", RuleRef: "7.3.8",
		Points: 100, IsCounted: false,
	},
	{
		ID: "w1aw_bulletin", Name: "W1AW Bulletin", RuleRef: "7.3.9",
		Points: 100, IsCounted: false,
	},
	{
		ID: "educational_activity", Name: "Educational Activity", RuleRef: "7.3.10",
		Points: 100, IsCounted: false,
	},
	{
		ID: "official_visit", Name: "Elected Official Visit", RuleRef: "7.3.11",
		Points: 100, IsCounted: false,
	},
	{
		ID: "agency_visit", Name: "Agency Representative Visit", RuleRef: "7.3.12",
		Points: 100, IsCounted: false,
	},
	{
		ID: "gota_bonus", Name: "GOTA Station Bonus", RuleRef: "7.3.13",
		Points: 5, IsCounted: true, MaxCount: 0, // unlimited + separate coach boolean
	},
	{
		ID: "web_submission", Name: "Web Submission", RuleRef: "7.3.14",
		Points: 50, IsCounted: false,
	},
	{
		ID: "youth_participation", Name: "Youth Participation", RuleRef: "7.3.15",
		Points: 20, IsCounted: true, MaxCount: 5,
	},
	{
		ID: "social_media", Name: "Social Media Promotion", RuleRef: "7.3.16",
		Points: 100, IsCounted: false,
	},
	{
		ID: "safety_officer", Name: "Safety Officer", RuleRef: "7.3.17",
		Points: 100, IsCounted: false,
	},
	{
		ID: "site_responsibilities", Name: "Site Responsibilities", RuleRef: "7.3.18",
		Points: 50, IsCounted: false,
	},
}

// ValidateBonusClaim validates a single bonus claim entry.
// Returns empty string on success, or an error message on failure.
// Validates that bonus_id exists in DefaultBonuses and that count
// is non-negative and within MaxCount for counted bonuses.
func ValidateBonusClaim(bonusID string, claimed bool, count int) string {
	// Look up the bonus item
	var item *BonusItem
	for i := range DefaultBonuses {
		if DefaultBonuses[i].ID == bonusID {
			item = &DefaultBonuses[i]
			break
		}
	}
	if item == nil {
		return "unknown bonus type: " + bonusID
	}

	if count < 0 {
		return "count must be non-negative for bonus: " + bonusID
	}

	if item.IsCounted && item.MaxCount > 0 && count > item.MaxCount {
		// MaxCount of 0 means unlimited (e.g., GOTA QSOs)
		// For non-zero MaxCount, enforce the limit
		// We only validate non-zero max here; 0 means unlimited.
		return "count exceeds maximum for bonus: " + bonusID
	}

	// Note: MaxCount=0 means unlimited (e.g., GOTA QSOs), no upper bound enforced.
	// The claim must still be >= 0 which is checked above.

	return ""
}

// CalculateBonusPoints computes the total bonus points from a map of claimed bonuses.
// Per ARRL rules section 7.3, bonus points are computed separately from raw QSO points
// and are added after the multiplier is applied.
//
// Special cases:
//   - emergency_power: claimed * count * 100
//   - message_handling: claimed * min(count, 10) * 10
//   - youth_participation: claimed * min(count, 5) * 20
//   - gota_bonus: claimed * count * 5 + (claimed && count >= 10 ? 100 : 0) coach bonus
//   - web_submission: claimed * 50
//   - safety_officer: claimed * 100
//   - site_responsibilities: claimed * 50
//   - all others: claimed * points
func CalculateBonusPoints(claims map[string]BonusClaim) int {
	total := 0

	for _, claim := range claims {
		if !claim.Claimed {
			continue
		}

		// Look up bonus item from defaults
		var item *BonusItem
		for i := range DefaultBonuses {
			if DefaultBonuses[i].ID == claim.BonusID {
				item = &DefaultBonuses[i]
				break
			}
		}
		if item == nil {
			continue // skip unknown bonus IDs
		}

		switch claim.BonusID {
		case "emergency_power":
			total += claim.Count * 100
		case "message_handling":
			capped := claim.Count
			if capped > 10 {
				capped = 10
			}
			total += capped * 10
		case "youth_participation":
			capped := claim.Count
			if capped > 5 {
				capped = 5
			}
			total += capped * 20
		case "gota_bonus":
			total += claim.Count * 5
			if claim.Count >= 10 {
				total += 100 // GOTA coach bonus
			}
		case "web_submission":
			total += 50
		case "safety_officer":
			total += 100
		case "site_responsibilities":
			total += 50
		default:
			total += item.Points
		}
	}

	return total
}
