package fleetdm

import (
	"context"
	"net/url"
	"strconv"

	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

// SoftwareTitleVersion represents a version entry within a software title.
type SoftwareTitleVersion struct {
	ID              uint     `json:"id"`
	Version         string   `json:"version"`
	Vulnerabilities []string `json:"vulnerabilities"` // List of CVE strings
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
	AppStoreApp      interface{}            `json:"app_store_app"`
	BundleIdentifier *string                `json:"bundle_identifier"`
	CountsUpdatedAt  *FleetTime             `json:"counts_updated_at"`
}

// ListSoftwareTitlesResponse is the expected structure for the list software titles API call.
type ListSoftwareTitlesResponse struct {
	SoftwareTitles []SoftwareTitle `json:"software_titles"`
	Meta           struct {
		HasNextResults     bool `json:"has_next_results"`
		HasPreviousResults bool `json:"has_previous_results"`
	} `json:"meta"`
	Count           int        `json:"count"`
	CountsUpdatedAt *FleetTime `json:"counts_updated_at"`
}

func tableFleetdmSoftwareTitle(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "fleetdm_software_title",
		Description: "Software titles from FleetDM. A software title groups multiple versions of the same software. Uses the /software/titles endpoint.",
		List: &plugin.ListConfig{
			Hydrate: listSoftwareTitles,
			KeyColumns: []*plugin.KeyColumn{
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
			},
		},
		Columns: []*plugin.Column{
			// Core software title information
			{Name: "id", Type: proto.ColumnType_INT, Description: "Unique ID of the software title."},
			{Name: "name", Type: proto.ColumnType_STRING, Description: "Name of the software title."},
			{Name: "display_name", Type: proto.ColumnType_STRING, Description: "Display name of the software title."},
			{Name: "icon_url", Type: proto.ColumnType_STRING, Transform: transform.FromField("IconURL"), Description: "URL of the software icon."},
			{Name: "source", Type: proto.ColumnType_STRING, Description: "Source of the software information (e.g., 'apps', 'deb_packages', 'chrome_extensions')."},
			{Name: "extension_for", Type: proto.ColumnType_STRING, Description: "If a browser extension, specifies which software it extends."},
			{Name: "browser", Type: proto.ColumnType_STRING, Description: "Browser name for browser extensions."},
			{Name: "hosts_count", Type: proto.ColumnType_INT, Description: "Number of hosts where this software title is installed."},
			{Name: "versions_count", Type: proto.ColumnType_INT, Description: "Number of distinct versions of this software title."},
			{Name: "bundle_identifier", Type: proto.ColumnType_STRING, Description: "Bundle identifier, typically for macOS and iOS software."},

			// Complex nested objects stored as JSON
			{Name: "versions", Type: proto.ColumnType_JSON, Description: "List of versions for this software title, including version IDs and associated vulnerabilities."},
			{Name: "software_package", Type: proto.ColumnType_JSON, Description: "Software package details if the software was added for install."},
			{Name: "app_store_app", Type: proto.ColumnType_JSON, Description: "App Store app details if the software is from an app store."},

			// Query parameters that can be used for filtering (key columns)
			{Name: "vulnerable_only", Type: proto.ColumnType_BOOL, Transform: transform.FromQual("vulnerable_only"), Description: "Filter for software titles with known vulnerabilities. Set in WHERE clause."},
			{Name: "team_id", Type: proto.ColumnType_INT, Transform: transform.FromQual("team_id"), Description: "Filter by team ID (Fleet Premium). Use 0 for hosts assigned to 'No team'. Set in WHERE clause."},
			{Name: "available_for_install", Type: proto.ColumnType_BOOL, Transform: transform.FromQual("available_for_install"), Description: "Filter for software available for install (added by the user). Set in WHERE clause."},
			{Name: "query", Type: proto.ColumnType_STRING, Transform: transform.FromQual("query"), Description: "Search query keywords. Searchable fields include title and CVE. Set in WHERE clause."},
			{Name: "self_service", Type: proto.ColumnType_BOOL, Transform: transform.FromQual("self_service"), Description: "Filter for self-service software only. Set in WHERE clause."},
			{Name: "packages_only", Type: proto.ColumnType_BOOL, Transform: transform.FromQual("packages_only"), Description: "Filter for install packages only, excluding app store apps (Fleet Premium). Set in WHERE clause."},
			{Name: "min_cvss_score", Type: proto.ColumnType_INT, Transform: transform.FromQual("min_cvss_score"), Description: "Filter for software with vulnerabilities having a CVSS v3.x base score higher than this value (Fleet Premium). Set in WHERE clause."},
			{Name: "max_cvss_score", Type: proto.ColumnType_INT, Transform: transform.FromQual("max_cvss_score"), Description: "Filter for software with vulnerabilities having a CVSS v3.x base score lower than this value (Fleet Premium). Set in WHERE clause."},
			{Name: "exploit", Type: proto.ColumnType_BOOL, Transform: transform.FromQual("exploit"), Description: "Filter for software with vulnerabilities that have been actively exploited in the wild — CISA known exploit (Fleet Premium). Set in WHERE clause."},
			{Name: "platform", Type: proto.ColumnType_STRING, Transform: transform.FromQual("platform"), Description: "Filter installable titles by platform. Options: 'macos', 'darwin', 'windows', 'linux', 'chrome', 'ios', 'ipados'. Requires team_id. Set in WHERE clause."},
			{Name: "exclude_fleet_maintained_apps", Type: proto.ColumnType_BOOL, Transform: transform.FromQual("exclude_fleet_maintained_apps"), Description: "Exclude Fleet-maintained apps from the results. Set in WHERE clause."},
		},
	}
}

func listSoftwareTitles(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := NewFleetDMClient(ctx, d.Connection)
	if err != nil {
		plugin.Logger(ctx).Error("fleetdm_software_title.listSoftwareTitles", "connection_error", err)
		return nil, err
	}

	page := 0
	perPage := 10000

	for {
		params := url.Values{}
		params.Add("page", strconv.Itoa(page))
		params.Add("per_page", strconv.Itoa(perPage))
		params.Add("order_key", "hosts_count")
		params.Add("order_direction", "desc")

		if d.EqualsQuals["vulnerable_only"] != nil {
			params.Add("vulnerable", strconv.FormatBool(d.EqualsQuals["vulnerable_only"].GetBoolValue()))
		}
		if d.EqualsQuals["team_id"] != nil {
			params.Add("team_id", strconv.FormatInt(d.EqualsQuals["team_id"].GetInt64Value(), 10))
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
			params.Add("min_cvss_score", strconv.FormatInt(d.EqualsQuals["min_cvss_score"].GetInt64Value(), 10))
		}
		if d.EqualsQuals["max_cvss_score"] != nil {
			params.Add("max_cvss_score", strconv.FormatInt(d.EqualsQuals["max_cvss_score"].GetInt64Value(), 10))
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
		_, err := client.Get(ctx, "software/titles", params, &response) // Endpoint is /api/v1/fleet/software/titles
		if err != nil {
			plugin.Logger(ctx).Error("fleetdm_software_title.listSoftwareTitles", "api_error", err, "page", page, "params", params.Encode())
			return nil, err
		}

		for _, title := range response.SoftwareTitles {
			d.StreamListItem(ctx, title)
			if d.RowsRemaining(ctx) == 0 {
				plugin.Logger(ctx).Debug("fleetdm_software_title.listSoftwareTitles", "limit_reached_sdk", "true")
				return nil, nil
			}
		}

		plugin.Logger(ctx).Info("fleetdm_software_title.listSoftwareTitles",
			"page_processed", page,
			"items_on_page", len(response.SoftwareTitles),
			"api_total_count", response.Count,
			"api_has_next_results", response.Meta.HasNextResults,
		)

		if len(response.SoftwareTitles) < perPage {
			plugin.Logger(ctx).Info("fleetdm_software_title.listSoftwareTitles", "pagination_ended_item_count_less_than_per_page", true, "current_page", page, "items_on_page", len(response.SoftwareTitles), "per_page", perPage)
			break
		}

		if !response.Meta.HasNextResults && len(response.SoftwareTitles) == perPage {
			plugin.Logger(ctx).Warn("fleetdm_software_title.listSoftwareTitles", "api_has_next_results_is_false_but_full_page_received", true, "current_page", page)
		}

		page++
		plugin.Logger(ctx).Debug("fleetdm_software_title.listSoftwareTitles", "incrementing_to_next_page", page)
	}

	plugin.Logger(ctx).Info("fleetdm_software_title.listSoftwareTitles", "list_software_titles_completed", true)
	return nil, nil
}
