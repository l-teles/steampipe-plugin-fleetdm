package fleetdm

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"

	"github.com/turbot/steampipe-plugin-sdk/v6/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin/transform"
)

// HostSummary is the single-row host count summary shown on Fleet's
// dashboard.
// Refer to: https://fleetdm.com/docs/rest-api/rest-api#get-hosts-summary
type HostSummary struct {
	TotalsHostsCount    int             `json:"totals_hosts_count"`
	OnlineCount         int             `json:"online_count"`
	OfflineCount        int             `json:"offline_count"`
	MIACount            int             `json:"mia_count"`
	Missing30DaysCount  int             `json:"missing_30_days_count"`
	NewCount            int             `json:"new_count"`
	AllLinuxCount       int             `json:"all_linux_count"`
	LowDiskSpaceCount   *int            `json:"low_disk_space_count"` // only with low_disk_space
	DEPAssignErrorCount int             `json:"dep_assign_error_count"`
	Platforms           json.RawMessage `json:"platforms"`
	BuiltinLabels       json.RawMessage `json:"builtin_labels"`
}

func tableFleetdmHostSummary(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "fleetdm_host_summary",
		Description: "Single-row summary of host counts by status and platform, optionally for one team or platform.",
		List: &plugin.ListConfig{
			Hydrate: listHostSummary,
			KeyColumns: []*plugin.KeyColumn{
				{Name: "team_id", Require: plugin.Optional},
				{Name: "platform", Require: plugin.Optional},
				{Name: "low_disk_space", Require: plugin.Optional},
			},
		},
		Columns: []*plugin.Column{
			{Name: "totals_hosts_count", Type: proto.ColumnType_INT, Description: "Total number of hosts."},
			{Name: "online_count", Type: proto.ColumnType_INT, Description: "Number of online hosts."},
			{Name: "offline_count", Type: proto.ColumnType_INT, Description: "Number of offline hosts."},
			{Name: "mia_count", Type: proto.ColumnType_INT, Transform: transform.FromField("MIACount"), Description: "Number of hosts missing in action (offline for 30+ days)."},
			{Name: "missing_30_days_count", Type: proto.ColumnType_INT, Description: "Number of hosts not seen in the last 30 days."},
			{Name: "new_count", Type: proto.ColumnType_INT, Description: "Number of hosts enrolled in the last 24 hours."},
			{Name: "all_linux_count", Type: proto.ColumnType_INT, Description: "Number of Linux hosts."},
			{Name: "low_disk_space_count", Type: proto.ColumnType_INT, Description: "Number of hosts below the low_disk_space threshold (Fleet Premium). Null unless low_disk_space is set."},
			{Name: "dep_assign_error_count", Type: proto.ColumnType_INT, Transform: transform.FromField("DEPAssignErrorCount"), Description: "Number of hosts with an Apple Business assignment error."},
			{Name: "platforms", Type: proto.ColumnType_JSON, Description: "Host counts per platform."},
			{Name: "builtin_labels", Type: proto.ColumnType_JSON, Description: "Built-in labels Fleet uses for the summary."},

			// Filters, echoed so the single row survives the WHERE clause.
			{Name: "team_id", Type: proto.ColumnType_INT, Transform: transform.FromQual("team_id"), Description: "Summarize only this team (fleet); 0 for hosts with no team. Set in WHERE clause."},
			{Name: "platform", Type: proto.ColumnType_STRING, Transform: transform.FromQual("platform"), Description: "Summarize only this platform (e.g. 'darwin', 'windows', 'linux'). Set in WHERE clause."},
			{Name: "low_disk_space", Type: proto.ColumnType_INT, Transform: transform.FromQual("low_disk_space"), Description: "Count hosts with less than this many GB free in low_disk_space_count (1-100, Fleet Premium). Set in WHERE clause."},
		},
	}
}

func listHostSummary(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := getClient(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("fleetdm_host_summary.listHostSummary", "connection_error", err)
		return nil, err
	}
	params := url.Values{}
	if q := d.EqualsQuals["team_id"]; q != nil {
		client.addFleetIDParam(ctx, d, params, strconv.FormatInt(q.GetInt64Value(), 10))
	}
	if q := d.EqualsQuals["platform"]; q != nil {
		params.Add("platform", q.GetStringValue())
	}
	if q := d.EqualsQuals["low_disk_space"]; q != nil {
		params.Add("low_disk_space", strconv.FormatInt(q.GetInt64Value(), 10))
	}
	var summary HostSummary
	if err := client.Get(ctx, "host_summary", params, &summary); err != nil {
		plugin.Logger(ctx).Error("fleetdm_host_summary.listHostSummary", "api_error", err)
		return nil, err
	}
	d.StreamListItem(ctx, summary)
	return nil, nil
}
