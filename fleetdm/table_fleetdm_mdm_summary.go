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

// MDMEnrollmentSummary is the single-row MDM enrollment summary.
// Refer to: https://fleetdm.com/docs/rest-api/rest-api#get-hosts-summary-of-mobile-device-management-mdm-information
type MDMEnrollmentSummary struct {
	CountsUpdatedAt  FleetTime `json:"counts_updated_at"`
	EnrollmentStatus struct {
		EnrolledManualHostsCount    int `json:"enrolled_manual_hosts_count"`
		EnrolledAutomatedHostsCount int `json:"enrolled_automated_hosts_count"`
		EnrolledPersonalHostsCount  int `json:"enrolled_personal_hosts_count"`
		PendingHostsCount           int `json:"pending_hosts_count"`
		UnenrolledHostsCount        int `json:"unenrolled_hosts_count"`
		HostsCount                  int `json:"hosts_count"`
	} `json:"mobile_device_management_enrollment_status"`
	Solutions json.RawMessage `json:"mobile_device_management_solution"`
}

func tableFleetdmMDMSummary(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "fleetdm_mdm_summary",
		Description: "Single-row summary of MDM enrollment across hosts, optionally for one team or platform.",
		List: &plugin.ListConfig{
			Hydrate: listMDMSummary,
			KeyColumns: []*plugin.KeyColumn{
				{Name: "team_id", Require: plugin.Optional},
				{Name: "platform", Require: plugin.Optional},
			},
		},
		Columns: []*plugin.Column{
			{Name: "hosts_count", Type: proto.ColumnType_INT, Transform: transform.FromField("EnrollmentStatus.HostsCount"), Description: "Number of hosts counted."},
			{Name: "enrolled_manual_hosts_count", Type: proto.ColumnType_INT, Transform: transform.FromField("EnrollmentStatus.EnrolledManualHostsCount"), Description: "Hosts enrolled in MDM manually."},
			{Name: "enrolled_automated_hosts_count", Type: proto.ColumnType_INT, Transform: transform.FromField("EnrollmentStatus.EnrolledAutomatedHostsCount"), Description: "Hosts enrolled in MDM automatically (ADE/Autopilot)."},
			{Name: "enrolled_personal_hosts_count", Type: proto.ColumnType_INT, Transform: transform.FromField("EnrollmentStatus.EnrolledPersonalHostsCount"), Description: "Personally owned hosts enrolled in MDM."},
			{Name: "pending_hosts_count", Type: proto.ColumnType_INT, Transform: transform.FromField("EnrollmentStatus.PendingHostsCount"), Description: "Hosts with a pending MDM enrollment."},
			{Name: "unenrolled_hosts_count", Type: proto.ColumnType_INT, Transform: transform.FromField("EnrollmentStatus.UnenrolledHostsCount"), Description: "Hosts not enrolled in MDM."},
			{Name: "mdm_solutions", Type: proto.ColumnType_JSON, Transform: transform.FromField("Solutions"), Description: "Host counts per MDM solution (server URL)."},
			{Name: "counts_updated_at", Type: proto.ColumnType_TIMESTAMP, Transform: transform.FromField("CountsUpdatedAt").Transform(flexibleTimeTransform), Description: "When the counts were last computed."},

			// Filters, echoed so the single row survives the WHERE clause.
			{Name: "team_id", Type: proto.ColumnType_INT, Transform: transform.FromQual("team_id"), Description: "Summarize only this team (fleet); 0 for hosts with no team. Set in WHERE clause."},
			{Name: "platform", Type: proto.ColumnType_STRING, Transform: transform.FromQual("platform"), Description: "Summarize only this platform ('darwin', 'windows', 'ios', 'ipados', 'android'). Set in WHERE clause."},
		},
	}
}

func listMDMSummary(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := getClient(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("fleetdm_mdm_summary.listMDMSummary", "connection_error", err)
		return nil, err
	}
	params := url.Values{}
	if q := d.EqualsQuals["team_id"]; q != nil {
		client.addFleetIDParam(ctx, d, params, strconv.FormatInt(q.GetInt64Value(), 10))
	}
	if q := d.EqualsQuals["platform"]; q != nil {
		params.Add("platform", q.GetStringValue())
	}
	var summary MDMEnrollmentSummary
	if err := client.Get(ctx, "hosts/summary/mdm", params, &summary); err != nil {
		if handleUnavailableFeature(ctx, d, "fleetdm_mdm_summary", err) {
			return nil, nil
		}
		plugin.Logger(ctx).Error("fleetdm_mdm_summary.listMDMSummary", "api_error", err)
		return nil, err
	}
	d.StreamListItem(ctx, summary)
	return nil, nil
}
