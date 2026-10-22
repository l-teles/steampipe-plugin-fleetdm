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

// HostSoftwareItem is one software title installed on (or available to) a host.
// Refer to: https://fleetdm.com/docs/rest-api/rest-api#get-hosts-software
type HostSoftwareItem struct {
	ID                uint        `json:"id"`
	Name              string      `json:"name"`
	DisplayName       string      `json:"display_name"`
	Source            string      `json:"source"`
	ExtensionFor      string      `json:"extension_for"`
	BundleIdentifier  *string     `json:"bundle_identifier"`
	IconURL           *string     `json:"icon_url"`
	Status            *string     `json:"status"` // install status of a Fleet-managed installer
	InstalledVersions interface{} `json:"installed_versions"`
	SoftwarePackage   interface{} `json:"software_package"`
	AppStoreApp       interface{} `json:"app_store_app"`
	HostID            uint        `json:"-"`
	Hostname          string      `json:"-"`
}

// ListHostSoftwareResponse for `GET /api/v1/fleet/hosts/{id}/software`.
type ListHostSoftwareResponse struct {
	Software []HostSoftwareItem `json:"software"`
	Meta     *ListMeta          `json:"meta"`
}

func tableFleetdmHostSoftware(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "fleetdm_host_software",
		Description: "Software on each host, with installed versions, paths and vulnerabilities. Without a host_id filter every host is queried.",
		List: &plugin.ListConfig{
			Hydrate: listHostSoftware,
			KeyColumns: []*plugin.KeyColumn{
				{Name: "host_id", Require: plugin.Optional},
			},
		},
		Columns: []*plugin.Column{
			{Name: "host_id", Type: proto.ColumnType_INT, Description: "ID of the host."},
			{Name: "hostname", Type: proto.ColumnType_STRING, Description: "Hostname of the host. Empty when filtering by host_id."},
			{Name: "software_title_id", Type: proto.ColumnType_INT, Transform: transform.FromField("ID"), Description: "ID of the software title."},
			{Name: "name", Type: proto.ColumnType_STRING, Description: "Name of the software."},
			{Name: "display_name", Type: proto.ColumnType_STRING, Description: "Custom display name of the software, if set."},
			{Name: "source", Type: proto.ColumnType_STRING, Description: "Source of the software (e.g. 'apps', 'programs', 'deb_packages', 'chrome_extensions')."},
			{Name: "extension_for", Type: proto.ColumnType_STRING, Description: "Browser or app the extension belongs to, for extensions."},
			{Name: "bundle_identifier", Type: proto.ColumnType_STRING, Description: "Bundle identifier (macOS)."},
			{Name: "status", Type: proto.ColumnType_STRING, Description: "Install status of a Fleet-managed installer, e.g. 'installed', 'pending_install', 'failed_install'."},
			{Name: "installed_versions", Type: proto.ColumnType_JSON, Description: "Installed versions with paths, signatures, last opened time and vulnerabilities."},
			{Name: "software_package", Type: proto.ColumnType_JSON, Description: "Fleet-managed installer package details, if any. package_url is null unless expose_secrets = true, as it can embed credentials."},
			{Name: "app_store_app", Type: proto.ColumnType_JSON, Description: "App Store (VPP) app details, if any."},
			{Name: "icon_url", Type: proto.ColumnType_STRING, Transform: transform.FromField("IconURL"), Description: "URL of the software icon."},
		},
	}
}

func listHostSoftware(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := getClient(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("fleetdm_host_software.listHostSoftware", "connection_error", err)
		return nil, err
	}
	exposed := secretsExposed(d)
	rowCap := perHostRowCap(d) // stop paging a host once LIMIT rows are in
	err = streamPerHost(ctx, d, client, func(ctx context.Context, host hostRef) ([]HostSoftwareItem, error) {
		var rows []HostSoftwareItem
		err := paginateCore(ctx, 100, func(ctx context.Context, page, perPage int) ([]HostSoftwareItem, *ListMeta, error) {
			params := url.Values{}
			params.Add("page", strconv.Itoa(page))
			params.Add("per_page", strconv.Itoa(perPage))
			var response ListHostSoftwareResponse
			if err := client.Get(ctx, fmt.Sprintf("hosts/%d/software", host.ID), params, &response); err != nil {
				return nil, nil, err
			}
			return response.Software, response.Meta, nil
		}, func(s HostSoftwareItem) bool {
			s.HostID, s.Hostname = host.ID, host.Hostname
			if !exposed {
				s.SoftwarePackage = nullURLKeys(s.SoftwarePackage, installerURLKeys...)
			}
			rows = append(rows, s)
			return rowCap == 0 || len(rows) < rowCap
		})
		return rows, err
	})
	if err != nil {
		plugin.Logger(ctx).Error("fleetdm_host_software.listHostSoftware", "api_error", err)
	}
	return nil, err
}
