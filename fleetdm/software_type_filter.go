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

// softwareTypeMatcher reports whether a row matches the source and
// extension_for quals, or is nil without them. Fleet before 4.94 ignores the
// filter and the fallback drops it, so rows are matched here too: rows Postgres
// would discard must not count towards a pushed-down LIMIT. A nil
// extensionFor is NULL, which matches nothing.
func softwareTypeMatcher(d *plugin.QueryData) func(source string, extensionFor *string) bool {
	source, ext := d.EqualsQuals["source"], d.EqualsQuals["extension_for"]
	if source == nil && ext == nil {
		return nil
	}
	return func(s string, e *string) bool {
		if source != nil && s != source.GetStringValue() {
			return false
		}
		return ext == nil || (e != nil && *e == ext.GetStringValue())
	}
}

// withSoftwareTypeFilter runs list with the software type filter, and again
// without it when Fleet rejects the filter (a source outside its allowlist).
// Callers still match rows with softwareTypeMatcher.
func withSoftwareTypeFilter(ctx context.Context, d *plugin.QueryData, table string, list func(filter url.Values) error) error {
	filter := softwareTypeFilter(d)
	err := list(filter)
	if len(filter) == 0 || !isInvalidArgument(err) {
		return err
	}
	plugin.Logger(ctx).Warn(table, "software_type_filter_rejected", err)
	return list(url.Values{})
}
