package fleetdm

import (
	"context"
	"net/url"
	"strconv"
	"time"

	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

// Carve represents a file carving session in FleetDM.
type Carve struct {
	ID         uint      `json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	HostID     uint      `json:"host_id"`
	Name       string    `json:"name"`
	BlockCount int64     `json:"block_count"`
	BlockSize  int64     `json:"block_size"`
	CarveSize  int64     `json:"carve_size"`
	CarveID    string    `json:"carve_id"`
	RequestID  string    `json:"request_id"`
	SessionID  string    `json:"session_id"`
	Expired    bool      `json:"expired"`
	MaxBlock   int64     `json:"max_block"`
	Error      *string   `json:"error,omitempty"` // Use pointer for optional field
}

// ListCarvesResponse is the structure for the list carves API response.
type ListCarvesResponse struct {
	Carves []Carve `json:"carves"`
}

func tableFleetdmCarve(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "fleetdm_carve",
		Description: "Information about file carving sessions in FleetDM.",
		List: &plugin.ListConfig{
			Hydrate: listCarves,
			// KeyColumns can be added here later if the API starts supporting filtering, e.g., by host_id
			// KeyColumns: []*plugin.KeyColumn{
			// 	{Name: "host_id", Require: plugin.Optional},
			// },
		},
		// No GetConfig as individual carves are not typically fetched by ID via a dedicated endpoint.
		Columns: []*plugin.Column{
			{Name: "id", Type: proto.ColumnType_INT, Description: "Unique ID of the carve session."},
			{Name: "name", Type: proto.ColumnType_STRING, Description: "The name of the carve session, typically including hostname and timestamp."},
			{Name: "host_id", Type: proto.ColumnType_INT, Description: "The ID of the host from which the file was carved."},
			{Name: "carve_id", Type: proto.ColumnType_STRING, Description: "The unique identifier for the carve data."},
			{Name: "session_id", Type: proto.ColumnType_STRING, Description: "The osquery session ID for the carve."},
			{Name: "request_id", Type: proto.ColumnType_STRING, Description: "The request ID, often from a distributed query."},
			{Name: "carve_size", Type: proto.ColumnType_INT, Description: "The total size of the carved file in bytes."},
			{Name: "block_count", Type: proto.ColumnType_INT, Description: "The number of blocks received for the carve."},
			{Name: "block_size", Type: proto.ColumnType_INT, Description: "The maximum size of each block in bytes."},
			{Name: "max_block", Type: proto.ColumnType_INT, Description: "The index of the last block received."},
			{Name: "expired", Type: proto.ColumnType_BOOL, Description: "Indicates if the carve session has expired."},
			{Name: "error", Type: proto.ColumnType_STRING, Description: "Any error message associated with the carve session."},
			{Name: "created_at", Type: proto.ColumnType_TIMESTAMP, Description: "Timestamp when the carve session was created."},

			// Connection config (server_url)
			{Name: "server_url", Type: proto.ColumnType_STRING, Hydrate: getServerURL, Transform: transform.FromValue(), Description: "FleetDM server URL from connection config."},
		},
	}
}

func listCarves(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := NewFleetDMClient(ctx, d.Connection)
	if err != nil {
		plugin.Logger(ctx).Error("fleetdm_carve.listCarves", "connection_error", err)
		return nil, err
	}

	page := 0
	perPage := 50

	for {
		params := url.Values{}
		params.Add("page", strconv.Itoa(page))
		params.Add("per_page", strconv.Itoa(perPage))
		params.Add("order_key", "id")
		params.Add("order_direction", "desc") // Get most recent carves first
		params.Add("expired", "true") // Also get expired carves

		var response ListCarvesResponse
		_, err := client.Get(ctx, "carves", params, &response)
		if err != nil {
			plugin.Logger(ctx).Error("fleetdm_carve.listCarves", "api_error", err, "page", page, "params", params.Encode())
			return nil, err
		}

		for _, carve := range response.Carves {
			d.StreamListItem(ctx, carve)
			if d.RowsRemaining(ctx) == 0 {
				plugin.Logger(ctx).Debug("fleetdm_carve.listCarves", "limit_reached", true)
				return nil, nil
			}
		}

		// The /carves endpoint does not specify a meta object for pagination,
		// so we rely on the number of items returned.
		if len(response.Carves) < perPage {
			plugin.Logger(ctx).Debug("fleetdm_carve.listCarves", "end_of_results", true, "carves_on_page", len(response.Carves))
			break
		}

		page++
		plugin.Logger(ctx).Debug("fleetdm_carve.listCarves", "next_page", page)
	}

	return nil, nil
}
