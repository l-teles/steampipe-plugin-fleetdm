package fleetdm

import (
	"context"
	"net/url"
	"strconv"

	"github.com/turbot/steampipe-plugin-sdk/v6/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin/transform"
)

// Invite is a pending user invitation. The API never returns the invite token.
// Refer to: https://fleetdm.com/docs/rest-api/rest-api#list-invites
type Invite struct {
	ID         uint       `json:"id"`
	InvitedBy  uint       `json:"invited_by"`
	Email      string     `json:"email"`
	Name       string     `json:"name"`
	Position   string     `json:"position"`
	SSOEnabled bool       `json:"sso_enabled"`
	MFAEnabled bool       `json:"mfa_enabled"`
	GlobalRole *string    `json:"global_role"`
	Teams      []UserTeam `json:"teams"`
	Fleets     []UserTeam `json:"fleets"` // teams since Fleet 4.82
	CreatedAt  FleetTime  `json:"created_at"`
	UpdatedAt  FleetTime  `json:"updated_at"`
}

// ListInvitesResponse for `GET /api/v1/fleet/invites`. There is no meta
// object; paging ends on the first empty page.
type ListInvitesResponse struct {
	Invites []Invite `json:"invites"`
}

func tableFleetdmInvite(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "fleetdm_invite",
		Description: "Pending user invitations in FleetDM.",
		List: &plugin.ListConfig{
			Hydrate: listInvites,
		},
		Columns: []*plugin.Column{
			{Name: "id", Type: proto.ColumnType_INT, Description: "Unique ID of the invite."},
			{Name: "email", Type: proto.ColumnType_STRING, Description: "Email address the invite was sent to."},
			{Name: "name", Type: proto.ColumnType_STRING, Description: "Name of the invited user."},
			{Name: "position", Type: proto.ColumnType_STRING, Description: "Job position of the invited user."},
			{Name: "global_role", Type: proto.ColumnType_STRING, Description: "Global role the user will get. Null for team-only roles."},
			{Name: "teams", Type: proto.ColumnType_JSON, Transform: transform.FromField("Fleets", "Teams"), Description: "Teams (fleets) and roles the user will get."},
			{Name: "sso_enabled", Type: proto.ColumnType_BOOL, Transform: transform.FromField("SSOEnabled"), Description: "Whether the user will sign in with SSO."},
			{Name: "mfa_enabled", Type: proto.ColumnType_BOOL, Transform: transform.FromField("MFAEnabled"), Description: "Whether the user will need multi-factor authentication."},
			{Name: "invited_by", Type: proto.ColumnType_INT, Description: "ID of the user who sent the invite."},
			{Name: "created_at", Type: proto.ColumnType_TIMESTAMP, Transform: transform.FromField("CreatedAt").Transform(flexibleTimeTransform), Description: "When the invite was created."},
			{Name: "updated_at", Type: proto.ColumnType_TIMESTAMP, Transform: transform.FromField("UpdatedAt").Transform(flexibleTimeTransform), Description: "When the invite was last updated."},
		},
	}
}

func listInvites(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := getClient(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("fleetdm_invite.listInvites", "connection_error", err)
		return nil, err
	}
	err = paginatedList(ctx, d, 100, func(ctx context.Context, page, perPage int) ([]Invite, *ListMeta, error) {
		params := url.Values{}
		params.Add("page", strconv.Itoa(page))
		params.Add("per_page", strconv.Itoa(perPage))
		var response ListInvitesResponse
		if err := client.Get(ctx, "invites", params, &response); err != nil {
			plugin.Logger(ctx).Error("fleetdm_invite.listInvites", "api_error", err, "page", page)
			return nil, nil, err
		}
		return response.Invites, nil, nil
	})
	return nil, err
}
