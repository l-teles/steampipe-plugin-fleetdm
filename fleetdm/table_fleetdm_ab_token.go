package fleetdm

import (
	"context"
	"encoding/json"

	"github.com/turbot/steampipe-plugin-sdk/v6/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin/transform"
)

// ABToken is an Apple Business (formerly Apple Business Manager) token used
// for automatic enrollment. The token itself is never returned.
// Refer to: https://fleetdm.com/docs/rest-api/rest-api#list-apple-business-ab-tokens
type ABToken struct {
	ID            uint            `json:"id"`
	AppleID       string          `json:"apple_id"`
	OrgName       string          `json:"org_name"`
	MDMServerUUID string          `json:"mdm_server_uuid"`
	MDMServerURL  string          `json:"mdm_server_url"`
	RenewDate     FleetTime       `json:"renew_date"`
	TermsExpired  bool            `json:"terms_expired"`
	TokenInvalid  bool            `json:"token_invalid"`
	Default       bool            `json:"default"`
	MacOSTeam     json.RawMessage `json:"macos_team"`
	MacOSFleet    json.RawMessage `json:"macos_fleet"` // macos_team since Fleet 4.82
	IOSTeam       json.RawMessage `json:"ios_team"`
	IOSFleet      json.RawMessage `json:"ios_fleet"`
	IPadOSTeam    json.RawMessage `json:"ipados_team"`
	IPadOSFleet   json.RawMessage `json:"ipados_fleet"`
	BYODTeam      json.RawMessage `json:"byod_team"`
	BYODFleet     json.RawMessage `json:"byod_fleet"`
}

// ListABTokensResponse for `GET /api/v1/fleet/ab_tokens` (formerly
// /abm_tokens; not paginated). Both envelope keys are declared.
type ListABTokensResponse struct {
	ABTokens  []ABToken `json:"ab_tokens"`
	ABMTokens []ABToken `json:"abm_tokens"`
}

func tableFleetdmABToken(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "fleetdm_ab_token",
		Description: "Apple Business (Manager) tokens used for automatic device enrollment (Fleet Premium). Token values are never returned.",
		List: &plugin.ListConfig{
			Hydrate: listABTokens,
		},
		Columns: []*plugin.Column{
			{Name: "id", Type: proto.ColumnType_INT, Description: "Unique ID of the token."},
			{Name: "org_name", Type: proto.ColumnType_STRING, Description: "Apple Business organization the token belongs to."},
			{Name: "apple_id", Type: proto.ColumnType_STRING, Description: "Apple ID of the account that created the token."},
			{Name: "renew_date", Type: proto.ColumnType_TIMESTAMP, Transform: transform.FromField("RenewDate").Transform(flexibleTimeTransform), Description: "When the token expires and must be renewed."},
			{Name: "terms_expired", Type: proto.ColumnType_BOOL, Description: "Whether Apple's terms must be accepted again before the token works."},
			{Name: "token_invalid", Type: proto.ColumnType_BOOL, Description: "Whether Apple rejects the token."},
			{Name: "default", Type: proto.ColumnType_BOOL, Description: "Whether this is the default token."},
			{Name: "mdm_server_uuid", Type: proto.ColumnType_STRING, Transform: transform.FromField("MDMServerUUID"), Description: "UUID of the MDM server in Apple Business."},
			{Name: "mdm_server_url", Type: proto.ColumnType_STRING, Transform: transform.FromField("MDMServerURL"), Description: "MDM server URL devices enroll to."},
			{Name: "macos_team", Type: proto.ColumnType_JSON, Transform: transform.FromField("MacOSFleet", "MacOSTeam"), Description: "Team (fleet) new macOS hosts are assigned to."},
			{Name: "ios_team", Type: proto.ColumnType_JSON, Transform: transform.FromField("IOSFleet", "IOSTeam"), Description: "Team (fleet) new iOS hosts are assigned to."},
			{Name: "ipados_team", Type: proto.ColumnType_JSON, Transform: transform.FromField("IPadOSFleet", "IPadOSTeam"), Description: "Team (fleet) new iPadOS hosts are assigned to."},
			{Name: "byod_team", Type: proto.ColumnType_JSON, Transform: transform.FromField("BYODFleet", "BYODTeam"), Description: "Team (fleet) new personally owned hosts are assigned to."},
		},
	}
}

func listABTokens(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := getClient(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("fleetdm_ab_token.listABTokens", "connection_error", err)
		return nil, err
	}
	var response ListABTokensResponse
	if err := client.GetCompat(ctx, d, famABTokens, nil, nil, &response); err != nil {
		if handleUnavailableFeature(ctx, d, "fleetdm_ab_token", err) {
			return nil, nil
		}
		plugin.Logger(ctx).Error("fleetdm_ab_token.listABTokens", "api_error", err)
		return nil, err
	}
	for _, t := range coalesceSlice(response.ABTokens, response.ABMTokens) {
		d.StreamListItem(ctx, t)
		if d.RowsRemaining(ctx) == 0 {
			break
		}
	}
	return nil, nil
}
