package fleetdm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/turbot/steampipe-plugin-sdk/v6/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin/transform"
)

// HostActivity is an activity that targeted a host, past or upcoming.
// Refer to: https://fleetdm.com/docs/rest-api/rest-api#list-host-past-activities
type HostActivity struct {
	Activity
	UUID     string `json:"uuid"` // set on upcoming activities
	HostID   uint   `json:"-"`
	Hostname string `json:"-"`
}

// ListHostActivitiesResponse for `GET /api/v1/fleet/hosts/{id}/activities`
// and `.../activities/upcoming`.
type ListHostActivitiesResponse struct {
	Activities []HostActivity `json:"activities"`
	Meta       *ListMeta      `json:"meta"`
}

func hostActivityColumns() []*plugin.Column {
	return []*plugin.Column{
		{Name: "host_id", Type: proto.ColumnType_INT, Description: "ID of the host."},
		{Name: "hostname", Type: proto.ColumnType_STRING, Description: "Hostname of the host. Empty when filtering by host_id."},
		{Name: "id", Type: proto.ColumnType_INT, Transform: transform.FromField("ID").NullIfZero(), Description: "ID of the activity (past activities); null for upcoming activities."},
		{Name: "uuid", Type: proto.ColumnType_STRING, Transform: transform.FromField("UUID"), Description: "UUID of the activity (upcoming activities)."},
		{Name: "type", Type: proto.ColumnType_STRING, Description: "Type of activity, e.g. 'ran_script', 'installed_software', 'locked_host'."},
		{Name: "created_at", Type: proto.ColumnType_TIMESTAMP, Transform: transform.FromField("CreatedAt").Transform(flexibleTimeTransform), Description: "When the activity was created."},
		{Name: "actor_full_name", Type: proto.ColumnType_STRING, Description: "Full name of the user who triggered the activity."},
		{Name: "actor_id", Type: proto.ColumnType_INT, Description: "ID of the user who triggered the activity."},
		{Name: "actor_email", Type: proto.ColumnType_STRING, Description: "Email of the user who triggered the activity."},
		{Name: "actor_api_only", Type: proto.ColumnType_BOOL, Description: "Whether the actor is an API-only user."},
		{Name: "fleet_initiated", Type: proto.ColumnType_BOOL, Description: "Whether Fleet itself triggered the activity."},
		{Name: "details", Type: proto.ColumnType_JSON, Description: "Activity details; the shape depends on the type."},
		// Plain columns for the IDs that link to result tables: Steampipe only
		// passes a required qual through a join from a column, not from a
		// JSON expression.
		{Name: "script_execution_id", Type: proto.ColumnType_STRING, Transform: transform.FromField("Details").Transform(detailString("script_execution_id")), Description: "Script execution ID (script activities); join to fleetdm_script_result."},
		{Name: "install_uuid", Type: proto.ColumnType_STRING, Transform: transform.FromField("Details").Transform(detailString("install_uuid")), Description: "Install ID (software install activities); join to fleetdm_software_install_result."},
		{Name: "command_uuid", Type: proto.ColumnType_STRING, Transform: transform.FromField("Details").Transform(detailString("command_uuid")), Description: "MDM command ID (MDM command activities); join to fleetdm_mdm_command_result."},
	}
}

// detailString reads one string field from an activity's details JSON.
func detailString(key string) transform.TransformFunc {
	return func(_ context.Context, d *transform.TransformData) (interface{}, error) {
		raw, ok := d.Value.(json.RawMessage)
		if !ok || len(raw) == 0 {
			return nil, nil
		}
		var m map[string]interface{}
		if err := json.Unmarshal(raw, &m); err != nil {
			return nil, nil
		}
		if v, ok := m[key].(string); ok && v != "" {
			return v, nil
		}
		return nil, nil
	}
}

func tableFleetdmHostActivity(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "fleetdm_host_activity",
		Description: "Past activities on each host (scripts run, software installed, locks, ...). Without a host_id filter every host is queried.",
		List: &plugin.ListConfig{
			Hydrate: listHostActivities,
			KeyColumns: []*plugin.KeyColumn{
				{Name: "host_id", Require: plugin.Optional},
			},
		},
		Columns: hostActivityColumns(),
	}
}

func tableFleetdmHostUpcomingActivity(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "fleetdm_host_upcoming_activity",
		Description: "Activities queued for each host that have not run yet. Without a host_id filter every host is queried.",
		List: &plugin.ListConfig{
			Hydrate: listHostUpcomingActivities,
			KeyColumns: []*plugin.KeyColumn{
				{Name: "host_id", Require: plugin.Optional},
			},
		},
		Columns: hostActivityColumns(),
	}
}

func listHostActivities(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	return nil, streamHostActivities(ctx, d, "fleetdm_host_activity", "hosts/%d/activities")
}

func listHostUpcomingActivities(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	return nil, streamHostActivities(ctx, d, "fleetdm_host_upcoming_activity", "hosts/%d/activities/upcoming")
}

func streamHostActivities(ctx context.Context, d *plugin.QueryData, table, pathFormat string) error {
	client, err := getClient(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error(table, "connection_error", err)
		return err
	}
	rowCap := perHostRowCap(d) // stop paging a host once LIMIT rows are in
	err = streamPerHost(ctx, d, client, func(ctx context.Context, host hostRef) ([]HostActivity, error) {
		var rows []HostActivity
		err := paginateCore(ctx, 100, func(ctx context.Context, page, perPage int) ([]HostActivity, *ListMeta, error) {
			params := url.Values{"page": {strconv.Itoa(page)}, "per_page": {strconv.Itoa(perPage)}}
			var response ListHostActivitiesResponse
			if err := client.Get(ctx, fmt.Sprintf(pathFormat, host.ID), params, &response); err != nil {
				return nil, nil, err
			}
			return response.Activities, response.Meta, nil
		}, func(a HostActivity) bool {
			a.HostID, a.Hostname = host.ID, host.Hostname
			rows = append(rows, a)
			return rowCap == 0 || len(rows) < rowCap
		})
		return rows, err
	})
	if err != nil {
		plugin.Logger(ctx).Error(table, "api_error", err)
	}
	return err
}
