package mappers

import "kota-siaga/internal/dto"

type ProvinceSource struct {
	ID       string  `json:"id"`
	Code     string  `json:"code"`
	Name     string  `json:"name"`
	AltName  string  `json:"alt_name"`
	Lat      float64 `json:"lat"`
	Lng      float64 `json:"lng"`
	IsActive int     `json:"is_active"`
}

type CitySource struct {
	ID         string  `json:"id"`
	ProvinceID string  `json:"province_id"`
	Code       string  `json:"code"`
	Name       string  `json:"name"`
	AltName    string  `json:"alt_name"`
	IsCity     int     `json:"is_city"`
	Lat        float64 `json:"lat"`
	Lng        float64 `json:"lng"`
	IsActive   int     `json:"is_active"`
}

type DistrictSource struct {
	ID        string  `json:"id"`
	RegencyID string  `json:"regency_id"`
	Code      string  `json:"code"`
	Name      string  `json:"name"`
	AltName   string  `json:"alt_name"`
	Lat       float64 `json:"lat"`
	Lng       float64 `json:"lng"`
	IsActive  int     `json:"is_active"`
}

type VillageSource struct {
	ID               string  `json:"id"`
	DistrictID       string  `json:"district_id"`
	Code             string  `json:"code"`
	Name             string  `json:"name"`
	AltName          string  `json:"alt_name"`
	PostalCode       string  `json:"postal_code"`
	IsCourierSupport int     `json:"is_courier_support"`
	Lat              float64 `json:"lat"`
	Lng              float64 `json:"lng"`
	IsActive         int     `json:"is_active"`
}

func MapProvince(source ProvinceSource) dto.Province {
	return dto.Province{
		ID:            source.ID,
		Code:          source.Code,
		Name:          source.Name,
		AlternateName: source.AltName,
		Latitude:      source.Lat,
		Longitude:     source.Lng,
		IsActive:      source.IsActive != 0,
	}
}

func MapCity(source CitySource) dto.City {
	return dto.City{
		ID:            source.ID,
		ProvinceID:    source.ProvinceID,
		Code:          source.Code,
		Name:          source.Name,
		AlternateName: source.AltName,
		IsCity:        source.IsCity != 0,
		Latitude:      source.Lat,
		Longitude:     source.Lng,
		IsActive:      source.IsActive != 0,
	}
}

func MapDistrict(source DistrictSource) dto.District {
	return dto.District{
		ID:            source.ID,
		RegencyID:     source.RegencyID,
		Code:          source.Code,
		Name:          source.Name,
		AlternateName: source.AltName,
		Latitude:      source.Lat,
		Longitude:     source.Lng,
		IsActive:      source.IsActive != 0,
	}
}

func MapVillage(source VillageSource) dto.Village {
	return dto.Village{
		ID:               source.ID,
		DistrictID:       source.DistrictID,
		Code:             source.Code,
		Name:             source.Name,
		AlternateName:    source.AltName,
		PostalCode:       source.PostalCode,
		IsCourierSupport: source.IsCourierSupport != 0,
		Latitude:         source.Lat,
		Longitude:        source.Lng,
		IsActive:         source.IsActive != 0,
	}
}

func MapProvincePage(source dto.Page[ProvinceSource]) dto.Page[dto.Province] {
	return dto.Page[dto.Province]{Data: mapItems(source.Data, MapProvince), Total: source.Total, Page: source.Page, PerPage: source.PerPage, TotalPages: source.TotalPages}
}

func MapCityPage(source dto.Page[CitySource]) dto.Page[dto.City] {
	return dto.Page[dto.City]{Data: mapItems(source.Data, MapCity), Total: source.Total, Page: source.Page, PerPage: source.PerPage, TotalPages: source.TotalPages}
}

func MapDistrictPage(source dto.Page[DistrictSource]) dto.Page[dto.District] {
	return dto.Page[dto.District]{Data: mapItems(source.Data, MapDistrict), Total: source.Total, Page: source.Page, PerPage: source.PerPage, TotalPages: source.TotalPages}
}

func MapVillagePage(source dto.Page[VillageSource]) dto.Page[dto.Village] {
	return dto.Page[dto.Village]{Data: mapItems(source.Data, MapVillage), Total: source.Total, Page: source.Page, PerPage: source.PerPage, TotalPages: source.TotalPages}
}

func mapItems[S any, T any](items []S, mapItem func(S) T) []T {
	mapped := make([]T, len(items))
	for i, item := range items {
		mapped[i] = mapItem(item)
	}
	return mapped
}
