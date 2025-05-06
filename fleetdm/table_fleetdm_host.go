package fleetdm

import (
	"context"
	"encoding/json" // For json.RawMessage
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

// HostIssues represents the 'issues' object for a host.
type HostIssues struct {
	FailingPoliciesCount         int `json:"failing_policies_count"`
	CriticalVulnerabilitiesCount int `json:"critical_vulnerabilities_count"`
	TotalIssuesCount             int `json:"total_issues_count"`
}

// HostMDM represents the 'mdm' object for a host.
type HostMDM struct {
	EnrollmentStatus       string  `json:"enrollment_status"`
	DEPProfileError        bool    `json:"dep_profile_error"`
	ServerURL              *string `json:"server_url"` // Pointer as it might be null
	Name                   *string `json:"name"`       // Pointer as it might be null
	EncryptionKeyAvailable bool    `json:"encryption_key_available"`
	ConnectedToFleet       *bool   `json:"connected_to_fleet"` // Pointer as it might be null
}

// Host represents a FleetDM host.
// Refer to: https://fleetdm.com/docs/rest-api/rest-api#host-object-details
type Host struct {
	ID                            int              `json:"id"`
	CreatedAt                     time.Time        `json:"created_at"`
	UpdatedAt                     time.Time        `json:"updated_at"`
	SoftwareUpdatedAt             *time.Time       `json:"software_updated_at"`
	DetailUpdatedAt               time.Time        `json:"detail_updated_at"`
	LabelUpdatedAt                time.Time        `json:"label_updated_at"`
	PolicyUpdatedAt               time.Time        `json:"policy_updated_at"`
	LastEnrolledAt                time.Time        `json:"last_enrolled_at"`
	SeenTime                      time.Time        `json:"seen_time"`
	RefetchRequested              bool             `json:"refetch_requested"`
	OsqueryHostID                 *string          `json:"osquery_host_id"`
	NodeKey                       *string          `json:"node_key"`
	UUID                          string           `json:"uuid"`
	Hostname                      string           `json:"hostname"`
	DisplayName                   string           `json:"display_name"`
	DisplayText                   string           `json:"display_text"`
	ComputerName                  string           `json:"computer_name"`
	Platform                      string           `json:"platform"`
	PlatformLike                  string           `json:"platform_like"`
	OsVersion                     string           `json:"os_version"`
	Build                         string           `json:"build"`
	CodeName                      string           `json:"code_name"`
	Uptime                        int64            `json:"uptime"` // Nanoseconds
	Memory                        int64            `json:"memory"` // bytes
	CPUType                       string           `json:"cpu_type"`
	CPUSubtype                    string           `json:"cpu_subtype"`
	CPUBrand                      string           `json:"cpu_brand"`
	CPUPhysicalCores              int              `json:"cpu_physical_cores"`
	CPULogicalCores               int              `json:"cpu_logical_cores"`
	HardwareVendor                string           `json:"hardware_vendor"`
	HardwareModel                 string           `json:"hardware_model"`
	HardwareVersion               string           `json:"hardware_version"`
	HardwareSerial                string           `json:"hardware_serial"`
	PrimaryIP                     string           `json:"primary_ip"`
	PrimaryMac                    string           `json:"primary_mac"`
	PublicIP                      string           `json:"public_ip"`
	OrbitVersion                  *string          `json:"orbit_version"`
	FleetDesktopVersion           *string          `json:"fleet_desktop_version"`
	ScriptsEnabled                *bool            `json:"scripts_enabled"`
	OsqueryVersion                *string          `json:"osquery_version"`
	TeamID                        *int             `json:"team_id"`
	TeamName                      *string          `json:"team_name"`
	DistributedInterval           *int             `json:"distributed_interval"`
	ConfigTLSRefresh              *int             `json:"config_tls_refresh"`
	LoggerTLSPeriod               *int             `json:"logger_tls_period"`
	PackStats                     *json.RawMessage `json:"pack_stats"`
	GigsDiskSpaceAvailable        float64          `json:"gigs_disk_space_available"`
	PercentDiskSpaceAvailable     float64          `json:"percent_disk_space_available"` // JSON sample shows int, using float64 for flexibility.
	GigsTotalDiskSpace            float64          `json:"gigs_total_disk_space"`
	Status                        string           `json:"status"` // online, offline, mia
	Issues                        *HostIssues      `json:"issues"`
	MDM                           *HostMDM         `json:"mdm"`
	RefetchCriticalQueriesUntil   *time.Time       `json:"refetch_critical_queries_until"`
	LastRestartedAt               *time.Time       `json:"last_restarted_at"`
}

// ListHostsResponse is the expected structure for the list hosts API call.
type ListHostsResponse struct {
	Hosts []Host `json:"hosts"`
}

func tableFleetdmHost(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "fleetdm_host",
		Description: "Information about hosts managed by FleetDM.",
		List: &plugin.ListConfig{
			Hydrate: listHosts,
			// TODO: Add KeyColumns for filtering if the API supports it directly for list calls.
			// Example:
			// KeyColumns: []*plugin.KeyColumn{
			//  {Name: "platform", Require: plugin.Optional},
			//  {Name: "team_id", Require: plugin.Optional},
			// },
		},
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("id"),
			Hydrate:    getHost,
		},
		Columns: []*plugin.Column{
			// Core Identification
			{Name: "id", Type: proto.ColumnType_INT, Description: "The unique ID of the host."},
			{Name: "hostname", Type: proto.ColumnType_STRING, Description: "The hostname of the host."},
			{Name: "uuid", Type: proto.ColumnType_STRING, Description: "The unique UUID of the host."},
			{Name: "display_name", Type: proto.ColumnType_STRING, Description: "The display name of the host."},
			{Name: "display_text", Type: proto.ColumnType_STRING, Description: "The display text for the host (often same as hostname or display_name)."},
			{Name: "computer_name", Type: proto.ColumnType_STRING, Description: "The computer name of the host."},
			{Name: "osquery_host_id", Type: proto.ColumnType_STRING, Description: "The osquery host identifier."},
			{Name: "node_key", Type: proto.ColumnType_STRING, Description: "The node key for the host."},

			// Status and Timestamps
			{Name: "status", Type: proto.ColumnType_STRING, Description: "The current status of the host (online, offline, mia)."},
			{Name: "seen_time", Type: proto.ColumnType_TIMESTAMP, Description: "Timestamp when the host was last seen by Fleet."},
			{Name: "created_at", Type: proto.ColumnType_TIMESTAMP, Description: "Timestamp when the host was created in Fleet."},
			{Name: "updated_at", Type: proto.ColumnType_TIMESTAMP, Description: "Timestamp when the host record was last updated in Fleet."},
			{Name: "software_updated_at", Type: proto.ColumnType_TIMESTAMP, Description: "Timestamp when the host software inventory was last updated."},
			{Name: "detail_updated_at", Type: proto.ColumnType_TIMESTAMP, Description: "Timestamp when the host details were last updated."},
			{Name: "label_updated_at", Type: proto.ColumnType_TIMESTAMP, Description: "Timestamp when the host labels were last updated."},
			{Name: "policy_updated_at", Type: proto.ColumnType_TIMESTAMP, Description: "Timestamp when the host policy status was last updated."},
			{Name: "last_enrolled_at", Type: proto.ColumnType_TIMESTAMP, Description: "Timestamp when the host last enrolled."},
			{Name: "last_restarted_at", Type: proto.ColumnType_TIMESTAMP, Description: "Timestamp of the last host restart event."},
			{Name: "refetch_requested", Type: proto.ColumnType_BOOL, Description: "Indicates if a refetch of host details has been requested."},
			{Name: "refetch_critical_queries_until", Type: proto.ColumnType_TIMESTAMP, Description: "Timestamp until which critical queries will be refetched for this host."},

			// OS and Platform
			{Name: "platform", Type: proto.ColumnType_STRING, Description: "The platform of the host (e.g., 'darwin', 'windows', 'linux')."},
			{Name: "platform_like", Type: proto.ColumnType_STRING, Description: "Platform-like classification (e.g., 'darwin')."},
			{Name: "os_version", Type: proto.ColumnType_STRING, Description: "The operating system version."},
			{Name: "build", Type: proto.ColumnType_STRING, Description: "The operating system build string."},
			{Name: "code_name", Type: proto.ColumnType_STRING, Description: "The OS code name."},
			{Name: "osquery_version", Type: proto.ColumnType_STRING, Description: "The version of osquery running on the host."},
			{Name: "orbit_version", Type: proto.ColumnType_STRING, Description: "The version of Orbit running on the host."},
			{Name: "fleet_desktop_version", Type: proto.ColumnType_STRING, Description: "The version of Fleet Desktop running on the host."},
			{Name: "scripts_enabled", Type: proto.ColumnType_BOOL, Description: "Indicates if running scripts is enabled for this host via Fleet."},

			// Hardware
			{Name: "uptime", Type: proto.ColumnType_INT, Description: "Uptime of the host in nanoseconds."},
			{Name: "memory", Type: proto.ColumnType_INT, Description: "Total physical memory in bytes."},
			{Name: "cpu_type", Type: proto.ColumnType_STRING, Description: "CPU type."},
			{Name: "cpu_subtype", Type: proto.ColumnType_STRING, Description: "CPU subtype."},
			{Name: "cpu_brand", Type: proto.ColumnType_STRING, Description: "CPU brand string."},
			{Name: "cpu_physical_cores", Type: proto.ColumnType_INT, Description: "Number of physical CPU cores."},
			{Name: "cpu_logical_cores", Type: proto.ColumnType_INT, Description: "Number of logical CPU cores."},
			{Name: "hardware_vendor", Type: proto.ColumnType_STRING, Description: "Hardware vendor."},
			{Name: "hardware_model", Type: proto.ColumnType_STRING, Description: "Hardware model."},
			{Name: "hardware_version", Type: proto.ColumnType_STRING, Description: "Hardware version."},
			{Name: "hardware_serial", Type: proto.ColumnType_STRING, Description: "Hardware serial number."},

			// Network
			{Name: "primary_ip", Type: proto.ColumnType_IPADDR, Description: "The primary IP address of the host."},
			{Name: "primary_mac", Type: proto.ColumnType_STRING, Description: "The primary MAC address of the host."},
			{Name: "public_ip", Type: proto.ColumnType_IPADDR, Description: "The public IP address of the host."},

			// Fleet Configuration
			{Name: "team_id", Type: proto.ColumnType_INT, Description: "The ID of the team the host belongs to, if any."},
			{Name: "team_name", Type: proto.ColumnType_STRING, Description: "The name of the team the host belongs to, if any."},
			{Name: "distributed_interval", Type: proto.ColumnType_INT, Description: "The distributed query interval for the host."},
			{Name: "config_tls_refresh", Type: proto.ColumnType_INT, Description: "The config TLS refresh interval."},
			{Name: "logger_tls_period", Type: proto.ColumnType_INT, Description: "The logger TLS period."},
			{Name: "pack_stats", Type: proto.ColumnType_JSON, Description: "Statistics for query packs on the host."},

			// Disk Space
			{Name: "gigs_disk_space_available", Type: proto.ColumnType_DOUBLE, Description: "Gigabytes of disk space available."},
			{Name: "percent_disk_space_available", Type: proto.ColumnType_DOUBLE, Description: "Percentage of disk space available. JSON sample shows int, using float64 for flexibility."},
			{Name: "gigs_total_disk_space", Type: proto.ColumnType_DOUBLE, Description: "Total gigabytes of disk space."},

			// Issues and MDM (as JSON columns for now, could be expanded)
			{Name: "issues", Type: proto.ColumnType_JSON, Description: "Host issues summary (failing policies, vulnerabilities)."},
			{Name: "mdm", Type: proto.ColumnType_JSON, Description: "Mobile Device Management (MDM) information for the host."},

			// Connection config
			{Name: "server_url", Type: proto.ColumnType_STRING, Hydrate: getServerURL, Transform: transform.FromValue(), Description: "FleetDM server URL from connection config."},
		},
	}
}

// listHosts fetches a list of hosts from the FleetDM API.
func listHosts(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := NewFleetDMClient(ctx, d.Connection)
	if err != nil {
		plugin.Logger(ctx).Error("fleetdm_host.listHosts", "connection_error", err)
		return nil, err
	}

	page := 0
	perPage := 100

	limit := d.QueryContext.Limit
	if limit != nil && *limit > 0 && *limit < int64(perPage) {
		// perPage = int(*limit) // Consider API minimums if any.
	}

	for {
		params := url.Values{}
		params.Add("page", strconv.Itoa(page))
		params.Add("per_page", strconv.Itoa(perPage))
		// TODO: Add any query qualifiers if API supports them (e.g., platform, team_id)
		// Example:
		// if d.EqualsQuals["platform"] != nil {
		// 	params.Add("platform", d.EqualsQuals["platform"].GetStringValue())
		// }
		// if d.EqualsQuals["team_id"] != nil {
		// 	params.Add("team_id", strconv.FormatInt(d.EqualsQuals["team_id"].GetInt64Value(), 10))
		// }


		// The /api/v1/fleet/hosts endpoint returns { "hosts": [...] }
		var response ListHostsResponse
		_, err := client.Get(ctx, "hosts", params, &response)
		if err != nil {
			plugin.Logger(ctx).Error("fleetdm_host.listHosts", "api_error", err, "page", page, "params", params.Encode())
			return nil, err
		}

		for _, host := range response.Hosts {
			d.StreamListItem(ctx, host)
			if d.RowsRemaining(ctx) == 0 {
				plugin.Logger(ctx).Debug("fleetdm_host.listHosts", "limit_reached", true)
				return nil, nil
			}
		}

		if len(response.Hosts) < perPage {
			plugin.Logger(ctx).Debug("fleetdm_host.listHosts", "end_of_results", true, "hosts_count_on_page", len(response.Hosts))
			break
		}

		page++
		plugin.Logger(ctx).Debug("fleetdm_host.listHosts", "next_page", page)
	}

	return nil, nil
}

// getHost fetches a single host by ID.
func getHost(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	id := d.EqualsQuals["id"].GetInt64Value()
	if id == 0 {
		return nil, nil // Invalid ID
	}

	client, err := NewFleetDMClient(ctx, d.Connection)
	if err != nil {
		plugin.Logger(ctx).Error("fleetdm_host.getHost", "connection_error", err)
		return nil, err
	}

	// The API endpoint GET /api/v1/fleet/hosts/{id} returns { "host": { ...host_object... } }
	var response struct {
		Host Host `json:"host"`
	}

	endpointPath := fmt.Sprintf("hosts/%d", id)
	_, err = client.Get(ctx, endpointPath, nil, &response)

	if err != nil {
		// TODO: Handle 404 Not Found specifically (return nil, nil for Steampipe)
		// Example:
		// if httpResp, ok := err.(interface{ StatusCode() int }); ok && httpResp.StatusCode() == http.StatusNotFound {
		// 	plugin.Logger(ctx).Info("fleetdm_host.getHost", "host_not_found", id)
		// 	return nil, nil
		// }
		plugin.Logger(ctx).Error("fleetdm_host.getHost", "api_error", err, "host_id", id, "endpoint", endpointPath)
		return nil, err
	}

	return response.Host, nil
}