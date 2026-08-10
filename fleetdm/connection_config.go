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
	// ExposeSecrets opts in to returning live secret material (e.g. team
	// enrollment secrets) in query results. Deliberately has no environment
	// variable fallback: widening secret exposure must be explicit in the .spc.
	ExposeSecrets *bool `cty:"expose_secrets"`
}

// ConfigSchema defines the schema for the plugin's connection configuration.
var ConfigSchema = map[string]*schema.Attribute{
	"server_url": {
		Type: schema.TypeString,
	},
	"api_token": {
		Type: schema.TypeString,
	},
	"expose_secrets": {
		Type: schema.TypeBool,
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
