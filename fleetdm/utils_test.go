package fleetdm

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/context_key"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

// ptr returns a pointer to v; a tiny helper for building config literals.
func ptr[T any](v T) *T { return &v }

// testCtx returns a context carrying a null hclog logger, since
// plugin.Logger(ctx) type-asserts the context value and panics on a bare
// context.
func testCtx(t *testing.T) context.Context {
	t.Helper()
	return context.WithValue(context.Background(), context_key.Logger, hclog.NewNullLogger())
}

// testConnection builds a *plugin.Connection with the given config values.
func testConnection(serverURL, apiToken *string, requestTimeout *int) *plugin.Connection {
	return &plugin.Connection{
		Name: "test",
		Config: fleetdmConfig{
			ServerURL:      serverURL,
			APIToken:       apiToken,
			RequestTimeout: requestTimeout,
		},
	}
}

func TestFleetTimeUnmarshalJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		wantZero bool
		want     time.Time
		wantErr  bool
	}{
		{name: "null literal", input: `null`, wantZero: true},
		{name: "empty string", input: `""`, wantZero: true},
		{
			name:  "valid RFC3339",
			input: `"2026-01-02T15:04:05Z"`,
			want:  time.Date(2026, 1, 2, 15, 4, 5, 0, time.UTC),
		},
		{name: "garbage", input: `"not-a-time"`, wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var ft FleetTime
			err := ft.UnmarshalJSON([]byte(tc.input))
			if tc.wantErr {
				if err == nil {
					t.Fatalf("UnmarshalJSON(%s) = nil error, want error", tc.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("UnmarshalJSON(%s) unexpected error: %v", tc.input, err)
			}
			if tc.wantZero {
				if !ft.IsZero() {
					t.Fatalf("UnmarshalJSON(%s) = %v, want zero time", tc.input, ft.Time)
				}
				return
			}
			if !ft.Equal(tc.want) {
				t.Fatalf("UnmarshalJSON(%s) = %v, want %v", tc.input, ft.Time, tc.want)
			}
		})
	}
}

func TestFleetTimeMarshalJSON(t *testing.T) {
	t.Parallel()

	zero, err := FleetTime{}.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON(zero) unexpected error: %v", err)
	}
	if string(zero) != "null" {
		t.Errorf("MarshalJSON(zero) = %s, want null", zero)
	}

	ft := FleetTime{Time: time.Date(2026, 1, 2, 15, 4, 5, 0, time.UTC)}
	got, err := ft.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON(non-zero) unexpected error: %v", err)
	}
	if string(got) != `"2026-01-02T15:04:05Z"` {
		t.Errorf("MarshalJSON(non-zero) = %s, want %q", got, `"2026-01-02T15:04:05Z"`)
	}
}

func TestFlexibleTimeTransform(t *testing.T) {
	t.Parallel()

	when := time.Date(2026, 1, 2, 15, 4, 5, 0, time.UTC)

	tests := []struct {
		name    string
		value   interface{}
		want    interface{}
		wantErr bool
	}{
		{name: "nil value", value: nil, want: nil},
		{name: "FleetTime zero", value: FleetTime{}, want: nil},
		{name: "FleetTime non-zero", value: FleetTime{Time: when}, want: when},
		{name: "*FleetTime nil", value: (*FleetTime)(nil), want: nil},
		{name: "*FleetTime zero", value: &FleetTime{}, want: nil},
		{name: "*FleetTime non-zero", value: &FleetTime{Time: when}, want: when},
		{name: "time.Time zero", value: time.Time{}, want: nil},
		{name: "time.Time non-zero", value: when, want: when},
		{name: "*time.Time nil", value: (*time.Time)(nil), want: nil},
		{name: "*time.Time non-zero", value: &when, want: when},
		{name: "unexpected type", value: 42, wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := flexibleTimeTransform(context.Background(), &transform.TransformData{Value: tc.value})
			if tc.wantErr {
				if err == nil {
					t.Fatalf("flexibleTimeTransform(%#v) = nil error, want error", tc.value)
				}
				return
			}
			if err != nil {
				t.Fatalf("flexibleTimeTransform(%#v) unexpected error: %v", tc.value, err)
			}
			if tc.want == nil {
				if got != nil {
					t.Fatalf("flexibleTimeTransform(%#v) = %v, want nil", tc.value, got)
				}
				return
			}
			gotTime, ok := got.(time.Time)
			if !ok {
				t.Fatalf("flexibleTimeTransform(%#v) returned %T, want time.Time", tc.value, got)
			}
			if !gotTime.Equal(tc.want.(time.Time)) {
				t.Fatalf("flexibleTimeTransform(%#v) = %v, want %v", tc.value, gotTime, tc.want)
			}
		})
	}
}

func TestTruncateBody(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		n     int
		want  string
	}{
		{name: "shorter than n", input: "short", n: 10, want: "short"},
		{name: "exactly n", input: "0123456789", n: 10, want: "0123456789"},
		{name: "longer than n", input: "0123456789X", n: 10, want: "0123456789… (truncated)"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := truncateBody([]byte(tc.input), tc.n); got != tc.want {
				t.Errorf("truncateBody(%q, %d) = %q, want %q", tc.input, tc.n, got, tc.want)
			}
		})
	}
}

func TestParseRetryAfter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  time.Duration
	}{
		{name: "empty", input: "", want: 0},
		{name: "delay seconds", input: "5", want: 5 * time.Second},
		{name: "zero seconds", input: "0", want: 0},
		{name: "negative seconds", input: "-3", want: 0},
		{name: "past HTTP-date", input: time.Now().Add(-30 * time.Second).UTC().Format(http.TimeFormat), want: 0},
		{name: "garbage", input: "not-a-duration", want: 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := parseRetryAfter(tc.input); got != tc.want {
				t.Errorf("parseRetryAfter(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}

	t.Run("future HTTP-date", func(t *testing.T) {
		t.Parallel()
		v := time.Now().Add(30 * time.Second).UTC().Format(http.TimeFormat)
		got := parseRetryAfter(v)
		if got <= 0 || got > 31*time.Second {
			t.Errorf("parseRetryAfter(%q) = %v, want a positive duration of at most ~30s", v, got)
		}
	})
}

func TestNewFleetDMClientURLNormalization(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		serverURL  string
		wantBase   string
		wantErrSub string // non-empty means an error containing this substring is expected
	}{
		{
			name:      "bare https host",
			serverURL: "https://fleet.example.com",
			wantBase:  "https://fleet.example.com/api/v1/fleet/",
		},
		{
			name:      "trailing slash",
			serverURL: "https://fleet.example.com/",
			wantBase:  "https://fleet.example.com/api/v1/fleet/",
		},
		{
			name:      "api suffix",
			serverURL: "https://fleet.example.com/api",
			wantBase:  "https://fleet.example.com/api/v1/fleet/",
		},
		{
			name:      "api/v1 suffix",
			serverURL: "https://fleet.example.com/api/v1",
			wantBase:  "https://fleet.example.com/api/v1/fleet/",
		},
		{
			name:      "api/v1/fleet suffix",
			serverURL: "https://fleet.example.com/api/v1/fleet",
			wantBase:  "https://fleet.example.com/api/v1/fleet/",
		},
		{
			name:      "http localhost allowed",
			serverURL: "http://localhost:8080",
			wantBase:  "http://localhost:8080/api/v1/fleet/",
		},
		{
			name:      "http loopback IP allowed",
			serverURL: "http://127.0.0.1:8080",
			wantBase:  "http://127.0.0.1:8080/api/v1/fleet/",
		},
		{
			name:       "http non-loopback rejected",
			serverURL:  "http://fleet.example.com",
			wantErrSub: "cleartext",
		},
		{
			name:       "unsupported scheme",
			serverURL:  "ftp://x",
			wantErrSub: "unsupported scheme",
		},
		{
			name:       "missing scheme",
			serverURL:  "fleet.example.com",
			wantErrSub: "absolute URL",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			conn := testConnection(ptr(tc.serverURL), ptr("t"), nil)
			client, err := NewFleetDMClient(testCtx(t), conn)
			if tc.wantErrSub != "" {
				if err == nil {
					t.Fatalf("NewFleetDMClient(%q) = nil error, want error containing %q", tc.serverURL, tc.wantErrSub)
				}
				if !strings.Contains(err.Error(), tc.wantErrSub) {
					t.Fatalf("NewFleetDMClient(%q) error = %q, want it to contain %q", tc.serverURL, err, tc.wantErrSub)
				}
				return
			}
			if err != nil {
				t.Fatalf("NewFleetDMClient(%q) unexpected error: %v", tc.serverURL, err)
			}
			if client.BaseURL != tc.wantBase {
				t.Errorf("NewFleetDMClient(%q) BaseURL = %q, want %q", tc.serverURL, client.BaseURL, tc.wantBase)
			}
		})
	}
}

// Env-dependent cases use t.Setenv, which is incompatible with t.Parallel.
func TestNewFleetDMClientEnvFallback(t *testing.T) {
	t.Run("empty config and empty env", func(t *testing.T) {
		t.Setenv("FLEETDM_URL", "")
		t.Setenv("FLEETDM_API_TOKEN", "")
		conn := &plugin.Connection{Name: "test", Config: fleetdmConfig{}}
		_, err := NewFleetDMClient(testCtx(t), conn)
		if err == nil {
			t.Fatal("NewFleetDMClient with no config and no env = nil error, want error")
		}
		if !strings.Contains(err.Error(), "FLEETDM_URL") {
			t.Errorf("error = %q, want it to mention FLEETDM_URL", err)
		}
	})

	t.Run("env used when config nil", func(t *testing.T) {
		t.Setenv("FLEETDM_URL", "https://env.example.com")
		t.Setenv("FLEETDM_API_TOKEN", "env-token")
		conn := &plugin.Connection{Name: "test", Config: nil}
		client, err := NewFleetDMClient(testCtx(t), conn)
		if err != nil {
			t.Fatalf("NewFleetDMClient from env unexpected error: %v", err)
		}
		if client.BaseURL != "https://env.example.com/api/v1/fleet/" {
			t.Errorf("BaseURL = %q, want %q", client.BaseURL, "https://env.example.com/api/v1/fleet/")
		}
		if client.APIToken != "env-token" {
			t.Errorf("APIToken = %q, want %q", client.APIToken, "env-token")
		}
	})

	t.Run("config wins over env", func(t *testing.T) {
		t.Setenv("FLEETDM_URL", "https://env.example.com")
		t.Setenv("FLEETDM_API_TOKEN", "env-token")
		conn := testConnection(ptr("https://spc.example.com"), ptr("spc-token"), nil)
		client, err := NewFleetDMClient(testCtx(t), conn)
		if err != nil {
			t.Fatalf("NewFleetDMClient unexpected error: %v", err)
		}
		if client.BaseURL != "https://spc.example.com/api/v1/fleet/" {
			t.Errorf("BaseURL = %q, want config value to win over env", client.BaseURL)
		}
		if client.APIToken != "spc-token" {
			t.Errorf("APIToken = %q, want config value to win over env", client.APIToken)
		}
	})
}

func TestNewFleetDMClientRequestTimeout(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		timeout *int
		want    time.Duration
		wantErr bool
	}{
		{name: "default when unset", timeout: nil, want: 30 * time.Second},
		{name: "explicit 5 seconds", timeout: ptr(5), want: 5 * time.Second},
		{name: "zero rejected", timeout: ptr(0), wantErr: true},
		{name: "negative rejected", timeout: ptr(-1), wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			conn := testConnection(ptr("https://fleet.example.com"), ptr("t"), tc.timeout)
			client, err := NewFleetDMClient(testCtx(t), conn)
			if tc.wantErr {
				if err == nil {
					t.Fatal("NewFleetDMClient = nil error, want error for non-positive request_timeout")
				}
				return
			}
			if err != nil {
				t.Fatalf("NewFleetDMClient unexpected error: %v", err)
			}
			if client.HTTPClient.Timeout != tc.want {
				t.Errorf("HTTPClient.Timeout = %v, want %v", client.HTTPClient.Timeout, tc.want)
			}
		})
	}
}
