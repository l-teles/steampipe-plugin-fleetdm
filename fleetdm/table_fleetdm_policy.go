package fleetdm

import (
	"context"
	"net/url"
	"strconv"
	"time"

	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

// Policy represents a FleetDM policy.
// Refer to: https://fleetdm.com/docs/rest-api/rest-api#policy-object
// And: https://fleetdm.com/docs/rest-api/rest-api#get-a-policy
type Policy struct {
	ID                    uint      `json:"id"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
	Name                  string    `json:"name"`
	Query                 string    `json:"query"` // This is the actual osquery query text
	Description           string    `json:"description"`
	AuthorID              *uint     `json:"author_id"` // Pointer as it can be null
	AuthorName            string    `json:"author_name"`
	AuthorEmail           string    `json:"author_email"`
	TeamID                *uint     `json:"team_id"`            // Null if global policy
	Resolution            string    `json:"resolution"`         // Instructions for failing hosts
	Platform              string    `json:"platform"`           // e.g., "windows", "linux", "darwin", "" for all
	PassingHostCount      int       `json:"passing_host_count"` // Number of hosts passing the policy
	FailingHostCount      int       `json:"failing_host_count"` // Number of hosts failing the policy
	Critical              bool      `json:"critical"`           // Whether the policy is critical (introduced in Fleet 4.41)
	CalendarEventsEnabled bool      `json:"calendar_events_enabled"` // Whether calendar events are enabled for this policy (Fleet 4.44+)
}

// ListPoliciesResponse is the structure for the list policies API response.
// Assuming `GET /api/v1/fleet/global/policies` returns `{"policies": [...]}`
type ListPoliciesResponse struct {
	Policies []Policy `json:"policies"`
}

// GetPolicyResponse is the structure for the get policy API response.
// Assuming `GET /api/v1/fleet/global/policies/{id}` returns `{"policy": {...}}`
type GetPolicyResponse struct {
	Policy Policy `json:"policy"`
}

func tableFleetdmPolicy(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "fleetdm_policy",
		Description: "Information about policies in FleetDM.",
		List: &plugin.ListConfig{
			Hydrate: listPolicies,
			// KeyColumns define how the table can be filtered.
			// The 'query' KeyColumn is used for the API's search parameter.
			// The 'team_id' KeyColumn is for filtering by team.
			// Note: If /global/policies endpoint does not support team_id filtering,
			// this KeyColumn might not work as expected for that specific endpoint.
			// A separate table or logic for team-specific policies might be needed if their endpoint is different, to be checked later.
			KeyColumns: []*plugin.KeyColumn{
				{Name: "filter_search_query", Require: plugin.Optional}, // Maps to API 'query' param for text search
				{Name: "team_id", Require: plugin.Optional},             // For filtering by team_id if supported by endpoint
			},
		},
		Columns: []*plugin.Column{
			{Name: "id", Type: proto.ColumnType_INT, Description: "Unique ID of the policy."},
			{Name: "name", Type: proto.ColumnType_STRING, Description: "Name of the policy."},
			{Name: "query_text", Type: proto.ColumnType_STRING, Transform: transform.FromField("Query"), Description: "The osquery query that defines the policy."},
			{Name: "description", Type: proto.ColumnType_STRING, Description: "Description of the policy."},
			{Name: "platform", Type: proto.ColumnType_STRING, Description: "Target platform for the policy (e.g., 'darwin', 'windows', 'linux', or empty for all)."},
			{Name: "team_id", Type: proto.ColumnType_INT, Description: "ID of the team the policy belongs to. Null if it's a global policy."},
			{Name: "passing_host_count", Type: proto.ColumnType_INT, Description: "Number of hosts currently passing this policy."},
			{Name: "failing_host_count", Type: proto.ColumnType_INT, Description: "Number of hosts currently failing this policy."},
			{Name: "resolution", Type: proto.ColumnType_STRING, Description: "Resolution steps or instructions for hosts failing this policy."},
			{Name: "author_id", Type: proto.ColumnType_INT, Description: "ID of the user who created the policy."},
			{Name: "author_name", Type: proto.ColumnType_STRING, Description: "Name of the user who created the policy."},
			{Name: "author_email", Type: proto.ColumnType_STRING, Description: "Email of the user who created the policy."},
			{Name: "critical", Type: proto.ColumnType_BOOL, Description: "Whether the policy is marked as critical."},
			{Name: "calendar_events_enabled", Type: proto.ColumnType_BOOL, Description: "Whether calendar events are enabled for this policy."},
			{Name: "created_at", Type: proto.ColumnType_TIMESTAMP, Description: "Timestamp when the policy was created."},
			{Name: "updated_at", Type: proto.ColumnType_TIMESTAMP, Description: "Timestamp when the policy was last updated."},

			// Key column for filtering via API 'query' parameter
			{Name: "filter_search_query", Type: proto.ColumnType_STRING, Transform: transform.FromQual("filter_search_query"), Description: "Search query string to filter policies by name or query text. Use in WHERE clause."},

			// Connection config (server_url)
			{Name: "server_url", Type: proto.ColumnType_STRING, Hydrate: getServerURL, Transform: transform.FromValue(), Description: "FleetDM server URL from connection config."},
		},
	}
}

func listPolicies(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := NewFleetDMClient(ctx, d.Connection)
	if err != nil {
		plugin.Logger(ctx).Error("fleetdm_policy.listPolicies", "connection_error", err)
		return nil, err
	}

	page := 0
	perPage := 50

	// limit := d.QueryContext.Limit
	// if limit != nil && *limit < int64(perPage) {
	// 	// perPage = int(*limit)
	// }

	for {
		params := url.Values{}
		params.Add("page", strconv.Itoa(page))
		params.Add("per_page", strconv.Itoa(perPage))

		if d.EqualsQuals["filter_search_query"] != nil {
			params.Add("query", d.EqualsQuals["filter_search_query"].GetStringValue())
		}
		if d.EqualsQuals["team_id"] != nil {
			// If /global/policies doesn't support team_id, this will be ignored by the API
			// or might cause an error. This needs to be verified against the actual API behavior for this endpoint.
			params.Add("team_id", strconv.FormatInt(d.EqualsQuals["team_id"].GetInt64Value(), 10))
			plugin.Logger(ctx).Debug("fleetdm_policy.listPolicies", "filtering_by_team_id_on_global_policies_endpoint", d.EqualsQuals["team_id"].GetInt64Value())
		}

		var response ListPoliciesResponse
		// Using "global/policies" as requested
		_, err := client.Get(ctx, "global/policies", params, &response)
		if err != nil {
			plugin.Logger(ctx).Error("fleetdm_policy.listPolicies", "api_error", err, "page", page, "params", params.Encode(), "endpoint", "global/policies")
			return nil, err
		}

		for _, policy := range response.Policies {
			d.StreamListItem(ctx, policy)
			if d.RowsRemaining(ctx) == 0 {
				plugin.Logger(ctx).Debug("fleetdm_policy.listPolicies", "limit_reached", true)
				return nil, nil
			}
		}

		if len(response.Policies) < perPage {
			plugin.Logger(ctx).Debug("fleetdm_policy.listPolicies", "end_of_results", true, "policies_on_page", len(response.Policies))
			break
		}

		page++
		plugin.Logger(ctx).Debug("fleetdm_policy.listPolicies", "next_page", page)
	}

	return nil, nil
}
