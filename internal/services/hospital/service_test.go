package hospitalservice

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	hospitalcache "kota-siaga/internal/cache/hospital"
	"kota-siaga/internal/dto"

	redismock "github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
)

type upstreamFake struct {
	page        dto.Page[dto.Hospital]
	err         error
	calls       int
	kabupatenID string
	pageNumber  int
	perPage     int
}

func (f *upstreamFake) ListHospitals(_ context.Context, kabupatenID string, page, perPage int) (dto.Page[dto.Hospital], error) {
	f.calls++
	f.kabupatenID = kabupatenID
	f.pageNumber = page
	f.perPage = perPage
	return f.page, f.err
}

func TestServiceRejectsInvalidKabupatenIDAndPaginationBeforeUpstream(t *testing.T) {
	fake := &upstreamFake{}
	service := NewService(fake, nil)
	ctx := context.Background()

	invalid := []struct {
		name string
		call func() error
	}{
		{name: "missing kabupaten ID", call: func() error { _, err := service.ListHospitals(ctx, "", 1, 20); return err }},
		{name: "non-numeric kabupaten ID", call: func() error { _, err := service.ListHospitals(ctx, "3273x", 1, 20); return err }},
		{name: "whitespace kabupaten ID", call: func() error { _, err := service.ListHospitals(ctx, " 3273", 1, 20); return err }},
		{name: "page zero", call: func() error { _, err := service.ListHospitals(ctx, "3273", 0, 20); return err }},
		{name: "per page zero", call: func() error { _, err := service.ListHospitals(ctx, "3273", 1, 0); return err }},
		{name: "per page over maximum", call: func() error { _, err := service.ListHospitals(ctx, "3273", 1, 201); return err }},
	}
	for _, tt := range invalid {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.call(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
	if fake.calls != 0 {
		t.Fatalf("validation called upstream %d times", fake.calls)
	}
}

func TestServicePreservesLeadingZeroKabupatenIDAndReturnsPage(t *testing.T) {
	want := dto.Page[dto.Hospital]{
		Data:  []dto.Hospital{{ID: "hospital-1", Name: "Official Hospital"}},
		Total: 1, Page: 1, PerPage: 200, TotalPages: 1,
	}
	fake := &upstreamFake{page: want}
	service := NewService(fake, nil)

	got, err := service.ListHospitals(context.Background(), "003273", 1, 200)
	if err != nil {
		t.Fatalf("ListHospitals() error = %v", err)
	}
	if got.Data[0].ID != "hospital-1" || got.Total != 1 || got.Page != 1 || got.PerPage != 200 || got.TotalPages != 1 {
		t.Fatalf("unexpected hospital page: %+v", got)
	}
	if fake.calls != 1 || fake.kabupatenID != "003273" || fake.pageNumber != 1 || fake.perPage != 200 {
		t.Fatalf("unexpected upstream call: calls=%d id=%q page=%d per_page=%d", fake.calls, fake.kabupatenID, fake.pageNumber, fake.perPage)
	}
}

func TestServiceUsesCachedPageWithoutUpstreamCall(t *testing.T) {
	redisClient, mock := redismock.NewClientMock()
	defer redisClient.Close()
	key := hospitalcache.Key("3273", 1, 20)
	cached := dto.Page[dto.Hospital]{
		Data:  []dto.Hospital{{ID: "cached", Name: "Cached Hospital"}},
		Total: 1, Page: 1, PerPage: 20, TotalPages: 1,
	}
	payload, err := json.Marshal(cached)
	if err != nil {
		t.Fatalf("marshal cached page: %v", err)
	}
	mock.ExpectGet(key).SetVal(string(payload))

	fake := &upstreamFake{err: errors.New("upstream must not be called")}
	service := NewService(fake, redisClient)
	got, err := service.ListHospitals(context.Background(), "3273", 1, 20)
	if err != nil {
		t.Fatalf("cached request error = %v", err)
	}
	if len(got.Data) != 1 || got.Data[0].ID != "cached" {
		t.Fatalf("unexpected cached result: %+v", got)
	}
	if fake.calls != 0 {
		t.Fatalf("cache hit called upstream %d times", fake.calls)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("redis expectations: %v", err)
	}
}

func TestServiceFailsOpenAfterRedisReadErrorAndCachesSuccessfulResult(t *testing.T) {
	t.Setenv("HOSPITAL_CACHE_TTL", "90m")
	redisClient, mock := redismock.NewClientMock()
	defer redisClient.Close()
	key := hospitalcache.Key("3273", 1, 20)
	mock.ExpectGet(key).SetErr(errors.New("redis unavailable"))
	fresh := dto.Page[dto.Hospital]{
		Data:  []dto.Hospital{{ID: "fresh", Name: "Fresh Hospital"}},
		Total: 1, Page: 1, PerPage: 20, TotalPages: 1,
	}
	payload, err := json.Marshal(fresh)
	if err != nil {
		t.Fatalf("marshal fresh page: %v", err)
	}
	mock.ExpectSet(key, string(payload), 90*time.Minute).SetVal("OK")

	service := NewService(&upstreamFake{page: fresh}, redisClient)
	got, err := service.ListHospitals(context.Background(), "3273", 1, 20)
	if err != nil || len(got.Data) != 1 || got.Data[0].ID != "fresh" {
		t.Fatalf("unexpected Redis fallback result: got=%+v err=%v", got, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("redis expectations: %v", err)
	}
}

func TestServiceFailsOpenAfterRedisWriteError(t *testing.T) {
	redisClient, mock := redismock.NewClientMock()
	defer redisClient.Close()
	key := hospitalcache.Key("3273", 1, 20)
	mock.ExpectGet(key).SetErr(redis.Nil)
	fresh := dto.Page[dto.Hospital]{
		Data:  []dto.Hospital{{ID: "fresh"}},
		Total: 1, Page: 1, PerPage: 20, TotalPages: 1,
	}
	payload, err := json.Marshal(fresh)
	if err != nil {
		t.Fatalf("marshal fresh page: %v", err)
	}
	mock.ExpectSet(key, string(payload), 24*time.Hour).SetErr(errors.New("redis unavailable"))

	service := NewService(&upstreamFake{page: fresh}, redisClient)
	got, err := service.ListHospitals(context.Background(), "3273", 1, 20)
	if err != nil || len(got.Data) != 1 || got.Data[0].ID != "fresh" {
		t.Fatalf("expected fresh result despite Redis write failure: got=%+v err=%v", got, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("redis expectations: %v", err)
	}
}

func TestServiceDoesNotCacheUpstreamFailure(t *testing.T) {
	redisClient, mock := redismock.NewClientMock()
	defer redisClient.Close()
	key := hospitalcache.Key("3273", 1, 20)
	mock.ExpectGet(key).SetErr(redis.Nil)
	wantErr := errors.New("upstream body and API key stay private")

	service := NewService(&upstreamFake{err: wantErr}, redisClient)
	_, err := service.ListHospitals(context.Background(), "3273", 1, 20)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected upstream error %v, got %v", wantErr, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected cache write after upstream failure: %v", err)
	}
}

func TestServiceReturnsMissingClientSentinel(t *testing.T) {
	service := NewService(nil, nil)
	_, err := service.ListHospitals(context.Background(), "3273", 1, 20)
	if !errors.Is(err, ErrHospitalClient) {
		t.Fatalf("expected ErrHospitalClient, got %v", err)
	}
}
