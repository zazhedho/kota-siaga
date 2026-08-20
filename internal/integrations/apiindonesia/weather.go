package apiindonesia

import (
	"context"
	"net/url"

	"kota-siaga/internal/dto"
	"kota-siaga/internal/mappers"
)

func (c *Client) GetWeather(ctx context.Context, adm4 string) ([]dto.WeatherForecast, error) {
	var envelope struct {
		Data []mappers.WeatherSource `json:"data"`
	}
	if err := c.GetJSON(ctx, "api/v1/cuaca", url.Values{"adm4": {adm4}}, &envelope); err != nil {
		return nil, err
	}
	return mappers.MapWeatherForecasts(envelope.Data), nil
}
