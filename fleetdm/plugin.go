package fleetdm

import (
	"context"

	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

// Plugin returns the FleetDM plugin.
func Plugin(ctx context.Context) *plugin.Plugin {
	p := &plugin.Plugin{
		Name: "steampipe-plugin-fleetdm",
		ConnectionConfigSchema: &plugin.ConnectionConfigSchema{
			NewInstance: ConfigInstance,
			Schema:      ConfigSchema,
		},
		DefaultTransform: transform.FromGo().NullIfZero(),
		TableMap: map[string]*plugin.Table{
			"fleetdm_activity":             tableFleetdmActivity(ctx),
			"fleetdm_app_store_app":        tableFleetdmAppStoreApp(ctx),
			"fleetdm_carve":                tableFleetdmCarve(ctx),
			"fleetdm_fleet_maintained_app": tableFleetdmFleetMaintainedApp(ctx),
			"fleetdm_host":                 tableFleetdmHost(ctx),
			"fleetdm_host_detail":          tableFleetdmHostDetail(ctx),
			"fleetdm_label":                tableFleetdmLabel(ctx),
			"fleetdm_os_version":           tableFleetdmOSVersion(ctx),
			"fleetdm_pack":                 tableFleetdmPack(ctx),
			"fleetdm_policy":               tableFleetdmPolicy(ctx),
			"fleetdm_query":                tableFleetdmQuery(ctx),
			"fleetdm_software_title":       tableFleetdmSoftwareTitle(ctx),
			"fleetdm_software_version":     tableFleetdmSoftwareVersion(ctx),
			"fleetdm_team":                 tableFleetdmTeam(ctx),
			"fleetdm_user":                 tableFleetdmUser(ctx),
		},
	}
	return p
}
