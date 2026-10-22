package fleetdm

import (
	"context"
	"fmt"
	"net/url"

	"github.com/turbot/steampipe-plugin-sdk/v6/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin/transform"
)

// HostConfigurationProfile is the delivery state of one configuration profile
// on one host, from the host's MDM details.
// Refer to: https://fleetdm.com/docs/rest-api/rest-api#get-host
type HostConfigurationProfile struct {
	ProfileUUID         string  `json:"profile_uuid"`
	Name                string  `json:"name"`
	Platform            string  `json:"platform"`
	Status              *string `json:"status"`
	OperationType       string  `json:"operation_type"`
	Detail              string  `json:"detail"`
	Scope               *string `json:"scope"`
	ManagedLocalAccount *string `json:"managed_local_account"`
	HostID              uint    `json:"-"`
	Hostname            string  `json:"-"`
}

// hostProfilesResponse is the part of `GET /api/v1/fleet/hosts/{id}` this
// table reads. GET /hosts/{id}/configuration_profiles is not used: it returns
// every Apple profile of the host's team, whatever the host's platform or
// labels.
type hostProfilesResponse struct {
	Host struct {
		Hostname string `json:"hostname"`
		MDM      struct {
			Profiles []HostConfigurationProfile `json:"profiles"`
		} `json:"mdm"`
	} `json:"host"`
}

func tableFleetdmHostConfigurationProfile(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "fleetdm_host_configuration_profile",
		Description: "Configuration profiles delivered to each MDM-enrolled host, with their status. Without a host_id filter every host is queried.",
		List: &plugin.ListConfig{
			Hydrate: listHostConfigurationProfiles,
			KeyColumns: []*plugin.KeyColumn{
				{Name: "host_id", Require: plugin.Optional},
			},
		},
		Columns: []*plugin.Column{
			{Name: "host_id", Type: proto.ColumnType_INT, Description: "ID of the host."},
			{Name: "hostname", Type: proto.ColumnType_STRING, Description: "Hostname of the host."},
			{Name: "profile_uuid", Type: proto.ColumnType_STRING, Transform: transform.FromField("ProfileUUID"), Description: "Unique ID of the profile."},
			{Name: "name", Type: proto.ColumnType_STRING, Description: "Name of the profile."},
			{Name: "platform", Type: proto.ColumnType_STRING, Description: "Platform of the host the profile targets."},
			{Name: "status", Type: proto.ColumnType_STRING, Description: "Delivery status, e.g. 'verified', 'verifying', 'pending' or 'failed'."},
			{Name: "operation_type", Type: proto.ColumnType_STRING, Description: "Whether the profile is being installed ('install') or removed ('remove')."},
			{Name: "detail", Type: proto.ColumnType_STRING, Description: "Details from the device, such as the error for a failed profile."},
			{Name: "scope", Type: proto.ColumnType_STRING, Description: "Payload scope, 'System' or 'User' (Apple hosts only)."},
			{Name: "managed_local_account", Type: proto.ColumnType_STRING, Description: "Managed local account the profile applies to, when Fleet reports one."},
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
		params := url.Values{}
		params.Set("exclude_software", "true")
		var response hostProfilesResponse
		if err := client.Get(ctx, fmt.Sprintf("hosts/%d", host.ID), params, &response); err != nil {
			return nil, err
		}
		profiles := response.Host.MDM.Profiles
		for i := range profiles {
			profiles[i].HostID, profiles[i].Hostname = host.ID, response.Host.Hostname
		}
		return profiles, nil
	})
	if err != nil {
		if handleUnavailableFeature(ctx, d, "fleetdm_host_configuration_profile", err) {
			return nil, nil
		}
		plugin.Logger(ctx).Error("fleetdm_host_configuration_profile.listHostConfigurationProfiles", "api_error", err)
	}
	return nil, err
}
