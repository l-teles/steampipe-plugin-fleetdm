package fleetdm

import (
	"context"
	"fmt"
	"maps"
	"net/url"
	"strconv"

	"github.com/turbot/steampipe-plugin-sdk/v6/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin/transform"
)

// SoftwareTitleVersion represents a version entry within a software title.
type SoftwareTitleVersion struct {
	ID              uint     `json:"id"`
	Version         string   `json:"version"`
	Release         string   `json:"release,omitempty"` // e.g. RPM release or the Go toolchain of a Go binary
	Vulnerabilities []string `json:"vulnerabilities"`   // List of CVE strings
	HostsCount      *uint    `json:"hosts_count,omitempty"`
}

// SoftwareTitle represents a software title in FleetDM.
// Refer to: https://fleetdm.com/docs/rest-api/rest-api#list-software
type SoftwareTitle struct {
	ID               uint                   `json:"id"`
	Name             string                 `json:"name"`
	DisplayName      string                 `json:"display_name"`
	IconURL          *string                `json:"icon_url"`
	Source           string                 `json:"source"`
	ExtensionFor     string                 `json:"extension_for"`
	Browser          string                 `json:"browser"`
	HostsCount       uint                   `json:"hosts_count"`
	VersionsCount    uint                   `json:"versions_count"`
	Versions         []SoftwareTitleVersion `json:"versions"`
	SoftwarePackage  interface{}            `json:"software_package"`
	Packages         interface{}            `json:"packages"`     // All installer packages, first-added first
	UpgradeCode      *string                `json:"upgrade_code"` // Windows MSI upgrade code
	AppStoreApp      interface{}            `json:"app_store_app"`
	AppStoreApps     interface{}            `json:"app_store_apps"` // Every App Store version on the team (Fleet 4.94+)
	BundleIdentifier *string                `json:"bundle_identifier"`
	CountsUpdatedAt  *FleetTime             `json:"counts_updated_at"`
}

// installerURLKeys are the installer fields that hold download or source URLs,
// which can embed credentials.
var installerURLKeys = []string{"package_url", "url"}

// redactInstallerURLs nulls installer URLs, and the managed configuration of
// App Store apps (which can hold app credentials), unless expose_secrets is set.
func redactInstallerURLs(d *plugin.QueryData, t *SoftwareTitle) {
	if secretsExposed(d) {
		return
	}
	t.SoftwarePackage = nullURLKeys(t.SoftwarePackage, installerURLKeys...)
	t.Packages = nullURLKeysEach(t.Packages, installerURLKeys...)
	t.AppStoreApp = nullURLKeys(t.AppStoreApp, "configuration")
	t.AppStoreApps = nullURLKeysEach(t.AppStoreApps, "configuration")
}

// ListSoftwareTitlesResponse is the expected structure for the list software titles API call.
type ListSoftwareTitlesResponse struct {
	SoftwareTitles  []SoftwareTitle `json:"software_titles"`
	Meta            *ListMeta       `json:"meta"`
	Count           int             `json:"count"`
	CountsUpdatedAt *FleetTime      `json:"counts_updated_at"`
}

func tableFleetdmSoftwareTitle(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "fleetdm_software_title",
		Description: "Software titles from FleetDM. A software title groups multiple versions of the same software. Uses the /software/titles endpoint.",
		List: &plugin.ListConfig{
			Hydrate: listSoftwareTitles,
			KeyColumns: append([]*plugin.KeyColumn{
				{Name: "id", Require: plugin.Optional},                            // Alone or with team_id, fetched directly
				{Name: "vulnerable_only", Require: plugin.Optional},               // Filter for vulnerable software
				{Name: "team_id", Require: plugin.Optional},                       // Filter by team ID (Fleet Premium)
				{Name: "available_for_install", Require: plugin.Optional},         // Filter for installable software
				{Name: "query", Require: plugin.Optional},                         // Search by title or CVE
				{Name: "self_service", Require: plugin.Optional},                  // Filter for self-service software
				{Name: "packages_only", Require: plugin.Optional},                 // Exclude app store apps (Fleet Premium)
				{Name: "min_cvss_score", Require: plugin.Optional},                // Min CVSS v3.x base score (Fleet Premium)
				{Name: "max_cvss_score", Require: plugin.Optional},                // Max CVSS v3.x base score (Fleet Premium)
				{Name: "exploit", Require: plugin.Optional},                       // Filter for CISA known exploits (Fleet Premium)
				{Name: "platform", Require: plugin.Optional},                      // Filter by platform (requires team_id)
				{Name: "exclude_fleet_maintained_apps", Require: plugin.Optional}, // Exclude Fleet-maintained apps
			}, softwareTypeKeyColumns()...), // source and extension_for (Fleet 4.94+)
		},
		Columns: []*plugin.Column{
			// Core software title information
			{Name: "id", Type: proto.ColumnType_INT, Description: "Unique ID of the software title."},
			{Name: "name", Type: proto.ColumnType_STRING, Description: "Name of the software title."},
			{Name: "display_name", Type: proto.ColumnType_STRING, Description: "Display name of the software title."},
			{Name: "icon_url", Type: proto.ColumnType_STRING, Transform: transform.FromField("IconURL"), Description: "URL of the software icon."},
			{Name: "source", Type: proto.ColumnType_STRING, Description: "Source of the software information (e.g., 'apps', 'deb_packages', 'chrome_extensions'). Filtered by Fleet 4.94+ when set in WHERE clause."},
			{Name: "extension_for", Type: proto.ColumnType_STRING, Description: "If a browser or IDE extension, specifies which software it extends. Filtered by Fleet 4.94+ when set in WHERE clause with source."},
			{Name: "browser", Type: proto.ColumnType_STRING, Description: "Browser name for browser extensions."},
			{Name: "hosts_count", Type: proto.ColumnType_INT, Description: "Number of hosts where this software title is installed."},
			{Name: "versions_count", Type: proto.ColumnType_INT, Description: "Number of distinct versions of this software title."},
			{Name: "bundle_identifier", Type: proto.ColumnType_STRING, Description: "Bundle identifier, typically for macOS and iOS software."},

			// Complex nested objects stored as JSON
			{Name: "versions", Type: proto.ColumnType_JSON, Description: "List of versions for this software title, including version IDs, release and associated vulnerabilities."},
			{Name: "software_package", Type: proto.ColumnType_JSON, Description: "Software package details if the software was added for install. Installer URLs are null unless expose_secrets = true, as they can embed credentials."},
			{Name: "packages", Type: proto.ColumnType_JSON, Description: "All installer packages uploaded for this title, first-added first. software_package only holds the first one. Installer URLs are null unless expose_secrets = true."},
			{Name: "upgrade_code", Type: proto.ColumnType_STRING, Description: "Windows MSI upgrade code, when known."},
			{Name: "app_store_app", Type: proto.ColumnType_JSON, Description: "App Store app details if the software is from an app store; the first-added version when there are several. configuration is null unless expose_secrets = true, as it can hold app credentials."},
			{Name: "app_store_apps", Type: proto.ColumnType_JSON, Description: "Every App Store version added to the team (Fleet 4.94+). Only set when filtering by id and team_id. configuration is null unless expose_secrets = true."},

			// Query parameters that can be used for filtering (key columns)
			{Name: "vulnerable_only", Type: proto.ColumnType_BOOL, Transform: transform.FromQual("vulnerable_only"), Description: "Filter for software titles with known vulnerabilities. Set in WHERE clause."},
			{Name: "team_id", Type: proto.ColumnType_INT, Transform: transform.FromQual("team_id"), Description: "Filter by team ID (Fleet Premium). Use 0 for hosts assigned to 'No team'. Set in WHERE clause."},
			{Name: "available_for_install", Type: proto.ColumnType_BOOL, Transform: transform.FromQual("available_for_install"), Description: "Filter for software available for install (added by the user). Set in WHERE clause."},
			{Name: "query", Type: proto.ColumnType_STRING, Transform: transform.FromQual("query"), Description: "Search query keywords. Searchable fields include title and CVE. Set in WHERE clause."},
			{Name: "self_service", Type: proto.ColumnType_BOOL, Transform: transform.FromQual("self_service"), Description: "Filter for self-service software only. Set in WHERE clause."},
			{Name: "packages_only", Type: proto.ColumnType_BOOL, Transform: transform.FromQual("packages_only"), Description: "Filter for install packages only, excluding app store apps (Fleet Premium). Set in WHERE clause."},
			{Name: "min_cvss_score", Type: proto.ColumnType_DOUBLE, Transform: transform.FromQual("min_cvss_score"), Description: "Filter for software with vulnerabilities having a CVSS v3.x base score higher than this value (Fleet Premium). Set in WHERE clause."},
			{Name: "max_cvss_score", Type: proto.ColumnType_DOUBLE, Transform: transform.FromQual("max_cvss_score"), Description: "Filter for software with vulnerabilities having a CVSS v3.x base score lower than this value (Fleet Premium). Set in WHERE clause."},
			{Name: "exploit", Type: proto.ColumnType_BOOL, Transform: transform.FromQual("exploit"), Description: "Filter for software with vulnerabilities that have been actively exploited in the wild — CISA known exploit (Fleet Premium). Set in WHERE clause."},
			{Name: "platform", Type: proto.ColumnType_STRING, Transform: transform.FromQual("platform"), Description: "Filter installable titles by platform. Options: 'macos', 'darwin', 'windows', 'linux', 'chrome', 'ios', 'ipados'. Requires team_id. Set in WHERE clause."},
			{Name: "exclude_fleet_maintained_apps", Type: proto.ColumnType_BOOL, Transform: transform.FromQual("exclude_fleet_maintained_apps"), Description: "Exclude Fleet-maintained apps from the results. Set in WHERE clause."},
		},
	}
}

func listSoftwareTitles(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	// A lookup by id (optionally per team) is one request; with other
	// filters, list so they still apply.
	if d.EqualsQuals["id"] != nil && onlyQuals(d, "id", "team_id") {
		return streamByID(ctx, d, "fleetdm_software_title", getSoftwareTitle)
	}
	client, err := getClient(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("fleetdm_software_title.listSoftwareTitles", "connection_error", err)
		return nil, err
	}

	// Bulk endpoint: large pages keep the page count low on big inventories.
	err = withSoftwareTypeFilter(ctx, d, "fleetdm_software_title.listSoftwareTitles", func(filter url.Values) error {
		return paginatedList(ctx, d, 1000, func(ctx context.Context, page, perPage int) ([]SoftwareTitle, *ListMeta, error) {
			params := url.Values{}
			maps.Copy(params, filter)
			params.Add("page", strconv.Itoa(page))
			params.Add("per_page", strconv.Itoa(perPage))
			params.Add("order_key", "hosts_count")
			params.Add("order_direction", "desc")

			if d.EqualsQuals["vulnerable_only"] != nil {
				params.Add("vulnerable", strconv.FormatBool(d.EqualsQuals["vulnerable_only"].GetBoolValue()))
			}
			if d.EqualsQuals["team_id"] != nil {
				client.addFleetIDParam(ctx, d, params, strconv.FormatInt(d.EqualsQuals["team_id"].GetInt64Value(), 10))
			}
			if d.EqualsQuals["available_for_install"] != nil {
				params.Add("available_for_install", strconv.FormatBool(d.EqualsQuals["available_for_install"].GetBoolValue()))
			}
			if d.EqualsQuals["query"] != nil {
				params.Add("query", d.EqualsQuals["query"].GetStringValue())
			}
			if d.EqualsQuals["self_service"] != nil {
				params.Add("self_service", strconv.FormatBool(d.EqualsQuals["self_service"].GetBoolValue()))
			}
			if d.EqualsQuals["packages_only"] != nil {
				params.Add("packages_only", strconv.FormatBool(d.EqualsQuals["packages_only"].GetBoolValue()))
			}

			// The API requires vulnerable=true when using min_cvss_score, max_cvss_score, or exploit.
			// Auto-set vulnerable=true if any of these are specified and vulnerable_only was not explicitly set.
			hasCVSSOrExploit := d.EqualsQuals["min_cvss_score"] != nil || d.EqualsQuals["max_cvss_score"] != nil || d.EqualsQuals["exploit"] != nil
			if hasCVSSOrExploit && d.EqualsQuals["vulnerable_only"] == nil {
				params.Add("vulnerable", "true")
			}

			if d.EqualsQuals["min_cvss_score"] != nil {
				params.Add("min_cvss_score", strconv.FormatFloat(d.EqualsQuals["min_cvss_score"].GetDoubleValue(), 'f', -1, 64))
			}
			if d.EqualsQuals["max_cvss_score"] != nil {
				params.Add("max_cvss_score", strconv.FormatFloat(d.EqualsQuals["max_cvss_score"].GetDoubleValue(), 'f', -1, 64))
			}
			if d.EqualsQuals["exploit"] != nil {
				params.Add("exploit", strconv.FormatBool(d.EqualsQuals["exploit"].GetBoolValue()))
			}
			if d.EqualsQuals["platform"] != nil {
				params.Add("platform", d.EqualsQuals["platform"].GetStringValue())
			}
			if d.EqualsQuals["exclude_fleet_maintained_apps"] != nil {
				params.Add("exclude_fleet_maintained_apps", strconv.FormatBool(d.EqualsQuals["exclude_fleet_maintained_apps"].GetBoolValue()))
			}

			var response ListSoftwareTitlesResponse
			if err := client.Get(ctx, "software/titles", params, &response); err != nil {
				plugin.Logger(ctx).Error("fleetdm_software_title.listSoftwareTitles", "api_error", err, "page", page, "params", params.Encode())
				return nil, nil, err
			}
			for i := range response.SoftwareTitles {
				redactInstallerURLs(d, &response.SoftwareTitles[i])
			}
			return response.SoftwareTitles, response.Meta, nil
		})
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// getSoftwareTitle fetches one software title by id, when it is the only filter (with team_id, for that team's counts).
func getSoftwareTitle(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := getClient(ctx, d)
	if err != nil {
		return nil, err
	}
	var response struct {
		SoftwareTitle *SoftwareTitle `json:"software_title"`
	}
	params := url.Values{}
	if q := d.EqualsQuals["team_id"]; q != nil {
		client.addFleetIDParam(ctx, d, params, strconv.FormatInt(q.GetInt64Value(), 10))
	}
	if err := client.Get(ctx, fmt.Sprintf("software/titles/%d", d.EqualsQuals["id"].GetInt64Value()), params, &response); err != nil {
		plugin.Logger(ctx).Error("fleetdm_software_title.getSoftwareTitle", "api_error", err)
		return nil, err
	}
	if response.SoftwareTitle != nil {
		redactInstallerURLs(d, response.SoftwareTitle)
		return *response.SoftwareTitle, nil
	}
	return nil, nil
}
