package fleetdm

import (
	"context"

	"github.com/turbot/steampipe-plugin-sdk/v6/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin/transform"
)

// AppleMDM describes the Apple Push Notification service (APNs) certificate
// Fleet uses for Apple MDM. When it expires, Apple devices stop receiving
// MDM commands.
// Refer to: https://fleetdm.com/docs/rest-api/rest-api#get-apple-mdm
type AppleMDM struct {
	CommonName   string    `json:"common_name"`
	SerialNumber string    `json:"serial_number"`
	Issuer       string    `json:"issuer"`
	RenewDate    FleetTime `json:"renew_date"`
}

func tableFleetdmAppleMDM(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "fleetdm_apple_mdm",
		Description: "Single-row details of the Apple Push Notification service (APNs) certificate used for Apple MDM. No rows when Apple MDM is off.",
		List: &plugin.ListConfig{
			Hydrate: listAppleMDM,
		},
		Columns: []*plugin.Column{
			{Name: "common_name", Type: proto.ColumnType_STRING, Description: "Common name of the APNs certificate."},
			{Name: "serial_number", Type: proto.ColumnType_STRING, Description: "Serial number of the APNs certificate."},
			{Name: "issuer", Type: proto.ColumnType_STRING, Description: "Issuer of the APNs certificate."},
			{Name: "renew_date", Type: proto.ColumnType_TIMESTAMP, Transform: transform.FromField("RenewDate").Transform(flexibleTimeTransform), Description: "When the APNs certificate expires and must be renewed."},
		},
	}
}

func listAppleMDM(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := getClient(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("fleetdm_apple_mdm.listAppleMDM", "connection_error", err)
		return nil, err
	}
	var mdm AppleMDM
	if err := client.Get(ctx, "mdm/apple", nil, &mdm); err != nil {
		if handleUnavailableFeature(ctx, d, "fleetdm_apple_mdm", err) {
			return nil, nil
		}
		plugin.Logger(ctx).Error("fleetdm_apple_mdm.listAppleMDM", "api_error", err)
		return nil, err
	}
	d.StreamListItem(ctx, mdm)
	return nil, nil
}
