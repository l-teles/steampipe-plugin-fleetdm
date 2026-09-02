package fleetdm

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
)

// apiErr builds a *FleetAPIError with the given status code and snippet.
func apiErr(status int, snippet string) *FleetAPIError {
	return &FleetAPIError{
		StatusCode: status,
		Status:     fmt.Sprintf("%d %s", status, http.StatusText(status)),
		URL:        "https://fleet.example.com/api/v1/fleet/test",
		Snippet:    snippet,
	}
}

// countingFetch returns a fetch func that returns err and counts calls.
func countingFetch(err error, calls *int) func() error {
	return func() error {
		*calls++
		return err
	}
}

func TestResolveCompatUncached(t *testing.T) {
	t.Parallel()

	notFound := apiErr(http.StatusNotFound, "not found")
	unauthorized := apiErr(http.StatusUnauthorized, "authentication required")

	tests := []struct {
		name         string
		newErr       error
		oldErr       error
		wantUseNew   *bool // nil means "not resolved"
		wantErr      error
		wantNewCalls int
		wantOldCalls int
	}{
		{
			name:         "new succeeds resolves useNew true",
			newErr:       nil,
			wantUseNew:   ptr(true),
			wantNewCalls: 1,
			wantOldCalls: 0,
		},
		{
			name:         "new 404 old succeeds resolves useNew false",
			newErr:       notFound,
			oldErr:       nil,
			wantUseNew:   ptr(false),
			wantNewCalls: 1,
			wantOldCalls: 1,
		},
		{
			name:         "new 404 old 404 returns old error unresolved",
			newErr:       notFound,
			oldErr:       notFound,
			wantErr:      notFound,
			wantNewCalls: 1,
			wantOldCalls: 1,
		},
		{
			name:         "new 401 returns error unresolved without fallback",
			newErr:       unauthorized,
			wantErr:      unauthorized,
			wantNewCalls: 1,
			wantOldCalls: 0,
		},
		{
			name:         "new 404 old 500 returns old error unresolved",
			newErr:       notFound,
			oldErr:       apiErr(http.StatusInternalServerError, "boom"),
			wantErr:      apiErr(http.StatusInternalServerError, "boom"),
			wantNewCalls: 1,
			wantOldCalls: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var newCalls, oldCalls int
			useNew, err := resolveCompat(countingFetch(tc.newErr, &newCalls), countingFetch(tc.oldErr, &oldCalls), nil)

			if tc.wantErr == nil && err != nil {
				t.Fatalf("resolveCompat unexpected error: %v", err)
			}
			if tc.wantErr != nil && (err == nil || err.Error() != tc.wantErr.Error()) {
				t.Fatalf("resolveCompat error = %v, want %v", err, tc.wantErr)
			}
			if (useNew == nil) != (tc.wantUseNew == nil) {
				t.Fatalf("resolveCompat useNew = %v, want %v", useNew, tc.wantUseNew)
			}
			if useNew != nil && *useNew != *tc.wantUseNew {
				t.Fatalf("resolveCompat *useNew = %v, want %v", *useNew, *tc.wantUseNew)
			}
			if newCalls != tc.wantNewCalls || oldCalls != tc.wantOldCalls {
				t.Fatalf("resolveCompat calls = new:%d old:%d, want new:%d old:%d", newCalls, oldCalls, tc.wantNewCalls, tc.wantOldCalls)
			}
		})
	}
}

func TestResolveCompatCached(t *testing.T) {
	t.Parallel()

	notFound := apiErr(http.StatusNotFound, "not found")

	tests := []struct {
		name         string
		cached       bool
		newErr       error
		oldErr       error
		wantErr      error
		wantNewCalls int
		wantOldCalls int
	}{
		{
			name:         "cached new calls only new path",
			cached:       true,
			wantNewCalls: 1,
			wantOldCalls: 0,
		},
		{
			name:         "cached old calls only old path",
			cached:       false,
			wantNewCalls: 0,
			wantOldCalls: 1,
		},
		{
			name:         "cached new does not flip on genuine 404",
			cached:       true,
			newErr:       notFound,
			wantErr:      notFound,
			wantNewCalls: 1,
			wantOldCalls: 0,
		},
		{
			name:         "cached old does not flip on genuine 404",
			cached:       false,
			oldErr:       notFound,
			wantErr:      notFound,
			wantNewCalls: 0,
			wantOldCalls: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var newCalls, oldCalls int
			useNew, err := resolveCompat(countingFetch(tc.newErr, &newCalls), countingFetch(tc.oldErr, &oldCalls), ptr(tc.cached))

			if useNew != nil {
				t.Fatalf("resolveCompat with cached generation resolved again: useNew = %v", *useNew)
			}
			if tc.wantErr == nil && err != nil {
				t.Fatalf("resolveCompat unexpected error: %v", err)
			}
			if tc.wantErr != nil && (err == nil || err.Error() != tc.wantErr.Error()) {
				t.Fatalf("resolveCompat error = %v, want %v", err, tc.wantErr)
			}
			if newCalls != tc.wantNewCalls || oldCalls != tc.wantOldCalls {
				t.Fatalf("resolveCompat calls = new:%d old:%d, want new:%d old:%d", newCalls, oldCalls, tc.wantNewCalls, tc.wantOldCalls)
			}
		})
	}
}

func TestCoalesceSlice(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		newer []string
		older []string
		want  []string
	}{
		{name: "both empty", newer: nil, older: nil, want: nil},
		{name: "newer wins", newer: []string{"a"}, older: []string{"b"}, want: []string{"a"}},
		{name: "older fallback", newer: nil, older: []string{"b"}, want: []string{"b"}},
		{name: "empty non-nil newer falls back", newer: []string{}, older: []string{"b"}, want: []string{"b"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := coalesceSlice(tc.newer, tc.older)
			if len(got) != len(tc.want) {
				t.Fatalf("coalesceSlice(%v, %v) = %v, want %v", tc.newer, tc.older, got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("coalesceSlice(%v, %v) = %v, want %v", tc.newer, tc.older, got, tc.want)
				}
			}
		})
	}
}

func TestAddFleetIDParam(t *testing.T) {
	t.Parallel()

	params := url.Values{}
	addFleetIDParam(params, "7")
	if got := params["fleet_id"]; len(got) != 1 || got[0] != "7" {
		t.Errorf("fleet_id = %v, want [7]", got)
	}
	if got := params["team_id"]; len(got) != 1 || got[0] != "7" {
		t.Errorf("team_id = %v, want [7]", got)
	}
}

func TestIsPremiumError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "402 payment required", err: apiErr(http.StatusPaymentRequired, ""), want: true},
		{name: "403 with license snippet", err: apiErr(http.StatusForbidden, "requires a Fleet license"), want: true},
		{name: "403 with Premium snippet case-insensitive", err: apiErr(http.StatusForbidden, "Requires Fleet PREMIUM"), want: true},
		{name: "400 with premium snippet", err: apiErr(http.StatusBadRequest, "premium feature"), want: true},
		{name: "403 without license snippet", err: apiErr(http.StatusForbidden, "forbidden"), want: false},
		{name: "500 with license snippet", err: apiErr(http.StatusInternalServerError, "license"), want: false},
		{name: "wrapped 402", err: fmt.Errorf("outer: %w", apiErr(http.StatusPaymentRequired, "")), want: true},
		{name: "plain error", err: errors.New("license premium"), want: false},
		{name: "nil error", err: nil, want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := isPremiumError(tc.err); got != tc.want {
				t.Errorf("isPremiumError(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}

func TestHandlePremiumError(t *testing.T) {
	t.Parallel()

	ctx := testCtx(t)
	// nil QueryData exercises the nil-safe dedupe path.
	if !handlePremiumError(ctx, nil, "fleetdm_test", apiErr(http.StatusPaymentRequired, "")) {
		t.Error("handlePremiumError(402) = false, want true")
	}
	if handlePremiumError(ctx, nil, "fleetdm_test", apiErr(http.StatusInternalServerError, "boom")) {
		t.Error("handlePremiumError(500) = true, want false")
	}
	if handlePremiumError(ctx, nil, "fleetdm_test", nil) {
		t.Error("handlePremiumError(nil) = true, want false")
	}
}

func TestRequireExposeSecrets(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		expose  *bool
		wantErr bool
	}{
		{name: "unset denies", expose: nil, wantErr: true},
		{name: "false denies", expose: ptr(false), wantErr: true},
		{name: "true allows", expose: ptr(true), wantErr: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			d := &plugin.QueryData{
				Connection: &plugin.Connection{
					Name:   "test",
					Config: fleetdmConfig{ExposeSecrets: tc.expose},
				},
			}
			err := requireExposeSecrets(d)
			if tc.wantErr && err == nil {
				t.Fatal("requireExposeSecrets = nil error, want error")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("requireExposeSecrets unexpected error: %v", err)
			}
		})
	}
}
