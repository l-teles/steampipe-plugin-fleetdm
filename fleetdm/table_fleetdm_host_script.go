package fleetdm

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/turbot/steampipe-plugin-sdk/v6/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin/transform"
)

// HostScript is a script available to a host, with its last execution.
// Refer to: https://fleetdm.com/docs/rest-api/rest-api#list-hosts-scripts
type HostScript struct {
	ScriptID      uint        `json:"script_id"`
	Name          string      `json:"name"`
	LastExecution interface{} `json:"last_execution"`
	HostID        uint        `json:"-"`
	Hostname      string      `json:"-"`
}

// ListHostScriptsResponse for `GET /api/v1/fleet/hosts/{id}/scripts`.
type ListHostScriptsResponse struct {
	Scripts []HostScript `json:"scripts"`
	Meta    *ListMeta    `json:"meta"`
}

func tableFleetdmHostScript(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "fleetdm_host_script",
		Description: "Scripts available to each host and their last execution. Without a host_id filter every host is queried.",
		List: &plugin.ListConfig{
			Hydrate: listHostScripts,
			KeyColumns: []*plugin.KeyColumn{
				{Name: "host_id", Require: plugin.Optional},
			},
		},
		Columns: []*plugin.Column{
			{Name: "host_id", Type: proto.ColumnType_INT, Description: "ID of the host."},
			{Name: "hostname", Type: proto.ColumnType_STRING, Description: "Hostname of the host. Empty when filtering by host_id."},
			{Name: "script_id", Type: proto.ColumnType_INT, Transform: transform.FromField("ScriptID"), Description: "ID of the script."},
			{Name: "name", Type: proto.ColumnType_STRING, Description: "File name of the script."},
			{Name: "last_execution", Type: proto.ColumnType_JSON, Description: "Last execution on this host (execution ID, time, status), or null if never run."},
		},
	}
}

func listHostScripts(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := getClient(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("fleetdm_host_script.listHostScripts", "connection_error", err)
		return nil, err
	}
	err = streamPerHost(ctx, d, client, func(ctx context.Context, host hostRef) ([]HostScript, error) {
		var rows []HostScript
		err := paginateCore(ctx, 100, func(ctx context.Context, page, perPage int) ([]HostScript, *ListMeta, error) {
			params := url.Values{}
			params.Add("page", strconv.Itoa(page))
			params.Add("per_page", strconv.Itoa(perPage))
			var response ListHostScriptsResponse
			if err := client.Get(ctx, fmt.Sprintf("hosts/%d/scripts", host.ID), params, &response); err != nil {
				return nil, nil, err
			}
			return response.Scripts, response.Meta, nil
		}, func(s HostScript) bool {
			s.HostID, s.Hostname = host.ID, host.Hostname
			rows = append(rows, s)
			return true
		})
		return rows, err
	})
	if err != nil {
		if handleUnavailableFeature(ctx, d, "fleetdm_host_script", err) {
			return nil, nil
		}
		plugin.Logger(ctx).Error("fleetdm_host_script.listHostScripts", "api_error", err)
	}
	return nil, err
}
