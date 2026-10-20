package fleetdm

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/turbot/steampipe-plugin-sdk/v6/connection"
	"github.com/turbot/steampipe-plugin-sdk/v6/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin"
)

func lookupQueryData(t *testing.T, serverURL string, quals plugin.KeyColumnEqualsQualMap) *plugin.QueryData {
	t.Helper()
	cache, err := connection.NewConnectionCache("test", 1000)
	if err != nil {
		t.Fatal(err)
	}
	return &plugin.QueryData{
		Connection:      testConnection(ptr(serverURL), ptr("token"), nil),
		ConnectionCache: cache,
		EqualsQuals:     quals,
	}
}

func TestVulnerabilityLookupReusesDetails(t *testing.T) {
	var cveRequests atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/fleet/vulnerabilities/CVE-2024-0001" {
			cveRequests.Add(1)
			_, _ = w.Write([]byte(`{"vulnerability":{"cve":"CVE-2024-0001"},"software":[{"name":"app"}],"os_versions":[]}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	ctx := testCtx(t)
	d := lookupQueryData(t, srv.URL, plugin.KeyColumnEqualsQualMap{
		"cve": &proto.QualValue{Value: &proto.QualValue_StringValue{StringValue: "CVE-2024-0001"}},
	})
	row, err := getVulnerability(ctx, d, nil)
	if err != nil || row == nil {
		t.Fatalf("getVulnerability = %v, %v", row, err)
	}
	details, err := getVulnerabilityDetails(ctx, d, &plugin.HydrateData{Item: row})
	if err != nil {
		t.Fatal(err)
	}
	if sw, _ := details.(VulnerabilityDetails).Software.([]any); len(sw) != 1 {
		t.Errorf("affected software = %v", details)
	}
	if n := cveRequests.Load(); n != 1 {
		t.Errorf("made %d requests for the CVE, want 1", n)
	}
}

func TestVulnerabilityDetailsUseTeamFilter(t *testing.T) {
	var gotFleetID string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/fleet/version":
			_, _ = w.Write([]byte(`{"version":"4.93.0"}`))
		case "/api/v1/fleet/vulnerabilities/CVE-2024-0001":
			gotFleetID = r.URL.Query().Get("fleet_id")
			_, _ = w.Write([]byte(`{"software":[],"os_versions":[]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	d := lookupQueryData(t, srv.URL, plugin.KeyColumnEqualsQualMap{
		"team_id": &proto.QualValue{Value: &proto.QualValue_Int64Value{Int64Value: 5}},
	})
	item := Vulnerability{CVE: "CVE-2024-0001"} // a list row, without details
	if _, err := getVulnerabilityDetails(testCtx(t), d, &plugin.HydrateData{Item: item}); err != nil {
		t.Fatal(err)
	}
	if gotFleetID != "5" {
		t.Errorf("details request fleet_id = %q, want 5", gotFleetID)
	}
}

func TestVulnerabilityLookupRejectsNonCVE(t *testing.T) {
	var requests atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
	}))
	defer srv.Close()

	d := lookupQueryData(t, srv.URL, plugin.KeyColumnEqualsQualMap{
		"cve": &proto.QualValue{Value: &proto.QualValue_StringValue{StringValue: "../../users"}},
	})
	row, err := getVulnerability(testCtx(t), d, nil)
	if err != nil || row != nil {
		t.Errorf("getVulnerability = %v, %v; want no row", row, err)
	}
	if n := requests.Load(); n != 0 {
		t.Errorf("made %d requests for a non-CVE qual", n)
	}
}

func TestResultLookupsRejectNonUUID(t *testing.T) {
	var requests atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
	}))
	defer srv.Close()

	for _, id := range []string{"../../users", "..", "x?per_page=1", ""} {
		qual := &proto.QualValue{Value: &proto.QualValue_StringValue{StringValue: id}}
		if _, err := listScriptResult(testCtx(t), lookupQueryData(t, srv.URL, plugin.KeyColumnEqualsQualMap{"execution_id": qual}), nil); err != nil {
			t.Errorf("script result %q: %v", id, err)
		}
		if _, err := listSoftwareInstallResult(testCtx(t), lookupQueryData(t, srv.URL, plugin.KeyColumnEqualsQualMap{"install_uuid": qual}), nil); err != nil {
			t.Errorf("install result %q: %v", id, err)
		}
	}
	if n := requests.Load(); n != 0 {
		t.Errorf("made %d requests for non-UUID quals", n)
	}
}
