package fleetdm

import (
	"context"
	"errors"
	"net/http"
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
		if handleUnavailableFeature(ctx, d, table, err) {
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

// scopeParam renders a team scope for addFleetIDParam; nil ("no team") is 0,
// which per-team endpoints accept for hosts with no team.
func scopeParam(id *uint) string {
	if id == nil {
		return "0"
	}
	return strconv.FormatUint(uint64(*id), 10)
}

// skipScope reports whether a per-team request error only means this scope is
// out of reach for an implicit all-teams listing: a team-scoped token gets 403
// on "no team", and a team deleted mid-listing gets 404. With an explicit
// team_id filter every error surfaces.
func skipScope(ctx context.Context, d *plugin.QueryData, table string, scope *uint, err error) bool {
	if d.EqualsQuals["team_id"] != nil {
		return false
	}
	var apiErr *FleetAPIError
	if !errors.As(err, &apiErr) {
		return false
	}
	if (scope == nil && apiErr.StatusCode == http.StatusForbidden) || (scope != nil && apiErr.StatusCode == http.StatusNotFound) {
		plugin.Logger(ctx).Warn("skipScope", "table", table, "no_team", scope == nil, "status", apiErr.StatusCode)
		return true
	}
	return false
}
