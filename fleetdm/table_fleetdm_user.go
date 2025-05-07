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

// User represents a FleetDM user.
// Refer to: https://fleetdm.com/docs/rest-api/rest-api#user-object
type User struct {
	ID                       uint       `json:"id"`
	CreatedAt                time.Time  `json:"created_at"`
	UpdatedAt                time.Time  `json:"updated_at"`
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
			// KeyColumns: plugin.KeyColumnEquals("team_id"), // TODO: Check if API supports filtering users by team_id directly
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
			{Name: "created_at", Type: proto.ColumnType_TIMESTAMP, Description: "Timestamp when the user was created."},
			{Name: "updated_at", Type: proto.ColumnType_TIMESTAMP, Description: "Timestamp when the user was last updated."},
			{Name: "teams", Type: proto.ColumnType_JSON, Description: "Teams the user belongs to, including their role in each team.", Transform: transform.FromField("Teams")},

			// Connection config (server_url)
			{Name: "server_url", Type: proto.ColumnType_STRING, Hydrate: getServerURL, Transform: transform.FromValue(), Description: "FleetDM server URL from connection config."},
		},
	}
}

func listUsers(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := NewFleetDMClient(ctx, d.Connection)
	if err != nil {
		plugin.Logger(ctx).Error("fleetdm_user.listUsers", "connection_error", err)
		return nil, err
	}

	// Pagination for users: The /api/v1/fleet/users endpoint supports `page` and `per_page`
	page := 0
	perPage := 50 // A reasonable default, adjust as needed or if API has specific limits/max

	limit := d.QueryContext.Limit
	if limit != nil && *limit < int64(perPage) {
		// perPage = int(*limit) // Be cautious if API has minimum per_page
	}

	for {
		params := url.Values{}
		params.Add("page", strconv.Itoa(page))
		params.Add("per_page", strconv.Itoa(perPage))

		var usersResponse ListUsersResponse

		httpResp, err := client.Get(ctx, "users", params, &usersResponse) // Pass ListUsersResponse struct
		if err != nil {
			plugin.Logger(ctx).Error("fleetdm_user.listUsers", "api_error", err, "page", page)
			return nil, err
		}

		// Check if usersResponse.Users is nil after a successful HTTP call, which might indicate an empty list or unexpected response format
		if httpResp.StatusCode == 200 && usersResponse.Users == nil {
			// This could happen if the API returned `[]` instead of `{"users": []}` and the decoder didn't error but also didn't populate.
			// Or if it returned `{"users": null}`.
			// Given the API docs, `{"users": [...]}` is expected, so `usersResponse.Users` should be populated.
			// If it's nil, it implies an empty list of users from the API for this page.
			plugin.Logger(ctx).Debug("fleetdm_user.listUsers", "users_array_is_nil_or_empty_on_page", page)
			// This is a valid state for the last page if it's empty, or if there are no users.
		}

		if len(usersResponse.Users) == 0 && page == 0 { // No users found at all on the first page
			plugin.Logger(ctx).Debug("fleetdm_user.listUsers", "no_users_found_at_all", true)
			return nil, nil // Stop if no users on the very first call
		}
		
		for _, user := range usersResponse.Users {
			d.StreamListItem(ctx, user)
			if d.RowsRemaining(ctx) == 0 {
				plugin.Logger(ctx).Debug("fleetdm_user.listUsers", "limit_reached", true)
				return nil, nil
			}
		}

		// Pagination check: if the number of users returned is less than per_page,
		// it's the last page. FleetDM's /users endpoint does not seem to use a 'meta.has_next_results' field.
		if len(usersResponse.Users) < perPage {
			plugin.Logger(ctx).Debug("fleetdm_user.listUsers", "end_of_results", true, "users_on_page", len(usersResponse.Users))
			break
		}

		page++
		plugin.Logger(ctx).Debug("fleetdm_user.listUsers", "next_page", page)
	}

	return nil, nil
}
