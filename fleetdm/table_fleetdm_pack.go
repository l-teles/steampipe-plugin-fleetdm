package fleetdm

import (
	"context"
	"encoding/json" // For json.RawMessage
	"net/url"
	"strconv"
	"time"

	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

// Pack represents a query pack in FleetDM.
// Refer to: https://fleetdm.com/docs/rest-api/rest-api#pack-object
type Pack struct {
	ID                 uint             `json:"id"`
	CreatedAt          time.Time        `json:"created_at"`
	UpdatedAt          time.Time        `json:"updated_at"`
	Name               string           `json:"name"`
	Description        string           `json:"description"`
	Platform           string           `json:"platform"` // Comma-separated list or empty for all
	Disabled           bool             `json:"disabled"`
	Type               string           `json:"type"`                // e.g., "global", "team"
	TeamID             *uint            `json:"team_id"`             // Null if global
	TargetCount        int              `json:"target_count"`        // Number of targets (hosts/labels)
	TotalScheduledQueriesCount int      `json:"total_scheduled_queries_count"` // Total scheduled queries in the pack
	Targets            *json.RawMessage `json:"targets,omitempty"`     // Only on GET /packs/{id}, complex object { hosts: [], labels: [], teams: [] }
	ScheduledQueries   []ScheduledQuery `json:"scheduled_queries,omitempty"` // Only on GET /packs/{id}
	AgentOptions       *json.RawMessage `json:"agent_options,omitempty"` // Present if pack is for a team
	HostIDs            []uint           `json:"host_ids,omitempty"`    // Host IDs this pack is targeted to (from GET /packs/{id})
	LabelIDs           []uint           `json:"label_ids,omitempty"`   // Label IDs this pack is targeted to (from GET /packs/{id})
	TeamIDs            []uint           `json:"team_ids,omitempty"`    // Team IDs this pack is targeted to (from GET /packs/{id}) - usually for global packs targeting teams
}

// ScheduledQuery represents a query within a pack.
// This is similar to QuerySaved but might have pack-specific attributes like interval.
type ScheduledQuery struct {
	ID                uint    `json:"id"` // This is the ID of the saved query itself
	Name              string  `json:"name"`
	Query             string  `json:"query"` // The actual SQL of the saved query
	Description       string  `json:"description"`
	Interval          uint    `json:"interval"`          // Interval for this query within the pack
	Platform          *string `json:"platform"`          // Platform for this query within the pack
	MinOsqueryVersion *string `json:"min_osquery_version"` // Min osquery version for this query within the pack
	Logging           string  `json:"logging"`           // snapshot, differential, differential_ignore_removals
	Removed           bool    `json:"removed"`           // Whether the query is removed (e.g. results are logged as removed)
	Snapshot          *bool   `json:"snapshot"`          // Whether to run as a snapshot query
	Shard             *uint   `json:"shard"`             // Shard number for the query
}

// ListPacksResponse for `GET /api/v1/fleet/packs`
type ListPacksResponse struct {
	Packs []Pack `json:"packs"`
	// Meta for pagination if API supports it for packs
}

// GetPackResponse for `GET /api/v1/fleet/packs/{id}`
type GetPackResponse struct {
	Pack Pack `json:"pack"`
}

func tableFleetdmPack(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "fleetdm_pack",
		Description: "Query packs in FleetDM.",
		List: &plugin.ListConfig{
			Hydrate: listPacks,
			// KeyColumns: plugin.KeyColumnEquals("team_id"), // If API supports filtering by team_id
		},
		Columns: []*plugin.Column{
			{Name: "id", Type: proto.ColumnType_INT, Description: "Unique ID of the pack."},
			{Name: "name", Type: proto.ColumnType_STRING, Description: "Name of the pack."},
			{Name: "description", Type: proto.ColumnType_STRING, Description: "Description of the pack."},
			{Name: "platform", Type: proto.ColumnType_STRING, Description: "Target platform(s) for the pack (comma-separated, or empty for all)."},
			{Name: "disabled", Type: proto.ColumnType_BOOL, Description: "Indicates if the pack is disabled."},
			{Name: "type", Type: proto.ColumnType_STRING, Description: "Type of the pack (e.g., 'global', 'team')."},
			{Name: "team_id", Type: proto.ColumnType_INT, Description: "ID of the team the pack belongs to. Null if it's a global pack."},
			{Name: "target_count", Type: proto.ColumnType_INT, Description: "Number of targets (hosts/labels/teams) for this pack."},
			{Name: "total_scheduled_queries_count", Type: proto.ColumnType_INT, Description: "Total number of scheduled queries in this pack."},
			{Name: "created_at", Type: proto.ColumnType_TIMESTAMP, Description: "Timestamp when the pack was created."},
			{Name: "updated_at", Type: proto.ColumnType_TIMESTAMP, Description: "Timestamp when the pack was last updated."},

			// Details available from GET /packs/{id}
			{Name: "targets", Type: proto.ColumnType_JSON, Description: "Target hosts, labels, and teams for this pack (details from GET)."},
			{Name: "scheduled_queries", Type: proto.ColumnType_JSON, Description: "Scheduled queries within this pack (details from GET)."},
			{Name: "agent_options", Type: proto.ColumnType_JSON, Description: "Agent options associated with the pack (if it's a team pack, from GET)."},
			{Name: "host_ids", Type: proto.ColumnType_JSON, Transform: transform.FromField("HostIDs"), Description: "List of host IDs targeted by this pack (from GET)."},
			{Name: "label_ids", Type: proto.ColumnType_JSON, Transform: transform.FromField("LabelIDs"), Description: "List of label IDs targeted by this pack (from GET)."},
			{Name: "team_ids_targeted", Type: proto.ColumnType_JSON, Transform: transform.FromField("TeamIDs"), Description: "List of team IDs targeted by this pack, typically for global packs (from GET)."},


			// Connection config (server_url)
			{Name: "server_url", Type: proto.ColumnType_STRING, Hydrate: getServerURL, Transform: transform.FromValue(), Description: "FleetDM server URL from connection config."},
		},
	}
}

func listPacks(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := NewFleetDMClient(ctx, d.Connection)
	if err != nil {
		plugin.Logger(ctx).Error("fleetdm_pack.listPacks", "connection_error", err)
		return nil, err
	}

	// Pagination for packs: The /api/v1/fleet/packs endpoint supports `page` and `per_page`
	page := 0
	perPage := 50 // API default is 20, max 100

	limit := d.QueryContext.Limit
	if limit != nil && *limit < int64(perPage) {
		// perPage = int(*limit)
	}

	for {
		params := url.Values{}
		params.Add("page", strconv.Itoa(page))
		params.Add("per_page", strconv.Itoa(perPage))
		// TODO: Add KeyColumn for team_id if API supports it:
		// if d.EqualsQuals["team_id"] != nil {
		// 	params.Add("team_id", strconv.FormatInt(d.EqualsQuals["team_id"].GetInt64Value(), 10))
		// }

		var response ListPacksResponse
		_, err := client.Get(ctx, "packs", params, &response)
		if err != nil {
			plugin.Logger(ctx).Error("fleetdm_pack.listPacks", "api_error", err, "page", page, "params", params.Encode())
			return nil, err
		}

		for _, pack := range response.Packs {
			// List endpoint for packs usually provides summary data.
			// Detailed fields like 'targets', 'scheduled_queries', 'agent_options' are from GET /packs/{id}.
			// These will be null/empty here and populated by getPack if a single item is fetched.
			d.StreamListItem(ctx, pack)
			if d.RowsRemaining(ctx) == 0 {
				plugin.Logger(ctx).Debug("fleetdm_pack.listPacks", "limit_reached", true)
				return nil, nil
			}
		}

		// Pagination check: if the number of packs returned is less than per_page,
		// it's likely the last page. The /packs endpoint does not specify a `meta.has_next_results`.
		if len(response.Packs) < perPage {
			plugin.Logger(ctx).Debug("fleetdm_pack.listPacks", "end_of_results", true, "packs_on_page", len(response.Packs))
			break
		}

		page++
		plugin.Logger(ctx).Debug("fleetdm_pack.listPacks", "next_page", page)
	}

	return nil, nil
}
