package fleetdm

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

// Policy represents a FleetDM policy.
// Refer to: https://fleetdm.com/docs/rest-api/rest-api#policy-object
// And: https://fleetdm.com/docs/rest-api/rest-api#get-a-policy
type Policy struct {
	ID                    uint      `json:"id"`
	CreatedAt             FleetTime `json:"created_at"`
	UpdatedAt             FleetTime `json:"updated_at"`
	Name                  string    `json:"name"`
	Query                 string    `json:"query"` // This is the actual osquery query text
	Description           string    `json:"description"`
	AuthorID              *uint     `json:"author_id"` // Pointer as it can be null
	AuthorName            string    `json:"author_name"`
	AuthorEmail           string    `json:"author_email"`
	TeamID                *uint     `json:"team_id"`                 // Null if global policy
	Resolution            string    `json:"resolution"`              // Instructions for failing hosts
	Platform              string    `json:"platform"`                // e.g., "windows", "linux", "darwin", "" for all
	PassingHostCount      int       `json:"passing_host_count"`      // Number of hosts passing the policy
	FailingHostCount      int       `json:"failing_host_count"`      // Number of hosts failing the policy
	Critical              bool      `json:"critical"`                // Whether the policy is critical (introduced in Fleet 4.41)
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
			// The API uses two different endpoints:
			// - Without team_id: GET /global/policies (supports page, per_page only)
			// - With team_id:    GET /teams/:id/policies, renamed upstream to
			//   /fleets/:id/policies (supports query, merge_inherited, page, per_page)
			KeyColumns: []*plugin.KeyColumn{
				{Name: "filter_search_query", Require: plugin.Optional}, // Maps to API 'query' param (team policies only)
				{Name: "team_id", Require: plugin.Optional},             // Switches to team policies endpoint
				{Name: "fleet_id", Require: plugin.Optional},            // Alias of team_id (newer Fleet naming)
				{Name: "merge_inherited", Require: plugin.Optional},     // Include global policies with team results (Fleet Premium)
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
			{Name: "created_at", Type: proto.ColumnType_TIMESTAMP, Transform: transform.FromField("CreatedAt").Transform(flexibleTimeTransform), Description: "Timestamp when the policy was created."},
			{Name: "updated_at", Type: proto.ColumnType_TIMESTAMP, Transform: transform.FromField("UpdatedAt").Transform(flexibleTimeTransform), Description: "Timestamp when the policy was last updated."},

			// Key column for filtering via API 'query' parameter
			{Name: "filter_search_query", Type: proto.ColumnType_STRING, Transform: transform.FromQual("filter_search_query"), Description: "Search query string to filter policies by name or query text. Only works when team_id is specified. Set in WHERE clause."},
			{Name: "fleet_id", Type: proto.ColumnType_INT, Transform: transform.FromQual("fleet_id"), Description: "Alias of team_id using the newer Fleet naming (teams were renamed to fleets). Set in WHERE clause to list a specific fleet's policies."},
			{Name: "merge_inherited", Type: proto.ColumnType_BOOL, Transform: transform.FromQual("merge_inherited"), Description: "If true, includes global policies in team policy results (Fleet Premium). Requires team_id. Set in WHERE clause."},
		},
	}
}

func listPolicies(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := getClient(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("fleetdm_policy.listPolicies", "connection_error", err)
		return nil, err
	}

	// Resolve the team/fleet scope: team_id and its newer alias fleet_id are
	// interchangeable, but conflicting values are a user error.
	var teamID *int64
	if d.EqualsQuals["team_id"] != nil {
		v := d.EqualsQuals["team_id"].GetInt64Value()
		teamID = &v
	}
	if d.EqualsQuals["fleet_id"] != nil {
		v := d.EqualsQuals["fleet_id"].GetInt64Value()
		if teamID != nil && *teamID != v {
			return nil, fmt.Errorf("team_id (%d) and fleet_id (%d) are aliases and must match when both are set", *teamID, v)
		}
		teamID = &v
	}
	if teamID != nil {
		plugin.Logger(ctx).Debug("fleetdm_policy.listPolicies", "using_team_endpoint_for_id", *teamID)
	}

	err = paginatedList(ctx, d, 100, func(ctx context.Context, page, perPage int) ([]Policy, *ListMeta, error) {
		params := url.Values{}
		params.Add("page", strconv.Itoa(page))
		params.Add("per_page", strconv.Itoa(perPage))

		// query and merge_inherited are only supported on the team policies endpoint
		if teamID != nil {
			if d.EqualsQuals["filter_search_query"] != nil {
				params.Add("query", d.EqualsQuals["filter_search_query"].GetStringValue())
			}
			if d.EqualsQuals["merge_inherited"] != nil {
				params.Add("merge_inherited", strconv.FormatBool(d.EqualsQuals["merge_inherited"].GetBoolValue()))
			}
		}

		var response ListPoliciesResponse
		var getErr error
		if teamID != nil {
			// Team-scoped: teams/:id/policies, renamed to fleets/:id/policies.
			getErr = client.GetCompat(ctx, d, famFleetPolicies, []any{*teamID}, params, &response)
		} else {
			// The global endpoint was not renamed.
			getErr = client.Get(ctx, "global/policies", params, &response)
		}
		if getErr != nil {
			plugin.Logger(ctx).Error("fleetdm_policy.listPolicies", "api_error", getErr, "page", page, "params", params.Encode())
			return nil, nil, getErr
		}
		// The policies endpoints do not document a meta object for pagination.
		return response.Policies, nil, nil
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}
