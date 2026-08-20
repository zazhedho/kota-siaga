package mappers

import "kota-siaga/internal/dto"

type WarningSource struct {
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
	IsActive    int    `json:"is_active"`
}

func MapWarning(source WarningSource) dto.Warning {
	return dto.Warning{
		ID:          source.ID,
		AlertID:     source.AlertID,
		Event:       source.Event,
		Urgency:     source.Urgency,
		Severity:    source.Severity,
		Certainty:   source.Certainty,
		Area:        source.Area,
		Province:    source.Province,
		Effective:   source.Effective,
		Expires:     source.Expires,
		Headline:    source.Headline,
		Description: source.Description,
		Instruction: source.Instruction,
		Source:      source.Source,
		IsActive:    source.IsActive != 0,
	}
}

func MapWarnings(sources []WarningSource) []dto.Warning {
	mapped := make([]dto.Warning, len(sources))
	for i, source := range sources {
		mapped[i] = MapWarning(source)
	}
	return mapped
}
