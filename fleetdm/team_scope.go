package fleetdm

import (
	"context"
	"strconv"

	"github.com/turbot/steampipe-plugin-sdk/v6/plugin"
)

// teamScopes returns the team scopes a per-team endpoint must be called with:
// only the team in a team_id qual, otherwise "no team" (nil) plus every team.
// Teams need Fleet Premium; on Fleet Free only "no team" is returned.
func teamScopes(ctx context.Context, d *plugin.QueryData, client *FleetDMClient, table string) ([]*uint, error) {
	if q := d.EqualsQuals["team_id"]; q != nil {
		v := q.GetInt64Value()
		if v < 0 {
			return nil, nil
		}
		id := uint(v)
		return []*uint{&id}, nil
	}
	scopes := []*uint{nil}
	teams, err := fetchAllTeams(ctx, d, client)
	if err != nil {
		if handlePremiumError(ctx, d, table, err) {
			return scopes, nil
		}
		return nil, err
	}
	for _, t := range teams {
		id := t.ID
		scopes = append(scopes, &id)
	}
	return scopes, nil
}

// scopeParam renders a team scope for addFleetIDParam.
func scopeParam(id *uint) string {
	return strconv.FormatUint(uint64(*id), 10)
}
