package fleetdm

import (
	"errors"
	"net/http"
	"testing"

	"github.com/turbot/steampipe-plugin-sdk/v6/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin"
)

func TestSkipScope(t *testing.T) {
	t.Parallel()

	team := uint(3)
	forbidden := &FleetAPIError{StatusCode: http.StatusForbidden}
	notFound := &FleetAPIError{StatusCode: http.StatusNotFound}
	unfiltered := &plugin.QueryData{EqualsQuals: plugin.KeyColumnEqualsQualMap{}}
	filtered := &plugin.QueryData{EqualsQuals: plugin.KeyColumnEqualsQualMap{
		"team_id": &proto.QualValue{Value: &proto.QualValue_Int64Value{Int64Value: 3}},
	}}

	tests := []struct {
		name  string
		d     *plugin.QueryData
		scope *uint
		err   error
		want  bool
	}{
		{"no-team 403 while listing all teams", unfiltered, nil, forbidden, true},
		{"deleted team while listing all teams", unfiltered, &team, notFound, true},
		{"team 403 still fails", unfiltered, &team, forbidden, false},
		{"no-team 404 still fails", unfiltered, nil, notFound, false},
		{"explicit team_id surfaces 404", filtered, &team, notFound, false},
		{"5xx still fails", unfiltered, nil, &FleetAPIError{StatusCode: http.StatusInternalServerError}, false},
		{"non-API error still fails", unfiltered, nil, errors.New("boom"), false},
	}
	for _, tc := range tests {
		if got := skipScope(testCtx(t), tc.d, "t", tc.scope, tc.err); got != tc.want {
			t.Errorf("%s: skipScope = %v, want %v", tc.name, got, tc.want)
		}
	}
}
