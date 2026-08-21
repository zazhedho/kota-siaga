package bmkg

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"kota-siaga/internal/dto"
)

const latestEarthquakePath = "DataMKG/TEWS/autogempa.json"

type earthquakeResponse struct {
	Infogempa struct {
		Earthquake earthquakeSource `json:"gempa"`
	} `json:"Infogempa"`
}

type earthquakeSource struct {
	DateTime    string `json:"DateTime"`
	Coordinates string `json:"Coordinates"`
	Magnitude   string `json:"Magnitude"`
	Depth       string `json:"Kedalaman"`
	Region      string `json:"Wilayah"`
	Potential   string `json:"Potensi"`
	FeltAreas   string `json:"Dirasakan"`
}

func (c *Client) ListLatest(ctx context.Context) ([]dto.Earthquake, error) {
	var response earthquakeResponse
	if err := c.getJSON(ctx, latestEarthquakePath, &response); err != nil {
		return nil, err
	}

	earthquake, err := mapEarthquake(response.Infogempa.Earthquake)
	if err != nil {
		return nil, fmt.Errorf("decode BMKG earthquake: %w", err)
	}
	return []dto.Earthquake{earthquake}, nil
}

func mapEarthquake(source earthquakeSource) (dto.Earthquake, error) {
	dateTime := strings.TrimSpace(source.DateTime)
	if dateTime == "" {
		return dto.Earthquake{}, errors.New("missing DateTime")
	}

	latitude, longitude, err := parseCoordinates(source.Coordinates)
	if err != nil {
		return dto.Earthquake{}, err
	}
	magnitude, err := parseMeasurement(source.Magnitude)
	if err != nil {
		return dto.Earthquake{}, fmt.Errorf("invalid Magnitude: %w", err)
	}
	depth, err := parseMeasurement(source.Depth)
	if err != nil {
		return dto.Earthquake{}, fmt.Errorf("invalid Kedalaman: %w", err)
	}

	feltAreas := strings.TrimSpace(source.FeltAreas)
	var feltAreasValue *string
	if feltAreas != "" {
		feltAreasValue = &feltAreas
	}

	return dto.Earthquake{
		ID:        "bmkg:" + dateTime,
		DateTime:  dateTime,
		Magnitude: magnitude,
		DepthKM:   depth,
		Latitude:  latitude,
		Longitude: longitude,
		Region:    strings.TrimSpace(source.Region),
		Potential: strings.TrimSpace(source.Potential),
		IsFelt:    feltAreasValue != nil,
		FeltAreas: feltAreasValue,
		Source:    "BMKG",
	}, nil
}

func parseCoordinates(value string) (float64, float64, error) {
	parts := strings.Split(value, ",")
	if len(parts) != 2 {
		return 0, 0, errors.New("invalid Coordinates")
	}

	latitude, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if err != nil {
		return 0, 0, errors.New("invalid Coordinates")
	}
	longitude, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err != nil {
		return 0, 0, errors.New("invalid Coordinates")
	}
	return latitude, longitude, nil
}

func parseMeasurement(value string) (float64, error) {
	value = strings.TrimSpace(strings.TrimSuffix(strings.ToLower(strings.TrimSpace(value)), "km"))
	if value == "" {
		return 0, errors.New("empty value")
	}
	return strconv.ParseFloat(strings.TrimSpace(value), 64)
}
