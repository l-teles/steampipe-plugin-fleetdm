package fleetdm

import (
	"context"
	"net/url"
	"strconv"

	"github.com/turbot/steampipe-plugin-sdk/v6/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin/transform"
)

// CustomVariable is a secret variable usable in scripts and profiles
// ($FLEET_SECRET_<NAME>). The API never returns its value.
// Refer to: https://fleetdm.com/docs/rest-api/rest-api#list-custom-variables
type CustomVariable struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	CreatedAt FleetTime `json:"created_at"`
	UpdatedAt FleetTime `json:"updated_at"`
}

// ListCustomVariablesResponse for `GET /api/v1/fleet/custom_variables`.
type ListCustomVariablesResponse struct {
	CustomVariables []CustomVariable `json:"custom_variables"`
	Meta            *ListMeta        `json:"meta"`
}

func tableFleetdmCustomVariable(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "fleetdm_custom_variable",
		Description: "Custom (secret) variables available to scripts and configuration profiles. Values are never returned.",
		List: &plugin.ListConfig{
			Hydrate: listCustomVariables,
		},
		Columns: []*plugin.Column{
			{Name: "id", Type: proto.ColumnType_INT, Description: "Unique ID of the variable."},
			{Name: "name", Type: proto.ColumnType_STRING, Description: "Name of the variable, referenced as $FLEET_SECRET_<name>."},
			{Name: "created_at", Type: proto.ColumnType_TIMESTAMP, Transform: transform.FromField("CreatedAt").Transform(flexibleTimeTransform), Description: "When the variable was created."},
			{Name: "updated_at", Type: proto.ColumnType_TIMESTAMP, Transform: transform.FromField("UpdatedAt").Transform(flexibleTimeTransform), Description: "When the variable was last updated."},
		},
	}
}

func listCustomVariables(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := getClient(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("fleetdm_custom_variable.listCustomVariables", "connection_error", err)
		return nil, err
	}
	err = paginatedList(ctx, d, 100, func(ctx context.Context, page, perPage int) ([]CustomVariable, *ListMeta, error) {
		params := url.Values{}
		params.Add("page", strconv.Itoa(page))
		params.Add("per_page", strconv.Itoa(perPage))
		var response ListCustomVariablesResponse
		if err := client.Get(ctx, "custom_variables", params, &response); err != nil {
			plugin.Logger(ctx).Error("fleetdm_custom_variable.listCustomVariables", "api_error", err, "page", page)
			return nil, nil, err
		}
		return response.CustomVariables, response.Meta, nil
	})
	return nil, err
}
