package model

type QSO struct {
	ID            int64  `json:"id"`
	ClientID      string `json:"client_id,omitempty"`
	Timestamp     string `json:"timestamp"`
	Callsign      string `json:"callsign"`
	Band          string `json:"band"`
	Mode          string `json:"mode"`
	SentExchange  string `json:"sent_exchange"`
	RecvExchange  string `json:"recv_exchange"`
	Operator      string `json:"operator,omitempty"`
	IsDupe        bool   `json:"is_dupe"`
	Points        int    `json:"points"`
	CreatedAt     string `json:"created_at"`
}

type CreateQSOInput struct {
	ClientID     string `json:"client_id,omitempty"`
	Callsign     string `json:"callsign"`
	Band         string `json:"band"`
	Mode         string `json:"mode"`
	RecvExchange string `json:"recv_exchange"`
	SentExchange string `json:"sent_exchange"`
	Operator     string `json:"operator,omitempty"`
}

func ValidateRequired(input CreateQSOInput) string {
	if input.Callsign == "" {
		return "callsign is required"
	}
	if input.Band == "" {
		return "band is required"
	}
	if input.Mode == "" {
		return "mode is required"
	}
	if input.RecvExchange == "" {
		return "recv_exchange is required"
	}
	return ""
}
