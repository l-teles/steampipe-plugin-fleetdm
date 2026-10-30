package fleetdm

import (
	"errors"
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"github.com/turbot/steampipe-plugin-sdk/v6/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin"
)

func stringQual(v string) *proto.QualValue {
	return &proto.QualValue{Value: &proto.QualValue_StringValue{StringValue: v}}
}

func TestSoftwareTypeFilter(t *testing.T) {
	cases := []struct {
		name  string
		quals plugin.KeyColumnEqualsQualMap
		want  url.Values
	}{
		{"none", plugin.KeyColumnEqualsQualMap{}, url.Values{}},
		{"source", plugin.KeyColumnEqualsQualMap{"source": stringQual("apps")}, url.Values{"source": {"apps"}}},
		{"both", plugin.KeyColumnEqualsQualMap{"source": stringQual("chrome_extensions"), "extension_for": stringQual("brave")},
			url.Values{"source": {"chrome_extensions"}, "extension_for": {"brave"}}},
		// Fleet rejects extension_for without source.
		{"extension_for alone", plugin.KeyColumnEqualsQualMap{"extension_for": stringQual("brave")}, url.Values{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := softwareTypeFilter(&plugin.QueryData{EqualsQuals: tc.quals})
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestWithSoftwareTypeFilterFallback(t *testing.T) {
	unprocessable := &FleetAPIError{StatusCode: http.StatusUnprocessableEntity, Status: "422 Unprocessable Entity"}
	badRequest := &FleetAPIError{StatusCode: http.StatusBadRequest, Status: "400 Bad Request"}
	forbidden := &FleetAPIError{StatusCode: http.StatusForbidden, Status: "403 Forbidden"}
	withSource := plugin.KeyColumnEqualsQualMap{"source": stringQual("not_a_source")}

	cases := []struct {
		name      string
		quals     plugin.KeyColumnEqualsQualMap
		errs      []error // returned by successive list calls
		wantCalls []url.Values
		wantErr   error
	}{
		{"filter accepted", withSource, []error{nil}, []url.Values{{"source": {"not_a_source"}}}, nil},
		{"filter rejected (422), retried without it", withSource, []error{unprocessable, nil},
			[]url.Values{{"source": {"not_a_source"}}, {}}, nil},
		{"filter rejected (400), retried without it", withSource, []error{badRequest, nil},
			[]url.Values{{"source": {"not_a_source"}}, {}}, nil},
		{"other errors are returned", withSource, []error{forbidden}, []url.Values{{"source": {"not_a_source"}}}, forbidden},
		{"no filter, no retry", plugin.KeyColumnEqualsQualMap{}, []error{unprocessable}, []url.Values{{}}, unprocessable},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var calls []url.Values
			err := withSoftwareTypeFilter(testCtx(t), &plugin.QueryData{EqualsQuals: tc.quals}, "test", func(filter url.Values) error {
				calls = append(calls, filter)
				return tc.errs[len(calls)-1]
			})
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("err = %v, want %v", err, tc.wantErr)
			}
			if !reflect.DeepEqual(calls, tc.wantCalls) {
				t.Fatalf("calls = %v, want %v", calls, tc.wantCalls)
			}
		})
	}
}
