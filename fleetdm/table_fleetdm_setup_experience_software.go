package fleetdm

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/turbot/steampipe-plugin-sdk/v6/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin/transform"
)

// SetupExperienceSoftware is a software title that can be installed during
// a team's setup experience (new device enrollment). TeamID is the scope it
// was listed under.
// Refer to: https://fleetdm.com/docs/rest-api/rest-api#list-software-available-during-setup-experience
type SetupExperienceSoftware struct {
	ID              uint        `json:"id"`
	Name            string      `json:"name"`
	DisplayName     string      `json:"display_name"`
	Source          string      `json:"source"`
	SoftwarePackage interface{} `json:"software_package"` // install_during_setup lives here
	AppStoreApp     interface{} `json:"app_store_app"`
	TeamID          *uint       `json:"-"`
	Platform        string      `json:"-"`
}

// setupExperiencePlatforms are the platforms the endpoint accepts; without a
// platform parameter Fleet lists macOS only.
var setupExperiencePlatforms = []string{"macos", "ios", "ipados", "windows", "linux", "android"}

// ListSetupExperienceSoftwareResponse for `GET /api/v1/fleet/setup_experience/software`.
type ListSetupExperienceSoftwareResponse struct {
	SoftwareTitles []SetupExperienceSoftware `json:"software_titles"`
	Meta           *ListMeta                 `json:"meta"`
}

func tableFleetdmSetupExperienceSoftware(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "fleetdm_setup_experience_software",
		Description: "Software available to install during each team's setup experience (Fleet Premium).",
		List: &plugin.ListConfig{
			Hydrate: listSetupExperienceSoftware,
			KeyColumns: []*plugin.KeyColumn{
				{Name: "team_id", Require: plugin.Optional},  // Only this team; 0 for no team
				{Name: "platform", Require: plugin.Optional}, // Only this platform
			},
		},
		Columns: []*plugin.Column{
			{Name: "id", Type: proto.ColumnType_INT, Description: "Software title ID."},
			{Name: "name", Type: proto.ColumnType_STRING, Description: "Name of the software."},
			{Name: "display_name", Type: proto.ColumnType_STRING, Description: "Custom display name of the software, if set."},
			{Name: "source", Type: proto.ColumnType_STRING, Description: "Source of the software (e.g. 'apps', 'programs', 'deb_packages')."},
			{Name: "team_id", Type: proto.ColumnType_INT, Description: "ID of the team (fleet); 0 for hosts with no team."},
			{Name: "platform", Type: proto.ColumnType_STRING, Description: "Setup experience platform: 'macos', 'ios', 'ipados', 'windows', 'linux' or 'android'. A title can appear once per platform."},
			{Name: "install_during_setup", Type: proto.ColumnType_BOOL, Transform: transform.From(installDuringSetup), Description: "Whether the software is installed during setup."},
			{Name: "software_package", Type: proto.ColumnType_JSON, Description: "Installer package details, if the title is a custom package. Installer URLs are null unless expose_secrets = true, as they can embed credentials."},
			{Name: "app_store_app", Type: proto.ColumnType_JSON, Description: "App Store app details, if the title is a VPP app."},
		},
	}
}

// installDuringSetup reads install_during_setup from whichever installer the
// title has (package or App Store app).
func installDuringSetup(_ context.Context, d *transform.TransformData) (interface{}, error) {
	s, ok := d.HydrateItem.(SetupExperienceSoftware)
	if !ok {
		return nil, nil
	}
	for _, inst := range []interface{}{s.SoftwarePackage, s.AppStoreApp} {
		if m, ok := inst.(map[string]interface{}); ok {
			if v, ok := m["install_during_setup"].(bool); ok {
				return v, nil
			}
		}
	}
	return nil, nil
}

func listSetupExperienceSoftware(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := getClient(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("fleetdm_setup_experience_software.listSetupExperienceSoftware", "connection_error", err)
		return nil, err
	}
	scopes, err := teamScopes(ctx, d, client, "fleetdm_setup_experience_software")
	if err != nil {
		return nil, err
	}
	platforms := setupExperiencePlatforms
	if q := d.EqualsQuals["platform"]; q != nil {
		platforms = []string{q.GetStringValue()}
	}
	exposed := secretsExposed(d)
scopeLoop:
	for _, scope := range scopes {
		teamID := uint(0)
		if scope != nil {
			teamID = *scope
		}
		for _, platform := range platforms {
			err := paginatedList(ctx, d, 100, func(ctx context.Context, page, perPage int) ([]SetupExperienceSoftware, *ListMeta, error) {
				params := url.Values{}
				params.Add("page", strconv.Itoa(page))
				params.Add("per_page", strconv.Itoa(perPage))
				params.Add("platform", platform)
				// The endpoint requires a team; 0 means no team.
				client.addFleetIDParam(ctx, d, params, scopeParam(scope))
				var response ListSetupExperienceSoftwareResponse
				if err := client.Get(ctx, "setup_experience/software", params, &response); err != nil {
					plugin.Logger(ctx).Error("fleetdm_setup_experience_software.listSetupExperienceSoftware", "api_error", err, "page", page, "platform", platform)
					return nil, nil, err
				}
				for i := range response.SoftwareTitles {
					response.SoftwareTitles[i].TeamID = &teamID
					response.SoftwareTitles[i].Platform = platform
					if !exposed {
						response.SoftwareTitles[i].SoftwarePackage = nullURLKeys(response.SoftwareTitles[i].SoftwarePackage, installerURLKeys...)
					}
				}
				return response.SoftwareTitles, response.Meta, nil
			})
			if err != nil {
				if handleUnavailableFeature(ctx, d, "fleetdm_setup_experience_software", err) {
					return nil, nil
				}
				// A server that does not support a platform rejects it; skip it.
				if isUnsupportedPlatformError(err) {
					plugin.Logger(ctx).Warn("fleetdm_setup_experience_software.listSetupExperienceSoftware", "skipping_platform", platform, "api_error", err)
					continue
				}
				if skipScope(ctx, d, "fleetdm_setup_experience_software", scope, err) {
					continue scopeLoop
				}
				return nil, err
			}
			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}
	}
	return nil, nil
}

// isUnsupportedPlatformError reports a 400 rejecting the platform parameter.
func isUnsupportedPlatformError(err error) bool {
	var apiErr *FleetAPIError
	return errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusBadRequest &&
		strings.Contains(strings.ToLower(apiErr.Snippet), "platform")
}
