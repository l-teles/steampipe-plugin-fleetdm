package fleetdm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os" // Added for os.Getenv
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/turbot/steampipe-plugin-sdk/v6/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin/transform"
	"golang.org/x/time/rate"
)

// FleetTime is a custom time type that handles empty strings in JSON unmarshalling.
// The Fleet DM API may return empty strings for time fields, which causes
// the standard time.Time JSON unmarshaller to fail.
type FleetTime struct {
	time.Time
}

// UnmarshalJSON implements the json.Unmarshaler interface for FleetTime.
func (ft *FleetTime) UnmarshalJSON(data []byte) error {
	s := string(data)
	if s == "null" || s == `""` {
		ft.Time = time.Time{}
		return nil
	}
	return ft.Time.UnmarshalJSON(data)
}

// MarshalJSON implements the json.Marshaler interface for FleetTime.
func (ft FleetTime) MarshalJSON() ([]byte, error) {
	if ft.IsZero() {
		return []byte("null"), nil
	}
	return ft.Time.MarshalJSON()
}

// flexibleTimeTransform converts FleetTime or time.Time values to time.Time for Steampipe TIMESTAMP columns.
func flexibleTimeTransform(_ context.Context, d *transform.TransformData) (interface{}, error) {
	if d.Value == nil {
		return nil, nil
	}
	switch v := d.Value.(type) {
	case FleetTime:
		if v.IsZero() {
			return nil, nil
		}
		return v.Time, nil
	case *FleetTime:
		if v == nil || v.IsZero() {
			return nil, nil
		}
		return v.Time, nil
	case time.Time:
		if v.IsZero() {
			return nil, nil
		}
		return v, nil
	case *time.Time:
		if v == nil || v.IsZero() {
			return nil, nil
		}
		return *v, nil
	default:
		return nil, fmt.Errorf("flexibleTimeTransform: unexpected type %T", d.Value)
	}
}

// ipAddrTransform adapts host-reported addresses to IPADDR columns: empty
// becomes NULL and an IPv6 zone ("fe80::1%en0", which inet cannot store) is
// dropped. Anything else unparseable becomes NULL rather than failing the row.
func ipAddrTransform(ctx context.Context, d *transform.TransformData) (interface{}, error) {
	s, ok := d.Value.(string)
	if !ok {
		return nil, nil
	}
	s = strings.TrimSpace(s)
	if i := strings.IndexByte(s, '%'); i >= 0 {
		s = s[:i]
	}
	if s == "" {
		return nil, nil
	}
	if net.ParseIP(s) == nil {
		plugin.Logger(ctx).Warn("ipAddrTransform", "unparseable_ip", s, "column", d.ColumnName)
		return nil, nil
	}
	return s, nil
}

// FleetDMClient is a client for the FleetDM API.
type FleetDMClient struct {
	BaseURL    string
	APIToken   string
	HTTPClient *http.Client
	// limiter paces every request when max_requests_per_second is set; it
	// covers requests made inside list calls (paging, per-host fan-out) that
	// Steampipe's hydrate-level limiters do not see.
	limiter *rate.Limiter
}

// defaultRequestTimeout is used when request_timeout is not set in the config.
const defaultRequestTimeout = 30 * time.Second

// getClient returns the FleetDMClient for this connection, building it once
// and memoizing it in the connection cache. This avoids re-parsing the config
// and re-allocating an http.Client per list call / per hydrate row.
func getClient(ctx context.Context, d *plugin.QueryData) (*FleetDMClient, error) {
	const cacheKey = "fleetdm_client"
	if v, ok := d.ConnectionCache.Get(ctx, cacheKey); ok {
		if c, ok := v.(*FleetDMClient); ok {
			return c, nil
		}
	}
	c, err := NewFleetDMClient(ctx, d.Connection)
	if err != nil {
		return nil, err
	}
	// The SDK applies its own TTL; a rebuild after expiry is cheap.
	if err := d.ConnectionCache.Set(ctx, cacheKey, c); err != nil {
		plugin.Logger(ctx).Warn("getClient", "connection_cache_set_error", err)
	}
	return c, nil
}

// NewFleetDMClient creates a new FleetDM API client.
func NewFleetDMClient(ctx context.Context, connection *plugin.Connection) (*FleetDMClient, error) {
	config := GetConfig(connection) // Gets config from .spc file

	serverURL := ""
	apiToken := ""

	// Get Server URL: .spc file takes precedence, then environment variable
	if config.ServerURL != nil && *config.ServerURL != "" {
		serverURL = *config.ServerURL
		plugin.Logger(ctx).Info("NewFleetDMClient", "server_url_source", ".spc_file")
	} else {
		envURL := os.Getenv("FLEETDM_URL")
		if envURL != "" {
			serverURL = envURL
			plugin.Logger(ctx).Info("NewFleetDMClient", "server_url_source", "env_FLEETDM_URL")
		}
	}

	// Get API Token: .spc file takes precedence, then environment variable
	if config.APIToken != nil && *config.APIToken != "" {
		apiToken = *config.APIToken
		plugin.Logger(ctx).Info("NewFleetDMClient", "api_token_source", ".spc_file")
	} else {
		envToken := os.Getenv("FLEETDM_API_TOKEN")
		if envToken != "" {
			apiToken = envToken
			plugin.Logger(ctx).Info("NewFleetDMClient", "api_token_source", "env_FLEETDM_API_TOKEN")
		}
	}

	// Validate that we have the necessary configuration
	if serverURL == "" {
		return nil, errors.New("server_url must be configured in fleetdm.spc or via FLEETDM_URL environment variable")
	}
	if apiToken == "" {
		return nil, errors.New("api_token must be configured in fleetdm.spc or via FLEETDM_API_TOKEN environment variable")
	}

	// Validate the server URL: the API token is sent as a bearer header, so
	// plain http would leak it in cleartext. Only allow http for loopback hosts.
	parsedURL, err := url.Parse(serverURL)
	if err != nil {
		return nil, fmt.Errorf("server_url %q is not a valid URL: %w", serverURL, err)
	}
	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return nil, fmt.Errorf("server_url %q must be an absolute URL including scheme, e.g. https://fleet.example.com", serverURL)
	}
	switch parsedURL.Scheme {
	case "https":
		// ok
	case "http":
		host := parsedURL.Hostname()
		if host != "localhost" && host != "127.0.0.1" && host != "::1" {
			return nil, fmt.Errorf("server_url %q uses http, which would send the API token in cleartext; use https (http is only allowed for localhost)", serverURL)
		}
	default:
		return nil, fmt.Errorf("server_url %q has unsupported scheme %q; use https", serverURL, parsedURL.Scheme)
	}

	// Normalize the baseURL
	baseURL := strings.TrimSuffix(parsedURL.String(), "/")
	if strings.HasSuffix(baseURL, "/api/v1/fleet") {
		baseURL += "/"
	} else if strings.HasSuffix(baseURL, "/api/v1") {
		baseURL += "/fleet/"
	} else if strings.HasSuffix(baseURL, "/api") {
		baseURL += "/v1/fleet/"
	} else {
		baseURL += "/api/v1/fleet/"
	}

	plugin.Logger(ctx).Debug("NewFleetDMClient", "final_derived_base_url", baseURL)

	timeout := defaultRequestTimeout
	if config.RequestTimeout != nil {
		if *config.RequestTimeout <= 0 {
			return nil, fmt.Errorf("request_timeout must be a positive number of seconds, got %d", *config.RequestTimeout)
		}
		timeout = time.Duration(*config.RequestTimeout) * time.Second
	}

	var limiter *rate.Limiter
	if config.MaxRequestsPerSecond != nil {
		rps := *config.MaxRequestsPerSecond
		if rps <= 0 {
			return nil, fmt.Errorf("max_requests_per_second must be a positive number, got %d", rps)
		}
		limiter = rate.NewLimiter(rate.Limit(rps), rps)
	}

	return &FleetDMClient{
		BaseURL:  baseURL,
		APIToken: apiToken,
		HTTPClient: &http.Client{
			Timeout: timeout,
			// Never follow redirects: Go re-sends the bearer token on a
			// same-host redirect even to plain http, or to a path the plugin
			// did not ask for. Get reports the 3xx and its target instead.
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		limiter: limiter,
	}, nil
}

// Response body caps. Error bodies only feed logs and snippets; success
// bodies are bounded generously so a misbehaving server cannot exhaust memory.
const (
	maxErrorBodyBytes   = 64 << 10
	maxSuccessBodyBytes = 256 << 20
)

// readBody reads at most limit bytes of r, failing if the body is larger.
func readBody(r io.Reader, limit int64) ([]byte, error) {
	b, err := io.ReadAll(io.LimitReader(r, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(b)) > limit {
		return nil, fmt.Errorf("response body exceeds %d bytes", limit)
	}
	return b, nil
}

// truncateBody returns at most n bytes of a response body as a string, for use
// in error messages surfaced to the SQL client. The full body is only ever
// written to the plugin log.
func truncateBody(b []byte, n int) string {
	if len(b) <= n {
		return string(b)
	}
	return string(b[:n]) + "… (truncated)"
}

// FleetAPIError is returned for non-2xx API responses. It carries the HTTP
// status code so retry/ignore predicates can classify failures without string
// matching.
type FleetAPIError struct {
	StatusCode int
	Status     string
	URL        string
	Snippet    string        // truncated response body, safe for SQL clients
	RetryAfter time.Duration // parsed from the Retry-After header, 0 if absent
}

func (e *FleetAPIError) Error() string {
	if e.Snippet == "" {
		return fmt.Sprintf("API request to %s failed with status %s", e.URL, e.Status)
	}
	return fmt.Sprintf("API request to %s failed with status %s: %s", e.URL, e.Status, e.Snippet)
}

// parseRetryAfter parses a Retry-After header value (delay-seconds or HTTP-date).
func parseRetryAfter(v string) time.Duration {
	if v == "" {
		return 0
	}
	if secs, err := strconv.Atoi(v); err == nil && secs > 0 {
		return time.Duration(secs) * time.Second
	}
	if t, err := http.ParseTime(v); err == nil {
		if d := time.Until(t); d > 0 {
			return d
		}
	}
	return 0
}

// maxRetryAfterSleep caps how long Get will honor a server's Retry-After
// in-line before handing the error to the SDK retry machinery.
const maxRetryAfterSleep = 10 * time.Second

// isTransient reports whether err is a rate limit, transient server error or
// network failure (timeout, reset connection, truncated response) worth
// retrying. A cancelled query is not.
func isTransient(err error) bool {
	var apiErr *FleetAPIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == http.StatusTooManyRequests || apiErr.StatusCode >= 500
	}
	if err == nil || errors.Is(err, context.Canceled) {
		return false
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}
	return errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, io.EOF) || errors.Is(err, syscall.ECONNRESET)
}

// shouldRetryError tells the SDK retry machinery to retry rate limits and
// transient server errors.
func shouldRetryError(_ context.Context, _ *plugin.QueryData, _ *plugin.HydrateData, err error) bool {
	return isTransient(err)
}

// Retry policy for requests made inside a list call after rows may already
// have streamed, where the SDK will no longer retry the whole hydrate.
var (
	transientRetryAttempts = 4
	transientRetryBase     = 500 * time.Millisecond
)

// retryTransient runs fn, retrying transient errors with exponential backoff.
// Other errors and context cancellation return immediately.
func retryTransient(ctx context.Context, fn func() error) error {
	delay := transientRetryBase
	for attempt := 1; ; attempt++ {
		err := fn()
		if err == nil || !isTransient(err) || attempt >= transientRetryAttempts {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
		delay *= 2
	}
}

// shouldIgnoreError maps 404s to "no rows" instead of a query error, so
// get-by-id lookups for deleted resources behave like SQL misses.
func shouldIgnoreError(_ context.Context, _ *plugin.QueryData, _ *plugin.HydrateData, err error) bool {
	var apiErr *FleetAPIError
	return errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound
}

// Get performs a GET request to the specified FleetDM API endpoint and
// unmarshals the response into target (which may be nil to discard the body).
// Non-2xx responses are returned as *FleetAPIError.
func (c *FleetDMClient) Get(ctx context.Context, endpoint string, queryParams url.Values, target interface{}) error {
	// Construct the full URL
	// Ensure endpoint doesn't start with a slash if BaseURL already ends with one
	trimmedEndpoint := strings.TrimPrefix(endpoint, "/")
	fullURLString := c.BaseURL + trimmedEndpoint

	fullURL, err := url.Parse(fullURLString)
	if err != nil {
		plugin.Logger(ctx).Error("FleetDMClient.Get", "url_parse_error", err, "base_url", c.BaseURL, "endpoint", endpoint)
		return fmt.Errorf("error parsing base URL '%s' and endpoint '%s': %w", c.BaseURL, endpoint, err)
	}
	if queryParams != nil {
		fullURL.RawQuery = queryParams.Encode()
	}

	plugin.Logger(ctx).Debug("FleetDMClient.Get", "url", fullURL.String())

	if c.limiter != nil {
		if err := c.limiter.Wait(ctx); err != nil {
			return err
		}
	}

	// Create the request
	req, err := http.NewRequestWithContext(ctx, "GET", fullURL.String(), nil)
	if err != nil {
		plugin.Logger(ctx).Error("FleetDMClient.Get", "request_creation_error", err, "url", fullURL.String())
		return fmt.Errorf("error creating HTTP request for %s: %w", fullURL.String(), err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+c.APIToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "steampipe-plugin-fleetdm/"+pluginVersion)

	// Perform the request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		plugin.Logger(ctx).Error("FleetDMClient.Get", "http_do_error", err, "url", fullURL.String())
		return fmt.Errorf("error performing HTTP request to %s: %w", fullURL.String(), err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			plugin.Logger(ctx).Error("FleetDMClient.Get", "close_error", cerr, "url", fullURL.String())
		}
	}()

	// Check for non-2xx status codes
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		apiErr := &FleetAPIError{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			URL:        fullURL.String(),
			RetryAfter: parseRetryAfter(resp.Header.Get("Retry-After")),
		}
		if resp.StatusCode >= 300 && resp.StatusCode < 400 {
			apiErr.Snippet = fmt.Sprintf("the server redirected to %q, which the plugin does not follow; set server_url to the final Fleet address", resp.Header.Get("Location"))
			plugin.Logger(ctx).Error("FleetDMClient.Get", "redirect_refused", resp.Header.Get("Location"), "url", fullURL.String(), "status_code", resp.StatusCode)
			return apiErr
		}
		bodyBytes, readErr := io.ReadAll(io.LimitReader(resp.Body, maxErrorBodyBytes))
		if readErr != nil {
			plugin.Logger(ctx).Error("FleetDMClient.Get", "read_error_body_failed", readErr, "url", fullURL.String(), "status_code", resp.StatusCode)
		} else {
			plugin.Logger(ctx).Error("FleetDMClient.Get", "api_error_response", string(bodyBytes), "url", fullURL.String(), "status_code", resp.StatusCode)
			apiErr.Snippet = truncateBody(bodyBytes, 200)
		}

		// Honor Retry-After in-line (bounded, context-aware) so the SDK's
		// exponential backoff retries against a server that is ready again.
		if apiErr.RetryAfter > 0 && (resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusServiceUnavailable) {
			sleep := min(apiErr.RetryAfter, maxRetryAfterSleep)
			select {
			case <-ctx.Done():
			case <-time.After(sleep):
			}
		}
		return apiErr
	}

	// Decode the JSON response
	if target != nil {
		bodyBytes, err := readBody(resp.Body, maxSuccessBodyBytes)
		if err != nil {
			plugin.Logger(ctx).Error("FleetDMClient.Get", "read_body_for_decode_error", err, "url", fullURL.String())
			return fmt.Errorf("error reading response body from %s: %w", fullURL.String(), err)
		}

		// 204 No Content (e.g. an unknown CVE) or an empty body has nothing
		// to decode; target keeps its zero value.
		if resp.StatusCode == http.StatusNoContent || len(bytes.TrimSpace(bodyBytes)) == 0 {
			return nil
		}

		// The body is left out of the error and log: successful responses
		// can carry secrets that expose_secrets would otherwise hide.
		if err := json.Unmarshal(bodyBytes, target); err != nil {
			plugin.Logger(ctx).Error("FleetDMClient.Get", "json_decode_error", err, "url", fullURL.String(), "response_body_bytes", len(bodyBytes))
			return fmt.Errorf("error decoding JSON response from %s: %w", fullURL.String(), err)
		}
	}

	return nil
}

// uuidPattern matches the UUIDs Fleet uses for script executions and software
// installs; any other value cannot match a row and is never put in a path.
var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// ListMeta mirrors FleetDM's pagination meta object. Response envelopes embed
// it as a *pointer* so "meta absent from the response" (nil) is
// distinguishable from has_next_results=false.
type ListMeta struct {
	HasNextResults     bool `json:"has_next_results"`
	HasPreviousResults bool `json:"has_previous_results"`
}

// fetchPageFunc fetches one page of results. It returns the page's items and
// the response meta object (nil when the endpoint has no meta object).
type fetchPageFunc[T any] func(ctx context.Context, page, perPage int) (items []T, meta *ListMeta, err error)

// forEachItem is invoked per item; returning false stops pagination early
// (e.g. the query's LIMIT was reached).
type forEachItem[T any] func(item T) bool

// paginateCore drives page/per_page iteration. Termination rules:
//  1. an empty page always terminates (no data progress);
//  2. when the response has a meta object, meta.has_next_results is
//     authoritative;
//  3. otherwise keep fetching until an empty page.
//
// Deliberately NOT a rule: stopping when len(items) < perPage. Servers may
// clamp per_page, which would silently truncate results.
func paginateCore[T any](ctx context.Context, perPage int, fetch fetchPageFunc[T], each forEachItem[T]) error {
	for page := 0; ; page++ {
		items, meta, err := fetch(ctx, page, perPage)
		if err != nil {
			return err
		}
		for _, item := range items {
			if !each(item) {
				return nil
			}
		}
		if len(items) == 0 {
			if meta != nil && meta.HasNextResults {
				plugin.Logger(ctx).Warn("paginateCore", "empty_page_with_has_next_results", true, "page", page)
			}
			return nil
		}
		if meta != nil && !meta.HasNextResults {
			return nil
		}
	}
}

// paginatedList is the Steampipe shim over paginateCore used by table list
// functions: it sizes per_page from the query's LIMIT when smaller, streams
// every item, and stops as soon as the SDK reports no rows remaining.
func paginatedList[T any](ctx context.Context, d *plugin.QueryData, defaultPerPage int, fetch fetchPageFunc[T]) error {
	perPage := defaultPerPage
	if limit := d.QueryContext.Limit; limit != nil && *limit > 0 && *limit < int64(perPage) {
		perPage = int(*limit)
	}
	return paginateCore(ctx, perPage, fetch, func(item T) bool {
		d.StreamListItem(ctx, item)
		return d.RowsRemaining(ctx) > 0
	})
}
