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

// certificateName is a certificate subject or issuer.
type certificateName struct {
	CommonName         string `json:"common_name"`
	Country            string `json:"country"`
	Organization       string `json:"organization"`
	OrganizationalUnit string `json:"organizational_unit"`
}

// HostCertificate is a certificate found in a host's keychain or store.
// Refer to: https://fleetdm.com/docs/rest-api/rest-api#list-hosts-certificates
type HostCertificate struct {
	ID                   uint             `json:"id"`
	CommonName           string           `json:"common_name"`
	Serial               string           `json:"serial"`
	NotValidBefore       FleetTime        `json:"not_valid_before"`
	NotValidAfter        FleetTime        `json:"not_valid_after"`
	CertificateAuthority bool             `json:"certificate_authority"`
	KeyAlgorithm         string           `json:"key_algorithm"`
	KeyStrength          int              `json:"key_strength"`
	KeyUsage             string           `json:"key_usage"`
	SigningAlgorithm     string           `json:"signing_algorithm"`
	Source               string           `json:"source"`
	Username             string           `json:"username"`
	Subject              *certificateName `json:"subject"`
	Issuer               *certificateName `json:"issuer"`
	HostID               uint             `json:"-"`
	Hostname             string           `json:"-"`
}

// ListHostCertificatesResponse for `GET /api/v1/fleet/hosts/{id}/certificates`.
type ListHostCertificatesResponse struct {
	Certificates []HostCertificate `json:"certificates"`
	Meta         *ListMeta         `json:"meta"`
}

func tableFleetdmHostCertificate(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "fleetdm_host_certificate",
		Description: "Certificates installed on each host (macOS keychains, Windows stores, ...). Without a host_id filter every host is queried.",
		List: &plugin.ListConfig{
			Hydrate: listHostCertificates,
			KeyColumns: []*plugin.KeyColumn{
				{Name: "host_id", Require: plugin.Optional},
			},
		},
		Columns: []*plugin.Column{
			{Name: "host_id", Type: proto.ColumnType_INT, Description: "ID of the host."},
			{Name: "hostname", Type: proto.ColumnType_STRING, Description: "Hostname of the host. Empty when filtering by host_id."},
			{Name: "id", Type: proto.ColumnType_INT, Description: "Fleet's ID for this certificate record."},
			{Name: "common_name", Type: proto.ColumnType_STRING, Description: "Common name of the certificate."},
			{Name: "subject_common_name", Type: proto.ColumnType_STRING, Transform: transform.FromField("Subject.CommonName"), Description: "Common name of the subject."},
			{Name: "issuer_common_name", Type: proto.ColumnType_STRING, Transform: transform.FromField("Issuer.CommonName"), Description: "Common name of the issuer."},
			{Name: "not_valid_before", Type: proto.ColumnType_TIMESTAMP, Transform: transform.FromField("NotValidBefore").Transform(flexibleTimeTransform), Description: "Start of the validity period."},
			{Name: "not_valid_after", Type: proto.ColumnType_TIMESTAMP, Transform: transform.FromField("NotValidAfter").Transform(flexibleTimeTransform), Description: "End of the validity period (expiry)."},
			{Name: "certificate_authority", Type: proto.ColumnType_BOOL, Description: "Whether the certificate is a CA certificate."},
			{Name: "serial", Type: proto.ColumnType_STRING, Description: "Serial number."},
			{Name: "key_algorithm", Type: proto.ColumnType_STRING, Description: "Key algorithm, e.g. 'rsaEncryption'."},
			{Name: "key_strength", Type: proto.ColumnType_INT, Description: "Key size in bits."},
			{Name: "key_usage", Type: proto.ColumnType_STRING, Description: "Key usage extensions."},
			{Name: "signing_algorithm", Type: proto.ColumnType_STRING, Description: "Signature algorithm, e.g. 'sha256WithRSAEncryption'."},
			{Name: "source", Type: proto.ColumnType_STRING, Description: "Where the certificate was found: 'system' or 'user'."},
			{Name: "username", Type: proto.ColumnType_STRING, Description: "User whose store holds the certificate, for user certificates."},
			{Name: "subject", Type: proto.ColumnType_JSON, Description: "Full subject (common name, country, organization, unit)."},
			{Name: "issuer", Type: proto.ColumnType_JSON, Description: "Full issuer (common name, country, organization, unit)."},
		},
	}
}

func listHostCertificates(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := getClient(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("fleetdm_host_certificate.listHostCertificates", "connection_error", err)
		return nil, err
	}
	err = streamPerHost(ctx, d, client, func(ctx context.Context, host hostRef) ([]HostCertificate, error) {
		var rows []HostCertificate
		err := paginateCore(ctx, 100, func(ctx context.Context, page, perPage int) ([]HostCertificate, *ListMeta, error) {
			params := url.Values{}
			params.Add("page", strconv.Itoa(page))
			params.Add("per_page", strconv.Itoa(perPage))
			var response ListHostCertificatesResponse
			if err := client.Get(ctx, fmt.Sprintf("hosts/%d/certificates", host.ID), params, &response); err != nil {
				return nil, nil, err
			}
			return response.Certificates, response.Meta, nil
		}, func(c HostCertificate) bool {
			c.HostID, c.Hostname = host.ID, host.Hostname
			rows = append(rows, c)
			return true
		})
		return rows, err
	})
	if err != nil {
		plugin.Logger(ctx).Error("fleetdm_host_certificate.listHostCertificates", "api_error", err)
	}
	return nil, err
}
