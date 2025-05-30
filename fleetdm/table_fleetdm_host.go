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

// HostUser represents a user on a specific host.
type HostUser struct {
	UID       int    `json:"uid"`
	Username  string `json:"username"`
	Type      string `json:"type"`
	Groupname string `json:"groupname"`
	Shell     string `json:"shell"`
}

// HostPolicy represents a policy's status on a specific host.
type HostPolicy struct {
	ID                    uint      `json:"id"`
	Name                  string    `json:"name"`
	Query                 string    `json:"query"`
	Critical              bool      `json:"critical"`
	Description           string    `json:"description"`
	AuthorID              *uint     `json:"author_id"`
	AuthorName            string    `json:"author_name"`
	AuthorEmail           string    `json:"author_email"`
	TeamID                *uint     `json:"team_id"`
	Resolution            string    `json:"resolution"`
	Platform              string    `json:"platform"`
	CalendarEventsEnabled bool      `json:"calendar_events_enabled"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
	Response              string    `json:"response"` // e.g., "pass", "fail"
}

// HostLabel represents a label applied to a specific host.
type HostLabel struct {
	ID                  uint      `json:"id"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
	Name                string    `json:"name"`
	Description         string    `json:"description"`
	Query               string    `json:"query"`
	Platform            string    `json:"platform"`
	LabelType           string    `json:"label_type"`
	LabelMembershipType string    `json:"label_membership_type"`
	AuthorID            *uint     `json:"author_id,omitempty"`
}

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
	ServerURL              *string `json:"server_url"`
	Name                   *string `json:"name"`
	EncryptionKeyAvailable bool    `json:"encryption_key_available"`
	ConnectedToFleet       *bool   `json:"connected_to_fleet"`
}

// DeviceMappingItem represents an item within the device_mapping array.
type DeviceMappingItem struct {
	Email  string `json:"email"`
	Source string `json:"source"`
}

// Host represents a FleetDM host.
// Refer to: https://fleetdm.com/docs/rest-api/rest-api#list-hosts
type Host struct {
	ID                            int                 `json:"id"`
	CreatedAt                     time.Time           `json:"created_at"`
	UpdatedAt                     time.Time           `json:"updated_at"`
	SoftwareUpdatedAt             *time.Time          `json:"software_updated_at"`
	DetailUpdatedAt               time.Time           `json:"detail_updated_at"`
	LabelUpdatedAt                time.Time           `json:"label_updated_at"`
	PolicyUpdatedAt               time.Time           `json:"policy_updated_at"`
	LastEnrolledAt                time.Time           `json:"last_enrolled_at"`
	SeenTime                      time.Time           `json:"seen_time"`
	RefetchRequested              bool                `json:"refetch_requested"`
	UUID                          string              `json:"uuid"`
	Hostname                      string              `json:"hostname"`
	DisplayName                   string              `json:"display_name"`
	DisplayText                   string              `json:"display_text"`
	ComputerName                  string              `json:"computer_name"`
	Platform                      string              `json:"platform"`
	PlatformLike                  string              `json:"platform_like"`
	OsVersion                     string              `json:"os_version"`
	Build                         string              `json:"build"`
	CodeName                      string              `json:"code_name"`
	Uptime                        int64               `json:"uptime"` // Nanoseconds
	Memory                        int64               `json:"memory"` // bytes
	CPUType                       string              `json:"cpu_type"`
	CPUSubtype                    string              `json:"cpu_subtype"`
	CPUBrand                      string              `json:"cpu_brand"`
	CPUPhysicalCores              int                 `json:"cpu_physical_cores"`
	CPULogicalCores               int                 `json:"cpu_logical_cores"`
	HardwareVendor                string              `json:"hardware_vendor"`
	HardwareModel                 string              `json:"hardware_model"`
	HardwareVersion               string              `json:"hardware_version"`
	HardwareSerial                string              `json:"hardware_serial"`
	PrimaryIP                     string              `json:"primary_ip"`
	PrimaryMac                    string              `json:"primary_mac"`
	PublicIP                      string              `json:"public_ip"`
	OrbitVersion                  *string             `json:"orbit_version"`
	FleetDesktopVersion           *string             `json:"fleet_desktop_version"`
	ScriptsEnabled                *bool               `json:"scripts_enabled"`
	OsqueryVersion                *string             `json:"osquery_version"`
	TeamID                        *int                `json:"team_id"`
	TeamName                      *string             `json:"team_name"`
	DistributedInterval           *int                `json:"distributed_interval"`
	ConfigTLSRefresh              *int                `json:"config_tls_refresh"`
	LoggerTLSPeriod               *int                `json:"logger_tls_period"`
	PackStats                     *json.RawMessage    `json:"pack_stats"`
	GigsDiskSpaceAvailable        float64             `json:"gigs_disk_space_available"`
	PercentDiskSpaceAvailable     float64             `json:"percent_disk_space_available"`
	GigsTotalDiskSpace            float64             `json:"gigs_total_disk_space"`
	Status                        string              `json:"status"`
	Issues                        *HostIssues         `json:"issues"`
	MDM                           *HostMDM            `json:"mdm"`
	RefetchCriticalQueriesUntil   *time.Time          `json:"refetch_critical_queries_until"`
	LastRestartedAt               *time.Time          `json:"last_restarted_at"`
	Users                         []HostUser          `json:"users,omitempty"`
	Policies                      []HostPolicy        `json:"policies,omitempty"`
	Labels                        []HostLabel         `json:"labels,omitempty"`
	DeviceMapping                 []DeviceMappingItem `json:"device_mapping,omitempty"` // Updated from *json.RawMessage
}

// ListHostsResponse is the expected structure for the list hosts API call.
type ListHostsResponse struct {
	Hosts []Host `json:"hosts"`
}

// Custom transform to ensure MDM struct is marshalled to JSON string
func mdmToJSONString(ctx context.Context, d *transform.TransformData) (interface{}, error) {
	if d.Value == nil {
		return nil, nil
	}
	mdmData, ok := d.Value.(*HostMDM)
	if !ok {
		return nil, fmt.Errorf("mdmToJSONString: type assertion to *HostMDM failed for type %T", d.Value)
	}
	if mdmData == nil {
		return nil, nil
	}

	// Marshal the HostMDM struct
	jsonBytes, err := json.Marshal(mdmData)
	if err != nil {
		return nil, err
	}
	return string(jsonBytes), nil
}

// Custom transform to ensure other JSONB fields are marshalled correctly
func arrayOrObjectToJSONString(ctx context.Context, d *transform.TransformData) (interface{}, error) {
	if d.Value == nil {
		return nil, nil
	}
	// This function can handle slices of structs (like Users, Policies, Labels, DeviceMapping)
	// or pointers to structs (like Issues, PackStats if it were a struct).
	// It will also handle *json.RawMessage directly if that's what PackStats remains.
	jsonBytes, err := json.Marshal(d.Value)
	if err != nil {
		plugin.Logger(ctx).Error("arrayOrObjectToJSONString", "json_marshal_error", err, "value_type", fmt.Sprintf("%T", d.Value))
		return nil, err
	}
	return string(jsonBytes), nil
}


func tableFleetdmHost(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "fleetdm_host",
		Description: "Information about hosts managed by FleetDM.",
		List: &plugin.ListConfig{
			Hydrate: listHosts,
		},
		Columns: []*plugin.Column{
			// Core Identification
			{Name: "id", Type: proto.ColumnType_INT, Description: "The unique ID of the host."},
			{Name: "hostname", Type: proto.ColumnType_STRING, Description: "The hostname of the host."},
			{Name: "uuid", Type: proto.ColumnType_STRING, Description: "The unique UUID of the host."},
			{Name: "display_name", Type: proto.ColumnType_STRING, Description: "The display name of the host."},
			{Name: "display_text", Type: proto.ColumnType_STRING, Description: "The display text for the host (often same as hostname or display_name)."},
			{Name: "computer_name", Type: proto.ColumnType_STRING, Description: "The computer name of the host."},

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
			{Name: "pack_stats", Type: proto.ColumnType_JSON, Transform: transform.FromField("PackStats").Transform(arrayOrObjectToJSONString), Description: "Statistics for query packs on the host."},
			// Disk Space
			{Name: "gigs_disk_space_available", Type: proto.ColumnType_DOUBLE, Description: "Gigabytes of disk space available."},
			{Name: "percent_disk_space_available", Type: proto.ColumnType_DOUBLE, Description: "Percentage of disk space available. JSON sample shows int, using float64 for flexibility."},
			{Name: "gigs_total_disk_space", Type: proto.ColumnType_DOUBLE, Description: "Total gigabytes of disk space."},
			// Issues, MDM, Users, Policies, Labels
			{Name: "issues", Type: proto.ColumnType_JSON, Transform: transform.FromField("Issues").Transform(arrayOrObjectToJSONString), Description: "Host issues summary (failing policies, vulnerabilities)."},
			{Name: "mdm", Type: proto.ColumnType_JSON, Transform: transform.FromField("MDM").Transform(mdmToJSONString), Description: "Mobile Device Management (MDM) information for the host."},
			{Name: "users", Type: proto.ColumnType_JSON, Transform: transform.FromField("Users").Transform(arrayOrObjectToJSONString), Description: "Users on this host (requires populate_users=true)."},
			{Name: "policies", Type: proto.ColumnType_JSON, Transform: transform.FromField("Policies").Transform(arrayOrObjectToJSONString), Description: "Policy compliance status for this host (requires populate_policies=true)."},
			{Name: "labels", Type: proto.ColumnType_JSON, Transform: transform.FromField("Labels").Transform(arrayOrObjectToJSONString), Description: "Labels applied to this host (requires populate_labels=true)."},
			{Name: "device_mapping", Type: proto.ColumnType_JSON, Transform: transform.FromField("DeviceMapping").Transform(arrayOrObjectToJSONString), Description: "Device mapping information (requires device_mapping=true)."},
			{Name: "server_url", Type: proto.ColumnType_STRING, Hydrate: getServerURL, Transform: transform.FromValue(), Description: "FleetDM server URL from connection config."},
		},
	}
}

func addHostPopulationParams(params url.Values) {
	params.Add("device_mapping", "true")
	params.Add("populate_policies", "true")
	params.Add("populate_users", "true")
	params.Add("populate_labels", "true")
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

	for {
		params := url.Values{}
		params.Add("page", strconv.Itoa(page))
		params.Add("per_page", strconv.Itoa(perPage))
		params.Add("order_key", "id")
		params.Add("order_direction", "desc") // Get latest hosts first, or 'asc' for consistent paging
		
		addHostPopulationParams(params)
		plugin.Logger(ctx).Debug("fleetdm_host.listHosts", "request_params", params.Encode())
		
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
