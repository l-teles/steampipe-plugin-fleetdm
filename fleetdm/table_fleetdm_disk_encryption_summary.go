package fleetdm

import (
	"context"
	"net/url"
	"strconv"

	"github.com/turbot/steampipe-plugin-sdk/v6/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin/transform"
)

// diskEncryptionCounts is the per-platform host count for one status.
type diskEncryptionCounts struct {
	MacOS   int `json:"macos"`
	Windows int `json:"windows"`
	Linux   int `json:"linux"`
}

// DiskEncryptionStatus is one row of the disk encryption summary.
type DiskEncryptionStatus struct {
	Status  string
	MacOS   int
	Windows int
	Linux   int
}

// diskEncryptionStatuses fixes the row order of the summary.
var diskEncryptionStatuses = []string{"verified", "verifying", "action_required", "enforcing", "failed", "removing_enforcement"}

func tableFleetdmDiskEncryptionSummary(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "fleetdm_disk_encryption_summary",
		Description: "Disk encryption status counts per platform, one row per status (Fleet Premium, MDM).",
		List: &plugin.ListConfig{
			Hydrate: listDiskEncryptionSummary,
			KeyColumns: []*plugin.KeyColumn{
				{Name: "team_id", Require: plugin.Optional},
			},
		},
		Columns: []*plugin.Column{
			{Name: "status", Type: proto.ColumnType_STRING, Description: "Disk encryption status: verified, verifying, action_required, enforcing, failed or removing_enforcement."},
			{Name: "macos", Type: proto.ColumnType_INT, Transform: transform.FromField("MacOS"), Description: "Number of macOS hosts in this status."},
			{Name: "windows", Type: proto.ColumnType_INT, Description: "Number of Windows hosts in this status."},
			{Name: "linux", Type: proto.ColumnType_INT, Description: "Number of Linux hosts in this status."},
			{Name: "team_id", Type: proto.ColumnType_INT, Transform: transform.FromQual("team_id"), Description: "Summarize only this team (fleet); 0 for hosts with no team. Set in WHERE clause."},
		},
	}
}

func listDiskEncryptionSummary(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := getClient(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("fleetdm_disk_encryption_summary.listDiskEncryptionSummary", "connection_error", err)
		return nil, err
	}
	params := url.Values{}
	if q := d.EqualsQuals["team_id"]; q != nil {
		client.addFleetIDParam(ctx, d, params, strconv.FormatInt(q.GetInt64Value(), 10))
	}
	var response map[string]diskEncryptionCounts
	if err := client.Get(ctx, "disk_encryption", params, &response); err != nil {
		if handleUnavailableFeature(ctx, d, "fleetdm_disk_encryption_summary", err) {
			return nil, nil
		}
		plugin.Logger(ctx).Error("fleetdm_disk_encryption_summary.listDiskEncryptionSummary", "api_error", err)
		return nil, err
	}
	for _, status := range diskEncryptionStatuses {
		c, ok := response[status]
		if !ok {
			continue
		}
		d.StreamListItem(ctx, DiskEncryptionStatus{Status: status, MacOS: c.MacOS, Windows: c.Windows, Linux: c.Linux})
	}
	return nil, nil
}
