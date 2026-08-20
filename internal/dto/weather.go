package dto

type WeatherForecast struct {
	ID                   string  `json:"id"`
	Adm4                 string  `json:"adm4"`
	Province             string  `json:"province"`
	Regency              string  `json:"regency"`
	District             string  `json:"district"`
	Village              string  `json:"village"`
	Datetime             string  `json:"datetime"`
	LocalDatetime        string  `json:"local_datetime"`
	Weather              string  `json:"weather"`
	WeatherCode          string  `json:"weather_code"`
	WeatherDescription   string  `json:"weather_description"`
	WeatherDescriptionEN string  `json:"weather_description_en"`
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
