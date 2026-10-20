package fleetdm

import (
	"context"
	"encoding/json" // Added import for json.RawMessage
	"fmt"
	"net/url"
	"strconv"

	"github.com/turbot/steampipe-plugin-sdk/v6/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin/transform"
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
	// Per-team settings blocks returned by the list endpoint.
	Features           *json.RawMessage `json:"features"`
	HostExpirySettings *json.RawMessage `json:"host_expiry_settings"`
	MDM                *json.RawMessage `json:"mdm"`
	Scripts            *json.RawMessage `json:"scripts"`
	Software           *json.RawMessage `json:"software"`
	Integrations       *json.RawMessage `json:"integrations"`     // may hold webhook URLs; gated
	WebhookSettings    *json.RawMessage `json:"webhook_settings"` // destination URLs; gated
}

// TeamSecret represents an enrollment secret for a team.
type TeamSecret struct {
	Secret    string    `json:"secret"`
	CreatedAt FleetTime `json:"created_at"`
	TeamID    uint      `json:"team_id"` // This might be redundant if secrets are always nested under a team object
}

// TeamUser is a member of a team and their role in it.
type TeamUser struct {
	ID         uint    `json:"id"`
	Name       string  `json:"name"`
	Email      string  `json:"email"`
	GlobalRole *string `json:"global_role"` // User's global role
	Role       string  `json:"role"`        // User's role within this specific team
}

// ListTeamsResponse is the structure for the list teams/fleets API response.
// Old servers respond with { "teams": [...] }; servers past the teams->fleets
// rename respond with { "fleets": [...] }. Both keys are declared so either
// generation unmarshals; use coalesceSlice(Fleets, Teams) to read the result.
type ListTeamsResponse struct {
	Teams  []Team `json:"teams"`
	Fleets []Team `json:"fleets"`
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
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("id"),
			Hydrate:    getTeam,
		},
		Columns: []*plugin.Column{
			{Name: "id", Type: proto.ColumnType_INT, Description: "Unique ID of the team."},
			{Name: "name", Type: proto.ColumnType_STRING, Description: "Name of the team."},
			{Name: "description", Type: proto.ColumnType_STRING, Description: "Description of the team."},
			{Name: "user_count", Type: proto.ColumnType_INT, Description: "Number of users in the team."},
			{Name: "host_count", Type: proto.ColumnType_INT, Description: "Number of hosts assigned to the team."},
			{Name: "created_at", Type: proto.ColumnType_TIMESTAMP, Transform: transform.FromField("CreatedAt").Transform(flexibleTimeTransform), Description: "Timestamp when the team was created."},
			{Name: "agent_options", Type: proto.ColumnType_JSON, Description: "Agent options configured for this team."},

			{Name: "secrets", Type: proto.ColumnType_JSON, Hydrate: getTeamSecrets, Transform: transform.FromValue(), Description: "Enrollment secrets associated with the team. NULL unless expose_secrets = true is set in the connection config, as these are live credentials that allow device enrollment."},
			{Name: "users", Type: proto.ColumnType_JSON, Hydrate: getTeamUsers, Transform: transform.FromValue(), Description: "Members of the team (id, name, email, global_role) and their role in it. One extra request per team, made only when this column is selected."},
			{Name: "features", Type: proto.ColumnType_JSON, Description: "Team feature toggles (host users, software inventory, historical data)."},
			{Name: "host_expiry_settings", Type: proto.ColumnType_JSON, Description: "Host expiry settings for the team."},
			{Name: "mdm", Type: proto.ColumnType_JSON, Transform: transform.FromField("MDM"), Description: "Team MDM settings (OS updates, disk encryption, setup experience, profiles)."},
			{Name: "scripts", Type: proto.ColumnType_JSON, Description: "Scripts configured for the team via GitOps."},
			{Name: "software", Type: proto.ColumnType_JSON, Description: "Software configured for the team (packages, App Store apps, Fleet-maintained apps)."},
			{Name: "integrations", Type: proto.ColumnType_JSON, Hydrate: itemIfSecretsExposed, Transform: transform.FromField("Integrations"), Description: "Team integrations (Jira, Zendesk, Google Calendar, conditional access). NULL unless expose_secrets = true, as webhook URLs can embed credentials."},
			{Name: "webhook_settings", Type: proto.ColumnType_JSON, Hydrate: itemIfSecretsExposed, Transform: transform.FromField("WebhookSettings"), Description: "Team webhook settings. NULL unless expose_secrets = true, as destination URLs can embed credentials."},

			// Query parameters that can be used for filtering (key columns)
			{Name: "query", Type: proto.ColumnType_STRING, Transform: transform.FromQual("query"), Description: "Search query keywords. Searchable field is team name. Set in WHERE clause."},
		},
	}
}

// getTeamSecrets gates the secrets column behind the expose_secrets connection
// config flag. Enrollment secrets are live credentials; by default they must
// not land in Steampipe query caches, exports, or dashboards. This is a pure
// gate over the already-fetched list item — no extra API call is made.
func getTeamSecrets(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	config := GetConfig(d.Connection)
	if config.ExposeSecrets == nil || !*config.ExposeSecrets {
		return nil, nil
	}
	team, ok := h.Item.(Team)
	if !ok {
		return nil, fmt.Errorf("getTeamSecrets: unexpected item type %T", h.Item)
	}
	if len(team.Secrets) == 0 {
		return nil, nil
	}
	return team.Secrets, nil
}

// listTeamUsersResponse is `GET /api/v1/fleet/fleets/{id}/users`. Each user
// embeds full copies of their teams (including settings gated elsewhere), so
// only the identity and the role for this team are kept.
type listTeamUsersResponse struct {
	Users []struct {
		ID         uint       `json:"id"`
		Name       string     `json:"name"`
		Email      string     `json:"email"`
		GlobalRole *string    `json:"global_role"`
		Teams      []teamRole `json:"teams"`
		Fleets     []teamRole `json:"fleets"` // teams since Fleet 4.82
	} `json:"users"`
}

// teamRole is the part of a user's team membership the users column keeps.
type teamRole struct {
	ID   uint   `json:"id"`
	Role string `json:"role"`
}

// getTeamUsers fetches a team's members; the team list does not include them.
func getTeamUsers(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	team, ok := h.Item.(Team)
	if !ok {
		return nil, fmt.Errorf("getTeamUsers: unexpected item type %T", h.Item)
	}
	client, err := getClient(ctx, d)
	if err != nil {
		return nil, err
	}
	var response listTeamUsersResponse
	// Without per_page Fleet returns every member in one response.
	if err := client.GetCompat(ctx, d, famFleetUsers, []any{team.ID}, nil, &response); err != nil {
		plugin.Logger(ctx).Error("fleetdm_team.getTeamUsers", "api_error", err, "team_id", team.ID)
		return nil, err
	}
	users := make([]TeamUser, 0, len(response.Users))
	for _, u := range response.Users {
		member := TeamUser{ID: u.ID, Name: u.Name, Email: u.Email, GlobalRole: u.GlobalRole}
		for _, t := range coalesceSlice(u.Fleets, u.Teams) {
			if t.ID == team.ID {
				member.Role = t.Role
			}
		}
		users = append(users, member)
	}
	return users, nil
}

func listTeams(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := getClient(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("fleetdm_team.listTeams", "connection_error", err)
		return nil, err
	}

	err = paginatedList(ctx, d, 100, func(ctx context.Context, page, perPage int) ([]Team, *ListMeta, error) {
		params := url.Values{}
		params.Add("page", strconv.Itoa(page))
		params.Add("per_page", strconv.Itoa(perPage))

		if d.EqualsQuals["query"] != nil {
			params.Add("query", d.EqualsQuals["query"].GetStringValue())
		}

		var response ListTeamsResponse
		if err := client.GetCompat(ctx, d, famFleets, nil, params, &response); err != nil {
			plugin.Logger(ctx).Error("fleetdm_team.listTeams", "api_error", err, "page", page)
			return nil, nil, err
		}
		// Neither the /teams nor the /fleets endpoint documents a meta object
		// for pagination. The list endpoint provides top-level team info only;
		// `users` and `secrets` are populated by the "Get team" endpoint.
		return coalesceSlice(response.Fleets, response.Teams), nil, nil
	})
	if err != nil {
		// Teams are a Fleet Premium feature.
		if handleUnavailableFeature(ctx, d, "fleetdm_team", err) {
			return nil, nil
		}
		return nil, err
	}

	return nil, nil
}

// fetchAllTeams returns every team (fleet) for tables that fan out per team.
// It is a nested full fetch rather than a stream, so it pages until the first
// empty page.
func fetchAllTeams(ctx context.Context, d *plugin.QueryData, client *FleetDMClient) ([]Team, error) {
	var all []Team
	err := paginateCore(ctx, 100, func(ctx context.Context, page, perPage int) ([]Team, *ListMeta, error) {
		params := url.Values{}
		params.Add("page", strconv.Itoa(page))
		params.Add("per_page", strconv.Itoa(perPage))
		var response ListTeamsResponse
		if err := client.GetCompat(ctx, d, famFleets, nil, params, &response); err != nil {
			return nil, nil, err
		}
		return coalesceSlice(response.Fleets, response.Teams), nil, nil
	}, func(team Team) bool {
		all = append(all, team)
		return true
	})
	return all, err
}

// getTeam fetches one team by id, instead of listing and filtering.
func getTeam(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := getClient(ctx, d)
	if err != nil {
		return nil, err
	}
	var response struct {
		Fleet *Team `json:"fleet"`
		Team  *Team `json:"team"`
	}
	if err := client.GetCompat(ctx, d, famFleet, []any{d.EqualsQuals["id"].GetInt64Value()}, nil, &response); err != nil {
		plugin.Logger(ctx).Error("fleetdm_team.getTeam", "api_error", err)
		return nil, err
	}
	if response.Fleet != nil {
		return *response.Fleet, nil
	}
	if response.Team != nil {
		return *response.Team, nil
	}
	return nil, nil
}
