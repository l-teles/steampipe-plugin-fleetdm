package fleetdm

import (
	"context"
	"encoding/json" // For json.RawMessage
	"net/url"
	"strconv"

	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

// Activity represents an audit log activity in FleetDM.
// Refer to: https://fleetdm.com/docs/rest-api/rest-api#activity-object
type Activity struct {
	ID              uint            `json:"id"`
	CreatedAt       FleetTime       `json:"created_at"`
	ActorFullName   string          `json:"actor_full_name"`
	ActorID         *uint           `json:"actor_id"` // Can be null for system activities
	ActorGravatar   string          `json:"actor_gravatar"`
	Type            string          `json:"type"`                        // e.g., "created_user", "deleted_pack", "live_query"
	Details         json.RawMessage `json:"details"`                     // JSON object, structure varies by type
	ActorEmail      *string         `json:"actor_email,omitempty"`       // Not in main doc, but often present
	ActorType       string          `json:"actor_type,omitempty"`        // e.g. "user", "system" - not in main doc but useful
	HostID          *uint           `json:"host_id,omitempty"`           // If activity relates to a specific host
	HostDisplayName *string         `json:"host_display_name,omitempty"` // If activity relates to a specific host
}

// ListActivitiesResponse for `GET /api/v1/fleet/activities`
// The API returns {"activities": [...]}
type ListActivitiesResponse struct {
	Activities []Activity `json:"activities"`
	Meta       *ListMeta  `json:"meta"`
	Count      int        `json:"count"` // Total count of activities matching the query
}

func tableFleetdmActivity(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "fleetdm_activity",
		Description: "Audit log activities in FleetDM.",
		List: &plugin.ListConfig{
			Hydrate: listActivities,
			KeyColumns: []*plugin.KeyColumn{
				{Name: "type", Require: plugin.Optional},             // Maps to API 'activity_type' param
				{Name: "query", Require: plugin.Optional},            // Search by actor_full_name or actor_email
				{Name: "start_created_at", Require: plugin.Optional}, // Filter activities after this date
				{Name: "end_created_at", Require: plugin.Optional},   // Filter activities before this date
			},
		},
		// No GetConfig for activities as individual activity GET is not standard.
		Columns: []*plugin.Column{
			{Name: "id", Type: proto.ColumnType_INT, Description: "Unique ID of the activity."},
			{Name: "created_at", Type: proto.ColumnType_TIMESTAMP, Transform: transform.FromField("CreatedAt").Transform(flexibleTimeTransform), Description: "Timestamp when the activity occurred."},
			{Name: "actor_full_name", Type: proto.ColumnType_STRING, Description: "Full name of the actor who performed the activity."},
			{Name: "actor_id", Type: proto.ColumnType_INT, Description: "ID of the actor (user). Null for system activities."},
			{Name: "actor_email", Type: proto.ColumnType_STRING, Description: "Email of the actor."},
			{Name: "actor_gravatar", Type: proto.ColumnType_STRING, Description: "Gravatar URL for the actor."},
			{Name: "type", Type: proto.ColumnType_STRING, Description: "Type of activity (e.g., 'created_user', 'ran_live_query')."},
			{Name: "details", Type: proto.ColumnType_JSON, Description: "JSON object containing details specific to the activity type."},
			{Name: "host_id", Type: proto.ColumnType_INT, Description: "ID of the host related to this activity, if applicable."},
			{Name: "host_display_name", Type: proto.ColumnType_STRING, Description: "Display name of the host related to this activity, if applicable."},

			// Query parameters that can be used for filtering (key columns)
			{Name: "query", Type: proto.ColumnType_STRING, Transform: transform.FromQual("query"), Description: "Search query keywords. Searchable fields include actor_full_name and actor_email. Set in WHERE clause."},
			{Name: "start_created_at", Type: proto.ColumnType_STRING, Transform: transform.FromQual("start_created_at"), Description: "Filter activities that happened after this date (e.g., '2024-01-01T00:00:00Z'). Set in WHERE clause."},
			{Name: "end_created_at", Type: proto.ColumnType_STRING, Transform: transform.FromQual("end_created_at"), Description: "Filter activities that happened before this date (e.g., '2024-12-31T23:59:59Z'). Set in WHERE clause."},
		},
	}
}

func listActivities(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := getClient(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("fleetdm_activity.listActivities", "connection_error", err)
		return nil, err
	}

	err = paginatedList(ctx, d, 50, func(ctx context.Context, page, perPage int) ([]Activity, *ListMeta, error) {
		params := url.Values{}
		params.Add("page", strconv.Itoa(page))
		params.Add("per_page", strconv.Itoa(perPage))
		params.Add("order_key", "id")
		params.Add("order_direction", "asc") // Most recent (highest ID) last

		if d.EqualsQuals["type"] != nil {
			params.Add("activity_type", d.EqualsQuals["type"].GetStringValue())
		}
		if d.EqualsQuals["query"] != nil {
			params.Add("query", d.EqualsQuals["query"].GetStringValue())
		}
		if d.EqualsQuals["start_created_at"] != nil {
			params.Add("start_created_at", d.EqualsQuals["start_created_at"].GetStringValue())
		}
		if d.EqualsQuals["end_created_at"] != nil {
			params.Add("end_created_at", d.EqualsQuals["end_created_at"].GetStringValue())
		}

		var response ListActivitiesResponse
		if err := client.Get(ctx, "activities", params, &response); err != nil {
			plugin.Logger(ctx).Error("fleetdm_activity.listActivities", "api_error", err, "page", page, "params", params.Encode())
			return nil, nil, err
		}
		return response.Activities, response.Meta, nil
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}
