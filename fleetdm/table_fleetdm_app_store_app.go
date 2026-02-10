package fleetdm

import (
	"context"
	"net/url"
	"strconv"

	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

// AppStoreApp represents an Apple App Store app in FleetDM.
// Refer to: https://fleetdm.com/docs/rest-api/rest-api#list-apple-app-store-apps
type AppStoreApp struct {
	AppStoreID       string      `json:"app_store_id"`
	Platform         string      `json:"platform"`
	SelfService      bool        `json:"self_service"`
	LabelsIncludeAny interface{} `json:"labels_include_any"`
	LabelsExcludeAny interface{} `json:"labels_exclude_any"`
	CreatedAt        *FleetTime  `json:"created_at"`
	Categories       interface{} `json:"categories"`
	DisplayName      *string     `json:"display_name"`
	BundleIdentifier string      `json:"bundle_identifier"`
	IconURL          string      `json:"icon_url"`
	Name             string      `json:"name"`
	LatestVersion    string      `json:"latest_version"`
}

// AppStoreAppWithTeam wraps an AppStoreApp with the team context it was queried for.
type AppStoreAppWithTeam struct {
	AppStoreApp
	TeamID   uint   `json:"team_id"`
	TeamName string `json:"team_name"`
}

// ListAppStoreAppsResponse is the expected structure for the list app store apps API call.
type ListAppStoreAppsResponse struct {
	AppStoreApps []AppStoreApp `json:"app_store_apps"`
}

func tableFleetdmAppStoreApp(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "fleetdm_app_store_app",
		Description: "Apple App Store apps (VPP) from FleetDM. Lists apps available for install on teams. The API requires a team_id, so the plugin auto-discovers all teams and queries each one. Uses the /software/app_store_apps endpoint.",
		List: &plugin.ListConfig{
			Hydrate: listAppStoreApps,
			KeyColumns: []*plugin.KeyColumn{
				{Name: "team_id", Require: plugin.Optional},
			},
		},
		Columns: []*plugin.Column{
			// Core App Store app information
			{Name: "app_store_id", Type: proto.ColumnType_STRING, Transform: transform.FromField("AppStoreID"), Description: "The Apple App Store ID of the app."},
			{Name: "name", Type: proto.ColumnType_STRING, Description: "Name of the app."},
			{Name: "display_name", Type: proto.ColumnType_STRING, Description: "Display name override for the app (null if not set)."},
			{Name: "icon_url", Type: proto.ColumnType_STRING, Transform: transform.FromField("IconURL"), Description: "URL of the app icon."},
			{Name: "platform", Type: proto.ColumnType_STRING, Description: "Platform for the app (e.g., 'darwin', 'ios', 'ipados')."},
			{Name: "latest_version", Type: proto.ColumnType_STRING, Description: "Latest available version of the app."},
			{Name: "bundle_identifier", Type: proto.ColumnType_STRING, Description: "Bundle identifier of the app (e.g., 'com.google.chrome.ios')."},
			{Name: "self_service", Type: proto.ColumnType_BOOL, Description: "Whether the app is available as self-service."},
			{Name: "categories", Type: proto.ColumnType_JSON, Description: "Categories the app belongs to."},
			{Name: "labels_include_any", Type: proto.ColumnType_JSON, Description: "Labels to include for targeting."},
			{Name: "labels_exclude_any", Type: proto.ColumnType_JSON, Description: "Labels to exclude for targeting."},
			{Name: "created_at", Type: proto.ColumnType_TIMESTAMP, Transform: transform.FromField("CreatedAt").Transform(flexibleTimeTransform), Description: "Timestamp when the app was added."},

			// Team association
			{Name: "team_id", Type: proto.ColumnType_INT, Transform: transform.FromField("TeamID"), Description: "The team ID this app was queried for. Set in WHERE clause to query a specific team."},
			{Name: "team_name", Type: proto.ColumnType_STRING, Description: "The name of the team this app was queried for."},
		},
	}
}

func listAppStoreApps(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := NewFleetDMClient(ctx, d.Connection)
	if err != nil {
		plugin.Logger(ctx).Error("fleetdm_app_store_app.listAppStoreApps", "connection_error", err)
		return nil, err
	}

	// Build the list of teams to query.
	// The /software/app_store_apps endpoint requires team_id.
	type teamInfo struct {
		ID   uint
		Name string
	}
	var teamsToQuery []teamInfo

	if d.EqualsQuals["team_id"] != nil {
		// User specified a team_id in the WHERE clause — only query that team.
		teamID := uint(d.EqualsQuals["team_id"].GetInt64Value())
		teamsToQuery = append(teamsToQuery, teamInfo{ID: teamID, Name: ""})
		plugin.Logger(ctx).Info("fleetdm_app_store_app.listAppStoreApps", "using_specific_team_id", teamID)
	} else {
		// Auto-discover all teams by calling GET /api/v1/fleet/teams.
		plugin.Logger(ctx).Info("fleetdm_app_store_app.listAppStoreApps", "discovering_all_teams", true)

		page := 0
		perPage := 10000

		for {
			params := url.Values{}
			params.Add("page", strconv.Itoa(page))
			params.Add("per_page", strconv.Itoa(perPage))

			var teamsResponse ListTeamsResponse
			_, err := client.Get(ctx, "teams", params, &teamsResponse)
			if err != nil {
				plugin.Logger(ctx).Error("fleetdm_app_store_app.listAppStoreApps", "teams_api_error", err, "page", page)
				return nil, err
			}

			for _, team := range teamsResponse.Teams {
				teamsToQuery = append(teamsToQuery, teamInfo{ID: team.ID, Name: team.Name})
			}

			if len(teamsResponse.Teams) < perPage {
				break
			}
			page++
		}

		plugin.Logger(ctx).Info("fleetdm_app_store_app.listAppStoreApps", "total_teams_discovered", len(teamsToQuery))
	}

	// For each team, query the app store apps endpoint.
	for _, team := range teamsToQuery {
		params := url.Values{}
		params.Add("team_id", strconv.FormatUint(uint64(team.ID), 10))

		var response ListAppStoreAppsResponse
		_, err := client.Get(ctx, "software/app_store_apps", params, &response)
		if err != nil {
			plugin.Logger(ctx).Error("fleetdm_app_store_app.listAppStoreApps", "api_error", err, "team_id", team.ID)
			return nil, err
		}

		plugin.Logger(ctx).Info("fleetdm_app_store_app.listAppStoreApps",
			"team_id", team.ID,
			"team_name", team.Name,
			"apps_returned", len(response.AppStoreApps),
		)

		for _, app := range response.AppStoreApps {
			d.StreamListItem(ctx, AppStoreAppWithTeam{
				AppStoreApp: app,
				TeamID:      team.ID,
				TeamName:    team.Name,
			})
			if d.RowsRemaining(ctx) == 0 {
				plugin.Logger(ctx).Debug("fleetdm_app_store_app.listAppStoreApps", "limit_reached_sdk", true)
				return nil, nil
			}
		}
	}

	plugin.Logger(ctx).Info("fleetdm_app_store_app.listAppStoreApps", "list_app_store_apps_completed", true)
	return nil, nil
}
