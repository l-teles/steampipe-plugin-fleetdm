package fleetdm

import (
	"context"
	"encoding/json"

	"github.com/turbot/steampipe-plugin-sdk/v6/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin/transform"
)

// VPPToken is an Apple Volume Purchasing Program token. The token itself is
// never returned.
// Refer to: https://fleetdm.com/docs/rest-api/rest-api#list-volume-purchasing-program-vpp-tokens
type VPPToken struct {
	ID          uint            `json:"id"`
	OrgName     string          `json:"org_name"`
	Location    string          `json:"location"`
	CountryCode string          `json:"country_code"`
	RenewDate   FleetTime       `json:"renew_date"`
	Teams       json.RawMessage `json:"teams"`
	Fleets      json.RawMessage `json:"fleets"` // teams since Fleet 4.82
}

// ListVPPTokensResponse for `GET /api/v1/fleet/vpp_tokens` (not paginated).
type ListVPPTokensResponse struct {
	VPPTokens []VPPToken `json:"vpp_tokens"`
}

func tableFleetdmVPPToken(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "fleetdm_vpp_token",
		Description: "Apple Volume Purchasing Program (VPP) tokens in FleetDM (Fleet Premium). Token values are never returned.",
		List: &plugin.ListConfig{
			Hydrate: listVPPTokens,
		},
		Columns: []*plugin.Column{
			{Name: "id", Type: proto.ColumnType_INT, Description: "Unique ID of the VPP token."},
			{Name: "org_name", Type: proto.ColumnType_STRING, Description: "Apple Business organization the token belongs to."},
			{Name: "location", Type: proto.ColumnType_STRING, Description: "Apple Business location of the token."},
			{Name: "country_code", Type: proto.ColumnType_STRING, Description: "Country code of the App Store the token buys from."},
			{Name: "renew_date", Type: proto.ColumnType_TIMESTAMP, Transform: transform.FromField("RenewDate").Transform(flexibleTimeTransform), Description: "When the token expires and must be renewed."},
			{Name: "teams", Type: proto.ColumnType_JSON, Transform: transform.FromField("Fleets", "Teams"), Description: "Teams (fleets) the token is assigned to."},
		},
	}
}

func listVPPTokens(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := getClient(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("fleetdm_vpp_token.listVPPTokens", "connection_error", err)
		return nil, err
	}
	var response ListVPPTokensResponse
	if err := client.Get(ctx, "vpp_tokens", nil, &response); err != nil {
		if handleUnavailableFeature(ctx, d, "fleetdm_vpp_token", err) {
			return nil, nil
		}
		plugin.Logger(ctx).Error("fleetdm_vpp_token.listVPPTokens", "api_error", err)
		return nil, err
	}
	for _, t := range response.VPPTokens {
		d.StreamListItem(ctx, t)
		if d.RowsRemaining(ctx) == 0 {
			break
		}
	}
	return nil, nil
}
