package main

import (

	"steampipe-plugin-fleetdm/fleetdm"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"


)

func main() {
	plugin.Serve(&plugin.ServeOpts{PluginFunc: fleetdm.Plugin})
}
