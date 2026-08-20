package utils

import (
	"errors"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

func TestJSONAndStringHelpers(t *testing.T) {
	if got := JsonEncode(map[string]string{"name": "Jane"}); !strings.Contains(got, "Jane") {
		t.Fatalf("unexpected JSON encoding: %q", got)
	}
	if got := NormalizePayload(struct {
		Name string `json:"name"`
	}{Name: "Jane"}); got == nil {
		t.Fatal("expected normalized payload")
	}
	if got := TitleCase("jane doe"); got != "Jane Doe" {
		t.Fatalf("unexpected title case: %q", got)
	}
	if got := CreateUUID(); got == "" {
		t.Fatal("expected uuid")
	}
	if got := JsonEncode(make(chan int)); got != "" {
		t.Fatalf("expected empty string for unsupported JSON value, got %q", got)
	}
	ch := make(chan int)
	if got := NormalizePayload(ch); got != ch {
		t.Fatal("expected unsupported payload to be returned unchanged")
	}
}

func TestGenerateLogIdAndRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

	id := uuid.New()
	ctx.Set(CtxKeyId, id)
	if got := GenerateLogId(ctx); got != id {
		t.Fatalf("expected stored uuid, got %s", got)
	}
	if got := GetRequestID(ctx); got != id.String() {
		t.Fatalf("expected request id string, got %q", got)
	}

	ctx.Set(CtxKeyId, "not-a-uuid")
	if got := GenerateLogId(ctx); got == uuid.Nil {
		t.Fatal("expected generated uuid for invalid string")
	}

	ctx.Set(CtxKeyId, " request-id ")
	if got := GetRequestID(ctx); got != "request-id" {
		t.Fatalf("expected trimmed request id, got %q", got)
	}
}

func TestValidateError(t *testing.T) {
	type request struct {
		Email string `json:"email" validate:"required,email"`
	}
	validate := validator.New()
	err := validate.Struct(request{})
	got := ValidateError(err, reflect.TypeOf(request{}), "json")
	if len(got) != 1 || got[0].Field != "email" || got[0].Message == "" {
		t.Fatalf("unexpected validation mapping: %+v", got)
	}

	got = ValidateError(errors.New("plain error"), reflect.TypeOf(request{}), "json")
	if len(got) != 1 || got[0].Message != "plain error" {
		t.Fatalf("unexpected plain error mapping: %+v", got)
	}
}

func TestValidateErrorMapsKnownTags(t *testing.T) {
	type request struct {
		Required string `json:"required" validate:"required"`
		Email    string `json:"email" validate:"email"`
		AlphaNum string `json:"alphanum" validate:"alphanum"`
		Min      string `json:"min" validate:"min=3"`
		Max      string `json:"max" validate:"max=2"`
		LTE      int    `json:"lte" validate:"lte=3"`
		GTE      int    `json:"gte" validate:"gte=3"`
		LTEField int    `json:"ltefield" validate:"ltefield=GTEField"`
		GTEField int    `json:"gtefield" validate:"gtefield=LTEField"`
		UUID     string `json:"uuid" validate:"uuid"`
	}
	validate := validator.New()
	err := validate.Struct(request{
		Email:    "bad",
		AlphaNum: "with space",
		Min:      "no",
		Max:      "too-long",
		LTE:      4,
		GTE:      2,
		LTEField: 10,
		GTEField: 1,
		UUID:     "bad",
	})
	got := ValidateError(err, reflect.TypeOf(request{}), "json")
	messages := map[string]string{}
	for _, item := range got {
		messages[item.Field] = item.Message
	}

	want := map[string]string{
		"required": "This field is required",
		"email":    "Invalid email",
		"alphanum": "Should be alphanumeric",
		"min":      "Minimum 3",
		"max":      "Maximum 2",
		"lte":      "Should be less than 3",
		"gte":      "Should be greater than 3",
		"ltefield": "Should be less than GTEField",
		"gtefield": "Should be greater than LTEField",
		"uuid":     "Invalid value",
	}
	for field, message := range want {
		if messages[field] != message {
			t.Fatalf("expected %s=%q, got %q in %#v", field, message, messages[field], messages)
		}
	}
}
