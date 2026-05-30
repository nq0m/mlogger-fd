package model

import "testing"

func TestDefaultStationConfig(t *testing.T) {
	cfg := DefaultStationConfig()

	if cfg.Callsign != "N0CALL" {
		t.Errorf("expected Callsign 'N0CALL', got '%s'", cfg.Callsign)
	}
	if cfg.Class != "1D" {
		t.Errorf("expected Class '1D', got '%s'", cfg.Class)
	}
	if cfg.ARRLSection != "EMA" {
		t.Errorf("expected ARRLSection 'EMA', got '%s'", cfg.ARRLSection)
	}
	if cfg.TransmitterCount != 1 {
		t.Errorf("expected TransmitterCount 1, got %d", cfg.TransmitterCount)
	}
	if cfg.PowerLevel != "LOW" {
		t.Errorf("expected PowerLevel 'LOW', got '%s'", cfg.PowerLevel)
	}
}

func TestValidateStationConfig(t *testing.T) {
	tests := []struct {
		name    string
		cfg     StationConfig
		wantMsg string
	}{
		{
			name: "valid config returns empty",
			cfg: StationConfig{
				Callsign:         "K1ABC",
				Class:            "1D",
				ARRLSection:      "EMA",
				TransmitterCount: 5,
				PowerLevel:       "LOW",
			},
			wantMsg: "",
		},
		{
			name: "empty callsign",
			cfg: StationConfig{
				Callsign: "",
				Class:    "1D", ARRLSection: "EMA",
				TransmitterCount: 1, PowerLevel: "LOW",
			},
			wantMsg: "callsign is required",
		},
		{
			name: "empty class",
			cfg: StationConfig{
				Callsign: "K1ABC",
				Class:    "", ARRLSection: "EMA",
				TransmitterCount: 1, PowerLevel: "LOW",
			},
			wantMsg: "class is required",
		},
		{
			name: "empty arrl_section",
			cfg: StationConfig{
				Callsign: "K1ABC",
				Class:    "1D", ARRLSection: "",
				TransmitterCount: 1, PowerLevel: "LOW",
			},
			wantMsg: "arrl_section is required",
		},
		{
			name: "transmitter_count zero",
			cfg: StationConfig{
				Callsign: "K1ABC",
				Class:    "1D", ARRLSection: "EMA",
				TransmitterCount: 0, PowerLevel: "LOW",
			},
			wantMsg: "transmitter_count must be between 1 and 20",
		},
		{
			name: "transmitter_count 21",
			cfg: StationConfig{
				Callsign: "K1ABC",
				Class:    "1D", ARRLSection: "EMA",
				TransmitterCount: 21, PowerLevel: "LOW",
			},
			wantMsg: "transmitter_count must be between 1 and 20",
		},
		{
			name: "invalid power_level MEDIUM",
			cfg: StationConfig{
				Callsign: "K1ABC",
				Class:    "1D", ARRLSection: "EMA",
				TransmitterCount: 1, PowerLevel: "MEDIUM",
			},
			wantMsg: "power_level must be LOW, HIGH, or QRP",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateStationConfig(tt.cfg)
			if got != tt.wantMsg {
				t.Errorf("ValidateStationConfig() = %q, want %q", got, tt.wantMsg)
			}
		})
	}
}
