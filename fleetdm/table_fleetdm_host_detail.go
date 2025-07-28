package fleetdm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

// --- New and Updated Structs based on detailed JSON sample ---

// HostMDMDetail represents the rich 'mdm' object.
type HostMDMDetail struct {
	EncryptionKeyAvailable bool             `json:"encryption_key_available"`
	EnrollmentStatus       string           `json:"enrollment_status"`
	Name                   *string          `json:"name"`
	ConnectedToFleet       *bool            `json:"connected_to_fleet"`
	ServerURL              *string          `json:"server_url"`
	DeviceStatus           string           `json:"device_status"`
	PendingAction          string           `json:"pending_action"`
	MacOSSettings          *json.RawMessage `json:"macos_settings"`
	MacOSSetup             *json.RawMessage `json:"macos_setup"`
	OsSettings             *json.RawMessage `json:"os_settings"`
	Profiles               *json.RawMessage `json:"profiles"`
}

// HostBattery represents a battery on a host.
type HostBattery struct {
	CycleCount int    `json:"cycle_count"`
	Health     string `json:"health"`
}

// HostGeometry represents the geometry part of geolocation.
type HostGeometry struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
}

// HostGeolocation represents the geolocation of a host.
type HostGeolocation struct {
	CountryISO string        `json:"country_iso"`
	CityName   string        `json:"city_name"`
	Geometry   *HostGeometry `json:"geometry"`
}

// HostMaintenanceWindow represents a configured maintenance window.
type HostMaintenanceWindow struct {
	StartsAt time.Time `json:"starts_at"`
	Timezone string    `json:"timezone"`
}

// HostOtherEmail represents an email entry for an end user.
type HostOtherEmail struct {
	Email  string `json:"email"`
	Source string `json:"source"`
}

// HostEndUser represents an end user associated with a device.
type HostEndUser struct {
	IdpInfoUpdatedAt time.Time        `json:"idp_info_updated_at"`
	IdpID            string           `json:"idp_id"`
	IdpUsername      string           `json:"idp_username"`
	IdpFullName      string           `json:"idp_full_name"`
	IdpGroups        []string         `json:"idp_groups"`
	OtherEmails      []HostOtherEmail `json:"other_emails"`
}

// HostSoftware represents a software item on a specific host.
type HostSoftware struct {
	ID               uint             `json:"id"`
	Name             string           `json:"name"`
	Version          string           `json:"version"`
	Source           string           `json:"source"`
	Browser          string           `json:"browser"`
	BundleIdentifier string           `json:"bundle_identifier"`
	LastOpenedAt     *time.Time       `json:"last_opened_at"`
	GeneratedCPE     string           `json:"generated_cpe"`
	Vulnerabilities  *json.RawMessage `json:"vulnerabilities"`
	InstalledPaths   []string         `json:"installed_paths"`
}

// HostDetail represents the full, rich host object from GET /hosts/:id
type HostDetail struct {
	ID                            int                    `json:"id"`
	CreatedAt                     time.Time              `json:"created_at"`
	UpdatedAt                     time.Time              `json:"updated_at"`
	SoftwareUpdatedAt             time.Time              `json:"software_updated_at"`
	DetailUpdatedAt               time.Time              `json:"detail_updated_at"`
	LabelUpdatedAt                time.Time              `json:"label_updated_at"`
	PolicyUpdatedAt               time.Time              `json:"policy_updated_at"`
	LastEnrolledAt                time.Time              `json:"last_enrolled_at"`
	LastMdmCheckedInAt            time.Time              `json:"last_mdm_checked_in_at"`
	LastMdmEnrolledAt             time.Time              `json:"last_mdm_enrolled_at"`
	LastRestartedAt               *time.Time             `json:"last_restarted_at"`
	SeenTime                      time.Time              `json:"seen_time"`
	RefetchRequested              bool                   `json:"refetch_requested"`
	Hostname                      string                 `json:"hostname"`
	UUID                          string                 `json:"uuid"`
	Platform                      string                 `json:"platform"`
	OsqueryVersion                string                 `json:"osquery_version"`
	OrbitVersion                  *string                `json:"orbit_version"`
	FleetDesktopVersion           *string                `json:"fleet_desktop_version"`
	ScriptsEnabled                *bool                  `json:"scripts_enabled"`
	OsVersion                     string                 `json:"os_version"`
	Build                         string                 `json:"build"`
	PlatformLike                  string                 `json:"platform_like"`
	CodeName                      string                 `json:"code_name"`
	Uptime                        int64                  `json:"uptime"`
	Memory                        int64                  `json:"memory"`
	CPUType                       string                 `json:"cpu_type"`
	CPUSubtype                    string                 `json:"cpu_subtype"`
	CPUBrand                      string                 `json:"cpu_brand"`
	CPUPhysicalCores              int                    `json:"cpu_physical_cores"`
	CPULogicalCores               int                    `json:"cpu_logical_cores"`
	HardwareVendor                string                 `json:"hardware_vendor"`
	HardwareModel                 string                 `json:"hardware_model"`
	HardwareVersion               string                 `json:"hardware_version"`
	HardwareSerial                string                 `json:"hardware_serial"`
	ComputerName                  string                 `json:"computer_name"`
	DisplayName                   string                 `json:"display_name"`
	PublicIP                      string                 `json:"public_ip"`
	PrimaryIP                     string                 `json:"primary_ip"`
	PrimaryMac                    string                 `json:"primary_mac"`
	DistributedInterval           int                    `json:"distributed_interval"`
	ConfigTLSRefresh              int                    `json:"config_tls_refresh"`
	LoggerTLSPeriod               int                    `json:"logger_tls_period"`
	TeamID                        *int                   `json:"team_id"`
	TeamName                      *string                `json:"team_name"`
	GigsDiskSpaceAvailable        float64                `json:"gigs_disk_space_available"`
	PercentDiskSpaceAvailable     float64                `json:"percent_disk_space_available"`
	GigsTotalDiskSpace            float64                `json:"gigs_total_disk_space"`
	DiskEncryptionEnabled         *bool                  `json:"disk_encryption_enabled"`
	Status                        string                 `json:"status"`
	DisplayText                   string                 `json:"display_text"`
	Additional                    *json.RawMessage       `json:"additional"`
	Issues                        *HostIssues            `json:"issues"`
	Batteries                     []HostBattery          `json:"batteries"`
	Geolocation                   *HostGeolocation       `json:"geolocation"`
	MaintenanceWindow             *HostMaintenanceWindow `json:"maintenance_window"`
	Users                         []HostUser             `json:"users"`
	EndUsers                      []HostEndUser          `json:"end_users"`
	Labels                        []HostLabel            `json:"labels"`
	Packs                         *json.RawMessage       `json:"packs"`
	Policies                      []HostPolicy           `json:"policies"`
	Software                      []HostSoftware         `json:"software"`
	MDM                           *HostMDMDetail         `json:"mdm"`
}

// Custom transform to ensure MDM struct is marshalled to JSON string
func mdmDetailToJSONString(ctx context.Context, d *transform.TransformData) (interface{}, error) {
	if d.Value == nil {
		return nil, nil
	}
	mdmData, ok := d.Value.(*HostMDMDetail)
	if !ok {
		return nil, fmt.Errorf("mdmDetailToJSONString: type assertion to *HostMDMDetail failed for type %T", d.Value)
	}
	if mdmData == nil {
		return nil, nil
	}

	// Marshal the HostMDMDetail struct
	jsonBytes, err := json.Marshal(mdmData)
	if err != nil {
		return nil, err
	}
	return string(jsonBytes), nil
}

// --- Table Definition ---

func tableFleetdmHostDetail(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "fleetdm_host_detail",
		Description: "Provides fully detailed information for each host by fetching details individually.",
		List: &plugin.ListConfig{
			Hydrate: listHostsForDetails,
		},
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("id"),
			Hydrate:    getHostDetails,
		},
		Columns: []*plugin.Column{
			// Columns from the basic host list call (NO HYDRATE)
			{Name: "id", Type: proto.ColumnType_INT, Description: "The unique ID of the host."},
			{Name: "hostname", Type: proto.ColumnType_STRING, Description: "The hostname of the host."},
			{Name: "display_name", Type: proto.ColumnType_STRING, Description: "The display name of the host."},
			{Name: "uuid", Type: proto.ColumnType_STRING, Description: "The unique UUID of the host."},
			{Name: "status", Type: proto.ColumnType_STRING, Description: "The current status of the host (online, offline, mia)."},
			{Name: "team_id", Type: proto.ColumnType_INT, Description: "The ID of the team the host belongs to, if any."},
			{Name: "team_name", Type: proto.ColumnType_STRING, Description: "The name of the team the host belongs to, if any."},
			{Name: "display_text", Type: proto.ColumnType_STRING, Description: "The display text for the host."},
			{Name: "computer_name", Type: proto.ColumnType_STRING, Description: "The computer name of the host."},
			{Name: "seen_time", Type: proto.ColumnType_TIMESTAMP, Description: "Timestamp when the host was last seen by Fleet."},
			{Name: "created_at", Type: proto.ColumnType_TIMESTAMP, Description: "Timestamp when the host was created in Fleet."},
			{Name: "updated_at", Type: proto.ColumnType_TIMESTAMP, Description: "Timestamp when the host record was last updated in Fleet."},
			{Name: "platform", Type: proto.ColumnType_STRING, Description: "The platform of the host (e.g., 'darwin', 'windows', 'linux')."},
			{Name: "os_version", Type: proto.ColumnType_STRING, Description: "The operating system version."},
			{Name: "osquery_version", Type: proto.ColumnType_STRING, Description: "The version of osquery running on the host."},
			{Name: "last_enrolled_at", Type: proto.ColumnType_TIMESTAMP, Description: "Timestamp when the host last enrolled."},
			{Name: "last_mdm_checked_in_at", Type: proto.ColumnType_TIMESTAMP, Description: "Timestamp when the host last checked in with MDM."},
			{Name: "last_mdm_enrolled_at", Type: proto.ColumnType_TIMESTAMP, Description: "Timestamp when the host was last enrolled in MDM."},
			{Name: "detail_updated_at", Type: proto.ColumnType_TIMESTAMP, Description: "Timestamp when the host details were last updated."},
			{Name: "label_updated_at", Type: proto.ColumnType_TIMESTAMP, Description: "Timestamp when the host labels were last updated."},
			{Name: "policy_updated_at", Type: proto.ColumnType_TIMESTAMP, Description: "Timestamp when the host policy status was last updated."},
			{Name: "software_updated_at", Type: proto.ColumnType_TIMESTAMP, Description: "Timestamp when the host software inventory was last updated."},
			{Name: "last_restarted_at", Type: proto.ColumnType_TIMESTAMP, Description: "Timestamp of the last host restart event."},
			{Name: "platform_like", Type: proto.ColumnType_STRING, Description: "Platform-like classification (e.g., 'darwin')."},
			{Name: "build", Type: proto.ColumnType_STRING, Description: "The operating system build string."},
			{Name: "code_name", Type: proto.ColumnType_STRING, Description: "The OS code name."},
			{Name: "orbit_version", Type: proto.ColumnType_STRING, Description: "The version of Orbit running on the host."},
			{Name: "fleet_desktop_version", Type: proto.ColumnType_STRING, Description: "The version of Fleet Desktop running on the host."},
			{Name: "scripts_enabled", Type: proto.ColumnType_BOOL, Description: "Indicates if running scripts is enabled for this host via Fleet."},
			{Name: "refetch_requested", Type: proto.ColumnType_BOOL, Description: "Indicates if a refetch of host details has been requested."},
			{Name: "hardware_model", Type: proto.ColumnType_STRING, Description: "Hardware model."},
			{Name: "hardware_serial", Type: proto.ColumnType_STRING, Description: "Hardware serial number."},
			{Name: "hardware_vendor", Type: proto.ColumnType_STRING, Description: "Hardware vendor."},
			{Name: "hardware_version", Type: proto.ColumnType_STRING, Description: "Hardware version."},
			{Name: "disk_encryption_enabled", Type: proto.ColumnType_BOOL, Description: "Indicates if disk encryption is enabled on the host."},
			{Name: "uptime", Type: proto.ColumnType_INT, Description: "Uptime of the host in nanoseconds."},
			{Name: "memory", Type: proto.ColumnType_INT, Description: "Total physical memory in bytes."},
			{Name: "cpu_type", Type: proto.ColumnType_STRING, Description: "CPU type."},
			{Name: "cpu_subtype", Type: proto.ColumnType_STRING, Description: "CPU subtype."},
			{Name: "cpu_brand", Type: proto.ColumnType_STRING, Description: "CPU brand string."},
			{Name: "cpu_physical_cores", Type: proto.ColumnType_INT, Description: "Number of physical CPU cores."},
			{Name: "cpu_logical_cores", Type: proto.ColumnType_INT, Description: "Number of logical CPU cores."},
			{Name: "gigs_disk_space_available", Type: proto.ColumnType_DOUBLE, Description: "Gigabytes of disk space available."},
			{Name: "percent_disk_space_available", Type: proto.ColumnType_DOUBLE, Description: "Percentage of disk space available."},
			{Name: "gigs_total_disk_space", Type: proto.ColumnType_DOUBLE, Description: "Total gigabytes of disk space."},
			{Name: "public_ip", Type: proto.ColumnType_IPADDR, Description: "The public IP address of the host."},
			{Name: "primary_ip", Type: proto.ColumnType_IPADDR, Description: "The primary IP address of the host."},
			{Name: "primary_mac", Type: proto.ColumnType_STRING, Description: "The primary MAC address of the host."},
			{Name: "distributed_interval", Type: proto.ColumnType_INT, Description: "The distributed query interval for the host."},
			{Name: "config_tls_refresh", Type: proto.ColumnType_INT, Description: "The config TLS refresh interval."},
			{Name: "logger_tls_period", Type: proto.ColumnType_INT, Description: "The logger TLS period."},

			// Columns that require the getHostDetails hydration call (HAS HYDRATE)
			{Name: "refetch_critical_queries_until", Type: proto.ColumnType_TIMESTAMP, Hydrate: getHostDetails, Description: "Timestamp until which critical queries will be refetched for this host."},
			{Name: "users", Type: proto.ColumnType_JSON, Hydrate: getHostDetails, Transform: transform.FromField("Users").Transform(arrayOrObjectToJSONString), Description: "Local users on this host."},
			{Name: "end_users", Type: proto.ColumnType_JSON, Hydrate: getHostDetails, Transform: transform.FromField("EndUsers").Transform(arrayOrObjectToJSONString), Description: "End users associated with this device via IdP or other mappings."},
			{Name: "policies", Type: proto.ColumnType_JSON, Hydrate: getHostDetails, Transform: transform.FromField("Policies").Transform(arrayOrObjectToJSONString), Description: "Policy compliance status for this host."},
			{Name: "labels", Type: proto.ColumnType_JSON, Hydrate: getHostDetails, Transform: transform.FromField("Labels").Transform(arrayOrObjectToJSONString), Description: "Labels applied to this host."},
			{Name: "software", Type: proto.ColumnType_JSON, Hydrate: getHostDetails, Transform: transform.FromField("Software").Transform(arrayOrObjectToJSONString), Description: "Software installed on this host."},
			{Name: "mdm", Type: proto.ColumnType_JSON, Hydrate: getHostDetails, Transform: transform.FromField("MDM").Transform(mdmDetailToJSONString), Description: "Mobile Device Management (MDM) information for the host."},
			{Name: "issues", Type: proto.ColumnType_JSON, Hydrate: getHostDetails, Transform: transform.FromField("Issues").Transform(arrayOrObjectToJSONString), Description: "Host issues summary."},
			{Name: "batteries", Type: proto.ColumnType_JSON, Hydrate: getHostDetails, Transform: transform.FromField("Batteries").Transform(arrayOrObjectToJSONString), Description: "Battery information for the host."},
			{Name: "geolocation", Type: proto.ColumnType_JSON, Hydrate: getHostDetails, Transform: transform.FromField("Geolocation").Transform(arrayOrObjectToJSONString), Description: "Geolocation information for the host."},
			{Name: "maintenance_window", Type: proto.ColumnType_JSON, Hydrate: getHostDetails, Transform: transform.FromField("MaintenanceWindow").Transform(arrayOrObjectToJSONString), Description: "Configured maintenance window for the host."},
			{Name: "additional", Type: proto.ColumnType_JSON, Hydrate: getHostDetails, Transform: transform.FromField("Additional").Transform(arrayOrObjectToJSONString), Description: "Additional custom details for the host."},
			{Name: "packs", Type: proto.ColumnType_JSON, Hydrate: getHostDetails, Transform: transform.FromField("Packs").Transform(arrayOrObjectToJSONString), Description: "Query packs applied to the host."},
		},
	}
}

// listHostsForDetails gets the minimal host object for hydration.
func listHostsForDetails(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := NewFleetDMClient(ctx, d.Connection)
	if err != nil {
		plugin.Logger(ctx).Error("fleetdm_host_detail.listHostsForDetails", "connection_error", err)
		return nil, err
	}

	page := 0
	perPage := 100

	for {
		params := url.Values{}
		params.Add("page", fmt.Sprintf("%d", page))
		params.Add("per_page", fmt.Sprintf("%d", perPage))
		params.Add("order_key", "id")
		params.Add("order_direction", "asc")

		var response ListHostsResponse
		_, err := client.Get(ctx, "hosts", params, &response)
		if err != nil {
			plugin.Logger(ctx).Error("fleetdm_host_detail.listHostsForDetails", "api_error", err, "page", page)
			return nil, err
		}

		for _, host := range response.Hosts {
			d.StreamListItem(ctx, host)
			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}

		if len(response.Hosts) < perPage {
			break
		}
		page++
	}

	return nil, nil
}

// getHostDetails is the hydrate function that fetches rich details for a single host.
func getHostDetails(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	var hostID int
	if h.Item != nil {
		// h.Item is the minimal Host object from listHostsForDetails
		hostFromList := h.Item.(Host) // Use the simpler Host struct here
		hostID = hostFromList.ID
	} else {
		hostID = int(d.EqualsQuals["id"].GetInt64Value())
	}
	
	if hostID == 0 {
		return nil, nil
	}
	
	plugin.Logger(ctx).Info("fleetdm_host_detail.getHostDetails", "hydrating_host_id", hostID)

	client, err := NewFleetDMClient(ctx, d.Connection)
	if err != nil {
		plugin.Logger(ctx).Error("fleetdm_host_detail.getHostDetails", "connection_error", err, "host_id", hostID)
		return nil, err
	}

	params := url.Values{}
	addHostPopulationParams(params)
	
	var response struct {
		Host HostDetail `json:"host"` // Use the new rich HostDetail struct
	}
	endpointPath := fmt.Sprintf("hosts/%d", hostID)
	_, err = client.Get(ctx, endpointPath, params, &response) 

	if err != nil {
		plugin.Logger(ctx).Error("fleetdm_host_detail.getHostDetails", "client_get_error", err, "host_id", hostID)
		return nil, err 
	}
	
	return response.Host, nil
}
