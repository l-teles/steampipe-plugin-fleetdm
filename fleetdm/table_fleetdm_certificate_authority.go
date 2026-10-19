package fleetdm

import (
	"context"

	"github.com/turbot/steampipe-plugin-sdk/v6/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin"
)

// CertificateAuthority is a configured certificate authority integration
// (DigiCert, NDES, custom SCEP, Smallstep, ...). Only the list summary is
// read: the per-ID endpoint returns credentials.
// Refer to: https://fleetdm.com/docs/rest-api/rest-api#list-certificate-authorities-cas
type CertificateAuthority struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// ListCertificateAuthoritiesResponse for `GET /api/v1/fleet/certificate_authorities`
// (not paginated).
type ListCertificateAuthoritiesResponse struct {
	CertificateAuthorities []CertificateAuthority `json:"certificate_authorities"`
}

func tableFleetdmCertificateAuthority(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "fleetdm_certificate_authority",
		Description: "Certificate authorities integrated with FleetDM (Fleet Premium).",
		List: &plugin.ListConfig{
			Hydrate: listCertificateAuthorities,
		},
		Columns: []*plugin.Column{
			{Name: "id", Type: proto.ColumnType_INT, Description: "Unique ID of the certificate authority."},
			{Name: "name", Type: proto.ColumnType_STRING, Description: "Name of the certificate authority."},
			{Name: "type", Type: proto.ColumnType_STRING, Description: "Type of certificate authority, e.g. 'digicert', 'ndes_scep_proxy', 'custom_scep_proxy', 'smallstep'."},
		},
	}
}

func listCertificateAuthorities(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := getClient(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("fleetdm_certificate_authority.listCertificateAuthorities", "connection_error", err)
		return nil, err
	}
	var response ListCertificateAuthoritiesResponse
	if err := client.Get(ctx, "certificate_authorities", nil, &response); err != nil {
		if handleUnavailableFeature(ctx, d, "fleetdm_certificate_authority", err) {
			return nil, nil
		}
		plugin.Logger(ctx).Error("fleetdm_certificate_authority.listCertificateAuthorities", "api_error", err)
		return nil, err
	}
	for _, ca := range response.CertificateAuthorities {
		d.StreamListItem(ctx, ca)
		if d.RowsRemaining(ctx) == 0 {
			break
		}
	}
	return nil, nil
}
