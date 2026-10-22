package fleetdm

import (
	"context"
	"net/url"
	"strconv"

	"github.com/turbot/steampipe-plugin-sdk/v6/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin/transform"
)

// CustomHostVital is an extra host attribute, filled in by a query, that
// labels can target (Fleet 4.90+).
// Refer to: https://fleetdm.com/docs/rest-api/rest-api#list-custom-host-vitals
type CustomHostVital struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	CreatedAt FleetTime `json:"created_at"`
	UpdatedAt FleetTime `json:"updated_at"`
}

// ListCustomHostVitalsResponse for `GET /api/v1/fleet/custom_host_vitals`.
type ListCustomHostVitalsResponse struct {
	CustomHostVitals []CustomHostVital `json:"custom_host_vitals"`
	Meta             *ListMeta         `json:"meta"`
}

func tableFleetdmCustomHostVital(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "fleetdm_custom_host_vital",
		Description: "Custom host vitals: extra host attributes that labels can target (Fleet 4.90+).",
		List: &plugin.ListConfig{
			Hydrate: listCustomHostVitals,
		},
		Columns: []*plugin.Column{
			{Name: "id", Type: proto.ColumnType_INT, Description: "Unique ID of the custom host vital."},
			{Name: "name", Type: proto.ColumnType_STRING, Description: "Name of the custom host vital."},
			{Name: "created_at", Type: proto.ColumnType_TIMESTAMP, Transform: transform.FromField("CreatedAt").Transform(flexibleTimeTransform), Description: "When the custom host vital was created."},
			{Name: "updated_at", Type: proto.ColumnType_TIMESTAMP, Transform: transform.FromField("UpdatedAt").Transform(flexibleTimeTransform), Description: "When the custom host vital was last updated."},
		},
	}
}

func listCustomHostVitals(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := getClient(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("fleetdm_custom_host_vital.listCustomHostVitals", "connection_error", err)
		return nil, err
	}
	err = paginatedList(ctx, d, 100, func(ctx context.Context, page, perPage int) ([]CustomHostVital, *ListMeta, error) {
		params := url.Values{}
		params.Add("page", strconv.Itoa(page))
		params.Add("per_page", strconv.Itoa(perPage))
		var response ListCustomHostVitalsResponse
		if err := client.Get(ctx, "custom_host_vitals", params, &response); err != nil {
			plugin.Logger(ctx).Error("fleetdm_custom_host_vital.listCustomHostVitals", "api_error", err, "page", page)
			return nil, nil, err
		}
		return response.CustomHostVitals, response.Meta, nil
	})
	return nil, err
}
