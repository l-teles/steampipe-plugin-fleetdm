package fleetdm

import (
	"context"

	"github.com/turbot/steampipe-plugin-sdk/v6/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin"
)

// RESTAPIEndpoint is one entry of Fleet's REST API catalog, used to scope
// API-only users (Fleet 4.90+).
type RESTAPIEndpoint struct {
	Method      string `json:"method"`
	Path        string `json:"path"`
	DisplayName string `json:"display_name"`
	Deprecated  bool   `json:"deprecated"`
}

// ListRESTAPIEndpointsResponse for `GET /api/v1/fleet/rest_api` (not
// paginated; the page parameter is ignored).
type ListRESTAPIEndpointsResponse struct {
	APIEndpoints []RESTAPIEndpoint `json:"api_endpoints"`
}

func tableFleetdmRESTAPIEndpoint(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "fleetdm_rest_api_endpoint",
		Description: "Fleet's REST API endpoints that API-only users can be limited to (Fleet Premium, 4.90+).",
		List: &plugin.ListConfig{
			Hydrate: listRESTAPIEndpoints,
		},
		Columns: []*plugin.Column{
			{Name: "method", Type: proto.ColumnType_STRING, Description: "HTTP method, e.g. 'GET'."},
			{Name: "path", Type: proto.ColumnType_STRING, Description: "Endpoint path, e.g. '/api/v1/fleet/hosts'."},
			{Name: "display_name", Type: proto.ColumnType_STRING, Description: "Human-readable name of the endpoint."},
			{Name: "deprecated", Type: proto.ColumnType_BOOL, Description: "Whether the endpoint is deprecated."},
		},
	}
}

func listRESTAPIEndpoints(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := getClient(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("fleetdm_rest_api_endpoint.listRESTAPIEndpoints", "connection_error", err)
		return nil, err
	}
	var response ListRESTAPIEndpointsResponse
	if err := client.Get(ctx, "rest_api", nil, &response); err != nil {
		if handleUnavailableFeature(ctx, d, "fleetdm_rest_api_endpoint", err) {
			return nil, nil
		}
		plugin.Logger(ctx).Error("fleetdm_rest_api_endpoint.listRESTAPIEndpoints", "api_error", err)
		return nil, err
	}
	for _, e := range response.APIEndpoints {
		d.StreamListItem(ctx, e)
		if d.RowsRemaining(ctx) == 0 {
			break
		}
	}
	return nil, nil
}
