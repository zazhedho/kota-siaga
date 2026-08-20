package apiindonesia

import (
	"context"
	"net/url"
	"strconv"

	"kota-siaga/internal/dto"
	"kota-siaga/internal/mappers"
)

type Hospital = dto.Hospital

func (c *Client) ListHospitals(ctx context.Context, kabupatenID string, page, perPage int) (dto.Page[dto.Hospital], error) {
	var source dto.Page[mappers.HospitalSource]
	query := url.Values{
		"kabupaten_id": {kabupatenID},
		"page":         {strconv.Itoa(page)},
		"per_page":     {strconv.Itoa(perPage)},
	}
	if err := c.GetJSON(ctx, "api/v1/rumah-sakit", query, &source); err != nil {
		return dto.Page[dto.Hospital]{}, err
	}
	return mappers.MapHospitalPage(source), nil
}
