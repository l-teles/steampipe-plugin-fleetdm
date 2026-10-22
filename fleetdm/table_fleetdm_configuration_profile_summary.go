package fleetdm

import (
	"context"
	"net/url"
	"strconv"

	"github.com/turbot/steampipe-plugin-sdk/v6/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin/transform"
)

// ConfigurationProfileSummary counts hosts by the delivery status of their
// configuration profiles.
// Refer to: https://fleetdm.com/docs/rest-api/rest-api#get-os-settings-status-summary
type ConfigurationProfileSummary struct {
	Verified  int `json:"verified"`
	Verifying int `json:"verifying"`
	Pending   int `json:"pending"`
	Failed    int `json:"failed"`
}

func tableFleetdmConfigurationProfileSummary(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "fleetdm_configuration_profile_summary",
		Description: "Single-row count of hosts by configuration profile (OS settings) delivery status, optionally for one team (MDM).",
		List: &plugin.ListConfig{
			Hydrate: listConfigurationProfileSummary,
			KeyColumns: []*plugin.KeyColumn{
				{Name: "team_id", Require: plugin.Optional},
			},
		},
		Columns: []*plugin.Column{
			{Name: "verified", Type: proto.ColumnType_INT, Description: "Hosts whose profiles are installed and verified."},
			{Name: "verifying", Type: proto.ColumnType_INT, Description: "Hosts whose profiles are installed and awaiting verification."},
			{Name: "pending", Type: proto.ColumnType_INT, Description: "Hosts with profiles still to be delivered."},
			{Name: "failed", Type: proto.ColumnType_INT, Description: "Hosts where at least one profile failed."},
			{Name: "team_id", Type: proto.ColumnType_INT, Transform: transform.FromQual("team_id"), Description: "Summarize only this team (fleet); 0 for hosts with no team. Set in WHERE clause."},
		},
	}
}

func listConfigurationProfileSummary(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := getClient(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("fleetdm_configuration_profile_summary.listConfigurationProfileSummary", "connection_error", err)
		return nil, err
	}
	params := url.Values{}
	if q := d.EqualsQuals["team_id"]; q != nil {
		client.addFleetIDParam(ctx, d, params, strconv.FormatInt(q.GetInt64Value(), 10))
	}
	var summary ConfigurationProfileSummary
	if err := client.Get(ctx, "configuration_profiles/summary", params, &summary); err != nil {
		if handleUnavailableFeature(ctx, d, "fleetdm_configuration_profile_summary", err) {
			return nil, nil
		}
		plugin.Logger(ctx).Error("fleetdm_configuration_profile_summary.listConfigurationProfileSummary", "api_error", err)
		return nil, err
	}
	d.StreamListItem(ctx, summary)
	return nil, nil
}
