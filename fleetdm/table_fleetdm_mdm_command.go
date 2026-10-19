package fleetdm

import (
	"context"
	"encoding/base64"
	"net/url"
	"strconv"

	"github.com/turbot/steampipe-plugin-sdk/v6/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin/transform"
)

// mdmCommandMaxPage is the last page Fleet serves for /commands; beyond it
// the API requires cursor paging.
const mdmCommandMaxPage = 100

// MDMCommand is an MDM command sent to a host.
// Refer to: https://fleetdm.com/docs/rest-api/rest-api#list-mdm-commands
type MDMCommand struct {
	HostUUID      string    `json:"host_uuid"`
	Hostname      string    `json:"hostname"`
	CommandUUID   string    `json:"command_uuid"`
	RequestType   string    `json:"request_type"`
	Name          *string   `json:"name"`
	Status        string    `json:"status"`
	CommandStatus string    `json:"command_status"`
	UpdatedAt     FleetTime `json:"updated_at"`
}

// ListMDMCommandsResponse for `GET /api/v1/fleet/commands`.
type ListMDMCommandsResponse struct {
	Results []MDMCommand `json:"results"`
	Meta    *ListMeta    `json:"meta"`
}

func tableFleetdmMDMCommand(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "fleetdm_mdm_command",
		Description: "MDM commands sent to a host. host_identifier (hostname, UUID or serial) is required.",
		List: &plugin.ListConfig{
			Hydrate: listMDMCommands,
			KeyColumns: []*plugin.KeyColumn{
				{Name: "host_identifier", Require: plugin.Required},
				{Name: "request_type", Require: plugin.Optional},
				{Name: "command_status", Require: plugin.Optional},
			},
		},
		Columns: []*plugin.Column{
			{Name: "host_identifier", Type: proto.ColumnType_STRING, Transform: transform.FromQual("host_identifier"), Description: "Host the commands were sent to: hostname, UUID or hardware serial. Required in WHERE clause."},
			{Name: "host_uuid", Type: proto.ColumnType_STRING, Transform: transform.FromField("HostUUID"), Description: "UUID of the host."},
			{Name: "hostname", Type: proto.ColumnType_STRING, Description: "Hostname of the host."},
			{Name: "command_uuid", Type: proto.ColumnType_STRING, Transform: transform.FromField("CommandUUID"), Description: "Unique ID of the command; use it to look up fleetdm_mdm_command_result."},
			{Name: "request_type", Type: proto.ColumnType_STRING, Description: "MDM request type, e.g. 'InstallProfile', 'DeviceLock', './Device/Vendor/MSFT/Reboot/RebootNow'."},
			{Name: "name", Type: proto.ColumnType_STRING, Description: "Profile name for InstallProfile/RemoveProfile commands Fleet enqueued (4.84+); null otherwise."},
			{Name: "status", Type: proto.ColumnType_STRING, Description: "Raw status from the device, e.g. 'Acknowledged', 'Pending', 'NotNow', 'Error'."},
			{Name: "command_status", Type: proto.ColumnType_STRING, Description: "Normalized status: 'ran', 'pending' or 'failed'. Filtering on it is done server-side for Apple and Android hosts."},
			{Name: "updated_at", Type: proto.ColumnType_TIMESTAMP, Transform: transform.FromField("UpdatedAt").Transform(flexibleTimeTransform), Description: "When the command's status last changed."},
		},
	}
}

func listMDMCommands(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := getClient(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("fleetdm_mdm_command.listMDMCommands", "connection_error", err)
		return nil, err
	}
	err = paginatedList(ctx, d, 1000, func(ctx context.Context, page, perPage int) ([]MDMCommand, *ListMeta, error) {
		if page >= mdmCommandMaxPage {
			plugin.Logger(ctx).Warn("fleetdm_mdm_command.listMDMCommands", "page_limit_reached", mdmCommandMaxPage)
			return nil, nil, nil
		}
		params := url.Values{}
		params.Add("page", strconv.Itoa(page))
		params.Add("per_page", strconv.Itoa(perPage))
		params.Add("host_identifier", d.EqualsQuals["host_identifier"].GetStringValue())
		if q := d.EqualsQuals["request_type"]; q != nil {
			params.Add("request_type", q.GetStringValue())
		}
		if q := d.EqualsQuals["command_status"]; q != nil {
			params.Add("command_status", q.GetStringValue())
		}
		var response ListMDMCommandsResponse
		if err := client.Get(ctx, "commands", params, &response); err != nil {
			plugin.Logger(ctx).Error("fleetdm_mdm_command.listMDMCommands", "api_error", err, "page", page)
			return nil, nil, err
		}
		return response.Results, response.Meta, nil
	})
	if err != nil && handleUnavailableFeature(ctx, d, "fleetdm_mdm_command", err) {
		return nil, nil
	}
	return nil, err
}

// MDMCommandResult is a host's response to an MDM command. Payload and
// result can carry sensitive values (e.g. a lock PIN), so they are gated.
// Refer to: https://fleetdm.com/docs/rest-api/rest-api#get-mdm-command-results
type MDMCommandResult struct {
	HostUUID        string      `json:"host_uuid"`
	Hostname        string      `json:"hostname"`
	CommandUUID     string      `json:"command_uuid"`
	RequestType     string      `json:"request_type"`
	Name            *string     `json:"name"`
	Status          string      `json:"status"`
	UpdatedAt       FleetTime   `json:"updated_at"`
	Payload         string      `json:"payload"` // base64
	Result          string      `json:"result"`  // base64
	ResultsMetadata interface{} `json:"results_metadata"`
}

// GetMDMCommandResultsResponse for `GET /api/v1/fleet/commands/results`.
type GetMDMCommandResultsResponse struct {
	Results []MDMCommandResult `json:"results"`
}

func tableFleetdmMDMCommandResult(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "fleetdm_mdm_command_result",
		Description: "Results of an MDM command on each targeted host. command_uuid is required; payload and result need expose_secrets = true.",
		List: &plugin.ListConfig{
			Hydrate: listMDMCommandResults,
			KeyColumns: []*plugin.KeyColumn{
				{Name: "command_uuid", Require: plugin.Required},
			},
		},
		Columns: []*plugin.Column{
			{Name: "command_uuid", Type: proto.ColumnType_STRING, Transform: transform.FromField("CommandUUID"), Description: "Unique ID of the command. Required in WHERE clause."},
			{Name: "host_uuid", Type: proto.ColumnType_STRING, Transform: transform.FromField("HostUUID"), Description: "UUID of the host."},
			{Name: "hostname", Type: proto.ColumnType_STRING, Description: "Hostname of the host."},
			{Name: "request_type", Type: proto.ColumnType_STRING, Description: "MDM request type."},
			{Name: "name", Type: proto.ColumnType_STRING, Description: "Profile name for profile commands Fleet enqueued; null otherwise."},
			{Name: "status", Type: proto.ColumnType_STRING, Description: "Status reported by the device, e.g. 'Acknowledged', 'Error'."},
			{Name: "updated_at", Type: proto.ColumnType_TIMESTAMP, Transform: transform.FromField("UpdatedAt").Transform(flexibleTimeTransform), Description: "When the result was received."},
			{Name: "results_metadata", Type: proto.ColumnType_JSON, Description: "Parsed metadata about the result, when Fleet provides it."},
			{Name: "payload", Type: proto.ColumnType_STRING, Hydrate: itemIfSecretsExposed, Transform: transform.FromField("Payload").Transform(base64ToString), Description: "Command sent to the device (decoded). NULL unless expose_secrets = true."},
			{Name: "result", Type: proto.ColumnType_STRING, Hydrate: itemIfSecretsExposed, Transform: transform.FromField("Result").Transform(base64ToString), Description: "Raw response from the device (decoded). NULL unless expose_secrets = true."},
		},
	}
}

// base64ToString decodes a base64 API field ([]byte in Fleet) to text,
// leaving values that are not base64 as they are.
func base64ToString(_ context.Context, d *transform.TransformData) (interface{}, error) {
	s, ok := d.Value.(string)
	if !ok || s == "" {
		return nil, nil
	}
	if b, err := base64.StdEncoding.DecodeString(s); err == nil {
		return string(b), nil
	}
	return s, nil
}

func listMDMCommandResults(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := getClient(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("fleetdm_mdm_command_result.listMDMCommandResults", "connection_error", err)
		return nil, err
	}
	params := url.Values{}
	params.Add("command_uuid", d.EqualsQuals["command_uuid"].GetStringValue())
	var response GetMDMCommandResultsResponse
	if err := client.Get(ctx, "commands/results", params, &response); err != nil {
		if handleUnavailableFeature(ctx, d, "fleetdm_mdm_command_result", err) {
			return nil, nil
		}
		plugin.Logger(ctx).Error("fleetdm_mdm_command_result.listMDMCommandResults", "api_error", err)
		return nil, err
	}
	for _, r := range response.Results {
		d.StreamListItem(ctx, r)
		if d.RowsRemaining(ctx) == 0 {
			break
		}
	}
	return nil, nil
}
