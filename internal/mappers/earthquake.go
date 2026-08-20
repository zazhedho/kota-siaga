package mappers

import "kota-siaga/internal/dto"

type EarthquakeSource struct {
	ID        string  `json:"id"`
	DateTime  string  `json:"datetime"`
	Magnitude float64 `json:"magnitude"`
	DepthKM   float64 `json:"depth_km"`
	Lat       float64 `json:"lat"`
	Lng       float64 `json:"lng"`
	Region    string  `json:"region"`
	Potential string  `json:"potential"`
	IsFelt    int     `json:"is_felt"`
	FeltAreas *string `json:"felt_areas"`
	Source    string  `json:"source"`
}

func MapEarthquake(source EarthquakeSource) dto.Earthquake {
	return dto.Earthquake{
		ID:        source.ID,
		DateTime:  source.DateTime,
		Magnitude: source.Magnitude,
		DepthKM:   source.DepthKM,
		Latitude:  source.Lat,
		Longitude: source.Lng,
		Region:    source.Region,
		Potential: source.Potential,
		IsFelt:    source.IsFelt != 0,
		FeltAreas: source.FeltAreas,
		Source:    source.Source,
	}
}

func MapEarthquakes(sources []EarthquakeSource) []dto.Earthquake {
	mapped := make([]dto.Earthquake, len(sources))
	for i, source := range sources {
		mapped[i] = MapEarthquake(source)
	}
	return mapped
}
