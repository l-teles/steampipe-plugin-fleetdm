package fleetdm

import (
	"context"
	"fmt"
	"net/url"

	"github.com/turbot/steampipe-plugin-sdk/v6/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin/transform"
)

// ScriptResult is the outcome of one script execution on a host. Contents and
// output can hold anything the script handles, so they are gated.
// Refer to: https://fleetdm.com/docs/rest-api/rest-api#get-script-result
type ScriptResult struct {
	ExecutionID    string    `json:"execution_id"`
	HostID         uint      `json:"host_id"`
	Hostname       string    `json:"hostname"`
	ScriptID       *uint     `json:"script_id"`
	ExitCode       *int64    `json:"exit_code"`
	Runtime        int       `json:"runtime"`
	HostTimeout    bool      `json:"host_timeout"`
	Message        string    `json:"message"`
	CreatedAt      FleetTime `json:"created_at"`
	ScriptContents string    `json:"script_contents"`
	Output         string    `json:"output"`
}

func tableFleetdmScriptResult(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "fleetdm_script_result",
		Description: "Result of a script execution. execution_id is required (see fleetdm_host_activity.script_execution_id); script_contents and output need expose_secrets = true.",
		List: &plugin.ListConfig{
			Hydrate: listScriptResult,
			KeyColumns: []*plugin.KeyColumn{
				{Name: "execution_id", Require: plugin.Required},
			},
		},
		Columns: []*plugin.Column{
			{Name: "execution_id", Type: proto.ColumnType_STRING, Transform: transform.FromField("ExecutionID"), Description: "ID of the script execution. Required in WHERE clause."},
			{Name: "host_id", Type: proto.ColumnType_INT, Description: "ID of the host the script ran on."},
			{Name: "hostname", Type: proto.ColumnType_STRING, Description: "Hostname of the host."},
			{Name: "script_id", Type: proto.ColumnType_INT, Transform: transform.FromField("ScriptID"), Description: "ID of the saved script; null for ad-hoc scripts."},
			{Name: "exit_code", Type: proto.ColumnType_INT, Description: "Exit code of the script; null while pending."},
			{Name: "runtime", Type: proto.ColumnType_INT, Description: "Runtime in seconds."},
			{Name: "host_timeout", Type: proto.ColumnType_BOOL, Description: "Whether the host did not respond in time."},
			{Name: "message", Type: proto.ColumnType_STRING, Description: "Status message from Fleet."},
			{Name: "created_at", Type: proto.ColumnType_TIMESTAMP, Transform: transform.FromField("CreatedAt").Transform(flexibleTimeTransform), Description: "When the script was queued."},
			{Name: "script_contents", Type: proto.ColumnType_STRING, Hydrate: itemIfSecretsExposed, Transform: transform.FromField("ScriptContents"), Description: "The script that ran. NULL unless expose_secrets = true."},
			{Name: "output", Type: proto.ColumnType_STRING, Hydrate: itemIfSecretsExposed, Transform: transform.FromField("Output"), Description: "Output of the script. NULL unless expose_secrets = true."},
		},
	}
}

func listScriptResult(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := getClient(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("fleetdm_script_result.listScriptResult", "connection_error", err)
		return nil, err
	}
	id := d.EqualsQuals["execution_id"].GetStringValue()
	var result ScriptResult
	if err := client.Get(ctx, fmt.Sprintf("scripts/results/%s", url.PathEscape(id)), nil, &result); err != nil {
		plugin.Logger(ctx).Error("fleetdm_script_result.listScriptResult", "api_error", err)
		return nil, err
	}
	d.StreamListItem(ctx, result)
	return nil, nil
}
