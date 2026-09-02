package fleetdm

import (
	"context"

	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
)

// VersionInfo represents the FleetDM server version and build information.
// Refer to: GET /api/v1/fleet/version
type VersionInfo struct {
	Version   string `json:"version"`
	Branch    string `json:"branch"`
	Revision  string `json:"revision"`
	GoVersion string `json:"go_version"`
	// BuildDate is kept as a string: the API returns a non-RFC3339 date
	// (e.g. "2021-03-27"), so it is not parsed into a timestamp.
	BuildDate string `json:"build_date"`
	BuildUser string `json:"build_user"`
}

func tableFleetdmVersion(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "fleetdm_version",
		Description: "FleetDM server version and build information.",
		List: &plugin.ListConfig{
			Hydrate: listVersion,
		},
		Columns: []*plugin.Column{
			{Name: "version", Type: proto.ColumnType_STRING, Description: "Version of the Fleet server."},
			{Name: "branch", Type: proto.ColumnType_STRING, Description: "Git branch the Fleet server was built from."},
			{Name: "revision", Type: proto.ColumnType_STRING, Description: "Git commit revision the Fleet server was built from."},
			{Name: "go_version", Type: proto.ColumnType_STRING, Description: "Go version the Fleet server was built with."},
			{Name: "build_date", Type: proto.ColumnType_STRING, Description: "Date the Fleet server was built (as reported by the API, not necessarily RFC3339)."},
			{Name: "build_user", Type: proto.ColumnType_STRING, Description: "User who built the Fleet server."},
		},
	}
}

// listVersion streams the single version row. GET /version has no pagination
// and no parameters; the response body is the version object itself (no
// envelope).
func listVersion(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := getClient(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("fleetdm_version.listVersion", "connection_error", err)
		return nil, err
	}

	var version VersionInfo
	if err := client.Get(ctx, "version", nil, &version); err != nil {
		plugin.Logger(ctx).Error("fleetdm_version.listVersion", "api_error", err)
		return nil, err
	}

	d.StreamListItem(ctx, version)
	return nil, nil
}
