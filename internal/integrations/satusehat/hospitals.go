package satusehat

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"kota-siaga/internal/dto"
)

const (
	hospitalFacilityType = "104"
	hospitalSearchLimit  = 2000
)

func (c *Client) ListHospitals(ctx context.Context, kabupatenID, search string, page, perPage int) (dto.Page[dto.Hospital], error) {
	search = strings.TrimSpace(search)
	var source hospitalPage
	query := url.Values{
		"limit":        {strconv.Itoa(perPage)},
		"page":         {strconv.Itoa(page)},
		"jenis_sarana": {hospitalFacilityType},
		"kode_kabkota": {normalizeRegencyCode(kabupatenID)},
	}
	if search != "" {
		query.Set("limit", strconv.Itoa(hospitalSearchLimit))
		query.Set("page", "1")
	}
	if err := c.GetJSON(ctx, "v1/mastersaranaindex/mastersarana", query, &source); err != nil {
		return dto.Page[dto.Hospital]{}, err
	}

	if search != "" {
		hospitals := mapHospitals(source.Data)
		for upstreamPage := 2; upstreamPage <= source.TotalPages; upstreamPage++ {
			query.Set("page", strconv.Itoa(upstreamPage))
			var next hospitalPage
			if err := c.GetJSON(ctx, "v1/mastersaranaindex/mastersarana", query, &next); err != nil {
				return dto.Page[dto.Hospital]{}, err
			}
			hospitals = append(hospitals, mapHospitals(next.Data)...)
		}

		search = strings.ToLower(search)
		filtered := make([]dto.Hospital, 0, len(hospitals))
		for _, hospital := range hospitals {
			if strings.Contains(strings.ToLower(hospital.Name), search) {
				filtered = append(filtered, hospital)
			}
		}

		total := len(filtered)
		totalPages := (total + perPage - 1) / perPage
		if page > totalPages {
			filtered = filtered[:0]
		} else {
			start := (page - 1) * perPage
			end := start + perPage
			if end > total {
				end = total
			}
			filtered = filtered[start:end]
		}
		return dto.Page[dto.Hospital]{
			Data:       filtered,
			Total:      total,
			Page:       page,
			PerPage:    perPage,
			TotalPages: totalPages,
		}, nil
	}

	currentPage := source.Page
	if currentPage < 1 {
		currentPage = page
	}
	totalPages := source.TotalPages
	if totalPages < 1 && len(source.Data) > 0 {
		totalPages = 1
	}
	return dto.Page[dto.Hospital]{
		Data:       mapHospitals(source.Data),
		Page:       currentPage,
		PerPage:    perPage,
		TotalPages: totalPages,
	}, nil
}

type hospitalPage struct {
	Page       int              `json:"page"`
	TotalPages int              `json:"total_page"`
	Data       []hospitalSource `json:"data"`
}

type hospitalSource struct {
	CodeSATUSEHAT string           `json:"kode_satusehat"`
	CodeFacility  string           `json:"kode_sarana"`
	Name          string           `json:"nama"`
	Phone         string           `json:"telp"`
	Address       string           `json:"alamat"`
	Province      facilityRegion   `json:"provinsi"`
	Regency       facilityRegion   `json:"kabkota"`
	FacilityType  facilityCategory `json:"jenis_sarana"`
	Subtype       facilityCategory `json:"subjenis"`
	Class         facilityCategory `json:"kelas_sarana"`
	IsActive      bool             `json:"status_aktif"`
}

type facilityRegion struct {
	Code string `json:"kode"`
	Name string `json:"nama"`
}

type facilityCategory struct {
	Code string `json:"kode"`
	Name string `json:"nama"`
}

func mapHospitals(source []hospitalSource) []dto.Hospital {
	hospitals := make([]dto.Hospital, len(source))
	for i, item := range source {
		id := strings.TrimSpace(item.CodeSATUSEHAT)
		if id == "" {
			id = strings.TrimSpace(item.CodeFacility)
		}
		hospitalType := strings.TrimSpace(item.Subtype.Name)
		if hospitalType == "" {
			hospitalType = strings.TrimSpace(item.FacilityType.Name)
		}
		hospitals[i] = dto.Hospital{
			ID:           id,
			Name:         strings.TrimSpace(item.Name),
			Type:         hospitalType,
			Class:        strings.TrimSpace(item.Class.Name),
			Address:      strings.TrimSpace(item.Address),
			Phone:        strings.TrimSpace(item.Phone),
			ProvinceName: strings.TrimSpace(item.Province.Name),
			RegencyName:  strings.TrimSpace(item.Regency.Name),
			IsActive:     item.IsActive,
		}
	}
	return hospitals
}

func normalizeRegencyCode(value string) string {
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return strings.TrimSpace(value)
	}
	return fmt.Sprintf("%04d", parsed)
}
