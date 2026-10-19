package fleetdm

import (
	"context"

	"github.com/turbot/steampipe-plugin-sdk/v6/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin/transform"
)

// EnrollSecret is a live credential that lets an agent enroll a host into
// Fleet (globally, or into a team).
// Refer to: https://fleetdm.com/docs/rest-api/rest-api#get-enroll-secrets
type EnrollSecret struct {
	Secret    string    `json:"secret"`
	CreatedAt FleetTime `json:"created_at"`
	TeamID    *uint     `json:"team_id"`
	FleetID   *uint     `json:"fleet_id"` // team_id since Fleet 4.82
}

// GetGlobalEnrollSecretsResponse for `GET /api/v1/fleet/spec/enroll_secret`.
type GetGlobalEnrollSecretsResponse struct {
	Spec struct {
		Secrets []EnrollSecret `json:"secrets"`
	} `json:"spec"`
}

// ListTeamEnrollSecretsResponse for `GET /api/v1/fleet/fleets/{id}/secrets`.
type ListTeamEnrollSecretsResponse struct {
	Secrets []EnrollSecret `json:"secrets"`
}

func tableFleetdmEnrollSecret(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "fleetdm_enroll_secret",
		Description: "Enroll secrets (global and per team). These are live credentials: the table only works with expose_secrets = true in the connection config.",
		List: &plugin.ListConfig{
			Hydrate: listEnrollSecrets,
			KeyColumns: []*plugin.KeyColumn{
				{Name: "team_id", Require: plugin.Optional}, // Only this team's secrets
			},
		},
		Columns: []*plugin.Column{
			{Name: "secret", Type: proto.ColumnType_STRING, Description: "The enroll secret."},
			{Name: "team_id", Type: proto.ColumnType_INT, Transform: transform.FromField("FleetID", "TeamID"), Description: "ID of the team (fleet) hosts enroll into. Null for global secrets."},
			{Name: "created_at", Type: proto.ColumnType_TIMESTAMP, Transform: transform.FromField("CreatedAt").Transform(flexibleTimeTransform), Description: "When the secret was created."},
		},
	}
}

func listEnrollSecrets(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	if err := requireExposeSecrets(d); err != nil {
		return nil, err
	}
	client, err := getClient(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("fleetdm_enroll_secret.listEnrollSecrets", "connection_error", err)
		return nil, err
	}
	scopes, err := teamScopes(ctx, d, client, "fleetdm_enroll_secret")
	if err != nil {
		return nil, err
	}
	for _, scope := range scopes {
		var secrets []EnrollSecret
		if scope == nil {
			var response GetGlobalEnrollSecretsResponse
			if err := client.Get(ctx, "spec/enroll_secret", nil, &response); err != nil {
				plugin.Logger(ctx).Error("fleetdm_enroll_secret.listEnrollSecrets", "api_error", err)
				return nil, err
			}
			secrets = response.Spec.Secrets
		} else {
			var response ListTeamEnrollSecretsResponse
			if err := client.GetCompat(ctx, d, famFleetSecrets, []any{*scope}, nil, &response); err != nil {
				plugin.Logger(ctx).Error("fleetdm_enroll_secret.listEnrollSecrets", "api_error", err, "team_id", *scope)
				return nil, err
			}
			secrets = response.Secrets
		}
		for _, s := range secrets {
			d.StreamListItem(ctx, s)
			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}
	}
	return nil, nil
}
