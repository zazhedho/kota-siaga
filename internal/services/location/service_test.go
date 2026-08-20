package locationservice

import (
	"context"
	"errors"
	"testing"

	"kota-siaga/internal/cache/location"
	"kota-siaga/internal/dto"

	"github.com/go-redis/redismock/v9"
)

type upstreamFake struct {
	provinces dto.Page[dto.Province]
	cities    dto.Page[dto.City]
	districts dto.Page[dto.District]
	villages  dto.Page[dto.Village]
	err       error
	calls     []string
}

func (f *upstreamFake) ListProvinces(_ context.Context, page, perPage int) (dto.Page[dto.Province], error) {
	f.calls = append(f.calls, "province")
	return f.provinces, f.err
}
func (f *upstreamFake) ListCities(_ context.Context, provinceID string, page, perPage int) (dto.Page[dto.City], error) {
	f.calls = append(f.calls, "city:"+provinceID)
	return f.cities, f.err
}
func (f *upstreamFake) ListDistricts(_ context.Context, regencyID string, page, perPage int) (dto.Page[dto.District], error) {
	f.calls = append(f.calls, "district:"+regencyID)
	return f.districts, f.err
}
func (f *upstreamFake) ListVillages(_ context.Context, districtID string, page, perPage int) (dto.Page[dto.Village], error) {
	f.calls = append(f.calls, "village:"+districtID)
	return f.villages, f.err
}

func TestServiceListsEachLocationLevel(t *testing.T) {
	fake := &upstreamFake{
		provinces: dto.Page[dto.Province]{Data: []dto.Province{{ID: "11", Name: "Aceh"}}, Total: 1, Page: 1, PerPage: 20},
		cities:    dto.Page[dto.City]{Data: []dto.City{{ID: "3201", ProvinceID: "32", Name: "Bogor"}}, Total: 1, Page: 1, PerPage: 20},
		districts: dto.Page[dto.District]{Data: []dto.District{{ID: "327301", RegencyID: "3273", Name: "Bogor Selatan"}}, Total: 1, Page: 1, PerPage: 20},
		villages:  dto.Page[dto.Village]{Data: []dto.Village{{ID: "3273010001", DistrictID: "3273010", Name: "Bondongan"}}, Total: 1, Page: 1, PerPage: 20},
	}
	svc := NewService(fake, nil)

	if got, err := svc.ListProvinces(context.Background(), 1, 20); err != nil || len(got.Data) != 1 || got.Data[0].Name != "Aceh" {
		t.Fatalf("province result: got=%+v err=%v", got, err)
	}
	if got, err := svc.ListCities(context.Background(), "32", 1, 20); err != nil || got.Data[0].ProvinceID != "32" {
		t.Fatalf("city result: got=%+v err=%v", got, err)
	}
	if got, err := svc.ListDistricts(context.Background(), "3273", 1, 20); err != nil || got.Data[0].RegencyID != "3273" {
		t.Fatalf("district result: got=%+v err=%v", got, err)
	}
	if got, err := svc.ListVillages(context.Background(), "3273010", 1, 20); err != nil || got.Data[0].DistrictID != "3273010" {
		t.Fatalf("village result: got=%+v err=%v", got, err)
	}

	if len(fake.calls) != 4 || fake.calls[1] != "city:32" || fake.calls[2] != "district:3273" || fake.calls[3] != "village:3273010" {
		t.Fatalf("unexpected upstream calls: %#v", fake.calls)
	}
}

func TestServiceRejectsInvalidPaginationAndParentIDsBeforeUpstream(t *testing.T) {
	fake := &upstreamFake{}
	svc := NewService(fake, nil)
	ctx := context.Background()

	invalid := []struct {
		name string
		call func() error
	}{
		{name: "page zero", call: func() error { _, err := svc.ListProvinces(ctx, 0, 20); return err }},
		{name: "per page zero", call: func() error { _, err := svc.ListProvinces(ctx, 1, 0); return err }},
		{name: "per page too large", call: func() error { _, err := svc.ListProvinces(ctx, 1, 101); return err }},
		{name: "missing parent", call: func() error { _, err := svc.ListCities(ctx, "", 1, 20); return err }},
		{name: "non numeric parent", call: func() error { _, err := svc.ListDistricts(ctx, "32x", 1, 20); return err }},
	}
	for _, tt := range invalid {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.call(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
	if len(fake.calls) != 0 {
		t.Fatalf("validation called upstream: %#v", fake.calls)
	}
}

func TestServiceUsesCachedPageWithoutUpstreamCall(t *testing.T) {
	redisClient, mock := redismock.NewClientMock()
	defer redisClient.Close()
	key := locationcache.ProvinceKey(1, 20)
	mock.ExpectGet(key).SetVal(`{"data":[{"id":"11","code":"11","name":"Cached"}],"total":1,"page":1,"per_page":20,"total_pages":1}`)

	fake := &upstreamFake{err: errors.New("upstream must not be called")}
	svc := NewService(fake, redisClient)
	got, err := svc.ListProvinces(context.Background(), 1, 20)
	if err != nil {
		t.Fatalf("cached request error = %v", err)
	}
	if len(got.Data) != 1 || got.Data[0].Name != "Cached" {
		t.Fatalf("unexpected cached result: %+v", got)
	}
	if len(fake.calls) != 0 {
		t.Fatalf("cache hit called upstream: %#v", fake.calls)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("redis expectations: %v", err)
	}
}

func TestServiceContinuesAfterRedisErrorAndCachesUpstreamResult(t *testing.T) {
	redisClient, mock := redismock.NewClientMock()
	defer redisClient.Close()
	key := locationcache.CityKey("32", 1, 20)
	mock.ExpectGet(key).SetErr(errors.New("redis unavailable"))
	mock.ExpectSet(key, `{"data":[{"id":"3201","province_id":"32","code":"","name":"Bogor","alternate_name":"","is_city":false,"latitude":0,"longitude":0,"is_active":false}],"total":1,"page":1,"per_page":20,"total_pages":1}`, 24*60*60*1000000000).SetVal("OK")

	fake := &upstreamFake{cities: dto.Page[dto.City]{Data: []dto.City{{ID: "3201", ProvinceID: "32", Name: "Bogor"}}, Total: 1, Page: 1, PerPage: 20, TotalPages: 1}}
	svc := NewService(fake, redisClient)
	got, err := svc.ListCities(context.Background(), "32", 1, 20)
	if err != nil {
		t.Fatalf("upstream fallback error = %v", err)
	}
	if len(got.Data) != 1 || got.Data[0].Name != "Bogor" || len(fake.calls) != 1 {
		t.Fatalf("unexpected fallback result/calls: %+v %#v", got, fake.calls)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("redis expectations: %v", err)
	}
}

func TestServiceReturnsUpstreamError(t *testing.T) {
	wantErr := errors.New("upstream unavailable")
	svc := NewService(&upstreamFake{err: wantErr}, nil)
	_, err := svc.ListVillages(context.Background(), "3273010", 1, 20)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected upstream error %v, got %v", wantErr, err)
	}
}
