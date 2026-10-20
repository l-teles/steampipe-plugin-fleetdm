package fleetdm

import (
	"context"
	"net/url"
	"strconv"

	"github.com/turbot/steampipe-plugin-sdk/v6/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin/transform"
)

// CertificateTemplate is a certificate Fleet requests from a certificate
// authority for a team's hosts. TeamID is the scope it was listed under.
// Refer to: https://fleetdm.com/docs/rest-api/rest-api#list-certificates
type CertificateTemplate struct {
	ID                       uint      `json:"id"`
	Name                     string    `json:"name"`
	SubjectName              string    `json:"subject_name"`
	SubjectAlternativeName   string    `json:"subject_alternative_name"`
	CertificateAuthorityID   uint      `json:"certificate_authority_id"`
	CertificateAuthorityName string    `json:"certificate_authority_name"`
	CreatedAt                FleetTime `json:"created_at"`
	TeamID                   *uint     `json:"-"`
}

// ListCertificateTemplatesResponse for `GET /api/v1/fleet/certificates`.
type ListCertificateTemplatesResponse struct {
	Certificates []CertificateTemplate `json:"certificates"`
	Meta         *ListMeta             `json:"meta"`
}

func tableFleetdmCertificateTemplate(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "fleetdm_certificate_template",
		Description: "Certificates Fleet requests from a certificate authority for each team's hosts, across all teams.",
		List: &plugin.ListConfig{
			Hydrate: listCertificateTemplates,
			KeyColumns: []*plugin.KeyColumn{
				{Name: "team_id", Require: plugin.Optional}, // Only this team's certificates; 0 for no team
			},
		},
		Columns: []*plugin.Column{
			{Name: "id", Type: proto.ColumnType_INT, Description: "Unique ID of the certificate template."},
			{Name: "name", Type: proto.ColumnType_STRING, Description: "Name of the certificate template."},
			{Name: "subject_name", Type: proto.ColumnType_STRING, Description: "Subject name of requested certificates."},
			{Name: "subject_alternative_name", Type: proto.ColumnType_STRING, Description: "Subject alternative name of requested certificates."},
			{Name: "certificate_authority_id", Type: proto.ColumnType_INT, Transform: transform.FromField("CertificateAuthorityID"), Description: "ID of the certificate authority that issues the certificates."},
			{Name: "certificate_authority_name", Type: proto.ColumnType_STRING, Description: "Name of the certificate authority that issues the certificates."},
			{Name: "team_id", Type: proto.ColumnType_INT, Description: "ID of the team (fleet) the certificate belongs to; 0 for hosts with no team."},
			{Name: "created_at", Type: proto.ColumnType_TIMESTAMP, Transform: transform.FromField("CreatedAt").Transform(flexibleTimeTransform), Description: "When the certificate template was created."},
		},
	}
}

func listCertificateTemplates(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := getClient(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("fleetdm_certificate_template.listCertificateTemplates", "connection_error", err)
		return nil, err
	}
	scopes, err := teamScopes(ctx, d, client, "fleetdm_certificate_template")
	if err != nil {
		return nil, err
	}
	for _, scope := range scopes {
		teamID := uint(0)
		if scope != nil {
			teamID = *scope
		}
		err := paginatedList(ctx, d, 100, func(ctx context.Context, page, perPage int) ([]CertificateTemplate, *ListMeta, error) {
			params := url.Values{}
			params.Add("page", strconv.Itoa(page))
			params.Add("per_page", strconv.Itoa(perPage))
			client.addFleetIDParam(ctx, d, params, scopeParam(scope))
			var response ListCertificateTemplatesResponse
			if err := client.Get(ctx, "certificates", params, &response); err != nil {
				plugin.Logger(ctx).Error("fleetdm_certificate_template.listCertificateTemplates", "api_error", err, "page", page)
				return nil, nil, err
			}
			for i := range response.Certificates {
				response.Certificates[i].TeamID = &teamID
			}
			return response.Certificates, response.Meta, nil
		})
		if err != nil {
			if handleUnavailableFeature(ctx, d, "fleetdm_certificate_template", err) {
				return nil, nil
			}
			if skipScope(ctx, d, "fleetdm_certificate_template", scope, err) {
				continue
			}
			return nil, err
		}
		if d.RowsRemaining(ctx) == 0 {
			break
		}
	}
	return nil, nil
}
