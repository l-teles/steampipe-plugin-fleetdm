package main

import (
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin"
	"steampipe-plugin-fleetdm/fleetdm"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{PluginFunc: fleetdm.Plugin})
}
