package apiindonesia

import (
	"context"
	"net/url"
	"strconv"

	"kota-siaga/internal/dto"
	"kota-siaga/internal/mappers"
)

type Province = dto.Province
type City = dto.City
type District = dto.District
type Village = dto.Village

func (c *Client) ListProvinces(ctx context.Context, page, perPage int) (dto.Page[dto.Province], error) {
	var source dto.Page[mappers.ProvinceSource]
	if err := c.GetJSON(ctx, "api/v1/wilayah/provinsi", locationQuery(page, perPage, "", ""), &source); err != nil {
		return dto.Page[dto.Province]{}, err
	}
	return mappers.MapProvincePage(source), nil
}

func (c *Client) ListCities(ctx context.Context, provinceID string, page, perPage int) (dto.Page[dto.City], error) {
	var source dto.Page[mappers.CitySource]
	if err := c.GetJSON(ctx, "api/v1/wilayah/kabupaten", locationQuery(page, perPage, "provinsi_id", provinceID), &source); err != nil {
		return dto.Page[dto.City]{}, err
	}
	return mappers.MapCityPage(source), nil
}

func (c *Client) ListDistricts(ctx context.Context, regencyID string, page, perPage int) (dto.Page[dto.District], error) {
	var source dto.Page[mappers.DistrictSource]
	if err := c.GetJSON(ctx, "api/v1/wilayah/kecamatan", locationQuery(page, perPage, "kabupaten_id", regencyID), &source); err != nil {
		return dto.Page[dto.District]{}, err
	}
	return mappers.MapDistrictPage(source), nil
}

func (c *Client) ListVillages(ctx context.Context, districtID string, page, perPage int) (dto.Page[dto.Village], error) {
	var source dto.Page[mappers.VillageSource]
	if err := c.GetJSON(ctx, "api/v1/wilayah/kelurahan", locationQuery(page, perPage, "kecamatan_id", districtID), &source); err != nil {
		return dto.Page[dto.Village]{}, err
	}
	return mappers.MapVillagePage(source), nil
}

func locationQuery(page, perPage int, parentKey, parentID string) url.Values {
	query := url.Values{
		"page":     {strconv.Itoa(page)},
		"per_page": {strconv.Itoa(perPage)},
	}
	if parentKey != "" {
		query.Set(parentKey, parentID)
	}
	return query
}
