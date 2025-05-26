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
			"fleetdm_activity": tableFleetdmActivity(ctx),
			"fleetdm_host":     tableFleetdmHost(ctx),
			"fleetdm_label":    tableFleetdmLabel(ctx),
			"fleetdm_pack":     tableFleetdmPack(ctx),
			"fleetdm_policy":   tableFleetdmPolicy(ctx),
			"fleetdm_query":    tableFleetdmQuery(ctx),
			"fleetdm_software": tableFleetdmSoftware(ctx),
			"fleetdm_team":     tableFleetdmTeam(ctx),
			"fleetdm_user":     tableFleetdmUser(ctx),
		},
	}
	return p
}
