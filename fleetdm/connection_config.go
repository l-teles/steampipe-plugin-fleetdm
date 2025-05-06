package fleetdm

import (
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/schema"
)

// fleetdmConfig contains the configuration for the FleetDM plugin.
// These settings are defined in a .spc file, typically ~/.steampipe/config/fleetdm.spc
type fleetdmConfig struct {
	ServerURL *string `cty:"server_url"`
	APIToken  *string `cty:"api_token"`
}

// ConfigSchema defines the schema for the plugin's connection configuration.
var ConfigSchema = map[string]*schema.Attribute{
	"server_url": {
		Type: schema.TypeString,
	},
	"api_token": {
		Type: schema.TypeString,
	},
}

// ConfigInstance returns a new instance of the fleetdmConfig struct.
func ConfigInstance() interface{} {
	return &fleetdmConfig{}
}

// GetConfig extracts and validates the fleetdmConfig from the connection configuration.
func GetConfig(connection *plugin.Connection) fleetdmConfig {
	if connection == nil || connection.Config == nil {
		return fleetdmConfig{}
	}
	config, _ := connection.Config.(fleetdmConfig)
	return config
}
