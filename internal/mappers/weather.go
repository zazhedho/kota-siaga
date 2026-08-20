package mappers

import "kota-siaga/internal/dto"

type WeatherSource struct {
	ID                   string  `json:"id"`
	Adm4                 string  `json:"adm4"`
	Provinsi             string  `json:"provinsi"`
	Kotkab               string  `json:"kotkab"`
	Kecamatan            string  `json:"kecamatan"`
	Desa                 string  `json:"desa"`
	Datetime             string  `json:"datetime"`
	LocalDatetime        string  `json:"local_datetime"`
	Weather              string  `json:"weather"`
	WeatherCode          string  `json:"weather_code"`
	WeatherDesc          string  `json:"weather_desc"`
	WeatherDescEN        string  `json:"weather_desc_en"`
	TemperatureC         float64 `json:"temperature_c"`
	HumidityPercent      float64 `json:"humidity_percent"`
	CloudCoverPercent    float64 `json:"cloud_cover_percent"`
	PrecipitationMM      float64 `json:"precipitation_mm"`
	WindDirection        string  `json:"wind_direction"`
	WindDirectionTo      string  `json:"wind_direction_to"`
	WindDirectionDegrees float64 `json:"wind_direction_degrees"`
	WindSpeed            float64 `json:"wind_speed"`
	VisibilityM          float64 `json:"visibility_m"`
	VisibilityText       string  `json:"visibility_text"`
	AnalysisDate         string  `json:"analysis_date"`
	Source               string  `json:"source"`
}

func MapWeather(source WeatherSource) dto.WeatherForecast {
	return dto.WeatherForecast{
		ID:                   source.ID,
		Adm4:                 source.Adm4,
		Province:             source.Provinsi,
		Regency:              source.Kotkab,
		District:             source.Kecamatan,
		Village:              source.Desa,
		Datetime:             source.Datetime,
		LocalDatetime:        source.LocalDatetime,
		Weather:              source.Weather,
		WeatherCode:          source.WeatherCode,
		WeatherDescription:   source.WeatherDesc,
		WeatherDescriptionEN: source.WeatherDescEN,
		TemperatureC:         source.TemperatureC,
		HumidityPercent:      source.HumidityPercent,
		CloudCoverPercent:    source.CloudCoverPercent,
		PrecipitationMM:      source.PrecipitationMM,
		WindDirection:        source.WindDirection,
		WindDirectionTo:      source.WindDirectionTo,
		WindDirectionDegrees: source.WindDirectionDegrees,
		WindSpeed:            source.WindSpeed,
		VisibilityM:          source.VisibilityM,
		VisibilityText:       source.VisibilityText,
		AnalysisDate:         source.AnalysisDate,
		Source:               source.Source,
	}
}

func MapWeatherForecasts(sources []WeatherSource) []dto.WeatherForecast {
	mapped := make([]dto.WeatherForecast, len(sources))
	for i, source := range sources {
		mapped[i] = MapWeather(source)
	}
	return mapped
}
