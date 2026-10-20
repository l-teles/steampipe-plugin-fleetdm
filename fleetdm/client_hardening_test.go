package fleetdm

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newConfiguredClient(t *testing.T, serverURL string) *FleetDMClient {
	t.Helper()
	c, err := NewFleetDMClient(testCtx(t), testConnection(ptr(serverURL), ptr("SECRET"), nil))
	if err != nil {
		t.Fatalf("NewFleetDMClient: %v", err)
	}
	return c
}

func TestGetRefusesRedirects(t *testing.T) {
	var leaked string
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		leaked = r.Header.Get("Authorization")
	}))
	defer target.Close()
	src := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL+"/api/v1/fleet/users", http.StatusMovedPermanently)
	}))
	defer src.Close()

	err := newConfiguredClient(t, src.URL).Get(testCtx(t), "scripts/results/x", nil, &struct{}{})
	var apiErr *FleetAPIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusMovedPermanently {
		t.Fatalf("want a 301 FleetAPIError, got %v", err)
	}
	if !strings.Contains(apiErr.Snippet, target.URL) {
		t.Errorf("error should name the redirect target, got %q", apiErr.Snippet)
	}
	if leaked != "" {
		t.Errorf("redirect target received Authorization %q", leaked)
	}
}

func TestGetDecodeErrorOmitsBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"secrets":[{"secret":"LIVE-ENROLL-SECRET"}],"n":"x"}`))
	}))
	defer srv.Close()

	var target struct {
		N int `json:"n"`
	}
	err := newConfiguredClient(t, srv.URL).Get(testCtx(t), "x", nil, &target)
	if err == nil {
		t.Fatal("want a decode error")
	}
	if strings.Contains(err.Error(), "LIVE-ENROLL-SECRET") {
		t.Errorf("decode error leaks the response body: %v", err)
	}
}

func TestReadBodyLimit(t *testing.T) {
	if _, err := readBody(strings.NewReader("12345"), 5); err != nil {
		t.Errorf("body at the limit: %v", err)
	}
	if _, err := readBody(strings.NewReader("123456"), 5); err == nil {
		t.Error("body over the limit should fail")
	}
}
