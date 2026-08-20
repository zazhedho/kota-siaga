package dto

type Warning struct {
	ID          string `json:"id"`
	AlertID     string `json:"alert_id"`
	Event       string `json:"event"`
	Urgency     string `json:"urgency"`
	Severity    string `json:"severity"`
	Certainty   string `json:"certainty"`
	Area        string `json:"area"`
	Province    string `json:"province"`
	Effective   string `json:"effective"`
	Expires     string `json:"expires"`
	Headline    string `json:"headline"`
	Description string `json:"description"`
	Instruction string `json:"instruction"`
	Source      string `json:"source"`
	IsActive    bool   `json:"is_active"`
}
