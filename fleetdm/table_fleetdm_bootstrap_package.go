package fleetdm

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"github.com/turbot/steampipe-plugin-sdk/v6/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin/transform"
)

// BootstrapPackage is the macOS package installed during automatic (ADE)
// enrollment for a team. The download token in the API response is not kept.
// Refer to: https://fleetdm.com/docs/rest-api/rest-api#get-metadata-about-a-bootstrap-package
type BootstrapPackage struct {
	Name      string    `json:"name"`
	Sha256    string    `json:"sha256"` // base64
	CreatedAt FleetTime `json:"created_at"`
	TeamID    uint      `json:"-"`
}

func tableFleetdmBootstrapPackage(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "fleetdm_bootstrap_package",
		Description: "macOS bootstrap package installed during automatic enrollment, per team (Fleet Premium, Apple MDM). Teams without one have no row.",
		List: &plugin.ListConfig{
			Hydrate: listBootstrapPackages,
			KeyColumns: []*plugin.KeyColumn{
				{Name: "team_id", Require: plugin.Optional}, // 0 for no team
			},
		},
		Columns: []*plugin.Column{
			{Name: "team_id", Type: proto.ColumnType_INT, Description: "ID of the team (fleet); 0 for hosts with no team."},
			{Name: "name", Type: proto.ColumnType_STRING, Description: "File name of the package."},
			{Name: "sha256", Type: proto.ColumnType_STRING, Transform: transform.FromField("Sha256").Transform(base64ToHex), Description: "SHA-256 of the package (hex)."},
			{Name: "created_at", Type: proto.ColumnType_TIMESTAMP, Transform: transform.FromField("CreatedAt").Transform(flexibleTimeTransform), Description: "When the package was uploaded."},
		},
	}
}

// base64ToHex renders a base64 digest as hex, as checksums are usually shown.
func base64ToHex(_ context.Context, d *transform.TransformData) (interface{}, error) {
	s, ok := d.Value.(string)
	if !ok || s == "" {
		return nil, nil
	}
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return s, nil
	}
	return hex.EncodeToString(b), nil
}

func listBootstrapPackages(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := getClient(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("fleetdm_bootstrap_package.listBootstrapPackages", "connection_error", err)
		return nil, err
	}
	scopes, err := teamScopes(ctx, d, client, "fleetdm_bootstrap_package")
	if err != nil {
		return nil, err
	}
	for _, scope := range scopes {
		var pkg BootstrapPackage
		// Packages for earlier teams may have streamed, so retry here rather
		// than fail the whole query on a transient error.
		err := retryTransient(ctx, func() error {
			return client.Get(ctx, fmt.Sprintf("bootstrap/%s/metadata", scopeParam(scope)), nil, &pkg)
		})
		if err != nil {
			// 404: this team has no bootstrap package.
			if is404(err) {
				continue
			}
			if handleUnavailableFeature(ctx, d, "fleetdm_bootstrap_package", err) {
				return nil, nil
			}
			if skipScope(ctx, d, "fleetdm_bootstrap_package", scope, err) {
				continue
			}
			plugin.Logger(ctx).Error("fleetdm_bootstrap_package.listBootstrapPackages", "api_error", err)
			return nil, err
		}
		if scope != nil {
			pkg.TeamID = *scope
		}
		d.StreamListItem(ctx, pkg)
	}
	return nil, nil
}
