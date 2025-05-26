package fleetdm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
)

// FleetDMClient is a client for the FleetDM API.
type FleetDMClient struct {
	BaseURL    string
	APIToken   string
	HTTPClient *http.Client
}

// NewFleetDMClient creates a new FleetDM API client.
func NewFleetDMClient(ctx context.Context, connection *plugin.Connection) (*FleetDMClient, error) {
	config := GetConfig(connection)

	if config.ServerURL == nil || *config.ServerURL == "" {
		return nil, errors.New("'server_url' must be set in the connection configuration. Edit your connection configuration file and then restart Steampipe")
	}
	if config.APIToken == nil || *config.APIToken == "" {
		return nil, errors.New("'api_token' must be set in the connection configuration. Edit your connection configuration file and then restart Steampipe")
	}

	// Normalize the baseURL provided by the user
	baseURL := strings.TrimSuffix(*config.ServerURL, "/")

	// Ensure BaseURL has the correct /api/v1/fleet/ prefix
	if strings.HasSuffix(baseURL, "/api/v1/fleet") {
		// Already has the full prefix, just ensure trailing slash
		baseURL += "/"
	} else if strings.HasSuffix(baseURL, "/api/v1") {
		// Has /api/v1, needs /fleet/
		baseURL += "/fleet/"
	} else if strings.HasSuffix(baseURL, "/api") {
		// Has /api, needs /v1/fleet/
		baseURL += "/v1/fleet/"
	} else {
		// Does not have /api part, append the full path
		baseURL += "/api/v1/fleet/"
	}

	plugin.Logger(ctx).Info("NewFleetDMClient", "configured_server_url", *config.ServerURL, "derived_base_url", baseURL)

	return &FleetDMClient{
		BaseURL:  baseURL,
		APIToken: *config.APIToken,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second, // Sensible default timeout
		},
	}, nil
}

// Get performs a GET request to the specified FleetDM API endpoint.
// The response is unmarshalled into the `target` interface.
func (c *FleetDMClient) Get(ctx context.Context, endpoint string, queryParams url.Values, target interface{}) (*http.Response, error) {
	// Construct the full URL
	// Ensure endpoint doesn't start with a slash if BaseURL already ends with one
	trimmedEndpoint := strings.TrimPrefix(endpoint, "/")
	fullURLString := c.BaseURL + trimmedEndpoint

	fullURL, err := url.Parse(fullURLString)
	if err != nil {
		plugin.Logger(ctx).Error("FleetDMClient.Get", "url_parse_error", err, "base_url", c.BaseURL, "endpoint", endpoint)
		return nil, fmt.Errorf("error parsing base URL '%s' and endpoint '%s': %w", c.BaseURL, endpoint, err)
	}
	if queryParams != nil {
		fullURL.RawQuery = queryParams.Encode()
	}

	plugin.Logger(ctx).Debug("FleetDMClient.Get", "url", fullURL.String())

	// Create the request
	req, err := http.NewRequestWithContext(ctx, "GET", fullURL.String(), nil)
	if err != nil {
		plugin.Logger(ctx).Error("FleetDMClient.Get", "request_creation_error", err, "url", fullURL.String())
		return nil, fmt.Errorf("error creating HTTP request for %s: %w", fullURL.String(), err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+c.APIToken)
	req.Header.Set("Accept", "application/json")

	// Perform the request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		plugin.Logger(ctx).Error("FleetDMClient.Get", "http_do_error", err, "url", fullURL.String())
		return resp, fmt.Errorf("error performing HTTP request to %s: %w", fullURL.String(), err)
	}

	// Check for non-2xx status codes
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		bodyBytes, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			plugin.Logger(ctx).Error("FleetDMClient.Get", "read_error_body_failed", readErr, "url", fullURL.String(), "status_code", resp.StatusCode)
			return resp, fmt.Errorf("API request to %s failed with status %s (unable to read error body)", fullURL.String(), resp.Status)
		}
		plugin.Logger(ctx).Error("FleetDMClient.Get", "api_error_response", string(bodyBytes), "url", fullURL.String(), "status_code", resp.StatusCode)
		return resp, fmt.Errorf("API request to %s failed with status %s: %s", fullURL.String(), resp.Status, string(bodyBytes))
	}

	// Decode the JSON response
	if target != nil {
		defer resp.Body.Close() // Ensure body is closed after decoding or if decoding fails
		// TeeReader would allow us to read the body for decoding and then potentially re-read it for logging if decoding fails.
		// However, for simplicity and common practice, we'll read it once.
		// If decoding fails, the original error from json.NewDecoder is usually sufficient.
		// For more detailed debugging, one might capture the body before attempting to decode.

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			plugin.Logger(ctx).Error("FleetDMClient.Get", "read_body_for_decode_error", err, "url", fullURL.String())
			return resp, fmt.Errorf("error reading response body from %s: %w", fullURL.String(), err)
		}

		// Now that we have the body bytes, we can attempt to unmarshal.
		// This also allows logging the raw body if unmarshalling fails.
		if err := json.Unmarshal(bodyBytes, target); err != nil {
			plugin.Logger(ctx).Error("FleetDMClient.Get", "json_decode_error", err, "url", fullURL.String(), "response_body_snippet", string(bodyBytes[:500])) // Log a snippet
			return resp, fmt.Errorf("error decoding JSON response from %s: %w. Response body: %s", fullURL.String(), err, string(bodyBytes))
		}
	}

	return resp, nil
}

// Hydrate function to get server_url from connection config
func getServerURL(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	config := GetConfig(d.Connection)
	if config.ServerURL == nil {
		return nil, fmt.Errorf("server_url is not configured")
	}
	return *config.ServerURL, nil
}
