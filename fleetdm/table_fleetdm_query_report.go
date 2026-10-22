package fleetdm

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"

	"github.com/turbot/steampipe-plugin-sdk/v6/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin/transform"
)

// QueryReportRow is one host's stored result row for a saved query (report).
// Refer to: https://fleetdm.com/docs/rest-api/rest-api#get-report
type QueryReportRow struct {
	HostID        uint            `json:"host_id"`
	HostName      string          `json:"host_name"`
	LastFetched   FleetTime       `json:"last_fetched"`
	Columns       json.RawMessage `json:"columns"`
	QueryID       uint            `json:"-"`
	ReportClipped bool            `json:"-"`
}

// GetQueryReportResponse for `GET /api/v1/fleet/reports/{id}/report`. Without
// paging parameters Fleet returns every stored row (Fleet caps a report at
// 1,000 rows); older Fleet versions do not page this endpoint at all.
type GetQueryReportResponse struct {
	Results       []QueryReportRow `json:"results"`
	ReportClipped bool             `json:"report_clipped"`
}

func tableFleetdmQueryReport(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "fleetdm_query_report",
		Description: "Results stored for saved queries (reports), one row per host result row; the result itself needs expose_secrets = true. Without a query_id filter every saved query, global and per team, is read.",
		List: &plugin.ListConfig{
			Hydrate: listQueryReports,
			KeyColumns: []*plugin.KeyColumn{
				{Name: "query_id", Require: plugin.Optional},
			},
		},
		Columns: []*plugin.Column{
			{Name: "query_id", Type: proto.ColumnType_INT, Transform: transform.FromField("QueryID"), Description: "ID of the saved query (report)."},
			{Name: "host_id", Type: proto.ColumnType_INT, Description: "ID of the host that returned the row."},
			{Name: "host_name", Type: proto.ColumnType_STRING, Description: "Display name of the host."},
			{Name: "last_fetched", Type: proto.ColumnType_TIMESTAMP, Transform: transform.FromField("LastFetched").Transform(flexibleTimeTransform), Description: "When the host last sent results for this query."},
			{Name: "columns", Type: proto.ColumnType_JSON, Hydrate: itemIfSecretsExposed, Transform: transform.FromField("Columns"), Description: "The result row: column name to value. NULL unless expose_secrets = true, as osquery results can hold credentials or personal data."},
			{Name: "report_clipped", Type: proto.ColumnType_BOOL, Description: "Whether Fleet stopped storing results because the report reached its row limit."},
		},
	}
}

func listQueryReports(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := getClient(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("fleetdm_query_report.listQueryReports", "connection_error", err)
		return nil, err
	}
	var ids []uint
	if q := d.EqualsQuals["query_id"]; q != nil {
		v := q.GetInt64Value()
		if v <= 0 {
			return nil, nil
		}
		ids = []uint{uint(v)}
	} else {
		// /reports without a team returns only global queries, so list every
		// team too.
		scopes, err := teamScopes(ctx, d, client, "fleetdm_query_report")
		if err != nil {
			return nil, err
		}
		seen := map[uint]bool{}
		for _, scope := range scopes {
			err := paginateCore(ctx, 100, func(ctx context.Context, page, perPage int) ([]QuerySaved, *ListMeta, error) {
				params := url.Values{"page": {strconv.Itoa(page)}, "per_page": {strconv.Itoa(perPage)}}
				if scope != nil {
					client.addFleetIDParam(ctx, d, params, scopeParam(scope))
				}
				var response ListQueriesResponse
				err := retryTransient(ctx, func() error {
					return client.GetCompat(ctx, d, famReports, nil, params, &response)
				})
				if err != nil {
					return nil, nil, err
				}
				return coalesceSlice(response.Reports, response.Queries), nil, nil
			}, func(q QuerySaved) bool {
				if !seen[q.ID] {
					seen[q.ID] = true
					ids = append(ids, q.ID)
				}
				return true
			})
			if err != nil {
				if skipScope(ctx, d, "fleetdm_query_report", scope, err) {
					continue
				}
				plugin.Logger(ctx).Error("fleetdm_query_report.listQueryReports", "api_error", err)
				return nil, err
			}
		}
	}
	for _, id := range ids {
		var response GetQueryReportResponse
		// Rows from earlier reports may have streamed, so retry here rather
		// than fail the whole query on a transient error.
		err := retryTransient(ctx, func() error {
			return client.GetCompat(ctx, d, famReportResults, []any{id}, nil, &response)
		})
		if err != nil {
			if is404(err) {
				continue
			}
			plugin.Logger(ctx).Error("fleetdm_query_report.listQueryReports", "api_error", err, "query_id", id)
			return nil, err
		}
		for _, row := range response.Results {
			row.QueryID, row.ReportClipped = id, response.ReportClipped
			d.StreamListItem(ctx, row)
			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}
	}
	return nil, nil
}
