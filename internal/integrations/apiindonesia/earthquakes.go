package apiindonesia

import (
	"context"

	"kota-siaga/internal/dto"
	"kota-siaga/internal/mappers"
)

func (c *Client) ListLatest(ctx context.Context) ([]dto.Earthquake, error) {
	var envelope struct {
		Data []mappers.EarthquakeSource `json:"data"`
	}
	if err := c.GetJSON(ctx, "api/v1/gempa/terkini", nil, &envelope); err != nil {
		return nil, err
	}
	return mappers.MapEarthquakes(envelope.Data), nil
}
