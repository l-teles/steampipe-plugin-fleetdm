package fleetdm

import (
	"context"
	"net/url"

	"github.com/turbot/steampipe-plugin-sdk/v6/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin/transform"
)

// SelfServiceCategory groups self-service software in Fleet Desktop.
// Refer to: https://fleetdm.com/docs/rest-api/rest-api#list-self-service-categories
type SelfServiceCategory struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	TeamID    *uint     `json:"team_id"`
	FleetID   *uint     `json:"fleet_id"` // team_id since Fleet 4.82
	CreatedAt FleetTime `json:"created_at"`
	UpdatedAt FleetTime `json:"updated_at"`
}

// ListSelfServiceCategoriesResponse for `GET /api/v1/fleet/software/self_service_categories`
// (not paginated; a team is required).
type ListSelfServiceCategoriesResponse struct {
	SelfServiceCategories []SelfServiceCategory `json:"self_service_categories"`
}

func tableFleetdmSoftwareSelfServiceCategory(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "fleetdm_software_self_service_category",
		Description: "Categories that group self-service software in Fleet Desktop, across all teams.",
		List: &plugin.ListConfig{
			Hydrate: listSelfServiceCategories,
			KeyColumns: []*plugin.KeyColumn{
				{Name: "team_id", Require: plugin.Optional}, // Only this team; 0 for no team
			},
		},
		Columns: []*plugin.Column{
			{Name: "id", Type: proto.ColumnType_INT, Description: "Unique ID of the category."},
			{Name: "name", Type: proto.ColumnType_STRING, Description: "Name of the category."},
			{Name: "team_id", Type: proto.ColumnType_INT, Transform: transform.FromField("FleetID", "TeamID"), Description: "ID of the team (fleet) the category belongs to; 0 for hosts with no team."},
			{Name: "created_at", Type: proto.ColumnType_TIMESTAMP, Transform: transform.FromField("CreatedAt").Transform(flexibleTimeTransform), Description: "When the category was created."},
			{Name: "updated_at", Type: proto.ColumnType_TIMESTAMP, Transform: transform.FromField("UpdatedAt").Transform(flexibleTimeTransform), Description: "When the category was last updated."},
		},
	}
}

func listSelfServiceCategories(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := getClient(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("fleetdm_software_self_service_category.listSelfServiceCategories", "connection_error", err)
		return nil, err
	}
	scopes, err := teamScopes(ctx, d, client, "fleetdm_software_self_service_category")
	if err != nil {
		return nil, err
	}
	for _, scope := range scopes {
		// The endpoint ignores paging and requires a team; 0 means no team.
		params := url.Values{}
		client.addFleetIDParam(ctx, d, params, scopeParam(scope))
		var response ListSelfServiceCategoriesResponse
		if err := client.Get(ctx, "software/self_service_categories", params, &response); err != nil {
			if handleUnavailableFeature(ctx, d, "fleetdm_software_self_service_category", err) {
				return nil, nil
			}
			if skipScope(ctx, d, "fleetdm_software_self_service_category", scope, err) {
				continue
			}
			plugin.Logger(ctx).Error("fleetdm_software_self_service_category.listSelfServiceCategories", "api_error", err)
			return nil, err
		}
		for _, c := range response.SelfServiceCategories {
			d.StreamListItem(ctx, c)
			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}
	}
	return nil, nil
}
