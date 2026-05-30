package model

type StationConfig struct {
	Callsign         string `json:"callsign"`
	Class            string `json:"class"`
	ARRLSection      string `json:"arrl_section"`
	TransmitterCount int    `json:"transmitter_count"`
	PowerLevel       string `json:"power_level"`
	UpdatedAt        string `json:"updated_at,omitempty"`
}

func DefaultStationConfig() StationConfig {
	return StationConfig{
		Callsign:         "N0CALL",
		Class:            "1D",
		ARRLSection:      "EMA",
		TransmitterCount: 1,
		PowerLevel:       "LOW",
	}
}

func ValidateStationConfig(cfg StationConfig) string {
	if cfg.Callsign == "" {
		return "callsign is required"
	}
	if cfg.Class == "" {
		return "class is required"
	}
	if cfg.ARRLSection == "" {
		return "arrl_section is required"
	}
	if cfg.TransmitterCount < 1 || cfg.TransmitterCount > 20 {
		return "transmitter_count must be between 1 and 20"
	}
	validPower := map[string]bool{"LOW": true, "HIGH": true, "QRP": true}
	if !validPower[cfg.PowerLevel] {
		return "power_level must be LOW, HIGH, or QRP"
	}
	return ""
}
