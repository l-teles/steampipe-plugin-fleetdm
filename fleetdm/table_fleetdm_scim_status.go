package fleetdm

import (
	"context"

	"github.com/turbot/steampipe-plugin-sdk/v6/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin/transform"
)

// ScimStatus reports the last request Fleet received from the identity
// provider's SCIM integration.
// Refer to: https://fleetdm.com/docs/rest-api/rest-api#get-scim-details
type ScimStatus struct {
	LastRequest *struct {
		Status      string    `json:"status"`
		Details     string    `json:"details"`
		RequestedAt FleetTime `json:"requested_at"`
	} `json:"last_request"`
}

func tableFleetdmScimStatus(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "fleetdm_scim_status",
		Description: "Single-row status of the identity provider's SCIM integration with Fleet (Fleet Premium).",
		List: &plugin.ListConfig{
			Hydrate: listScimStatus,
		},
		Columns: []*plugin.Column{
			{Name: "last_request_status", Type: proto.ColumnType_STRING, Transform: transform.FromField("LastRequest.Status"), Description: "Outcome of the last SCIM request ('success' or 'error'). Null if Fleet never received one."},
			{Name: "last_request_details", Type: proto.ColumnType_STRING, Transform: transform.FromField("LastRequest.Details"), Description: "Details of the last SCIM request, such as an error message."},
			{Name: "last_request_at", Type: proto.ColumnType_TIMESTAMP, Transform: transform.FromField("LastRequest.RequestedAt").Transform(flexibleTimeTransform), Description: "When the last SCIM request was received."},
		},
	}
}

func listScimStatus(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := getClient(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("fleetdm_scim_status.listScimStatus", "connection_error", err)
		return nil, err
	}
	var status ScimStatus
	if err := client.Get(ctx, "scim/details", nil, &status); err != nil {
		if handleUnavailableFeature(ctx, d, "fleetdm_scim_status", err) {
			return nil, nil
		}
		plugin.Logger(ctx).Error("fleetdm_scim_status.listScimStatus", "api_error", err)
		return nil, err
	}
	d.StreamListItem(ctx, status)
	return nil, nil
}
