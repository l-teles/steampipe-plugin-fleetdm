package fleetdm

import (
	"context"
	"net/url"
	"strconv"

	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

// Label represents a FleetDM label.
// Refer to: https://fleetdm.com/docs/rest-api/rest-api#label-object
type Label struct {
	ID                  uint      `json:"id"`
	CreatedAt           FleetTime `json:"created_at"`
	UpdatedAt           FleetTime `json:"updated_at"`
	Name                string    `json:"name"`
	Description         string    `json:"description"`
	Query               string    `json:"query"`                 // The SQL query for dynamic labeling
	Platform            string    `json:"platform"`              // e.g., "darwin", "windows", "linux", "" for all
	LabelType           string    `json:"label_type"`            // "regular" or "builtin"
	LabelMembershipType string    `json:"label_membership_type"` // "dynamic" or "manual" (manual not via API yet for creation)
	HostCount           int       `json:"host_count"`
	DisplayText         string    `json:"display_text"` // Usually same as name
	BuiltIn             bool      `json:"built_in"`     // Derived from label_type == "builtin"
	// Hosts field is not typically included in list/get label, but on a separate endpoint like /labels/{id}/hosts
}

// ListLabelsResponse for `GET /api/v1/fleet/labels`
// The API returns {"labels": [...]}
type ListLabelsResponse struct {
	Labels []Label `json:"labels"`
	// Meta for pagination if API supports it
}

// GetLabelResponse for `GET /api/v1/fleet/labels/{id}`
// The API returns {"label": {...}}
type GetLabelResponse struct {
	Label Label `json:"label"`
}

func tableFleetdmLabel(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "fleetdm_label",
		Description: "Labels used for grouping hosts in FleetDM.",
		List: &plugin.ListConfig{
			Hydrate: listLabels,
			KeyColumns: []*plugin.KeyColumn{
				{Name: "team_id", Require: plugin.Optional}, // Filter by team (Fleet Premium). Use 'global' for global-only labels.
			},
		},
		Columns: []*plugin.Column{
			{Name: "id", Type: proto.ColumnType_INT, Description: "Unique ID of the label."},
			{Name: "name", Type: proto.ColumnType_STRING, Description: "Name of the label."},
			{Name: "display_text", Type: proto.ColumnType_STRING, Description: "Display text for the label, usually the same as the name."},
			{Name: "description", Type: proto.ColumnType_STRING, Description: "Description of the label."},
			{Name: "query_sql", Type: proto.ColumnType_STRING, Transform: transform.FromField("Query"), Description: "The SQL query used for dynamic labeling."},
			{Name: "platform", Type: proto.ColumnType_STRING, Description: "Target platform(s) for the label (e.g., 'darwin', 'windows', 'linux', or empty for all)."},
			{Name: "label_type", Type: proto.ColumnType_STRING, Description: "Type of the label, e.g., 'regular' or 'builtin'."},
			{Name: "label_membership_type", Type: proto.ColumnType_STRING, Description: "Membership type, e.g., 'dynamic' or 'manual'."},
			{Name: "host_count", Type: proto.ColumnType_INT, Description: "Number of hosts associated with this label."},
			{Name: "built_in", Type: proto.ColumnType_BOOL, Description: "Indicates if the label is a built-in label."},
			{Name: "created_at", Type: proto.ColumnType_TIMESTAMP, Transform: transform.FromField("CreatedAt").Transform(flexibleTimeTransform), Description: "Timestamp when the label was created."},
			{Name: "updated_at", Type: proto.ColumnType_TIMESTAMP, Transform: transform.FromField("UpdatedAt").Transform(flexibleTimeTransform), Description: "Timestamp when the label was last updated."},

			// Query parameters that can be used for filtering (key columns)
			{Name: "team_id", Type: proto.ColumnType_STRING, Transform: transform.FromQual("team_id"), Description: "Filter by team (Fleet Premium). Use 'global' for global-only labels. Set in WHERE clause."},
		},
	}
}

func listLabels(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := NewFleetDMClient(ctx, d.Connection)
	if err != nil {
		plugin.Logger(ctx).Error("fleetdm_label.listLabels", "connection_error", err)
		return nil, err
	}

	// Pagination for labels: The /api/v1/fleet/labels endpoint supports `page` and `per_page`
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

		if d.EqualsQuals["team_id"] != nil {
			params.Add("team_id", d.EqualsQuals["team_id"].GetStringValue())
		}

		var response ListLabelsResponse
		_, err := client.Get(ctx, "labels", params, &response)
		if err != nil {
			plugin.Logger(ctx).Error("fleetdm_label.listLabels", "api_error", err, "page", page, "params", params.Encode())
			return nil, err
		}

		for _, label := range response.Labels {
			d.StreamListItem(ctx, label)
			if d.RowsRemaining(ctx) == 0 {
				plugin.Logger(ctx).Debug("fleetdm_label.listLabels", "limit_reached", true)
				return nil, nil
			}
		}

		// Pagination check: if the number of labels returned is less than per_page,
		// it's likely the last page. The /labels endpoint does not specify a `meta.has_next_results`.
		if len(response.Labels) < perPage {
			plugin.Logger(ctx).Debug("fleetdm_label.listLabels", "end_of_results", true, "labels_on_page", len(response.Labels))
			break
		}

		page++
		plugin.Logger(ctx).Debug("fleetdm_label.listLabels", "next_page", page)
	}

	return nil, nil
}
