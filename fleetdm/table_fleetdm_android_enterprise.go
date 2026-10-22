package fleetdm

import (
	"context"

	"github.com/turbot/steampipe-plugin-sdk/v6/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin/transform"
)

// AndroidEnterprise is the Android Enterprise Fleet is bound to for Android
// MDM.
// Refer to: https://fleetdm.com/docs/rest-api/rest-api#get-android-enterprise
type AndroidEnterprise struct {
	EnterpriseID string `json:"android_enterprise_id"`
}

func tableFleetdmAndroidEnterprise(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "fleetdm_android_enterprise",
		Description: "Single row with the Android Enterprise Fleet uses for Android MDM. No rows when Android MDM is not set up.",
		List: &plugin.ListConfig{
			Hydrate: listAndroidEnterprise,
		},
		Columns: []*plugin.Column{
			{Name: "android_enterprise_id", Type: proto.ColumnType_STRING, Transform: transform.FromField("EnterpriseID"), Description: "ID of the Android Enterprise."},
		},
	}
}

func listAndroidEnterprise(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := getClient(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("fleetdm_android_enterprise.listAndroidEnterprise", "connection_error", err)
		return nil, err
	}
	var e AndroidEnterprise
	if err := client.Get(ctx, "android_enterprise", nil, &e); err != nil {
		// 404: no enterprise is bound.
		if is404(err) || handleUnavailableFeature(ctx, d, "fleetdm_android_enterprise", err) {
			return nil, nil
		}
		plugin.Logger(ctx).Error("fleetdm_android_enterprise.listAndroidEnterprise", "api_error", err)
		return nil, err
	}
	if e.EnterpriseID != "" {
		d.StreamListItem(ctx, e)
	}
	return nil, nil
}
