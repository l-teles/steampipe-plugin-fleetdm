package fleetdm

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/turbot/steampipe-plugin-sdk/v6/connection"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin"
)

func TestVersionBefore(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		status  int
		version string
		want    bool
	}{
		{"older", http.StatusOK, "4.73.0", true},
		{"same", http.StatusOK, "4.74.0", false},
		{"newer", http.StatusOK, "4.93.0", false},
		{"unparseable", http.StatusOK, "dev", false},
		{"lookup fails", http.StatusInternalServerError, "", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var lookups atomic.Int32
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				lookups.Add(1)
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(`{"version":"` + tc.version + `"}`))
			}))
			defer srv.Close()
			cache, err := connection.NewConnectionCache("test", 1000)
			if err != nil {
				t.Fatal(err)
			}
			d := &plugin.QueryData{ConnectionCache: cache}
			c := &FleetDMClient{BaseURL: srv.URL + "/api/v1/fleet/", APIToken: "t", HTTPClient: srv.Client()}
			if got := c.versionBefore(testCtx(t), d, 4, 74); got != tc.want {
				t.Errorf("versionBefore(4.74) on %q = %v, want %v", tc.version, got, tc.want)
			}
		})
	}
}
