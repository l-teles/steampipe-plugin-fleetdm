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
	Meta       struct {   // FleetDM API for activities includes a meta object for pagination
		HasNextResults     bool   `json:"has_next_results"`
		HasPreviousResults bool   `json:"has_previous_results"`
		NextCursor         string `json:"next_cursor"` // The API docs mention 'after' parameter with a timestamp, but also page/per_page. Let's assume page/per_page for now as per examples.
	} `json:"meta"`
	Count int `json:"count"` // Total count of activities matching the query
}

func tableFleetdmActivity(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "fleetdm_activity",
		Description: "Audit log activities in FleetDM.",
		List: &plugin.ListConfig{
			Hydrate: listActivities,
			KeyColumns: []*plugin.KeyColumn{
				{Name: "type", Require: plugin.Optional},
				// The API also supports 'after' (timestamp string) for cursor-like pagination,
				// and 'order_key', 'order_direction'.
				// For simplicity, we'll use page-based pagination.
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
		},
	}
}

func listActivities(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := NewFleetDMClient(ctx, d.Connection)
	if err != nil {
		plugin.Logger(ctx).Error("fleetdm_activity.listActivities", "connection_error", err)
		return nil, err
	}

	// Pagination for activities: The /api/v1/fleet/activities endpoint supports `page` and `per_page`
	// It also supports `after` (timestamp string like "2022-11-22T17:39:00Z") for cursor-based pagination.
	// We'll use page/per_page for consistency with other tables.
	page := 0
	perPage := 50 // API default is 20, max 100

	// limit := d.QueryContext.Limit
	// if limit != nil && *limit < int64(perPage) {
	// 	// perPage = int(*limit)
	// }

	for {
		params := url.Values{}
		params.Add("page", strconv.Itoa(page))
		params.Add("per_page", strconv.Itoa(perPage))
		params.Add("order_key", "id")
		params.Add("order_direction", "asc") // Most recent (highest ID) last

		if d.EqualsQuals["type"] != nil {
			params.Add("type", d.EqualsQuals["type"].GetStringValue())
		}

		var response ListActivitiesResponse
		_, err := client.Get(ctx, "activities", params, &response)
		if err != nil {
			plugin.Logger(ctx).Error("fleetdm_activity.listActivities", "api_error", err, "page", page, "params", params.Encode())
			return nil, err
		}

		for _, activity := range response.Activities {
			d.StreamListItem(ctx, activity)
			if d.RowsRemaining(ctx) == 0 {
				plugin.Logger(ctx).Debug("fleetdm_activity.listActivities", "limit_reached", true)
				return nil, nil
			}
		}

		// Pagination check
		if !response.Meta.HasNextResults && response.Meta.NextCursor == "" {
			plugin.Logger(ctx).Debug("fleetdm_activity.listActivities", "end_of_results_by_meta", true, "activities_on_page", len(response.Activities), "has_next_meta", response.Meta.HasNextResults)
			break
		}
		if len(response.Activities) < perPage { // Fallback if meta isn't conclusive with page/per_page
			plugin.Logger(ctx).Debug("fleetdm_activity.listActivities", "end_of_results_by_count", true, "activities_on_page", len(response.Activities))
			break
		}

		page++
		plugin.Logger(ctx).Debug("fleetdm_activity.listActivities", "next_page", page)
	}

	return nil, nil
}
