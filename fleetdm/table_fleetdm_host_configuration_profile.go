package fleetdm

import (
	"context"
	"fmt"

	"github.com/turbot/steampipe-plugin-sdk/v6/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin/transform"
)

// HostConfigurationProfile is a configuration profile assigned to a host.
// Delivery status per host is in fleetdm_host_detail.mdm.
// Refer to: https://fleetdm.com/docs/rest-api/rest-api#get-host-os-settings-configuration-profiles
type HostConfigurationProfile struct {
	ProfileUUID string    `json:"profile_uuid"`
	Name        string    `json:"name"`
	Platform    string    `json:"platform"`
	Identifier  string    `json:"identifier"`
	Scope       string    `json:"scope"`
	Checksum    string    `json:"checksum"`
	TeamID      *uint     `json:"team_id"`
	FleetID     *uint     `json:"fleet_id"` // team_id since Fleet 4.82
	CreatedAt   FleetTime `json:"created_at"`
	UpdatedAt   FleetTime `json:"updated_at"`
	HostID      uint      `json:"-"`
	Hostname    string    `json:"-"`
}

// ListHostConfigurationProfilesResponse for
// `GET /api/v1/fleet/hosts/{id}/configuration_profiles` (not paginated).
type ListHostConfigurationProfilesResponse struct {
	Profiles []HostConfigurationProfile `json:"profiles"`
}

func tableFleetdmHostConfigurationProfile(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "fleetdm_host_configuration_profile",
		Description: "Configuration profiles assigned to each host (MDM). Without a host_id filter every host is queried.",
		List: &plugin.ListConfig{
			Hydrate: listHostConfigurationProfiles,
			KeyColumns: []*plugin.KeyColumn{
				{Name: "host_id", Require: plugin.Optional},
			},
		},
		Columns: []*plugin.Column{
			{Name: "host_id", Type: proto.ColumnType_INT, Description: "ID of the host."},
			{Name: "hostname", Type: proto.ColumnType_STRING, Description: "Hostname of the host. Empty when filtering by host_id."},
			{Name: "profile_uuid", Type: proto.ColumnType_STRING, Transform: transform.FromField("ProfileUUID"), Description: "Unique ID of the profile."},
			{Name: "name", Type: proto.ColumnType_STRING, Description: "Name of the profile."},
			{Name: "platform", Type: proto.ColumnType_STRING, Description: "Platform of the profile."},
			{Name: "identifier", Type: proto.ColumnType_STRING, Description: "Payload identifier (Apple profiles only)."},
			{Name: "scope", Type: proto.ColumnType_STRING, Description: "Payload scope, 'System' or 'User' (Apple profiles only)."},
			{Name: "checksum", Type: proto.ColumnType_STRING, Description: "Checksum of the profile contents (base64)."},
			{Name: "team_id", Type: proto.ColumnType_INT, Transform: transform.FromField("FleetID", "TeamID"), Description: "ID of the team (fleet) the profile belongs to."},
			{Name: "created_at", Type: proto.ColumnType_TIMESTAMP, Transform: transform.FromField("CreatedAt").Transform(flexibleTimeTransform), Description: "When the profile was created."},
			{Name: "updated_at", Type: proto.ColumnType_TIMESTAMP, Transform: transform.FromField("UpdatedAt").Transform(flexibleTimeTransform), Description: "When the profile was last uploaded."},
		},
	}
}

func listHostConfigurationProfiles(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := getClient(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("fleetdm_host_configuration_profile.listHostConfigurationProfiles", "connection_error", err)
		return nil, err
	}
	err = streamPerHost(ctx, d, client, func(ctx context.Context, host hostRef) ([]HostConfigurationProfile, error) {
		// The endpoint ignores paging.
		var response ListHostConfigurationProfilesResponse
		if err := client.Get(ctx, fmt.Sprintf("hosts/%d/configuration_profiles", host.ID), nil, &response); err != nil {
			return nil, err
		}
		for i := range response.Profiles {
			response.Profiles[i].HostID, response.Profiles[i].Hostname = host.ID, host.Hostname
		}
		return response.Profiles, nil
	})
	if err != nil {
		if handleUnavailableFeature(ctx, d, "fleetdm_host_configuration_profile", err) {
			return nil, nil
		}
		plugin.Logger(ctx).Error("fleetdm_host_configuration_profile.listHostConfigurationProfiles", "api_error", err)
	}
	return nil, err
}
