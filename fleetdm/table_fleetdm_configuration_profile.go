package fleetdm

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"

	"github.com/turbot/steampipe-plugin-sdk/v6/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin/transform"
)

// ConfigurationProfile is an MDM configuration profile (Apple, Windows or
// Android). Profile contents are not part of the list response.
// Refer to: https://fleetdm.com/docs/rest-api/rest-api#list-custom-os-settings-configuration-profiles
type ConfigurationProfile struct {
	ProfileUUID      string          `json:"profile_uuid"`
	Name             string          `json:"name"`
	Platform         string          `json:"platform"`
	Identifier       string          `json:"identifier"` // Apple only
	Scope            string          `json:"scope"`      // Apple only: "System" or "User"
	Checksum         string          `json:"checksum"`
	TeamID           *uint           `json:"team_id"`
	FleetID          *uint           `json:"fleet_id"` // team_id since Fleet 4.82
	CreatedAt        FleetTime       `json:"created_at"`
	UpdatedAt        FleetTime       `json:"updated_at"` // When the profile was last uploaded
	LabelsIncludeAll json.RawMessage `json:"labels_include_all"`
	LabelsIncludeAny json.RawMessage `json:"labels_include_any"`
	LabelsExcludeAny json.RawMessage `json:"labels_exclude_any"`
}

// ListConfigurationProfilesResponse for `GET /api/v1/fleet/configuration_profiles`.
type ListConfigurationProfilesResponse struct {
	Profiles []ConfigurationProfile `json:"profiles"`
	Meta     *ListMeta              `json:"meta"`
}

func tableFleetdmConfigurationProfile(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "fleetdm_configuration_profile",
		Description: "MDM configuration profiles (OS settings) in FleetDM, across all teams.",
		List: &plugin.ListConfig{
			Hydrate: listConfigurationProfiles,
			KeyColumns: []*plugin.KeyColumn{
				{Name: "team_id", Require: plugin.Optional}, // Only this team's profiles
			},
		},
		Columns: []*plugin.Column{
			{Name: "profile_uuid", Type: proto.ColumnType_STRING, Description: "Unique ID of the profile."},
			{Name: "name", Type: proto.ColumnType_STRING, Description: "Name of the profile."},
			{Name: "platform", Type: proto.ColumnType_STRING, Description: "Platform the profile applies to: 'darwin', 'windows' or 'android'."},
			{Name: "identifier", Type: proto.ColumnType_STRING, Description: "Payload identifier (Apple profiles only)."},
			{Name: "scope", Type: proto.ColumnType_STRING, Description: "Payload scope, 'System' or 'User' (Apple profiles only)."},
			{Name: "checksum", Type: proto.ColumnType_STRING, Description: "Checksum of the profile contents (base64)."},
			{Name: "team_id", Type: proto.ColumnType_INT, Transform: transform.FromField("FleetID", "TeamID"), Description: "ID of the team (fleet) the profile belongs to. Null for profiles with no team."},
			{Name: "labels_include_all", Type: proto.ColumnType_JSON, Description: "Profile only applies to hosts with all of these labels."},
			{Name: "labels_include_any", Type: proto.ColumnType_JSON, Description: "Profile only applies to hosts with any of these labels."},
			{Name: "labels_exclude_any", Type: proto.ColumnType_JSON, Description: "Profile does not apply to hosts with any of these labels."},
			{Name: "created_at", Type: proto.ColumnType_TIMESTAMP, Transform: transform.FromField("CreatedAt").Transform(flexibleTimeTransform), Description: "When the profile was created."},
			{Name: "updated_at", Type: proto.ColumnType_TIMESTAMP, Transform: transform.FromField("UpdatedAt").Transform(flexibleTimeTransform), Description: "When the profile was last uploaded."},
		},
	}
}

func listConfigurationProfiles(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := getClient(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("fleetdm_configuration_profile.listConfigurationProfiles", "connection_error", err)
		return nil, err
	}

	scopes, err := teamScopes(ctx, d, client, "fleetdm_configuration_profile")
	if err != nil {
		return nil, err
	}
	for _, scope := range scopes {
		err := paginatedList(ctx, d, 100, func(ctx context.Context, page, perPage int) ([]ConfigurationProfile, *ListMeta, error) {
			params := url.Values{}
			params.Add("page", strconv.Itoa(page))
			params.Add("per_page", strconv.Itoa(perPage))
			if scope != nil {
				client.addFleetIDParam(ctx, d, params, scopeParam(scope))
			}
			var response ListConfigurationProfilesResponse
			if err := client.Get(ctx, "configuration_profiles", params, &response); err != nil {
				plugin.Logger(ctx).Error("fleetdm_configuration_profile.listConfigurationProfiles", "api_error", err, "page", page)
				return nil, nil, err
			}
			return response.Profiles, response.Meta, nil
		})
		if err != nil {
			if handleUnavailableFeature(ctx, d, "fleetdm_configuration_profile", err) {
				return nil, nil
			}
			return nil, err
		}
		if d.RowsRemaining(ctx) == 0 {
			break
		}
	}
	return nil, nil
}
