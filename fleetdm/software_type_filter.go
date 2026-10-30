package fleetdm

import (
	"context"
	"net/url"

	"github.com/turbot/steampipe-plugin-sdk/v6/plugin"
)

// softwareTypeKeyColumns are the optional key columns behind softwareTypeFilter.
func softwareTypeKeyColumns() []*plugin.KeyColumn {
	return []*plugin.KeyColumn{
		{Name: "source", Require: plugin.Optional},
		{Name: "extension_for", Require: plugin.Optional},
	}
}

// softwareTypeFilter returns the source and extension_for parameters (Fleet
// 4.94+) for `=` quals on those columns. extension_for is only sent with
// source, as Fleet rejects it alone.
func softwareTypeFilter(d *plugin.QueryData) url.Values {
	params := url.Values{}
	source := d.EqualsQuals["source"]
	if source == nil {
		return params
	}
	params.Set("source", source.GetStringValue())
	if ext := d.EqualsQuals["extension_for"]; ext != nil {
		params.Set("extension_for", ext.GetStringValue())
	}
	return params
}

// withSoftwareTypeFilter runs list with the software type filter, and again
// without it when Fleet rejects the filter with a 400 (a source outside its
// allowlist). Fleet before 4.94 ignores the parameters. Either way Postgres
// still applies the WHERE clause to the rows returned.
func withSoftwareTypeFilter(ctx context.Context, d *plugin.QueryData, table string, list func(filter url.Values) error) error {
	filter := softwareTypeFilter(d)
	err := list(filter)
	if len(filter) == 0 || !is400(err) {
		return err
	}
	plugin.Logger(ctx).Warn(table, "software_type_filter_rejected", err)
	return list(url.Values{})
}
