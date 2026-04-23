package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSearch_MissingQuery_ReturnsBadRequest(t *testing.T) {
	h := &APIHandler{}
	req := httptest.NewRequest(http.MethodGet, "/api/search", nil)
	rr := httptest.NewRecorder()

	h.Search(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response JSON: %v", err)
	}
	if body["error"] != "query param 'q' is required" {
		t.Fatalf("unexpected error message: %q", body["error"])
	}
}

func TestLoginStatus_MissingKey_ReturnsBadRequest(t *testing.T) {
	h := &APIHandler{}
	req := httptest.NewRequest(http.MethodGet, "/api/login/status", nil)
	rr := httptest.NewRecorder()

	h.LoginStatus(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response JSON: %v", err)
	}
	if body["error"] != "key is required" {
		t.Fatalf("unexpected error message: %q", body["error"])
	}
}

func TestParseSearchLimit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		raw      string
		fallback int
		max      int
		want     int
	}{
		{name: "empty uses fallback", raw: "", fallback: 5, max: 50, want: 5},
		{name: "valid positive value", raw: "8", fallback: 5, max: 50, want: 8},
		{name: "non-numeric uses fallback", raw: "abc", fallback: 5, max: 50, want: 5},
		{name: "negative uses fallback", raw: "-1", fallback: 5, max: 50, want: 5},
		{name: "zero uses fallback", raw: "0", fallback: 5, max: 50, want: 5},
		{name: "caps at max", raw: "999", fallback: 5, max: 50, want: 50},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := parseSearchLimit(tc.raw, tc.fallback, tc.max)
			if got != tc.want {
				t.Fatalf("parseSearchLimit(%q, %d, %d) = %d; want %d",
					tc.raw, tc.fallback, tc.max, got, tc.want)
			}
		})
	}
}
