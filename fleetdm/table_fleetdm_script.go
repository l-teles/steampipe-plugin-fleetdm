package fleetdm

import (
	"context"
	"net/url"
	"strconv"

	"github.com/turbot/steampipe-plugin-sdk/v6/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin/transform"
)

// Script is a saved script that Fleet can run on hosts. Contents are not part
// of the list response.
// Refer to: https://fleetdm.com/docs/rest-api/rest-api#list-scripts
type Script struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	TeamID    *uint     `json:"team_id"`
	FleetID   *uint     `json:"fleet_id"` // team_id since Fleet 4.82
	CreatedAt FleetTime `json:"created_at"`
	UpdatedAt FleetTime `json:"updated_at"`
}

// ListScriptsResponse for `GET /api/v1/fleet/scripts`.
type ListScriptsResponse struct {
	Scripts []Script  `json:"scripts"`
	Meta    *ListMeta `json:"meta"`
}

func tableFleetdmScript(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "fleetdm_script",
		Description: "Saved scripts in FleetDM, across all teams.",
		List: &plugin.ListConfig{
			Hydrate: listScripts,
			KeyColumns: []*plugin.KeyColumn{
				{Name: "team_id", Require: plugin.Optional}, // Only this team's scripts
			},
		},
		Columns: []*plugin.Column{
			{Name: "id", Type: proto.ColumnType_INT, Description: "Unique ID of the script."},
			{Name: "name", Type: proto.ColumnType_STRING, Description: "File name of the script."},
			{Name: "team_id", Type: proto.ColumnType_INT, Transform: transform.FromField("FleetID", "TeamID"), Description: "ID of the team (fleet) the script belongs to. Null for scripts with no team."},
			{Name: "created_at", Type: proto.ColumnType_TIMESTAMP, Transform: transform.FromField("CreatedAt").Transform(flexibleTimeTransform), Description: "When the script was uploaded."},
			{Name: "updated_at", Type: proto.ColumnType_TIMESTAMP, Transform: transform.FromField("UpdatedAt").Transform(flexibleTimeTransform), Description: "When the script was last modified."},
		},
	}
}

func listScripts(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := getClient(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("fleetdm_script.listScripts", "connection_error", err)
		return nil, err
	}

	scopes, err := teamScopes(ctx, d, client, "fleetdm_script")
	if err != nil {
		return nil, err
	}
	for _, scope := range scopes {
		err := paginatedList(ctx, d, 100, func(ctx context.Context, page, perPage int) ([]Script, *ListMeta, error) {
			params := url.Values{}
			params.Add("page", strconv.Itoa(page))
			params.Add("per_page", strconv.Itoa(perPage))
			if scope != nil {
				client.addFleetIDParam(ctx, d, params, scopeParam(scope))
			}
			var response ListScriptsResponse
			if err := client.Get(ctx, "scripts", params, &response); err != nil {
				plugin.Logger(ctx).Error("fleetdm_script.listScripts", "api_error", err, "page", page)
				return nil, nil, err
			}
			return response.Scripts, response.Meta, nil
		})
		if err != nil {
			if skipScope(ctx, d, "fleetdm_script", scope, err) {
				continue
			}
			return nil, err
		}
		if d.RowsRemaining(ctx) == 0 {
			break
		}
	}
	return nil, nil
}
