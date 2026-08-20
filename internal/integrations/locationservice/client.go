package locationservice

import (
	"context"
	"errors"
	"net/url"
	"strings"

	"kota-siaga/internal/dto"
	"kota-siaga/internal/integrations/apiindonesia"
	"kota-siaga/pkg/config"
)

type Client struct {
	transport *apiindonesia.Client
}

type sourceLocation struct {
	Code       string `json:"code"`
	FullCode   string `json:"full_code"`
	Name       string `json:"name"`
	ParentCode string `json:"parent_code"`
	PostalCode string `json:"postal_code"`
}

type sourceResponse struct {
	Data []sourceLocation `json:"data"`
}

func NewClient(clientConfig config.LocationServiceConfig) (*Client, error) {
	if err := config.ValidateLocationServiceBaseURL(clientConfig.BaseURL); err != nil {
		return nil, errors.New("invalid location service base URL")
	}

	transport, err := apiindonesia.NewClient(apiindonesia.APIIndonesiaConfig{
		BaseURL: clientConfig.BaseURL,
		Timeout: clientConfig.Timeout,
	})
	if err != nil {
		return nil, errors.New("invalid location service base URL")
	}
	return &Client{transport: transport}, nil
}

func (c *Client) ListProvinces(ctx context.Context, page, perPage int) (dto.Page[dto.Province], error) {
	items, err := c.list(ctx, "api/locations/provinces", nil)
	if err != nil {
		return dto.Page[dto.Province]{}, err
	}

	provinces := make([]dto.Province, len(items))
	for i, item := range items {
		provinces[i] = mapProvince(item)
	}
	return paginate(provinces, page, perPage), nil
}

func (c *Client) ListCities(ctx context.Context, provinceID string, page, perPage int) (dto.Page[dto.City], error) {
	items, err := c.list(ctx, "api/locations/regencies", url.Values{"province_code": {dottedCode(provinceID)}})
	if err != nil {
		return dto.Page[dto.City]{}, err
	}

	cities := make([]dto.City, len(items))
	for i, item := range items {
		cities[i] = mapCity(item, provinceID)
	}
	return paginate(cities, page, perPage), nil
}

func (c *Client) ListDistricts(ctx context.Context, regencyID string, page, perPage int) (dto.Page[dto.District], error) {
	items, err := c.list(ctx, "api/locations/districts", url.Values{"regency_code": {dottedCode(regencyID)}})
	if err != nil {
		return dto.Page[dto.District]{}, err
	}

	districts := make([]dto.District, len(items))
	for i, item := range items {
		districts[i] = mapDistrict(item, regencyID)
	}
	return paginate(districts, page, perPage), nil
}

func (c *Client) ListVillages(ctx context.Context, districtID string, page, perPage int) (dto.Page[dto.Village], error) {
	items, err := c.list(ctx, "api/locations/villages", url.Values{"district_code": {dottedCode(districtID)}})
	if err != nil {
		return dto.Page[dto.Village]{}, err
	}

	villages := make([]dto.Village, len(items))
	for i, item := range items {
		villages[i] = mapVillage(item, districtID)
	}
	return paginate(villages, page, perPage), nil
}

func (c *Client) list(ctx context.Context, requestPath string, query url.Values) ([]sourceLocation, error) {
	if c == nil || c.transport == nil {
		return nil, errors.New("location service client is not configured")
	}

	var response sourceResponse
	if err := c.transport.GetJSON(ctx, requestPath, query, &response); err != nil {
		return nil, err
	}
	return response.Data, nil
}

func mapProvince(source sourceLocation) dto.Province {
	code := providerCode(source)
	return dto.Province{
		ID:       compactCode(code),
		Code:     code,
		Name:     source.Name,
		IsActive: true,
	}
}

func mapCity(source sourceLocation, parentID string) dto.City {
	code := providerCode(source)
	return dto.City{
		ID:         compactCode(code),
		ProvinceID: compactParentID(source.ParentCode, parentID),
		Code:       code,
		Name:       source.Name,
		IsActive:   true,
	}
}

func mapDistrict(source sourceLocation, parentID string) dto.District {
	code := providerCode(source)
	return dto.District{
		ID:        compactCode(code),
		RegencyID: compactParentID(source.ParentCode, parentID),
		Code:      code,
		Name:      source.Name,
		IsActive:  true,
	}
}

func mapVillage(source sourceLocation, parentID string) dto.Village {
	code := providerCode(source)
	return dto.Village{
		ID:         compactCode(code),
		DistrictID: compactParentID(source.ParentCode, parentID),
		Code:       code,
		Name:       source.Name,
		PostalCode: source.PostalCode,
		IsActive:   true,
	}
}

func providerCode(source sourceLocation) string {
	if source.Code != "" {
		return source.Code
	}
	return source.FullCode
}

func compactParentID(providerParent, fallback string) string {
	if providerParent == "" {
		return compactCode(fallback)
	}
	return compactCode(providerParent)
}

func compactCode(code string) string {
	return strings.ReplaceAll(code, ".", "")
}

func dottedCode(code string) string {
	if strings.Contains(code, ".") {
		return code
	}

	switch len(code) {
	case 4:
		return code[:2] + "." + code[2:]
	case 6, 7:
		return code[:2] + "." + code[2:4] + "." + code[4:6]
	case 10:
		return code[:2] + "." + code[2:4] + "." + code[4:6] + "." + code[6:]
	default:
		return code
	}
}

func paginate[T any](items []T, page, perPage int) dto.Page[T] {
	total := len(items)
	totalPages := 0
	if total > 0 {
		totalPages = (total + perPage - 1) / perPage
	}

	data := make([]T, 0)
	if page >= 1 && perPage >= 1 && page <= totalPages {
		start := (page - 1) * perPage
		end := start + perPage
		if end > total {
			end = total
		}
		data = items[start:end]
	}

	return dto.Page[T]{
		Data:       data,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}
}
