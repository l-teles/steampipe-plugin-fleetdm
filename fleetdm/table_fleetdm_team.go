package fleetdm

import (
	"context"
	"encoding/json" // Added import for json.RawMessage
	"net/url"
	"strconv"

	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

// Team represents a FleetDM team.
// Refer to: https://fleetdm.com/docs/rest-api/rest-api#team-object
type Team struct {
	ID           uint             `json:"id"`
	CreatedAt    FleetTime        `json:"created_at"`
	Name         string           `json:"name"`
	Description  string           `json:"description"`
	UserCount    int              `json:"user_count"`    // Calculated field, number of users in the team
	HostCount    int              `json:"host_count"`    // Calculated field, number of hosts in the team
	Secrets      []TeamSecret     `json:"secrets"`       // Agent enrollment secrets
	Users        []TeamUser       `json:"users"`         // Users in the team with their roles
	AgentOptions *json.RawMessage `json:"agent_options"` // Agent options for this team (can be complex JSON)
	// TODO: Add other fields like 'policies_count', 'mdm', etc. if they become available directly on the team object
	// Or consider hydrating them if they require separate API calls.
}

// TeamSecret represents an enrollment secret for a team.
type TeamSecret struct {
	Secret    string    `json:"secret"`
	CreatedAt FleetTime `json:"created_at"`
	TeamID    uint      `json:"team_id"` // This might be redundant if secrets are always nested under a team object
}

// TeamUser represents a user within a team and their role.
// This is similar to UserTeam in the user table but might be structured slightly differently
// in the /teams endpoint response if it includes more/less detail.
// The API doc for "Get team" shows `users` array with `id`, `name`, `email`, `global_role`, `role`.
type TeamUser struct {
	ID         uint    `json:"id"`
	Name       string  `json:"name"`
	Email      string  `json:"email"`
	GlobalRole *string `json:"global_role"` // User's global role
	Role       string  `json:"role"`        // User's role within this specific team
}

// ListTeamsResponse is the structure for the list teams API response.
// The API `GET /api/v1/fleet/teams` returns an array of teams directly.
// For consistency, we'll use a wrapper, but the actual API might just be `[]Team`.
// Update: The API doc for "List teams" (https://fleetdm.com/docs/rest-api/rest-api#list-all-teams)
// shows a response like: { "teams": [ { ...team_object... } ] }
type ListTeamsResponse struct {
	Teams []Team `json:"teams"`
	// Meta  struct { // If pagination meta is introduced for teams
	// 	HasNextResults bool `json:"has_next_results"`
	// } `json:"meta"`
}

func tableFleetdmTeam(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "fleetdm_team",
		Description: "Information about teams in FleetDM.",
		List: &plugin.ListConfig{
			Hydrate: listTeams,
			KeyColumns: []*plugin.KeyColumn{
				{Name: "query", Require: plugin.Optional}, // Search by team name
			},
		},
		Columns: []*plugin.Column{
			{Name: "id", Type: proto.ColumnType_INT, Description: "Unique ID of the team."},
			{Name: "name", Type: proto.ColumnType_STRING, Description: "Name of the team."},
			{Name: "description", Type: proto.ColumnType_STRING, Description: "Description of the team."},
			{Name: "user_count", Type: proto.ColumnType_INT, Description: "Number of users in the team."},
			{Name: "host_count", Type: proto.ColumnType_INT, Description: "Number of hosts assigned to the team."},
			{Name: "created_at", Type: proto.ColumnType_TIMESTAMP, Transform: transform.FromField("CreatedAt").Transform(flexibleTimeTransform), Description: "Timestamp when the team was created."},
			{Name: "agent_options", Type: proto.ColumnType_JSON, Description: "Agent options configured for this team."},

			// Secrets and Users are complex objects/arrays, exposing as JSON.
			// Could be expanded into separate tables or hydrated further.
			{Name: "secrets", Type: proto.ColumnType_JSON, Description: "Enrollment secrets associated with the team."},
			{Name: "users", Type: proto.ColumnType_JSON, Description: "Users belonging to this team and their roles. Fetched via GetTeam hydrate function."},

			// Query parameters that can be used for filtering (key columns)
			{Name: "query", Type: proto.ColumnType_STRING, Transform: transform.FromQual("query"), Description: "Search query keywords. Searchable field is team name. Set in WHERE clause."},
		},
	}
}

func listTeams(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := NewFleetDMClient(ctx, d.Connection)
	if err != nil {
		plugin.Logger(ctx).Error("fleetdm_team.listTeams", "connection_error", err)
		return nil, err
	}

	// Pagination for teams: The /api/v1/fleet/teams endpoint supports `page` and `per_page`
	page := 0
	perPage := 10000

	// limit := d.QueryContext.Limit
	// if limit != nil && *limit < int64(perPage) {
	// 	// perPage = int(*limit)
	// }

	for {
		params := url.Values{}
		params.Add("page", strconv.Itoa(page))
		params.Add("per_page", strconv.Itoa(perPage))

		if d.EqualsQuals["query"] != nil {
			params.Add("query", d.EqualsQuals["query"].GetStringValue())
		}

		var response ListTeamsResponse
		_, err := client.Get(ctx, "teams", params, &response)
		if err != nil {
			plugin.Logger(ctx).Error("fleetdm_team.listTeams", "api_error", err, "page", page)
			return nil, err
		}

		for _, team := range response.Teams {
			// The list endpoint for teams might not include all details like full user list or secrets.
			// These are typically available in the "Get team" endpoint.
			// We can stream the basic info here, and rely on GetTeam or a separate hydrate for richer details if needed.
			// For now, we assume the list endpoint provides sufficient top-level info.
			// The `users` field in the `Team` struct will be populated by `getTeam` when a single team is fetched.
			// For `listTeams`, the `users` field might be empty or contain minimal info from the list endpoint.
			// The API doc for "List teams" does not show `users` or `secrets` in the response items.
			// These are shown in "Get team". So, for list, these fields will be nil/empty.
			// We can add a hydrate function to populate them if `plugin.GetConfig` is not used.
			// Or, document that these fields are primarily for `Get`.
			// For simplicity in list, we stream what `GET /teams` provides.
			// The `users` column in the table definition will be hydrated by `getTeam` for individual `GET`s.
			// If we want `users` for `LIST`, we'd need a separate hydrate call for each team.
			d.StreamListItem(ctx, team)
			if d.RowsRemaining(ctx) == 0 {
				plugin.Logger(ctx).Debug("fleetdm_team.listTeams", "limit_reached", true)
				return nil, nil
			}
		}

		// Pagination check: if the number of teams returned is less than per_page,
		// it's the last page. The `/teams` endpoint does not specify a `meta.has_next_results`.
		if len(response.Teams) < perPage {
			plugin.Logger(ctx).Debug("fleetdm_team.listTeams", "end_of_results", true, "teams_on_page", len(response.Teams))
			break
		}

		page++
		plugin.Logger(ctx).Debug("fleetdm_team.listTeams", "next_page", page)
	}

	return nil, nil
}
