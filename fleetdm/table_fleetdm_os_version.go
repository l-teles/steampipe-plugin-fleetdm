package fleetdm

import (
	"context"
	"net/url"
	"strconv"

	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

// OSVersionVulnerability represents a vulnerability associated with an OS version.
type OSVersionVulnerability struct {
	CVE               string     `json:"cve"`
	DetailsLink       string     `json:"details_link"`
	CreatedAt         *FleetTime `json:"created_at,omitempty"`
	CVSSScore         *float64   `json:"cvss_score,omitempty"`
	EPSSProbability   *float64   `json:"epss_probability,omitempty"`
	CISAKnownExploit  *bool      `json:"cisa_known_exploit,omitempty"`
	CVEPublished      *FleetTime `json:"cve_published,omitempty"`
	CVEDescription    *string    `json:"cve_description,omitempty"`
	ResolvedInVersion *string    `json:"resolved_in_version,omitempty"`
}

// OSVersion represents an operating system version in FleetDM.
// Refer to: https://fleetdm.com/docs/rest-api/rest-api#list-operating-systems
type OSVersion struct {
	OSVersionID          uint                     `json:"os_version_id"`
	HostsCount           uint                     `json:"hosts_count"`
	Name                 string                   `json:"name"`
	NameOnly             string                   `json:"name_only"`
	Version              string                   `json:"version"`
	Platform             string                   `json:"platform"`
	GeneratedCPEs        []string                 `json:"generated_cpes"`
	Vulnerabilities      []OSVersionVulnerability `json:"vulnerabilities"`
	VulnerabilitiesCount uint                     `json:"vulnerabilities_count"`
}

// ListOSVersionsResponse is the expected structure for the list OS versions API call.
type ListOSVersionsResponse struct {
	OSVersions []OSVersion `json:"os_versions"`
	Meta       struct {
		HasNextResults     bool `json:"has_next_results"`
		HasPreviousResults bool `json:"has_previous_results"`
	} `json:"meta"`
	Count           int        `json:"count"`
	CountsUpdatedAt *FleetTime `json:"counts_updated_at"`
}

func tableFleetdmOSVersion(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "fleetdm_os_version",
		Description: "Operating system versions from FleetDM. Lists all OS versions across managed hosts with vulnerability information. Uses the /os_versions endpoint.",
		List: &plugin.ListConfig{
			Hydrate: listOSVersions,
			KeyColumns: []*plugin.KeyColumn{
				{Name: "team_id", Require: plugin.Optional},
				{Name: "platform", Require: plugin.Optional},
				{Name: "os_name", Require: plugin.Optional},
				{Name: "os_version_filter", Require: plugin.Optional},
			},
		},
		Columns: []*plugin.Column{
			// Core OS version information
			{Name: "os_version_id", Type: proto.ColumnType_INT, Transform: transform.FromField("OSVersionID"), Description: "Unique ID of the OS version."},
			{Name: "hosts_count", Type: proto.ColumnType_INT, Description: "Number of hosts running this OS version."},
			{Name: "name", Type: proto.ColumnType_STRING, Description: "Full name of the OS version (e.g., 'macOS 26.2', 'Microsoft Windows 11 Pro 24H2 10.0.26100.7623')."},
			{Name: "name_only", Type: proto.ColumnType_STRING, Description: "OS name without version (e.g., 'macOS', 'Microsoft Windows 11 Pro 24H2')."},
			{Name: "version", Type: proto.ColumnType_STRING, Description: "Version string of the OS (e.g., '26.2', '10.0.26100.7623')."},
			{Name: "platform", Type: proto.ColumnType_STRING, Description: "Platform of the OS (e.g., 'darwin', 'windows', 'ubuntu')."},
			{Name: "generated_cpes", Type: proto.ColumnType_JSON, Transform: transform.FromField("GeneratedCPEs"), Description: "Generated Common Platform Enumeration (CPE) strings for the OS."},
			{Name: "vulnerabilities", Type: proto.ColumnType_JSON, Description: "Vulnerabilities associated with this OS version."},
			{Name: "vulnerabilities_count", Type: proto.ColumnType_INT, Description: "Number of known vulnerabilities for this OS version."},

			// Query parameters that can be used for filtering
			{Name: "team_id", Type: proto.ColumnType_INT, Transform: transform.FromQual("team_id"), Description: "Filter by team ID. Set in WHERE clause."},
			{Name: "os_name", Type: proto.ColumnType_STRING, Transform: transform.FromQual("os_name"), Description: "Filter by OS name (must be used with os_version_filter). Set in WHERE clause."},
			{Name: "os_version_filter", Type: proto.ColumnType_STRING, Transform: transform.FromQual("os_version_filter"), Description: "Filter by OS version string (must be used with os_name). Set in WHERE clause."},
		},
	}
}

func listOSVersions(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := NewFleetDMClient(ctx, d.Connection)
	if err != nil {
		plugin.Logger(ctx).Error("fleetdm_os_version.listOSVersions", "connection_error", err)
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

		if d.EqualsQuals["team_id"] != nil {
			params.Add("team_id", strconv.FormatInt(d.EqualsQuals["team_id"].GetInt64Value(), 10))
		}
		if d.EqualsQuals["platform"] != nil {
			params.Add("platform", d.EqualsQuals["platform"].GetStringValue())
		}
		if d.EqualsQuals["os_name"] != nil {
			params.Add("os_name", d.EqualsQuals["os_name"].GetStringValue())
		}
		if d.EqualsQuals["os_version_filter"] != nil {
			params.Add("os_version", d.EqualsQuals["os_version_filter"].GetStringValue())
		}

		var response ListOSVersionsResponse
		_, err := client.Get(ctx, "os_versions", params, &response) // Endpoint is /api/v1/fleet/os_versions
		if err != nil {
			plugin.Logger(ctx).Error("fleetdm_os_version.listOSVersions", "api_error", err, "page", page, "params", params.Encode())
			return nil, err
		}

		for _, osVer := range response.OSVersions {
			d.StreamListItem(ctx, osVer)
			if d.RowsRemaining(ctx) == 0 {
				plugin.Logger(ctx).Debug("fleetdm_os_version.listOSVersions", "limit_reached_sdk", "true")
				return nil, nil
			}
		}

		plugin.Logger(ctx).Info("fleetdm_os_version.listOSVersions",
			"page_processed", page,
			"items_on_page", len(response.OSVersions),
			"api_total_count", response.Count,
			"api_has_next_results", response.Meta.HasNextResults,
		)

		if len(response.OSVersions) < perPage {
			plugin.Logger(ctx).Info("fleetdm_os_version.listOSVersions", "pagination_ended_item_count_less_than_per_page", true, "current_page", page, "items_on_page", len(response.OSVersions), "per_page", perPage)
			break
		}

		if !response.Meta.HasNextResults && len(response.OSVersions) == perPage {
			plugin.Logger(ctx).Warn("fleetdm_os_version.listOSVersions", "api_has_next_results_is_false_but_full_page_received", true, "current_page", page)
		}

		page++
		plugin.Logger(ctx).Debug("fleetdm_os_version.listOSVersions", "incrementing_to_next_page", page)
	}

	plugin.Logger(ctx).Info("fleetdm_os_version.listOSVersions", "list_os_versions_completed", true)
	return nil, nil
}
