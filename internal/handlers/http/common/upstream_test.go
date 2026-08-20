package handlercommon

import "testing"

func TestSafeUpstreamCodeKeepsSafeCodesAndRedactsUnsafeValues(t *testing.T) {
	tests := []struct {
		name string
		code string
		want string
	}{
		{name: "safe code", code: " quota_exceeded ", want: "QUOTA_EXCEEDED"},
		{name: "empty code", code: "", want: "UNKNOWN"},
		{name: "unsafe characters", code: "body with secret", want: "UPSTREAM_ERROR"},
		{name: "too long", code: "ABCDEFGHIJKLMNOPQRSTUVWXYZABCDEFGHIJKLMNOPQRSTUVWXYZABCDEFGHIJKLM", want: "UNKNOWN"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SafeUpstreamCode(tt.code); got != tt.want {
				t.Fatalf("SafeUpstreamCode(%q) = %q, want %q", tt.code, got, tt.want)
			}
		})
	}
}
