package fleetdm

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/turbot/steampipe-plugin-sdk/v6/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin/transform"
)

// LabelHost is one host that is a member of a label.
// Refer to: https://fleetdm.com/docs/rest-api/rest-api#list-hosts-in-a-label
type LabelHost struct {
	ID          uint    `json:"id"`
	Hostname    string  `json:"hostname"`
	DisplayName string  `json:"display_name"`
	Platform    string  `json:"platform"`
	OsVersion   string  `json:"os_version"`
	Status      string  `json:"status"`
	TeamID      *uint   `json:"team_id"`
	FleetID     *uint   `json:"fleet_id"` // team_id since Fleet 4.82
	TeamName    *string `json:"team_name"`
	FleetName   *string `json:"fleet_name"`
	LabelID     uint    `json:"-"`
	LabelName   string  `json:"-"`
}

// ListLabelHostsResponse for `GET /api/v1/fleet/labels/{id}/hosts` (no meta;
// paging ends on an empty page).
type ListLabelHostsResponse struct {
	Hosts []LabelHost `json:"hosts"`
}

func tableFleetdmLabelHost(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "fleetdm_label_host",
		Description: "Label membership: one row per host in each label. Without a label_id filter every label is read.",
		List: &plugin.ListConfig{
			Hydrate: listLabelHosts,
			KeyColumns: []*plugin.KeyColumn{
				{Name: "label_id", Require: plugin.Optional},
			},
		},
		Columns: []*plugin.Column{
			{Name: "label_id", Type: proto.ColumnType_INT, Transform: transform.FromField("LabelID"), Description: "ID of the label."},
			{Name: "label_name", Type: proto.ColumnType_STRING, Description: "Name of the label. Empty when filtering by label_id."},
			{Name: "host_id", Type: proto.ColumnType_INT, Transform: transform.FromField("ID"), Description: "ID of the host."},
			{Name: "hostname", Type: proto.ColumnType_STRING, Description: "Hostname of the host."},
			{Name: "display_name", Type: proto.ColumnType_STRING, Description: "Display name of the host."},
			{Name: "platform", Type: proto.ColumnType_STRING, Description: "Platform of the host."},
			{Name: "os_version", Type: proto.ColumnType_STRING, Description: "Operating system version of the host."},
			{Name: "status", Type: proto.ColumnType_STRING, Description: "Status of the host (online, offline, mia, new)."},
			{Name: "team_id", Type: proto.ColumnType_INT, Transform: transform.FromField("FleetID", "TeamID"), Description: "ID of the team (fleet) the host belongs to."},
			{Name: "team_name", Type: proto.ColumnType_STRING, Transform: transform.FromField("FleetName", "TeamName"), Description: "Name of the team (fleet) the host belongs to."},
		},
	}
}

func listLabelHosts(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := getClient(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("fleetdm_label_host.listLabelHosts", "connection_error", err)
		return nil, err
	}
	var labels []Label
	if q := d.EqualsQuals["label_id"]; q != nil {
		v := q.GetInt64Value()
		if v <= 0 {
			return nil, nil
		}
		labels = []Label{{ID: uint(v)}}
	} else {
		err := paginateCore(ctx, 100, func(ctx context.Context, page, perPage int) ([]Label, *ListMeta, error) {
			params := url.Values{"page": {strconv.Itoa(page)}, "per_page": {strconv.Itoa(perPage)}}
			var response ListLabelsResponse
			if err := client.Get(ctx, "labels", params, &response); err != nil {
				return nil, nil, err
			}
			return response.Labels, nil, nil
		}, func(l Label) bool {
			labels = append(labels, l)
			return true
		})
		if err != nil {
			plugin.Logger(ctx).Error("fleetdm_label_host.listLabelHosts", "api_error", err)
			return nil, err
		}
	}
	for _, label := range labels {
		err := paginatedList(ctx, d, 500, func(ctx context.Context, page, perPage int) ([]LabelHost, *ListMeta, error) {
			params := url.Values{"page": {strconv.Itoa(page)}, "per_page": {strconv.Itoa(perPage)}, "order_key": {"id"}}
			var response ListLabelHostsResponse
			if err := client.Get(ctx, fmt.Sprintf("labels/%d/hosts", label.ID), params, &response); err != nil {
				return nil, nil, err
			}
			for i := range response.Hosts {
				response.Hosts[i].LabelID, response.Hosts[i].LabelName = label.ID, label.Name
			}
			return response.Hosts, nil, nil
		})
		if err != nil && !is404(err) {
			plugin.Logger(ctx).Error("fleetdm_label_host.listLabelHosts", "api_error", err, "label_id", label.ID)
			return nil, err
		}
		if d.RowsRemaining(ctx) == 0 {
			break
		}
	}
	return nil, nil
}
