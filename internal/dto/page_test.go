package dto

import (
	"encoding/json"
	"testing"
)

func TestPageUnmarshalsNestedMeta(t *testing.T) {
	var got Page[string]
	if err := json.Unmarshal([]byte(`{"data":["one","two"],"meta":{"total":7,"page":2,"per_page":2,"total_pages":4}}`), &got); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if len(got.Data) != 2 || got.Data[0] != "one" || got.Data[1] != "two" {
		t.Fatalf("unexpected data: %#v", got.Data)
	}
	if got.Total != 7 || got.Page != 2 || got.PerPage != 2 || got.TotalPages != 4 {
		t.Fatalf("unexpected pagination metadata: %#v", got)
	}
}

func TestPageUnmarshalsFlatShape(t *testing.T) {
	var got Page[string]
	if err := json.Unmarshal([]byte(`{"data":["cached"],"total":1,"page":1,"per_page":20,"total_pages":1}`), &got); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if len(got.Data) != 1 || got.Data[0] != "cached" {
		t.Fatalf("unexpected data: %#v", got.Data)
	}
	if got.Total != 1 || got.Page != 1 || got.PerPage != 20 || got.TotalPages != 1 {
		t.Fatalf("unexpected pagination metadata: %#v", got)
	}
}
