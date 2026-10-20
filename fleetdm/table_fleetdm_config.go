package fleetdm

import (
	"context"
	"encoding/json"

	"github.com/turbot/steampipe-plugin-sdk/v6/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin/transform"
)

// AppConfig is Fleet's global configuration, one JSON column per section.
// Fleet masks the SMTP password itself; webhook and integration URLs and the
// logging destinations are returned in full, so those sections are gated.
// Refer to: https://fleetdm.com/docs/rest-api/rest-api#get-configuration
type AppConfig struct {
	OrgInfo                json.RawMessage `json:"org_info"`
	ServerSettings         json.RawMessage `json:"server_settings"`
	SMTPSettings           json.RawMessage `json:"smtp_settings"`
	HostExpirySettings     json.RawMessage `json:"host_expiry_settings"`
	ActivityExpirySettings json.RawMessage `json:"activity_expiry_settings"`
	Features               json.RawMessage `json:"features"`
	MDM                    json.RawMessage `json:"mdm"`
	Scripts                json.RawMessage `json:"scripts"`
	FleetDesktop           json.RawMessage `json:"fleet_desktop"`
	GitOps                 json.RawMessage `json:"gitops"`
	ConditionalAccess      json.RawMessage `json:"conditional_access"`
	License                json.RawMessage `json:"license"`
	UpdateInterval         json.RawMessage `json:"update_interval"`
	Vulnerabilities        json.RawMessage `json:"vulnerabilities"`
	VulnerabilitySettings  json.RawMessage `json:"vulnerability_settings"`
	MaxSoftwarePackageSize *int64          `json:"max_software_package_size"`
	AgentOptions           json.RawMessage `json:"agent_options"`
	Integrations           json.RawMessage `json:"integrations"`     // gated
	WebhookSettings        json.RawMessage `json:"webhook_settings"` // gated
	Logging                json.RawMessage `json:"logging"`          // gated
}

func tableFleetdmConfig(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "fleetdm_config",
		Description: "Fleet's global configuration as a single row, one JSON column per settings section.",
		List: &plugin.ListConfig{
			Hydrate: listConfig,
		},
		Columns: []*plugin.Column{
			{Name: "org_info", Type: proto.ColumnType_JSON, Description: "Organization name, logos and contact URL."},
			{Name: "server_settings", Type: proto.ColumnType_JSON, Description: "Server URL and server-side options."},
			{Name: "smtp_settings", Type: proto.ColumnType_JSON, Transform: transform.FromField("SMTPSettings"), Description: "Email (SMTP) settings; Fleet masks the password."},
			{Name: "host_expiry_settings", Type: proto.ColumnType_JSON, Description: "Global host expiry settings."},
			{Name: "activity_expiry_settings", Type: proto.ColumnType_JSON, Description: "Activity log retention settings."},
			{Name: "features", Type: proto.ColumnType_JSON, Description: "Global feature toggles (host users, software inventory, ...)."},
			{Name: "mdm", Type: proto.ColumnType_JSON, Transform: transform.FromField("MDM"), Description: "Global MDM settings: OS updates, disk encryption, setup experience, Apple/Windows/Android state."},
			{Name: "scripts", Type: proto.ColumnType_JSON, Description: "Scripts configured globally via GitOps."},
			{Name: "fleet_desktop", Type: proto.ColumnType_JSON, Description: "Fleet Desktop settings."},
			{Name: "gitops", Type: proto.ColumnType_JSON, Transform: transform.FromField("GitOps"), Description: "GitOps mode settings."},
			{Name: "conditional_access", Type: proto.ColumnType_JSON, Description: "Conditional access (Microsoft Entra, Okta) settings."},
			{Name: "license", Type: proto.ColumnType_JSON, Description: "License tier, organization, device count and expiration."},
			{Name: "update_interval", Type: proto.ColumnType_JSON, Description: "Agent refresh intervals."},
			{Name: "vulnerabilities", Type: proto.ColumnType_JSON, Description: "Vulnerability processing settings."},
			{Name: "vulnerability_settings", Type: proto.ColumnType_JSON, Description: "Vulnerability database settings."},
			{Name: "max_software_package_size", Type: proto.ColumnType_INT, Description: "Largest software installer Fleet accepts, in bytes."},
			{Name: "agent_options", Type: proto.ColumnType_JSON, Description: "Global osquery and Orbit agent options."},
			{Name: "integrations", Type: proto.ColumnType_JSON, Hydrate: itemIfSecretsExposed, Transform: transform.FromField("Integrations"), Description: "Jira, Zendesk, Google Calendar and other integrations. NULL unless expose_secrets = true, as URLs can embed credentials."},
			{Name: "webhook_settings", Type: proto.ColumnType_JSON, Hydrate: itemIfSecretsExposed, Transform: transform.FromField("WebhookSettings"), Description: "Global webhook settings. NULL unless expose_secrets = true, as destination URLs can embed credentials."},
			{Name: "logging", Type: proto.ColumnType_JSON, Hydrate: itemIfSecretsExposed, Transform: transform.FromField("Logging"), Description: "Log destinations for results, status and audit logs. NULL unless expose_secrets = true, as webhook destinations can embed credentials."},
		},
	}
}

func listConfig(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := getClient(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("fleetdm_config.listConfig", "connection_error", err)
		return nil, err
	}
	var config AppConfig
	if err := client.Get(ctx, "config", nil, &config); err != nil {
		plugin.Logger(ctx).Error("fleetdm_config.listConfig", "api_error", err)
		return nil, err
	}
	d.StreamListItem(ctx, config)
	return nil, nil
}
