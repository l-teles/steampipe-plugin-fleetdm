package fleetdm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os" // Added for os.Getenv
	"strings"
	"time"

	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
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

// FleetDMClient is a client for the FleetDM API.
type FleetDMClient struct {
	BaseURL    string
	APIToken   string
	HTTPClient *http.Client
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

	// Normalize the baseURL
	baseURL := strings.TrimSuffix(serverURL, "/")
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


	return &FleetDMClient{
		BaseURL:  baseURL,
		APIToken: apiToken,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second, 
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
		defer func() {
			if err := resp.Body.Close(); err != nil {
				plugin.Logger(ctx).Error("FleetDMClient.Get", "close_error", err, "url", fullURL.String())
			}
		}()
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
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			plugin.Logger(ctx).Error("FleetDMClient.Get", "read_body_for_decode_error", err, "url", fullURL.String())
			return resp, fmt.Errorf("error reading response body from %s: %w", fullURL.String(), err)
		}


		if err := json.Unmarshal(bodyBytes, target); err != nil {
			plugin.Logger(ctx).Error("FleetDMClient.Get", "json_decode_error", err, "url", fullURL.String(), "response_body_snippet", string(bodyBytes[:500])) // Log a snippet
			return resp, fmt.Errorf("error decoding JSON response from %s: %w. Response body: %s", fullURL.String(), err, string(bodyBytes))
		}
	}

	return resp, nil
}