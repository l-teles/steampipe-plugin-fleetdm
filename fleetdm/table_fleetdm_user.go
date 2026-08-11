package fleetdm

import (
	"context"
	"net/url"
	"strconv"

	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

// User represents a FleetDM user.
// Refer to: https://fleetdm.com/docs/rest-api/rest-api#user-object
type User struct {
	ID                       uint       `json:"id"`
	CreatedAt                FleetTime  `json:"created_at"`
	UpdatedAt                FleetTime  `json:"updated_at"`
	Name                     string     `json:"name"`
	Email                    string     `json:"email"`
	AdminForcedPasswordReset bool       `json:"admin_forced_password_reset"`
	GravatarURL              string     `json:"gravatar_url"`
	SSOEnabled               bool       `json:"sso_enabled"`
	GlobalRole               *string    `json:"global_role"` // e.g., "admin", "maintainer", "observer"
	Teams                    []UserTeam `json:"teams"`       // Teams the user belongs to and their role in each
	APIOnly                  bool       `json:"api_only"`    // True if the user is an API-only user
}

// UserTeam represents a team a user belongs to and their role.
type UserTeam struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"` // Role within the team, e.g., "admin", "maintainer", "observer"
}

// ListUsersResponse is the structure for the list users API response.
// The API doc (https://fleetdm.com/docs/rest-api/rest-api#list-all-users) shows a response like:
// { "users": [ { ...user_object... } ] }
type ListUsersResponse struct {
	Users []User `json:"users"`
}

func tableFleetdmUser(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "fleetdm_user",
		Description: "Information about users in FleetDM.",
		List: &plugin.ListConfig{
			Hydrate: listUsers,
			KeyColumns: []*plugin.KeyColumn{
				{Name: "query", Require: plugin.Optional},   // Search by name or email
				{Name: "team_id", Require: plugin.Optional}, // Filter by team (Fleet Premium)
			},
		},
		Columns: []*plugin.Column{
			{Name: "id", Type: proto.ColumnType_INT, Description: "Unique ID of the user."},
			{Name: "name", Type: proto.ColumnType_STRING, Description: "Full name of the user."},
			{Name: "email", Type: proto.ColumnType_STRING, Description: "Email address of the user."},
			{Name: "global_role", Type: proto.ColumnType_STRING, Description: "Global role of the user (e.g., admin, maintainer, observer). Null if not a global role."},
			{Name: "api_only", Type: proto.ColumnType_BOOL, Description: "Indicates if the user is an API-only user."},
			{Name: "sso_enabled", Type: proto.ColumnType_BOOL, Description: "Indicates if Single Sign-On is enabled for the user."},
			{Name: "admin_forced_password_reset", Type: proto.ColumnType_BOOL, Description: "Indicates if an admin has forced a password reset for the user."},
			{Name: "gravatar_url", Type: proto.ColumnType_STRING, Description: "URL for the user's Gravatar image."},
			{Name: "created_at", Type: proto.ColumnType_TIMESTAMP, Transform: transform.FromField("CreatedAt").Transform(flexibleTimeTransform), Description: "Timestamp when the user was created."},
			{Name: "updated_at", Type: proto.ColumnType_TIMESTAMP, Transform: transform.FromField("UpdatedAt").Transform(flexibleTimeTransform), Description: "Timestamp when the user was last updated."},
			{Name: "teams", Type: proto.ColumnType_JSON, Description: "Teams the user belongs to, including their role in each team.", Transform: transform.FromField("Teams")},

			// Query parameters that can be used for filtering (key columns)
			{Name: "query", Type: proto.ColumnType_STRING, Transform: transform.FromQual("query"), Description: "Search query keywords. Searchable fields include name and email. Set in WHERE clause."},
			{Name: "team_id", Type: proto.ColumnType_INT, Transform: transform.FromQual("team_id"), Description: "Filter by team ID (Fleet Premium). Set in WHERE clause."},
		},
	}
}

func listUsers(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := getClient(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("fleetdm_user.listUsers", "connection_error", err)
		return nil, err
	}

	err = paginatedList(ctx, d, 100, func(ctx context.Context, page, perPage int) ([]User, *ListMeta, error) {
		params := url.Values{}
		params.Add("page", strconv.Itoa(page))
		params.Add("per_page", strconv.Itoa(perPage))

		if d.EqualsQuals["query"] != nil {
			params.Add("query", d.EqualsQuals["query"].GetStringValue())
		}
		if d.EqualsQuals["team_id"] != nil {
			addFleetIDParam(params, strconv.FormatInt(d.EqualsQuals["team_id"].GetInt64Value(), 10))
		}

		var usersResponse ListUsersResponse
		if err := client.Get(ctx, "users", params, &usersResponse); err != nil {
			plugin.Logger(ctx).Error("fleetdm_user.listUsers", "api_error", err, "page", page)
			return nil, nil, err
		}
		// The /users endpoint does not document a meta object for pagination.
		return usersResponse.Users, nil, nil
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}
