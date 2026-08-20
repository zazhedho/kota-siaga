package apiindonesia

import (
	"context"
	"net/url"

	"kota-siaga/internal/dto"
	"kota-siaga/internal/mappers"
)

func (c *Client) ListWarnings(ctx context.Context, province string) ([]dto.Warning, error) {
	var envelope struct {
		Data []mappers.WarningSource `json:"data"`
	}
	if err := c.GetJSON(ctx, "api/v1/peringatan-dini", url.Values{"provinsi": {province}}, &envelope); err != nil {
		return nil, err
	}
	return mappers.MapWarnings(envelope.Data), nil
}
