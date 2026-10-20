package fleetdm

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/turbot/steampipe-plugin-sdk/v6/plugin"
)

// newTestClient builds a FleetDMClient pointed at the given test server.
func newTestClient(srv *httptest.Server) *FleetDMClient {
	return &FleetDMClient{
		BaseURL:    srv.URL + "/api/v1/fleet/",
		APIToken:   "test-token",
		HTTPClient: srv.Client(),
	}
}

func TestGetSendsExpectedHeaders(t *testing.T) {
	t.Parallel()

	var gotAuth, gotAccept, gotUA string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotAccept = r.Header.Get("Accept")
		gotUA = r.Header.Get("User-Agent")
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{}`)); err != nil {
			t.Errorf("write response: %v", err)
		}
	}))
	defer srv.Close()

	c := newTestClient(srv)
	if err := c.Get(testCtx(t), "hosts", nil, nil); err != nil {
		t.Fatalf("Get unexpected error: %v", err)
	}

	if want := "Bearer test-token"; gotAuth != want {
		t.Errorf("Authorization = %q, want %q", gotAuth, want)
	}
	if want := "application/json"; gotAccept != want {
		t.Errorf("Accept = %q, want %q", gotAccept, want)
	}
	if want := "steampipe-plugin-fleetdm/" + pluginVersion; gotUA != want {
		t.Errorf("User-Agent = %q, want %q", gotUA, want)
	}
}

func TestGetDecodesJSONResponse(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{"id": 1, "hostname": "host-A"}`)); err != nil {
			t.Errorf("write response: %v", err)
		}
	}))
	defer srv.Close()

	var target struct {
		ID       int    `json:"id"`
		Hostname string `json:"hostname"`
	}
	c := newTestClient(srv)
	if err := c.Get(testCtx(t), "hosts/1", url.Values{"page": []string{"0"}}, &target); err != nil {
		t.Fatalf("Get unexpected error: %v", err)
	}
	if target.ID != 1 || target.Hostname != "host-A" {
		t.Errorf("decoded target = %+v, want ID=1 Hostname=host-A", target)
	}
}

// Regression test: a short non-JSON 200 body used to panic via bodyBytes[:500]
// when the body was shorter than the snippet length. It must return a decode
// error instead.
func TestGetShortNonJSONBodyReturnsError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		if _, err := w.Write([]byte("<html>not json 25b</html>")); err != nil {
			t.Errorf("write response: %v", err)
		}
	}))
	defer srv.Close()

	var target map[string]interface{}
	c := newTestClient(srv)
	err := c.Get(testCtx(t), "hosts", nil, &target)
	if err == nil {
		t.Fatal("Get with non-JSON body = nil error, want decode error")
	}
	if !strings.Contains(err.Error(), "error decoding JSON response") {
		t.Errorf("error = %q, want a JSON decode error", err)
	}
}

func TestGet404ReturnsFleetAPIErrorWithTruncatedSnippet(t *testing.T) {
	t.Parallel()

	longBody := strings.Repeat("a", 300) + "TAIL-MARKER"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		if _, err := w.Write([]byte(longBody)); err != nil {
			t.Errorf("write response: %v", err)
		}
	}))
	defer srv.Close()

	c := newTestClient(srv)
	err := c.Get(testCtx(t), "hosts/999", nil, nil)
	if err == nil {
		t.Fatal("Get on 404 = nil error, want *FleetAPIError")
	}

	var apiErr *FleetAPIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error type = %T, want *FleetAPIError via errors.As", err)
	}
	if apiErr.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode = %d, want 404", apiErr.StatusCode)
	}
	if !strings.HasSuffix(apiErr.Snippet, "… (truncated)") {
		t.Errorf("Snippet = %q, want it to end with the truncation suffix", apiErr.Snippet)
	}
	if strings.Contains(err.Error(), "TAIL-MARKER") {
		t.Errorf("error string contains the body tail; snippet was not truncated: %q", err)
	}
}

func TestGet429HonorsRetryAfterInline(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "1")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	c := newTestClient(srv)
	start := time.Now()
	err := c.Get(testCtx(t), "hosts", nil, nil)
	elapsed := time.Since(start)

	var apiErr *FleetAPIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error type = %T, want *FleetAPIError", err)
	}
	if apiErr.StatusCode != http.StatusTooManyRequests {
		t.Errorf("StatusCode = %d, want 429", apiErr.StatusCode)
	}
	if apiErr.RetryAfter != time.Second {
		t.Errorf("RetryAfter = %v, want 1s", apiErr.RetryAfter)
	}
	if elapsed < 900*time.Millisecond {
		t.Errorf("Get returned after %v, want >= ~900ms (in-line Retry-After sleep)", elapsed)
	}
}

func TestGet429SleepHonorsContextCancellation(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "60")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	// Cancel shortly after the request completes so Get is inside the
	// Retry-After sleep select when ctx.Done fires.
	ctx, cancel := context.WithCancel(testCtx(t))
	defer cancel()
	timer := time.AfterFunc(100*time.Millisecond, cancel)
	defer timer.Stop()

	c := newTestClient(srv)
	start := time.Now()
	err := c.Get(ctx, "hosts", nil, nil)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("Get on 429 = nil error, want error")
	}
	if elapsed >= 2*time.Second {
		t.Errorf("Get returned after %v, want < 2s (sleep select must honor ctx.Done)", elapsed)
	}
}

func TestGetNilTargetDiscardsBody(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{"hosts": []}`)); err != nil {
			t.Errorf("write response: %v", err)
		}
	}))
	defer srv.Close()

	c := newTestClient(srv)
	if err := c.Get(testCtx(t), "hosts", nil, nil); err != nil {
		t.Fatalf("Get with nil target unexpected error: %v", err)
	}
}

func TestGetNoContentLeavesTargetUntouched(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	var target struct {
		Vulnerability *Vulnerability `json:"vulnerability"`
	}
	if err := newTestClient(srv).Get(testCtx(t), "vulnerabilities/CVE-2024-0001", nil, &target); err != nil {
		t.Fatalf("Get on 204 unexpected error: %v", err)
	}
	if target.Vulnerability != nil {
		t.Errorf("target.Vulnerability = %+v, want nil", target.Vulnerability)
	}
}

func TestMaxRequestsPerSecond(t *testing.T) {
	t.Parallel()

	newConn := func(rps *int) *plugin.Connection {
		conn := &plugin.Connection{Name: "test"}
		conn.SetConfig(fleetdmConfig{ServerURL: ptr("https://fleet.example.com"), APIToken: ptr("t"), MaxRequestsPerSecond: rps})
		return conn
	}

	t.Run("non-positive rejected", func(t *testing.T) {
		t.Parallel()
		for _, v := range []int{0, -5} {
			if _, err := NewFleetDMClient(testCtx(t), newConn(ptr(v))); err == nil {
				t.Errorf("max_requests_per_second = %d accepted, want error", v)
			}
		}
	})

	t.Run("unset means no limiter", func(t *testing.T) {
		t.Parallel()
		c, err := NewFleetDMClient(testCtx(t), newConn(nil))
		if err != nil {
			t.Fatalf("NewFleetDMClient: %v", err)
		}
		if c.limiter != nil {
			t.Error("limiter set without max_requests_per_second")
		}
	})

	t.Run("requests are paced", func(t *testing.T) {
		t.Parallel()
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(`{}`))
		}))
		defer srv.Close()
		c, err := NewFleetDMClient(testCtx(t), newConn(ptr(20)))
		if err != nil {
			t.Fatalf("NewFleetDMClient: %v", err)
		}
		c.BaseURL, c.HTTPClient = srv.URL+"/api/v1/fleet/", srv.Client()

		// A burst of 20 is free; the next 10 need about half a second at 20/s.
		start := time.Now()
		for range 30 {
			if err := c.Get(testCtx(t), "version", nil, nil); err != nil {
				t.Fatalf("Get: %v", err)
			}
		}
		if elapsed := time.Since(start); elapsed < 400*time.Millisecond {
			t.Errorf("30 requests at 20/s took %v, want at least ~500ms", elapsed)
		}
	})
}
