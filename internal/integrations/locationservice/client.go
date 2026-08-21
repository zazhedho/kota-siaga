package locationservice

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
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
	Level      string `json:"level"`
	ParentCode string `json:"parent_code"`
	PostalCode string `json:"postal_code"`
}

type sourceResponse struct {
	Data []sourceLocation `json:"data"`
}

type sourceDetailResponse struct {
	Data sourceLocation `json:"data"`
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

func (c *Client) SearchLocations(ctx context.Context, query string, limit int) ([]dto.LocationSearchItem, error) {
	items, err := c.list(ctx, "api/locations/search", url.Values{
		"q":     {query},
		"limit": {strconv.Itoa(limit)},
	})
	if err != nil {
		return nil, err
	}

	cache := make(map[string]sourceLocation)
	results := make([]dto.LocationSearchItem, 0, len(items))
	for _, item := range items {
		code := providerCode(item)
		result := dto.LocationSearchItem{
			ID:         compactCode(code),
			Code:       code,
			Name:       item.Name,
			Level:      item.Level,
			PostalCode: item.PostalCode,
		}
		if isSearchableLevel(item.Level) {
			path, err := c.resolveLocation(ctx, code, cache)
			if err != nil {
				return nil, err
			}
			result.Hierarchy = formatHierarchy(path)
		}
		results = append(results, result)
		if limit > 0 && len(results) == limit {
			break
		}
	}
	return results, nil
}

func (c *Client) ResolveLocation(ctx context.Context, code string) (dto.LocationPath, error) {
	return c.resolveLocation(ctx, code, make(map[string]sourceLocation))
}

func (c *Client) resolveLocation(ctx context.Context, code string, cache map[string]sourceLocation) (dto.LocationPath, error) {
	selected, err := c.getCached(ctx, code, cache)
	if err != nil {
		return dto.LocationPath{}, err
	}

	switch strings.ToLower(strings.TrimSpace(selected.Level)) {
	case "district":
		city, err := c.getCached(ctx, selected.ParentCode, cache)
		if err != nil {
			return dto.LocationPath{}, err
		}
		province, err := c.getCached(ctx, city.ParentCode, cache)
		if err != nil {
			return dto.LocationPath{}, err
		}
		return dto.LocationPath{
			Level:    "district",
			Province: mapProvince(province),
			City:     mapCity(city, providerCode(province)),
			District: mapDistrict(selected, providerCode(city)),
		}, nil
	case "village":
		district, err := c.getCached(ctx, selected.ParentCode, cache)
		if err != nil {
			return dto.LocationPath{}, err
		}
		city, err := c.getCached(ctx, district.ParentCode, cache)
		if err != nil {
			return dto.LocationPath{}, err
		}
		province, err := c.getCached(ctx, city.ParentCode, cache)
		if err != nil {
			return dto.LocationPath{}, err
		}
		village := mapVillage(selected, providerCode(district))
		return dto.LocationPath{
			Level:    "village",
			Province: mapProvince(province),
			City:     mapCity(city, providerCode(province)),
			District: mapDistrict(district, providerCode(city)),
			Village:  &village,
		}, nil
	default:
		return dto.LocationPath{}, fmt.Errorf("unsupported location level %q", selected.Level)
	}
}

func (c *Client) getCached(ctx context.Context, code string, cache map[string]sourceLocation) (sourceLocation, error) {
	key := dottedCode(strings.TrimSpace(code))
	if item, ok := cache[key]; ok {
		return item, nil
	}
	item, err := c.get(ctx, code)
	if err != nil {
		return sourceLocation{}, err
	}
	cache[key] = item
	return item, nil
}

func isSearchableLevel(level string) bool {
	level = strings.ToLower(strings.TrimSpace(level))
	return level == "district" || level == "village"
}

func formatHierarchy(path dto.LocationPath) string {
	if path.Village != nil {
		return fmt.Sprintf("%s — %s, %s, %s", path.Village.Name, path.District.Name, path.City.Name, path.Province.Name)
	}
	return fmt.Sprintf("%s — %s, %s", path.District.Name, path.City.Name, path.Province.Name)
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

func (c *Client) get(ctx context.Context, code string) (sourceLocation, error) {
	if c == nil || c.transport == nil {
		return sourceLocation{}, errors.New("location service client is not configured")
	}
	if strings.TrimSpace(code) == "" {
		return sourceLocation{}, errors.New("location code is empty")
	}

	var response sourceDetailResponse
	if err := c.transport.GetJSON(ctx, "api/locations/"+dottedCode(code), nil, &response); err != nil {
		return sourceLocation{}, err
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
