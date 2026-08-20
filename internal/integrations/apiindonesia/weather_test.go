package apiindonesia

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"kota-siaga/internal/dto"
)

func TestGetWeatherUsesExactEndpointAndMapsForecastFields(t *testing.T) {
	const adm4 = "32.73.01.1001"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/cuaca" {
			t.Errorf("expected /api/v1/cuaca, got %q", r.URL.Path)
		}
		if r.URL.RawQuery != "adm4="+adm4 {
			t.Errorf("expected exact adm4 query, got %q", r.URL.RawQuery)
		}
		_, _ = fmt.Fprint(w, `{"data":[{"id":"forecast-1","adm4":"32.73.01.1001","provinsi":"Jawa Barat","kotkab":"Kota Bogor","kecamatan":"Bogor Selatan","desa":"Bondongan","datetime":"2026-08-20T00:00:00Z","local_datetime":"2026-08-20T07:00:00+07:00","weather":"Cerah","weather_code":"1","weather_desc":"Cerah Berawan","weather_desc_en":"Partly Cloudy","temperature_c":24.5,"humidity_percent":81.2,"cloud_cover_percent":32.5,"precipitation_mm":0.4,"wind_direction":"N","wind_direction_to":"S","wind_direction_degrees":180,"wind_speed":4.75,"visibility_m":9500.5,"visibility_text":"Good","analysis_date":"2026-08-19","source":"BMKG"}]}`)
	}))
	defer server.Close()

	client, err := NewClient(APIIndonesiaConfig{BaseURL: server.URL, Timeout: time.Second})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	got, err := client.GetWeather(context.Background(), adm4)
	if err != nil {
		t.Fatalf("GetWeather() error = %v", err)
	}
	want := dto.WeatherForecast{
		ID:                   "forecast-1",
		Adm4:                 adm4,
		Province:             "Jawa Barat",
		Regency:              "Kota Bogor",
		District:             "Bogor Selatan",
		Village:              "Bondongan",
		Datetime:             "2026-08-20T00:00:00Z",
		LocalDatetime:        "2026-08-20T07:00:00+07:00",
		Weather:              "Cerah",
		WeatherCode:          "1",
		WeatherDescription:   "Cerah Berawan",
		WeatherDescriptionEN: "Partly Cloudy",
		TemperatureC:         24.5,
		HumidityPercent:      81.2,
		CloudCoverPercent:    32.5,
		PrecipitationMM:      0.4,
		WindDirection:        "N",
		WindDirectionTo:      "S",
		WindDirectionDegrees: 180,
		WindSpeed:            4.75,
		VisibilityM:          9500.5,
		VisibilityText:       "Good",
		AnalysisDate:         "2026-08-19",
		Source:               "BMKG",
	}
	if len(got) != 1 || got[0] != want {
		t.Fatalf("unexpected normalized forecast: %+v", got)
	}
}

func TestGetWeatherHonorsCanceledContext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, `{"data":[]}`)
	}))
	defer server.Close()

	client, err := NewClient(APIIndonesiaConfig{BaseURL: server.URL, Timeout: time.Second})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = client.GetWeather(ctx, "32.73.01.1001")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
}
