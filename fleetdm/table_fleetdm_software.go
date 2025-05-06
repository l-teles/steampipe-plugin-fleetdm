package fleetdm

import (
	"context"
	"net/url"
	"strconv"
	"time"

	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

// SoftwareVulnerability represents a vulnerability associated with a software item.
type SoftwareVulnerability struct {
	CVE                       string     `json:"cve"`
	DetailsLink               string     `json:"details_link"`
	CVSSScore                 *float64   `json:"cvss_score,omitempty"`                  // Common Vulnerability Scoring System
	EPSSProbability           *float64   `json:"epss_probability,omitempty"`            // Exploit Prediction Scoring System
	CISAKnownExploit          *bool      `json:"cisa_known_exploit,omitempty"`          // CISA Known Exploited Vulnerabilities Catalog
	CVEPublished              *time.Time `json:"cve_published,omitempty"`               // Date CVE was published
	ResolvedInVersion         *string    `json:"resolved_in_version,omitempty"`         // Version the vulnerability is resolved in
	CurrentlyExploited        *bool      `json:"currently_exploited,omitempty"`         // Premium feature: From Recorded Future
	Exploitability7Day        *int       `json:"exploitability_7_day,omitempty"`        // Premium feature
	Exploitability30Day       *int       `json:"exploitability_30_day,omitempty"`       // Premium feature
	Exploitability60Day       *int       `json:"exploitability_60_day,omitempty"`       // Premium feature
	Exploitability90Day       *int       `json:"exploitability_90_day,omitempty"`       // Premium feature
	ExploitedActivity7Day     *int       `json:"exploited_activity_7_day,omitempty"`    // Premium feature
	ExploitedActivity30Day    *int       `json:"exploited_activity_30_day,omitempty"`   // Premium feature
	ExploitedActivity60Day    *int       `json:"exploited_activity_60_day,omitempty"`   // Premium feature
	ExploitedActivity90Day    *int       `json:"exploited_activity_90_day,omitempty"`   // Premium feature
	ExploitedMalware7Day      *int       `json:"exploited_malware_7_day,omitempty"`     // Premium feature
	ExploitedMalware30Day     *int       `json:"exploited_malware_30_day,omitempty"`    // Premium feature
	ExploitedMalware60Day     *int       `json:"exploited_malware_60_day,omitempty"`    // Premium feature
	ExploitedMalware90Day     *int       `json:"exploited_malware_90_day,omitempty"`    // Premium feature
	ExploitedNetwork7Day      *int       `json:"exploited_network_7_day,omitempty"`     // Premium feature
	ExploitedNetwork30Day     *int       `json:"exploited_network_30_day,omitempty"`    // Premium feature
	ExploitedNetwork60Day     *int       `json:"exploited_network_60_day,omitempty"`    // Premium feature
	ExploitedNetwork90Day     *int       `json:"exploited_network_90_day,omitempty"`    // Premium feature
	ExploitedPublic7Day       *int       `json:"exploited_public_7_day,omitempty"`      // Premium feature
	ExploitedPublic30Day      *int       `json:"exploited_public_30_day,omitempty"`     // Premium feature
	ExploitedPublic60Day      *int       `json:"exploited_public_60_day,omitempty"`     // Premium feature
	ExploitedPublic90Day      *int       `json:"exploited_public_90_day,omitempty"`     // Premium feature
	ExploitedRansomware7Day   *int       `json:"exploited_ransomware_7_day,omitempty"`  // Premium feature
	ExploitedRansomware30Day  *int       `json:"exploited_ransomware_30_day,omitempty"` // Premium feature
	ExploitedRansomware60Day  *int       `json:"exploited_ransomware_60_day,omitempty"` // Premium feature
	ExploitedRansomware90Day  *int       `json:"exploited_ransomware_90_day,omitempty"` // Premium feature
	ExploitedRemote7Day       *int       `json:"exploited_remote_7_day,omitempty"`      // Premium feature
	ExploitedRemote30Day      *int       `json:"exploited_remote_30_day,omitempty"`     // Premium feature
	ExploitedRemote60Day      *int       `json:"exploited_remote_60_day,omitempty"`     // Premium feature
	ExploitedRemote90Day      *int       `json:"exploited_remote_90_day,omitempty"`     // Premium feature
	ExploitedUnauthenticated7Day *int    `json:"exploited_unauthenticated_7_day,omitempty"` // Premium feature
	ExploitedUnauthenticated30Day *int   `json:"exploited_unauthenticated_30_day,omitempty"`// Premium feature
	ExploitedUnauthenticated60Day *int   `json:"exploited_unauthenticated_60_day,omitempty"`// Premium feature
	ExploitedUnauthenticated90Day *int   `json:"exploited_unauthenticated_90_day,omitempty"`// Premium feature
}

// Software represents a software item in FleetDM.
// Refer to: https://fleetdm.com/docs/rest-api/rest-api#software-item
type Software struct {
	ID                 uint                    `json:"id"`
	Name               string                  `json:"name"`
	Version            string                  `json:"version"`
	Source             string                  `json:"source"`
	BundleIdentifier   *string                 `json:"bundle_identifier"` // macOS, iOS
	GeneratedCPE       string                  `json:"generated_cpe"`
	HostCount          uint                    `json:"host_count"` // Number of hosts with this software
	Vulnerabilities    []SoftwareVulnerability `json:"vulnerabilities"`
	CountsUpdatedAt    time.Time               `json:"counts_updated_at"` // Timestamp for when host_count was last updated
	LastOpenedAt       *time.Time              `json:"last_opened_at"`    // This is typically per-host, might be null or aggregated differently in the global software list
	Release            *string                 `json:"release,omitempty"` // e.g., for RPMs
	Vendor             *string                 `json:"vendor,omitempty"`  // e.g., for RPMs
	Arch               *string                 `json:"arch,omitempty"`    // e.g., for RPMs
	ExtensionID        *string                 `json:"extension_id,omitempty"` // For browser extensions
	Browser            *string                 `json:"browser,omitempty"`      // For browser extensions
	Path               *string                 `json:"path,omitempty"`         // e.g., for Programs
	InstalledPath      *string                 `json:"installed_path,omitempty"` // e.g., for Homebrew packages
}

// ListSoftwareResponse is the expected structure for the list software API call.
type ListSoftwareResponse struct {
	Software []Software `json:"software"`
	Meta     struct {
		HasNextResults     bool `json:"has_next_results"`
		HasPreviousResults bool `json:"has_previous_results"`
	} `json:"meta"`
	Count            int       `json:"count"` // Total count of all software items matching the query
	CountsUpdatedAt  time.Time `json:"counts_updated_at"`
}

func tableFleetdmSoftware(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "fleetdm_software",
		Description: "Software inventory from FleetDM.",
		List: &plugin.ListConfig{
			Hydrate: listSoftware,
			KeyColumns: []*plugin.KeyColumn{ // Corrected: Use a slice of *plugin.KeyColumn
				{Name: "vulnerable_only", Require: plugin.Optional}, // Allow filtering for vulnerable software
				{Name: "os_id", Require: plugin.Optional},           // Filter by OS ID
				{Name: "os_name", Require: plugin.Optional},         // Filter by OS name
				{Name: "os_version", Require: plugin.Optional},      // Filter by OS version
				{Name: "team_id", Require: plugin.Optional},         // Filter by team ID
			},
		},
		// Get: &plugin.GetConfig{ // Individual software GET endpoint is /api/v1/fleet/software/{id}
		// 	KeyColumns: plugin.SingleColumn("id"),
		// 	Hydrate:    getSoftware,
		// },
		Columns: []*plugin.Column{
			// Core software information
			{Name: "id", Type: proto.ColumnType_INT, Description: "Unique ID of the software item."},
			{Name: "name", Type: proto.ColumnType_STRING, Description: "Name of the software."},
			{Name: "version", Type: proto.ColumnType_STRING, Description: "Version of the software."},
			{Name: "source", Type: proto.ColumnType_STRING, Description: "Source of the software information (e.g., 'apps', 'deb_packages', 'chrome_extensions')."},
			{Name: "host_count", Type: proto.ColumnType_INT, Description: "Number of hosts where this software is installed."},
			{Name: "generated_cpe", Type: proto.ColumnType_STRING, Description: "Generated Common Platform Enumeration (CPE) string for the software."},
			{Name: "bundle_identifier", Type: proto.ColumnType_STRING, Description: "Bundle identifier, typically for macOS and iOS software."},
			{Name: "release", Type: proto.ColumnType_STRING, Description: "Release information, e.g., for RPM packages."},
			{Name: "vendor", Type: proto.ColumnType_STRING, Description: "Vendor information, e.g., for RPM packages."},
			{Name: "arch", Type: proto.ColumnType_STRING, Description: "Architecture information, e.g., for RPM packages."},
			{Name: "extension_id", Type: proto.ColumnType_STRING, Description: "Extension ID for browser extensions."},
			{Name: "browser", Type: proto.ColumnType_STRING, Description: "Browser name for browser extensions."},
			{Name: "path", Type: proto.ColumnType_STRING, Description: "Install path for certain software types like Programs."},
			{Name: "installed_path", Type: proto.ColumnType_STRING, Description: "Installed path, e.g., for Homebrew packages."},
			{Name: "last_opened_at", Type: proto.ColumnType_TIMESTAMP, Description: "Timestamp when the software was last opened (may be aggregated or host-specific)."},
			{Name: "counts_updated_at", Type: proto.ColumnType_TIMESTAMP, Description: "Timestamp when the host_count for this software item was last updated."},

			// Vulnerabilities - stored as JSONB as it's an array of complex objects
			// Users can query into this using JSON functions in SQL.
			{Name: "vulnerabilities", Type: proto.ColumnType_JSON, Description: "Vulnerabilities associated with this software."},

			// Query parameters that can be used for filtering
			{Name: "vulnerable_only", Type: proto.ColumnType_BOOL, Transform: transform.FromQual("vulnerable_only"), Description: "Filter for software with known vulnerabilities. Set in WHERE clause."},
			{Name: "os_id", Type: proto.ColumnType_INT, Transform: transform.FromQual("os_id"), Description: "Filter by OS ID. Set in WHERE clause."},
			{Name: "os_name", Type: proto.ColumnType_STRING, Transform: transform.FromQual("os_name"), Description: "Filter by OS name. Set in WHERE clause."},
			{Name: "os_version", Type: proto.ColumnType_STRING, Transform: transform.FromQual("os_version"), Description: "Filter by OS version. Set in WHERE clause."},
			{Name: "team_id", Type: proto.ColumnType_INT, Transform: transform.FromQual("team_id"), Description: "Filter by team ID. Set in WHERE clause."},

			// Connection config (server_url)
			{Name: "server_url", Type: proto.ColumnType_STRING, Hydrate: getServerURL, Transform: transform.FromValue(), Description: "FleetDM server URL from connection config."},
		},
	}
}

func listSoftware(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := NewFleetDMClient(ctx, d.Connection)
	if err != nil {
		plugin.Logger(ctx).Error("fleetdm_software.listSoftware", "connection_error", err)
		return nil, err
	}

	// Pagination parameters for software endpoint
	page := 0
	perPage := 100 // Default and often max for FleetDM software listing

	// Limiting the results
	limit := d.QueryContext.Limit
	if limit != nil && *limit < int64(perPage) {
		// perPage = int(*limit) // Be cautious if API has minimum per_page
	}

	for {
		params := url.Values{}
		params.Add("page", strconv.Itoa(page))
		params.Add("per_page", strconv.Itoa(perPage))
		// params.Add("order_key", "name") // Example: default sort, API might have its own default
		// params.Add("order_direction", "asc")

		// Add query qualifiers if provided
		if d.EqualsQuals["vulnerable_only"] != nil {
			params.Add("vulnerable", strconv.FormatBool(d.EqualsQuals["vulnerable_only"].GetBoolValue()))
		}
		if d.EqualsQuals["os_id"] != nil {
			params.Add("os_id", strconv.FormatInt(d.EqualsQuals["os_id"].GetInt64Value(), 10))
		}
		if d.EqualsQuals["os_name"] != nil {
			params.Add("os_name", d.EqualsQuals["os_name"].GetStringValue())
		}
		if d.EqualsQuals["os_version"] != nil {
			params.Add("os_version", d.EqualsQuals["os_version"].GetStringValue())
		}
		if d.EqualsQuals["team_id"] != nil {
			params.Add("team_id", strconv.FormatInt(d.EqualsQuals["team_id"].GetInt64Value(), 10))
		}

		var response ListSoftwareResponse
		_, err := client.Get(ctx, "software", params, &response) // Endpoint is /api/v1/fleet/software
		if err != nil {
			plugin.Logger(ctx).Error("fleetdm_software.listSoftware", "api_error", err, "page", page, "params", params.Encode())
			return nil, err
		}

		for _, swItem := range response.Software {
			d.StreamListItem(ctx, swItem)
			if d.RowsRemaining(ctx) == 0 {
				plugin.Logger(ctx).Debug("fleetdm_software.listSoftware", "limit_reached", true)
				return nil, nil
			}
		}

		if !response.Meta.HasNextResults {
			plugin.Logger(ctx).Debug("fleetdm_software.listSoftware", "end_of_results", true, "total_software_count_from_api", response.Count)
			break
		}

		page++
		plugin.Logger(ctx).Debug("fleetdm_software.listSoftware", "next_page", page)
	}

	return nil, nil
}

// TODO: Implement getSoftware if you add a GetConfig to the table
// func getSoftware(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
// 	id := d.EqualsQuals["id"].GetInt64Value()
// 	if id == 0 {
// 		return nil, nil
// 	}

// 	client, err := NewFleetDMClient(ctx, d.Connection)
// 	if err != nil {
// 		plugin.Logger(ctx).Error("fleetdm_software.getSoftware", "connection_error", err)
// 		return nil, err
// 	}

// 	var softwareItem Software // Assuming the Get endpoint returns a single Software item directly, not nested
// 	_, err = client.Get(ctx, fmt.Sprintf("software/%d", id), nil, &softwareItem)

// 	if err != nil {
// 		// Handle 404 Not Found if necessary
// 		plugin.Logger(ctx).Error("fleetdm_software.getSoftware", "api_error", err, "software_id", id)
// 		return nil, err
// 	}
// 	return softwareItem, nil
// }
