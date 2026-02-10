package fleetdm

import (
	"context"
	"net/url"
	"strconv"

	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

// FleetMaintainedApp represents a Fleet-maintained app.
// Refer to: https://fleetdm.com/docs/rest-api/rest-api#list-fleet-maintained-apps
type FleetMaintainedApp struct {
	ID              uint     `json:"id"`
	Name            string   `json:"name"`
	Slug            string   `json:"slug"`
	Platform        string   `json:"platform"`
	Version         *string  `json:"version,omitempty"`
	SoftwareTitleID *uint    `json:"software_title_id"`
	Categories      []string `json:"categories"`
}

// ListFleetMaintainedAppsResponse is the expected structure for the list Fleet-maintained apps API call.
type ListFleetMaintainedAppsResponse struct {
	FleetMaintainedApps []FleetMaintainedApp `json:"fleet_maintained_apps"`
	Meta                struct {
		HasNextResults     bool `json:"has_next_results"`
		HasPreviousResults bool `json:"has_previous_results"`
	} `json:"meta"`
}

func tableFleetdmFleetMaintainedApp(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "fleetdm_fleet_maintained_app",
		Description: "Fleet-maintained apps available in FleetDM. These are pre-packaged software installers maintained by Fleet. Uses the /software/fleet_maintained_apps endpoint.",
		List: &plugin.ListConfig{
			Hydrate: listFleetMaintainedApps,
			KeyColumns: []*plugin.KeyColumn{
				{Name: "team_id", Require: plugin.Optional},
			},
		},
		Columns: []*plugin.Column{
			// Core Fleet-maintained app information
			{Name: "id", Type: proto.ColumnType_INT, Description: "Unique ID of the Fleet-maintained app."},
			{Name: "name", Type: proto.ColumnType_STRING, Description: "Name of the Fleet-maintained app."},
			{Name: "slug", Type: proto.ColumnType_STRING, Description: "Slug identifier for the app (e.g., '1password/darwin')."},
			{Name: "platform", Type: proto.ColumnType_STRING, Description: "Platform for the app (e.g., 'darwin', 'windows')."},
			{Name: "version", Type: proto.ColumnType_STRING, Description: "Latest available version of the app."},
			{Name: "software_title_id", Type: proto.ColumnType_INT, Transform: transform.FromField("SoftwareTitleID"), Description: "Software title ID if the app has been added to the specified team."},
			{Name: "categories", Type: proto.ColumnType_JSON, Description: "Categories the app belongs to."},

			// Query parameters that can be used for filtering
			{Name: "team_id", Type: proto.ColumnType_INT, Transform: transform.FromQual("team_id"), Description: "Filter by team ID. When specified, each app includes the software_title_id if already added to that team. Set in WHERE clause."},
		},
	}
}

func listFleetMaintainedApps(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := NewFleetDMClient(ctx, d.Connection)
	if err != nil {
		plugin.Logger(ctx).Error("fleetdm_fleet_maintained_app.listFleetMaintainedApps", "connection_error", err)
		return nil, err
	}

	page := 0
	perPage := 10000

	for {
		params := url.Values{}
		params.Add("page", strconv.Itoa(page))
		params.Add("per_page", strconv.Itoa(perPage))

		if d.EqualsQuals["team_id"] != nil {
			params.Add("team_id", strconv.FormatInt(d.EqualsQuals["team_id"].GetInt64Value(), 10))
		}

		var response ListFleetMaintainedAppsResponse
		_, err := client.Get(ctx, "software/fleet_maintained_apps", params, &response) // Endpoint is /api/v1/fleet/software/fleet_maintained_apps
		if err != nil {
			plugin.Logger(ctx).Error("fleetdm_fleet_maintained_app.listFleetMaintainedApps", "api_error", err, "page", page, "params", params.Encode())
			return nil, err
		}

		for _, app := range response.FleetMaintainedApps {
			d.StreamListItem(ctx, app)
			if d.RowsRemaining(ctx) == 0 {
				plugin.Logger(ctx).Debug("fleetdm_fleet_maintained_app.listFleetMaintainedApps", "limit_reached_sdk", "true")
				return nil, nil
			}
		}

		plugin.Logger(ctx).Info("fleetdm_fleet_maintained_app.listFleetMaintainedApps",
			"page_processed", page,
			"items_on_page", len(response.FleetMaintainedApps),
			"api_has_next_results", response.Meta.HasNextResults,
		)

		if len(response.FleetMaintainedApps) < perPage {
			plugin.Logger(ctx).Info("fleetdm_fleet_maintained_app.listFleetMaintainedApps", "pagination_ended_item_count_less_than_per_page", true, "current_page", page, "items_on_page", len(response.FleetMaintainedApps), "per_page", perPage)
			break
		}

		if !response.Meta.HasNextResults && len(response.FleetMaintainedApps) == perPage {
			plugin.Logger(ctx).Warn("fleetdm_fleet_maintained_app.listFleetMaintainedApps", "api_has_next_results_is_false_but_full_page_received", true, "current_page", page)
		}

		page++
		plugin.Logger(ctx).Debug("fleetdm_fleet_maintained_app.listFleetMaintainedApps", "incrementing_to_next_page", page)
	}

	plugin.Logger(ctx).Info("fleetdm_fleet_maintained_app.listFleetMaintainedApps", "list_fleet_maintained_apps_completed", true)
	return nil, nil
}
