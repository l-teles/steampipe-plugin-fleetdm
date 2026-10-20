package fleetdm

import (
	"context"
	"fmt"
	"net/url"

	"github.com/turbot/steampipe-plugin-sdk/v6/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin/transform"
)

// SoftwareInstallResult is the outcome of installing a software package on a
// host. Script and query outputs are gated.
// Refer to: https://fleetdm.com/docs/rest-api/rest-api#get-software-install-result
type SoftwareInstallResult struct {
	InstallUUID             string     `json:"install_uuid"`
	HostID                  uint       `json:"host_id"`
	SoftwareTitle           string     `json:"software_title"`
	SoftwareTitleID         *uint      `json:"software_title_id"`
	SoftwarePackage         string     `json:"software_package"`
	HashSHA256              *string    `json:"hash_sha256"`
	Source                  *string    `json:"source"`
	Status                  string     `json:"status"`
	SelfService             bool       `json:"self_service"`
	PolicyID                *uint      `json:"policy_id"`
	AttemptNumber           *int       `json:"attempt_number"`
	CreatedAt               FleetTime  `json:"created_at"`
	UpdatedAt               *FleetTime `json:"updated_at"`
	Output                  *string    `json:"output"`
	PreInstallQueryOutput   *string    `json:"pre_install_query_output"`
	PostInstallScriptOutput *string    `json:"post_install_script_output"`
}

// GetSoftwareInstallResultResponse for `GET /api/v1/fleet/software/install/{uuid}/results`.
type GetSoftwareInstallResultResponse struct {
	Results *SoftwareInstallResult `json:"results"`
}

func tableFleetdmSoftwareInstallResult(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "fleetdm_software_install_result",
		Description: "Result of a software install on a host. install_uuid is required (see fleetdm_host_activity.install_uuid); outputs need expose_secrets = true.",
		List: &plugin.ListConfig{
			Hydrate: listSoftwareInstallResult,
			KeyColumns: []*plugin.KeyColumn{
				{Name: "install_uuid", Require: plugin.Required},
			},
		},
		Columns: []*plugin.Column{
			{Name: "install_uuid", Type: proto.ColumnType_STRING, Transform: transform.FromField("InstallUUID"), Description: "ID of the install. Required in WHERE clause."},
			{Name: "host_id", Type: proto.ColumnType_INT, Description: "ID of the host."},
			{Name: "software_title", Type: proto.ColumnType_STRING, Description: "Name of the software title."},
			{Name: "software_title_id", Type: proto.ColumnType_INT, Description: "ID of the software title."},
			{Name: "software_package", Type: proto.ColumnType_STRING, Description: "Installer package file name."},
			{Name: "hash_sha256", Type: proto.ColumnType_STRING, Transform: transform.FromField("HashSHA256"), Description: "SHA-256 hash of the installer package."},
			{Name: "source", Type: proto.ColumnType_STRING, Description: "Source of the software."},
			{Name: "status", Type: proto.ColumnType_STRING, Description: "Install status, e.g. 'pending_install', 'installed', 'failed_install'."},
			{Name: "self_service", Type: proto.ColumnType_BOOL, Description: "Whether the end user started the install from Fleet Desktop."},
			{Name: "policy_id", Type: proto.ColumnType_INT, Description: "Policy whose automation triggered the install, if any."},
			{Name: "attempt_number", Type: proto.ColumnType_INT, Description: "Which attempt this result is, when Fleet retried the install."},
			{Name: "created_at", Type: proto.ColumnType_TIMESTAMP, Transform: transform.FromField("CreatedAt").Transform(flexibleTimeTransform), Description: "When the install was requested."},
			{Name: "updated_at", Type: proto.ColumnType_TIMESTAMP, Transform: transform.FromField("UpdatedAt").Transform(flexibleTimeTransform), Description: "When the install status last changed."},
			{Name: "output", Type: proto.ColumnType_STRING, Hydrate: itemIfSecretsExposed, Transform: transform.FromField("Output"), Description: "Install script output. NULL unless expose_secrets = true."},
			{Name: "pre_install_query_output", Type: proto.ColumnType_STRING, Hydrate: itemIfSecretsExposed, Transform: transform.FromField("PreInstallQueryOutput"), Description: "Pre-install query output. NULL unless expose_secrets = true."},
			{Name: "post_install_script_output", Type: proto.ColumnType_STRING, Hydrate: itemIfSecretsExposed, Transform: transform.FromField("PostInstallScriptOutput"), Description: "Post-install script output. NULL unless expose_secrets = true."},
		},
	}
}

func listSoftwareInstallResult(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := getClient(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("fleetdm_software_install_result.listSoftwareInstallResult", "connection_error", err)
		return nil, err
	}
	id := d.EqualsQuals["install_uuid"].GetStringValue()
	if !uuidPattern.MatchString(id) {
		return nil, nil
	}
	var response GetSoftwareInstallResultResponse
	if err := client.Get(ctx, fmt.Sprintf("software/install/%s/results", url.PathEscape(id)), nil, &response); err != nil {
		// Software installers are a Fleet Premium feature.
		if handleUnavailableFeature(ctx, d, "fleetdm_software_install_result", err) {
			return nil, nil
		}
		plugin.Logger(ctx).Error("fleetdm_software_install_result.listSoftwareInstallResult", "api_error", err)
		return nil, err
	}
	if response.Results != nil {
		d.StreamListItem(ctx, *response.Results)
	}
	return nil, nil
}
