package fleetdm

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"
	"time"

	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

// QuerySaved represents a saved query in FleetDM.
// Refer to: https://fleetdm.com/docs/rest-api/rest-api#saved-query-object
type QuerySaved struct {
	ID                 uint            `json:"id"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
	Name               string          `json:"name"`
	Description        string          `json:"description"`
	Query              string          `json:"query"` // The actual SQL query
	AuthorID           *uint           `json:"author_id"`
	AuthorName         string          `json:"author_name"`
	AuthorEmail        string          `json:"author_email"`
	ObserverCanRun     bool            `json:"observer_can_run"` // Whether observers can run this query
	TeamID             *uint           `json:"team_id"`          // Null if global
	AutomationsEnabled bool            `json:"automations_enabled"`
	Interval           *uint           `json:"interval"` // For scheduled queries, in seconds
	Platform           *string         `json:"platform"` // Comma-separated list or empty for all
	MinOsqueryVersion  *string         `json:"min_osquery_version"`
	Logging            *string         `json:"logging"` // "snapshot", "differential", "differential_ignore_removals"
	Stats              json.RawMessage `json:"stats"`   // Performance statistics, complex object
	Packs              []QueryPack     `json:"packs"`   // Packs this query belongs to (available on GET /queries/{id})
}

// QueryPack minimal info for a pack a query belongs to.
type QueryPack struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"` // e.g. "global", "team"
}

// ListQueriesResponse is the structure for the list queries API response.
// `GET /api/v1/fleet/queries` returns `{"queries": [...]}`
type ListQueriesResponse struct {
	Queries []QuerySaved `json:"queries"`
	// Meta  struct { // If pagination meta is introduced
	// 	HasNextResults bool `json:"has_next_results"`
	// } `json:"meta"`
}

// GetQueryResponse is the structure for the get query API response.
// `GET /api/v1/fleet/queries/{id}` returns `{"query": {...}}`
type GetQueryResponse struct {
	Query QuerySaved `json:"query"`
}

func tableFleetdmQuery(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "fleetdm_query",
		Description: "Saved queries in FleetDM.",
		List: &plugin.ListConfig{
			Hydrate: listQueries,
			KeyColumns: []*plugin.KeyColumn{ // Corrected: Use a slice of *plugin.KeyColumn
				{Name: "query_text_filter", Require: plugin.Optional}, // Maps to 'query' API parameter for search
				{Name: "team_id", Require: plugin.Optional},
				// TODO: Add other optional key columns if needed
				// {Name: "order_key", Require: plugin.Optional},
				// {Name: "order_direction", Require: plugin.Optional},
			},
		},
		Columns: []*plugin.Column{
			{Name: "id", Type: proto.ColumnType_INT, Description: "Unique ID of the saved query."},
			{Name: "name", Type: proto.ColumnType_STRING, Description: "Name of the saved query."},
			{Name: "query_sql", Type: proto.ColumnType_STRING, Transform: transform.FromField("Query"), Description: "The SQL content of the saved query."},
			{Name: "description", Type: proto.ColumnType_STRING, Description: "Description of the saved query."},
			{Name: "team_id", Type: proto.ColumnType_INT, Description: "ID of the team the query belongs to. Null if it's a global query."},
			{Name: "author_id", Type: proto.ColumnType_INT, Description: "ID of the user who created the query."},
			{Name: "author_name", Type: proto.ColumnType_STRING, Description: "Name of the user who created the query."},
			{Name: "author_email", Type: proto.ColumnType_STRING, Description: "Email of the user who created the query."},
			{Name: "observer_can_run", Type: proto.ColumnType_BOOL, Description: "Indicates if users with the observer role can run this query."},
			{Name: "automations_enabled", Type: proto.ColumnType_BOOL, Description: "Indicates if automations (scheduling) are enabled for this query."},
			{Name: "interval", Type: proto.ColumnType_INT, Description: "Interval in seconds for scheduled execution. Null if not scheduled."},
			{Name: "platform", Type: proto.ColumnType_STRING, Description: "Target platform(s) for the query (comma-separated, or empty for all)."},
			{Name: "min_osquery_version", Type: proto.ColumnType_STRING, Description: "Minimum osquery version required to run this query."},
			{Name: "logging_type", Type: proto.ColumnType_STRING, Transform: transform.FromField("Logging"), Description: "Type of logging for query results (e.g., snapshot, differential)."},
			{Name: "stats", Type: proto.ColumnType_JSON, Description: "Performance statistics for the query execution."},
			{Name: "packs", Type: proto.ColumnType_JSON, Description: "Packs this query belongs to (details available on GET)."},
			{Name: "created_at", Type: proto.ColumnType_TIMESTAMP, Description: "Timestamp when the query was created."},
			{Name: "updated_at", Type: proto.ColumnType_TIMESTAMP, Description: "Timestamp when the query was last updated."},

			// Key column for filtering via API 'query' parameter
			{Name: "query_text_filter", Type: proto.ColumnType_STRING, Transform: transform.FromQual("query_text_filter"), Description: "Search query string to filter saved queries by name or SQL. Use in WHERE clause."},

			// Connection config (server_url)
			{Name: "server_url", Type: proto.ColumnType_STRING, Hydrate: getServerURL, Transform: transform.FromValue(), Description: "FleetDM server URL from connection config."},
		},
	}
}

func listQueries(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := NewFleetDMClient(ctx, d.Connection)
	if err != nil {
		plugin.Logger(ctx).Error("fleetdm_query.listQueries", "connection_error", err)
		return nil, err
	}

	page := 0
	perPage := 50 // API default is 20, max 100

	// limit := d.QueryContext.Limit
	// if limit != nil && *limit < int64(perPage) {
	// 	// perPage = int(*limit)
	// }

	for {
		params := url.Values{}
		params.Add("page", strconv.Itoa(page))
		params.Add("per_page", strconv.Itoa(perPage))

		if d.EqualsQuals["query_text_filter"] != nil {
			params.Add("query", d.EqualsQuals["query_text_filter"].GetStringValue())
		}
		if d.EqualsQuals["team_id"] != nil {
			params.Add("team_id", strconv.FormatInt(d.EqualsQuals["team_id"].GetInt64Value(), 10))
		}
		// TODO: Support order_key and order_direction via Quals

		var response ListQueriesResponse
		_, err := client.Get(ctx, "queries", params, &response)
		if err != nil {
			plugin.Logger(ctx).Error("fleetdm_query.listQueries", "api_error", err, "page", page, "params", params.Encode())
			return nil, err
		}

		for _, query := range response.Queries {
			// The list endpoint for queries might not include 'packs'.
			// 'packs' are listed in the response for GET /api/v1/fleet/queries/{id}.
			// So, for list, 'packs' will be nil/empty.
			d.StreamListItem(ctx, query)
			if d.RowsRemaining(ctx) == 0 {
				plugin.Logger(ctx).Debug("fleetdm_query.listQueries", "limit_reached", true)
				return nil, nil
			}
		}

		if len(response.Queries) < perPage {
			plugin.Logger(ctx).Debug("fleetdm_query.listQueries", "end_of_results", true, "queries_on_page", len(response.Queries))
			break
		}

		page++
		plugin.Logger(ctx).Debug("fleetdm_query.listQueries", "next_page", page)
	}

	return nil, nil
}
