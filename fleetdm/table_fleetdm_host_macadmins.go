package fleetdm

import (
	"context"
	"fmt"

	"github.com/turbot/steampipe-plugin-sdk/v6/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin/transform"
)

// HostMacadmins is the macadmins data for a host: Munki (macOS) and the MDM
// enrollment the host itself reports (macOS and Windows).
// Refer to: https://fleetdm.com/docs/rest-api/rest-api#get-hosts-macadmins-data
type HostMacadmins struct {
	Munki *struct {
		Version string `json:"version"`
	} `json:"munki"`
	MunkiIssues            interface{} `json:"munki_issues"`
	MobileDeviceManagement *struct {
		EnrollmentStatus string `json:"enrollment_status"`
		ServerURL        string `json:"server_url"`
		Name             string `json:"name"`
		ID               *uint  `json:"id"`
	} `json:"mobile_device_management"`
	HostID   uint   `json:"-"`
	Hostname string `json:"-"`
}

// GetHostMacadminsResponse for `GET /api/v1/fleet/hosts/{id}/macadmins`;
// macadmins is null for hosts without such data (e.g. Linux).
type GetHostMacadminsResponse struct {
	Macadmins *HostMacadmins `json:"macadmins"`
}

func tableFleetdmHostMacadmins(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "fleetdm_host_macadmins",
		Description: "Munki (macOS) and self-reported MDM enrollment (macOS, Windows) per host. Without a host_id filter every host is queried; hosts without such data (e.g. Linux) have no row.",
		List: &plugin.ListConfig{
			Hydrate: listHostMacadmins,
			KeyColumns: []*plugin.KeyColumn{
				{Name: "host_id", Require: plugin.Optional},
			},
		},
		Columns: []*plugin.Column{
			{Name: "host_id", Type: proto.ColumnType_INT, Description: "ID of the host."},
			{Name: "hostname", Type: proto.ColumnType_STRING, Description: "Hostname of the host. Empty when filtering by host_id."},
			{Name: "munki_version", Type: proto.ColumnType_STRING, Transform: transform.FromField("Munki.Version"), Description: "Munki version, if Munki is installed."},
			{Name: "munki_issues", Type: proto.ColumnType_JSON, Description: "Munki errors and warnings reported by the host."},
			{Name: "mdm_enrollment_status", Type: proto.ColumnType_STRING, Transform: transform.FromField("MobileDeviceManagement.EnrollmentStatus"), Description: "MDM enrollment status the host reports, e.g. 'On (manual)', 'On (automatic)', 'Off'."},
			{Name: "mdm_server_url", Type: proto.ColumnType_STRING, Transform: transform.FromField("MobileDeviceManagement.ServerURL"), Description: "MDM server the host is enrolled to."},
			{Name: "mdm_name", Type: proto.ColumnType_STRING, Transform: transform.FromField("MobileDeviceManagement.Name"), Description: "Name of the MDM solution."},
		},
	}
}

func listHostMacadmins(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := getClient(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("fleetdm_host_macadmins.listHostMacadmins", "connection_error", err)
		return nil, err
	}
	err = streamPerHost(ctx, d, client, func(ctx context.Context, host hostRef) ([]HostMacadmins, error) {
		var response GetHostMacadminsResponse
		if err := client.Get(ctx, fmt.Sprintf("hosts/%d/macadmins", host.ID), nil, &response); err != nil {
			return nil, err
		}
		if response.Macadmins == nil {
			return nil, nil
		}
		m := *response.Macadmins
		m.HostID, m.Hostname = host.ID, host.Hostname
		return []HostMacadmins{m}, nil
	})
	if err != nil {
		plugin.Logger(ctx).Error("fleetdm_host_macadmins.listHostMacadmins", "api_error", err)
	}
	return nil, err
}
