package fleetdm

import (
	"context"
	"net/url"
	"strconv"

	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

// SoftwareVulnerability represents a vulnerability associated with a software item.
type SoftwareVulnerability struct {
	CVE                           string     `json:"cve"`
	DetailsLink                   string     `json:"details_link"`
	CVSSScore                     *float64   `json:"cvss_score,omitempty"`                       // Common Vulnerability Scoring System
	EPSSProbability               *float64   `json:"epss_probability,omitempty"`                 // Exploit Prediction Scoring System
	CISAKnownExploit              *bool      `json:"cisa_known_exploit,omitempty"`               // CISA Known Exploited Vulnerabilities Catalog
	CVEPublished                  *FleetTime `json:"cve_published,omitempty"`                    // Date CVE was published
	ResolvedInVersion             *string    `json:"resolved_in_version,omitempty"`              // Version the vulnerability is resolved in
	CurrentlyExploited            *bool      `json:"currently_exploited,omitempty"`              // Premium feature: From Recorded Future
	Exploitability7Day            *int       `json:"exploitability_7_day,omitempty"`             // Premium feature
	Exploitability30Day           *int       `json:"exploitability_30_day,omitempty"`            // Premium feature
	Exploitability60Day           *int       `json:"exploitability_60_day,omitempty"`            // Premium feature
	Exploitability90Day           *int       `json:"exploitability_90_day,omitempty"`            // Premium feature
	ExploitedActivity7Day         *int       `json:"exploited_activity_7_day,omitempty"`         // Premium feature
	ExploitedActivity30Day        *int       `json:"exploited_activity_30_day,omitempty"`        // Premium feature
	ExploitedActivity60Day        *int       `json:"exploited_activity_60_day,omitempty"`        // Premium feature
	ExploitedActivity90Day        *int       `json:"exploited_activity_90_day,omitempty"`        // Premium feature
	ExploitedMalware7Day          *int       `json:"exploited_malware_7_day,omitempty"`          // Premium feature
	ExploitedMalware30Day         *int       `json:"exploited_malware_30_day,omitempty"`         // Premium feature
	ExploitedMalware60Day         *int       `json:"exploited_malware_60_day,omitempty"`         // Premium feature
	ExploitedMalware90Day         *int       `json:"exploited_malware_90_day,omitempty"`         // Premium feature
	ExploitedNetwork7Day          *int       `json:"exploited_network_7_day,omitempty"`          // Premium feature
	ExploitedNetwork30Day         *int       `json:"exploited_network_30_day,omitempty"`         // Premium feature
	ExploitedNetwork60Day         *int       `json:"exploited_network_60_day,omitempty"`         // Premium feature
	ExploitedNetwork90Day         *int       `json:"exploited_network_90_day,omitempty"`         // Premium feature
	ExploitedPublic7Day           *int       `json:"exploited_public_7_day,omitempty"`           // Premium feature
	ExploitedPublic30Day          *int       `json:"exploited_public_30_day,omitempty"`          // Premium feature
	ExploitedPublic60Day          *int       `json:"exploited_public_60_day,omitempty"`          // Premium feature
	ExploitedPublic90Day          *int       `json:"exploited_public_90_day,omitempty"`          // Premium feature
	ExploitedRansomware7Day       *int       `json:"exploited_ransomware_7_day,omitempty"`       // Premium feature
	ExploitedRansomware30Day      *int       `json:"exploited_ransomware_30_day,omitempty"`      // Premium feature
	ExploitedRansomware60Day      *int       `json:"exploited_ransomware_60_day,omitempty"`      // Premium feature
	ExploitedRansomware90Day      *int       `json:"exploited_ransomware_90_day,omitempty"`      // Premium feature
	ExploitedRemote7Day           *int       `json:"exploited_remote_7_day,omitempty"`           // Premium feature
	ExploitedRemote30Day          *int       `json:"exploited_remote_30_day,omitempty"`          // Premium feature
	ExploitedRemote60Day          *int       `json:"exploited_remote_60_day,omitempty"`          // Premium feature
	ExploitedRemote90Day          *int       `json:"exploited_remote_90_day,omitempty"`          // Premium feature
	ExploitedUnauthenticated7Day  *int       `json:"exploited_unauthenticated_7_day,omitempty"`  // Premium feature
	ExploitedUnauthenticated30Day *int       `json:"exploited_unauthenticated_30_day,omitempty"` // Premium feature
	ExploitedUnauthenticated60Day *int       `json:"exploited_unauthenticated_60_day,omitempty"` // Premium feature
	ExploitedUnauthenticated90Day *int       `json:"exploited_unauthenticated_90_day,omitempty"` // Premium feature
}

// Software represents a software item in FleetDM.
// Refer to: https://fleetdm.com/docs/rest-api/rest-api#software-item
type Software struct {
	ID               uint                    `json:"id"`
	Name             string                  `json:"name"`
	Version          string                  `json:"version"`
	Source           string                  `json:"source"`
	ExtensionFor     *string                 `json:"extension_for"` // For browser extensions - extension for which app
	Browser          *string                 `json:"browser,omitempty"`       // For browser extensions
	Vendor           *string                 `json:"vendor,omitempty"`        // e.g., for RPMs
	GeneratedCPE     string                  `json:"generated_cpe"`
	BundleIdentifier *string                 `json:"bundle_identifier"` // macOS, iOS
	HostCount        uint                    `json:"hosts_count"`       // Number of hosts with this software (note: API uses "hosts_count" plural)
	Vulnerabilities  []SoftwareVulnerability `json:"vulnerabilities"`
	UpgradeCode      *string                 `json:"upgrade_code"`           // Windows installer upgrade code
	DisplayName      *string                 `json:"display_name"`           // Display name for the software
	LastOpenedAt     *FleetTime              `json:"last_opened_at"`         // This is typically per-host, might be null or aggregated differently in the global software list
	Release          *string                 `json:"release,omitempty"`      // e.g., for RPMs
	Arch             *string                 `json:"arch,omitempty"`         // e.g., for RPMs
	ExtensionID      *string                 `json:"extension_id,omitempty"` // For browser extensions
}

// ListSoftwareResponse is the expected structure for the list software API call.
type ListSoftwareResponse struct {
	Software []Software `json:"software"`
	Meta     struct {
		HasNextResults     bool   `json:"has_next_results"`
		HasPreviousResults bool   `json:"has_previous_results"`
		NextCursor         string `json:"next_cursor"`
	} `json:"meta"`
	Count           int       `json:"count"` // Total count of all software items matching the query
}

func tableFleetdmSoftwareVersion(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "fleetdm_software_version",
		Description: "Software versions inventory from FleetDM. Uses the /software/versions endpoint.",
		List: &plugin.ListConfig{
			Hydrate: listSoftwareVersions,
			KeyColumns: []*plugin.KeyColumn{
				{Name: "vulnerable_only", Require: plugin.Optional}, // Filter for vulnerable software
				{Name: "team_id", Require: plugin.Optional},         // Filter by team ID (Fleet Premium)
				{Name: "query", Require: plugin.Optional},           // Search by name, version, or CVE
				{Name: "min_cvss_score", Require: plugin.Optional},  // Min CVSS v3.x base score (Fleet Premium)
				{Name: "max_cvss_score", Require: plugin.Optional},  // Max CVSS v3.x base score (Fleet Premium)
				{Name: "exploit", Require: plugin.Optional},         // Filter for CISA known exploits (Fleet Premium)
			},
		},
		Columns: []*plugin.Column{
			// Core software information
			{Name: "id", Type: proto.ColumnType_INT, Description: "Unique ID of the software item."},
			{Name: "name", Type: proto.ColumnType_STRING, Description: "Name of the software."},
			{Name: "version", Type: proto.ColumnType_STRING, Description: "Version of the software."},
			{Name: "source", Type: proto.ColumnType_STRING, Description: "Source of the software information (e.g., 'apps', 'deb_packages', 'chrome_extensions')."},
			{Name: "host_count", Type: proto.ColumnType_INT, Transform: transform.FromField("HostCount"), Description: "Number of hosts where this software is installed."},
			{Name: "generated_cpe", Type: proto.ColumnType_STRING, Transform: transform.FromField("GeneratedCPE"), Description: "Generated Common Platform Enumeration (CPE) string for the software."},
			{Name: "bundle_identifier", Type: proto.ColumnType_STRING, Description: "Bundle identifier, typically for macOS and iOS software."},
			{Name: "upgrade_code", Type: proto.ColumnType_STRING, Description: "Windows installer upgrade code."},
			{Name: "display_name", Type: proto.ColumnType_STRING, Description: "Display name for the software."},
			{Name: "extension_for", Type: proto.ColumnType_STRING, Description: "For browser extensions - indicates which application the extension is for."},
			{Name: "release", Type: proto.ColumnType_STRING, Description: "Release information, e.g., for RPM packages."},
			{Name: "vendor", Type: proto.ColumnType_STRING, Description: "Vendor information, e.g., for RPM packages."},
			{Name: "arch", Type: proto.ColumnType_STRING, Description: "Architecture information, e.g., for RPM packages."},
			{Name: "extension_id", Type: proto.ColumnType_STRING, Description: "Extension ID for browser extensions."},
			{Name: "browser", Type: proto.ColumnType_STRING, Description: "Browser name for browser extensions."},
			{Name: "last_opened_at", Type: proto.ColumnType_TIMESTAMP, Transform: transform.FromField("LastOpenedAt").Transform(flexibleTimeTransform), Description: "Timestamp when the software was last opened (may be aggregated or host-specific)."},

			// Vulnerabilities - stored as JSONB as it's an array of complex objects
			// Users can query into this using JSON functions in SQL.
			{Name: "vulnerabilities", Type: proto.ColumnType_JSON, Description: "Vulnerabilities associated with this software."},

			// Query parameters that can be used for filtering (key columns)
			{Name: "vulnerable_only", Type: proto.ColumnType_BOOL, Transform: transform.FromQual("vulnerable_only"), Description: "Filter for software with known vulnerabilities. Set in WHERE clause."},
			{Name: "team_id", Type: proto.ColumnType_INT, Transform: transform.FromQual("team_id"), Description: "Filter by team ID (Fleet Premium). Use 0 for hosts assigned to 'No team'. Set in WHERE clause."},
			{Name: "query", Type: proto.ColumnType_STRING, Transform: transform.FromQual("query"), Description: "Search query keywords. Searchable fields include name, version, and CVE. Set in WHERE clause."},
			{Name: "min_cvss_score", Type: proto.ColumnType_INT, Transform: transform.FromQual("min_cvss_score"), Description: "Filter for software with vulnerabilities having a CVSS v3.x base score higher than this value (Fleet Premium). Set in WHERE clause."},
			{Name: "max_cvss_score", Type: proto.ColumnType_INT, Transform: transform.FromQual("max_cvss_score"), Description: "Filter for software with vulnerabilities having a CVSS v3.x base score lower than this value (Fleet Premium). Set in WHERE clause."},
			{Name: "exploit", Type: proto.ColumnType_BOOL, Transform: transform.FromQual("exploit"), Description: "Filter for software with vulnerabilities that have been actively exploited in the wild — CISA known exploit (Fleet Premium). Set in WHERE clause."},
		},
	}
}

func listSoftwareVersions(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := NewFleetDMClient(ctx, d.Connection)
	if err != nil {
		plugin.Logger(ctx).Error("fleetdm_software_version.listSoftwareVersions", "connection_error", err)
		return nil, err
	}

	page := 0
	// perPage is the number of items to request per API call.
	perPage := 10000 // This seems to have no limit and 100 was making it super slow, 1000 slow so let's go with 10000 🤠.

	for {
		params := url.Values{}
		params.Add("page", strconv.Itoa(page))
		params.Add("per_page", strconv.Itoa(perPage))
		params.Add("order_key", "id")
		params.Add("order_direction", "asc")

		if d.EqualsQuals["vulnerable_only"] != nil {
			params.Add("vulnerable", strconv.FormatBool(d.EqualsQuals["vulnerable_only"].GetBoolValue()))
		}
		if d.EqualsQuals["team_id"] != nil {
			params.Add("team_id", strconv.FormatInt(d.EqualsQuals["team_id"].GetInt64Value(), 10))
		}
		if d.EqualsQuals["query"] != nil {
			params.Add("query", d.EqualsQuals["query"].GetStringValue())
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

		var response ListSoftwareResponse
		_, err := client.Get(ctx, "software/versions", params, &response) // Endpoint is /api/v1/fleet/software/versions
		if err != nil {
			plugin.Logger(ctx).Error("fleetdm_software_version.listSoftwareVersions", "api_error", err, "page", page, "params", params.Encode())
			return nil, err
		}

		for _, swItem := range response.Software {
			d.StreamListItem(ctx, swItem)
			if d.RowsRemaining(ctx) == 0 {
				plugin.Logger(ctx).Debug("fleetdm_software_version.listSoftwareVersions", "limit_reached_sdk", "true")
				return nil, nil
			}
		}

		// Log pagination details from the API response
		plugin.Logger(ctx).Info("fleetdm_software_version.listSoftwareVersions",
			"page_processed", page,
			"items_on_page", len(response.Software),
			"api_total_count", response.Count, // Total items matching filter, not just on this page
			"api_has_next_results", response.Meta.HasNextResults,
			"api_next_cursor", response.Meta.NextCursor,
		)

		// Determine if there are more pages to fetch.
		// Primary condition: Continue if the API returned a full page of items.
		if len(response.Software) < perPage {
			plugin.Logger(ctx).Info("fleetdm_software_version.listSoftwareVersions", "pagination_ended_item_count_less_than_per_page", true, "current_page", page, "items_on_page", len(response.Software), "per_page", perPage)
			break
		}

		// Secondary check (optional, but good for observation): Log if HasNextResults is false but we got a full page.
		// This might indicate an inconsistency in the API's meta field.
		if !response.Meta.HasNextResults && len(response.Software) == perPage {
			plugin.Logger(ctx).Warn("fleetdm_software_version.listSoftwareVersions", "api_has_next_results_is_false_but_full_page_received", true, "current_page", page)
			// Depending on API behavior, you might still want to try fetching the next page,
			// or trust len(response.Software) < perPage as the more reliable indicator.
			// For now, we will break if len(response.Software) < perPage, making HasNextResults secondary.
		}

		// If HasNextResults is explicitly false, and we trust it, we can break early.
		// However, considering a previous issue existed, let's prioritize the item count.
		// if !response.Meta.HasNextResults {
		// 	plugin.Logger(ctx).Info("fleetdm_software_version.listSoftwareVersions", "pagination_ended_by_api_has_next_results_false", true, "current_page", page)
		// 	break
		// }

		page++
		plugin.Logger(ctx).Debug("fleetdm_software_version.listSoftwareVersions", "incrementing_to_next_page", page)
	}

	plugin.Logger(ctx).Info("fleetdm_software_version.listSoftwareVersions", "list_software_versions_completed", true)
	return nil, nil
}
