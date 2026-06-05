package model

import (
	"encoding/json"
	"testing"
)

func TestDefaultBonusesCount(t *testing.T) {
	if len(DefaultBonuses) != 18 {
		t.Errorf("DefaultBonuses has %d items, want 18", len(DefaultBonuses))
	}
}

func TestDefaultBonusesContainsEmergencyPower(t *testing.T) {
	found := false
	for _, b := range DefaultBonuses {
		if b.ID == "emergency_power" {
			found = true
			if b.Name != "100% Emergency Power" {
				t.Errorf("emergency_power Name = %q, want %q", b.Name, "100% Emergency Power")
			}
			if b.Points != 100 {
				t.Errorf("emergency_power Points = %d, want 100", b.Points)
			}
			if !b.IsCounted {
				t.Error("emergency_power IsCounted should be true")
			}
			if b.MaxCount != 20 {
				t.Errorf("emergency_power MaxCount = %d, want 20", b.MaxCount)
			}
		}
	}
	if !found {
		t.Error("emergency_power not found in DefaultBonuses")
	}
}

func TestDefaultBonusesContainsAllIDs(t *testing.T) {
	expected := []string{
		"emergency_power", "media_publicity", "public_location", "public_info_table",
		"message_to_sm", "message_handling", "satellite_qso", "alternate_power",
		"w1aw_bulletin", "educational_activity", "official_visit", "agency_visit",
		"gota_bonus", "web_submission", "youth_participation", "social_media",
		"safety_officer", "site_responsibilities",
	}

	bonusMap := make(map[string]bool)
	for _, b := range DefaultBonuses {
		bonusMap[b.ID] = true
	}

	for _, id := range expected {
		if !bonusMap[id] {
			t.Errorf("DefaultBonuses missing bonus_id %q", id)
		}
	}
}

func TestValidateBonusClaim_Valid(t *testing.T) {
	err := ValidateBonusClaim("emergency_power", true, 3)
	if err != "" {
		t.Errorf("ValidateBonusClaim(emergency_power, true, 3) = %q, want empty", err)
	}
}

func TestValidateBonusClaim_UnknownID(t *testing.T) {
	err := ValidateBonusClaim("invalid_bonus", true, 0)
	if err == "" {
		t.Error("ValidateBonusClaim with unknown ID should return error")
	}
}

func TestValidateBonusClaim_NegativeCount(t *testing.T) {
	err := ValidateBonusClaim("emergency_power", true, -1)
	if err == "" {
		t.Error("ValidateBonusClaim with negative count should return error")
	}
}

func TestValidateBonusClaim_ExceedsMaxCount(t *testing.T) {
	err := ValidateBonusClaim("emergency_power", true, 21)
	if err == "" {
		t.Error("ValidateBonusClaim with count exceeding MaxCount should return error")
	}
}

func TestValidateBonusClaim_BooleanBonusIgnoresCount(t *testing.T) {
	err := ValidateBonusClaim("media_publicity", true, 0)
	if err != "" {
		t.Errorf("ValidateBonusClaim(media_publicity, true, 0) = %q, want empty", err)
	}
}

func TestCalculateBonusPoints_Empty(t *testing.T) {
	claims := map[string]BonusClaim{}
	got := CalculateBonusPoints(claims)
	if got != 0 {
		t.Errorf("CalculateBonusPoints(empty) = %d, want 0", got)
	}
}

func TestCalculateBonusPoints_EmergencyPower(t *testing.T) {
	claims := map[string]BonusClaim{
		"emergency_power": {BonusID: "emergency_power", Claimed: true, Count: 3},
	}
	got := CalculateBonusPoints(claims)
	want := 300
	if got != want {
		t.Errorf("CalculateBonusPoints(emergency_power x3) = %d, want %d", got, want)
	}
}

func TestCalculateBonusPoints_MessageHandling(t *testing.T) {
	claims := map[string]BonusClaim{
		"message_handling": {BonusID: "message_handling", Claimed: true, Count: 5},
	}
	got := CalculateBonusPoints(claims)
	want := 50
	if got != want {
		t.Errorf("CalculateBonusPoints(message_handling x5) = %d, want %d", got, want)
	}
}

func TestCalculateBonusPoints_MessageHandlingCap(t *testing.T) {
	claims := map[string]BonusClaim{
		"message_handling": {BonusID: "message_handling", Claimed: true, Count: 15},
	}
	got := CalculateBonusPoints(claims)
	want := 100
	if got != want {
		t.Errorf("CalculateBonusPoints(message_handling x15 capped) = %d, want %d", got, want)
	}
}

func TestCalculateBonusPoints_YouthParticipation(t *testing.T) {
	claims := map[string]BonusClaim{
		"youth_participation": {BonusID: "youth_participation", Claimed: true, Count: 3},
	}
	got := CalculateBonusPoints(claims)
	want := 60
	if got != want {
		t.Errorf("CalculateBonusPoints(youth x3) = %d, want %d", got, want)
	}
}

func TestCalculateBonusPoints_YouthParticipationCap(t *testing.T) {
	claims := map[string]BonusClaim{
		"youth_participation": {BonusID: "youth_participation", Claimed: true, Count: 6},
	}
	got := CalculateBonusPoints(claims)
	want := 100
	if got != want {
		t.Errorf("CalculateBonusPoints(youth x6 capped) = %d, want %d", got, want)
	}
}

func TestCalculateBonusPoints_GOTABonus(t *testing.T) {
	claims := map[string]BonusClaim{
		"gota_bonus": {BonusID: "gota_bonus", Claimed: true, Count: 8},
	}
	got := CalculateBonusPoints(claims)
	want := 40
	if got != want {
		t.Errorf("CalculateBonusPoints(gota x8) = %d, want %d", got, want)
	}
}

func TestCalculateBonusPoints_GOTACoachBonus(t *testing.T) {
	claims := map[string]BonusClaim{
		"gota_bonus": {BonusID: "gota_bonus", Claimed: true, Count: 12},
	}
	got := CalculateBonusPoints(claims)
	want := 160
	if got != want {
		t.Errorf("CalculateBonusPoints(gota x12 + coach) = %d, want %d", got, want)
	}
}

func TestCalculateBonusPoints_WebSubmission(t *testing.T) {
	claims := map[string]BonusClaim{
		"web_submission": {BonusID: "web_submission", Claimed: true, Count: 0},
	}
	got := CalculateBonusPoints(claims)
	want := 50
	if got != want {
		t.Errorf("CalculateBonusPoints(web_submission) = %d, want %d", got, want)
	}
}

func TestCalculateBonusPoints_SafetyOfficer(t *testing.T) {
	claims := map[string]BonusClaim{
		"safety_officer": {BonusID: "safety_officer", Claimed: true, Count: 0},
	}
	got := CalculateBonusPoints(claims)
	want := 100
	if got != want {
		t.Errorf("CalculateBonusPoints(safety_officer) = %d, want %d", got, want)
	}
}

func TestCalculateBonusPoints_SiteResponsibilities(t *testing.T) {
	claims := map[string]BonusClaim{
		"site_responsibilities": {BonusID: "site_responsibilities", Claimed: true, Count: 0},
	}
	got := CalculateBonusPoints(claims)
	want := 50
	if got != want {
		t.Errorf("CalculateBonusPoints(site_responsibilities) = %d, want %d", got, want)
	}
}

func TestCalculateBonusPoints_BooleanBonus(t *testing.T) {
	claims := map[string]BonusClaim{
		"media_publicity": {BonusID: "media_publicity", Claimed: true, Count: 0},
	}
	got := CalculateBonusPoints(claims)
	want := 100
	if got != want {
		t.Errorf("CalculateBonusPoints(media_publicity) = %d, want %d", got, want)
	}
}

func TestCalculateBonusPoints_NotClaimed(t *testing.T) {
	claims := map[string]BonusClaim{
		"media_publicity": {BonusID: "media_publicity", Claimed: false, Count: 0},
	}
	got := CalculateBonusPoints(claims)
	want := 0
	if got != want {
		t.Errorf("CalculateBonusPoints(unclaimed) = %d, want %d", got, want)
	}
}

func TestCalculateBonusPoints_MultipleBonuses(t *testing.T) {
	claims := map[string]BonusClaim{
		"emergency_power":  {BonusID: "emergency_power", Claimed: true, Count: 3},
		"media_publicity":  {BonusID: "media_publicity", Claimed: true, Count: 0},
		"message_handling": {BonusID: "message_handling", Claimed: true, Count: 5},
		"web_submission":   {BonusID: "web_submission", Claimed: true, Count: 0},
		"gota_bonus":       {BonusID: "gota_bonus", Claimed: true, Count: 12},
	}
	got := CalculateBonusPoints(claims)
	want := 660
	if got != want {
		t.Errorf("CalculateBonusPoints(multiple) = %d, want %d", got, want)
	}
}

func TestBonusItemJSONTags(t *testing.T) {
	b := BonusItem{
		ID:           "test",
		Name:         "Test Bonus",
		RuleRef:      "1.0",
		Points:       100,
		IsCounted:    true,
		MaxCount:     5,
		DefaultCount: 1,
	}

	data, err := json.Marshal(b)
	if err != nil {
		t.Fatalf("BonusItem JSON marshal failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("BonusItem JSON unmarshal failed: %v", err)
	}

	if result["id"] != "test" {
		t.Errorf("json tag 'id' = %v, want 'test'", result["id"])
	}
	if result["name"] != "Test Bonus" {
		t.Errorf("json tag 'name' = %v, want 'Test Bonus'", result["name"])
	}
	if result["rule_ref"] != "1.0" {
		t.Errorf("json tag 'rule_ref' = %v, want '1.0'", result["rule_ref"])
	}
	if result["points"].(float64) != 100 {
		t.Errorf("json tag 'points' = %v, want 100", result["points"])
	}
	if result["is_counted"] != true {
		t.Errorf("json tag 'is_counted' = %v, want true", result["is_counted"])
	}
	if result["max_count"].(float64) != 5 {
		t.Errorf("json tag 'max_count' = %v, want 5", result["max_count"])
	}
	if result["default_count"].(float64) != 1 {
		t.Errorf("json tag 'default_count' = %v, want 1", result["default_count"])
	}
}

func TestBonusClaimJSONTags(t *testing.T) {
	b := BonusClaim{
		BonusID:   "test_bonus",
		Claimed:   true,
		Count:     3,
		UpdatedAt: "2026-06-04T12:00:00Z",
	}

	data, err := json.Marshal(b)
	if err != nil {
		t.Fatalf("BonusClaim JSON marshal failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("BonusClaim JSON unmarshal failed: %v", err)
	}

	if result["bonus_id"] != "test_bonus" {
		t.Errorf("json tag 'bonus_id' = %v, want 'test_bonus'", result["bonus_id"])
	}
	if result["claimed"] != true {
		t.Errorf("json tag 'claimed' = %v, want true", result["claimed"])
	}
	if result["count"].(float64) != 3 {
		t.Errorf("json tag 'count' = %v, want 3", result["count"])
	}
	if result["updated_at"] != "2026-06-04T12:00:00Z" {
		t.Errorf("json tag 'updated_at' = %v, want '2026-06-04T12:00:00Z'", result["updated_at"])
	}
}
