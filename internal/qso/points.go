package qso

import "strings"

var twoPointModes = map[string]bool{
	"CW":    true,
	"RTTY":  true,
	"FT8":   true,
	"FT4":   true,
	"PSK31": true,
	"MFSK":  true,
	"JT65":  true,
	"JT9":   true,
	"OLIVIA": true,
	"DOMINO": true,
}

var onePointModes = map[string]bool{
	"SSB": true,
	"FM":  true,
	"AM":  true,
}

func CalculatePoints(mode string, isDupe bool) int {
	if isDupe {
		return 0
	}
	mode = strings.ToUpper(mode)
	if twoPointModes[mode] {
		return 2
	}
	return 1
}
