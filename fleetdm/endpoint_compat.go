package fleetdm

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
)

// endpointFamily identifies a group of FleetDM API endpoints that were renamed
// upstream (teams -> fleets, queries -> reports). Old servers (<= v4.x) only
// serve the old paths; new servers may drop the old paths entirely, so the
// plugin resolves which generation a server speaks at query time and caches
// the answer per connection.
type endpointFamily string

const (
	famFleets        endpointFamily = "fleets"         // fleets <- teams
	famReports       endpointFamily = "reports"        // reports <- queries
	famFleetPolicies endpointFamily = "fleet_policies" // fleets/%d/policies <- teams/%d/policies
)

// endpointPaths maps a family to {newPathFormat, oldPathFormat}. The formats
// may contain fmt verbs filled from GetCompat's pathArgs.
var endpointPaths = map[endpointFamily][2]string{
	famFleets:        {"fleets", "teams"},
	famReports:       {"reports", "queries"},
	famFleetPolicies: {"fleets/%d/policies", "teams/%d/policies"},
}

// is404 reports whether err is a *FleetAPIError with a 404 status.
func is404(err error) bool {
	var apiErr *FleetAPIError
	return errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound
}

// resolveCompat is the pure decision core behind GetCompat, extracted so it
// can be unit-tested without SDK plumbing.
//
// cached is the previously resolved generation for the family (nil if not yet
// resolved). When cached, exactly one request is made against that generation
// and its result is returned verbatim — a genuine 404 (e.g. a deleted id) must
// surface/ignore normally, never flip the generation.
//
// When not cached, the NEW path is tried first. Resolution only happens on a
// definitive signal:
//   - new path succeeds            -> useNew=true
//   - new path 404, old succeeds   -> useNew=false
//   - new path 404, old also errs  -> old error returned, nothing resolved
//     (could be a genuine 404 on both generations)
//   - new path non-404 error       -> error returned, nothing resolved
//     (401/403/5xx say nothing about the generation)
//
// The returned useNew is non-nil only when this call resolved the generation.
func resolveCompat(tryNew, tryOld func() error, cached *bool) (useNew *bool, err error) {
	if cached != nil {
		if *cached {
			return nil, tryNew()
		}
		return nil, tryOld()
	}

	if err := tryNew(); err != nil {
		if !is404(err) {
			return nil, err
		}
		if oldErr := tryOld(); oldErr != nil {
			return nil, oldErr
		}
		resolved := false
		return &resolved, nil
	}
	resolved := true
	return &resolved, nil
}

// GetCompat performs a GET against a renamed endpoint family, transparently
// falling back from the new path to the old one on 404 and caching the
// resolved generation in the connection cache (which is already scoped
// per-connection, so the key needs no connection name).
func (c *FleetDMClient) GetCompat(ctx context.Context, d *plugin.QueryData, fam endpointFamily, pathArgs []any, params url.Values, target any) error {
	paths, ok := endpointPaths[fam]
	if !ok {
		return fmt.Errorf("endpoint_compat: unknown endpoint family %q", fam)
	}
	newPath, oldPath := paths[0], paths[1]
	if len(pathArgs) > 0 {
		newPath = fmt.Sprintf(newPath, pathArgs...)
		oldPath = fmt.Sprintf(oldPath, pathArgs...)
	}

	cacheKey := "endpoint_family/" + string(fam)
	var cached *bool
	if v, ok := d.ConnectionCache.Get(ctx, cacheKey); ok {
		if b, ok := v.(bool); ok {
			cached = &b
		}
	}

	useNew, err := resolveCompat(
		func() error { return c.Get(ctx, newPath, params, target) },
		func() error { return c.Get(ctx, oldPath, params, target) },
		cached,
	)
	if useNew != nil {
		plugin.Logger(ctx).Info("endpoint_compat", "family", fam, "use_new", *useNew)
		if cerr := d.ConnectionCache.Set(ctx, cacheKey, *useNew); cerr != nil {
			plugin.Logger(ctx).Warn("endpoint_compat", "connection_cache_set_error", cerr, "family", fam)
		}
	}
	return err
}

// coalesceSlice returns newer when it has elements, otherwise older. Used to
// read response envelopes that renamed their key (e.g. "fleets" vs "teams")
// without caring which generation answered.
func coalesceSlice[T any](newer, older []T) []T {
	if len(newer) > 0 {
		return newer
	}
	return older
}

// addFleetIDParam sets BOTH the new "fleet_id" and the old "team_id" query
// parameters to id. Servers of either generation ignore the unknown one; the
// helper centralizes that policy so table code stays generation-agnostic.
func addFleetIDParam(params url.Values, id string) {
	params.Set("fleet_id", id)
	params.Set("team_id", id)
}

// requireExposeSecrets gates tables that return live secret material behind
// the expose_secrets connection config flag.
func requireExposeSecrets(d *plugin.QueryData) error {
	config := GetConfig(d.Connection)
	if config.ExposeSecrets != nil && *config.ExposeSecrets {
		return nil
	}
	return errors.New("this table returns live secrets; set expose_secrets = true in the fleetdm connection config to query it")
}

// isPremiumError classifies err as a Fleet Premium licensing failure: a 402,
// or a 400/403 whose body mentions a license/premium requirement.
func isPremiumError(err error) bool {
	var apiErr *FleetAPIError
	if !errors.As(err, &apiErr) {
		return false
	}
	switch apiErr.StatusCode {
	case http.StatusPaymentRequired:
		return true
	case http.StatusBadRequest, http.StatusForbidden:
		snippet := strings.ToLower(apiErr.Snippet)
		return strings.Contains(snippet, "license") || strings.Contains(snippet, "premium")
	}
	return false
}

// handlePremiumError returns true when err is a Fleet Premium licensing
// failure that should be swallowed (treat as zero rows) instead of failing
// the query. It logs one Warn per connection+table, deduped via the
// connection cache.
func handlePremiumError(ctx context.Context, d *plugin.QueryData, table string, err error) bool {
	if !isPremiumError(err) {
		return false
	}
	cacheKey := "premium_warned/" + table
	if d != nil && d.ConnectionCache != nil {
		if _, warned := d.ConnectionCache.Get(ctx, cacheKey); warned {
			return true
		}
		if cerr := d.ConnectionCache.Set(ctx, cacheKey, true); cerr != nil {
			plugin.Logger(ctx).Warn("handlePremiumError", "connection_cache_set_error", cerr, "table", table)
		}
	}
	plugin.Logger(ctx).Warn("handlePremiumError", "table", table, "premium_feature_unavailable", err.Error())
	return true
}
