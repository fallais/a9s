package resources

import (
	"context"
	"fmt"
	"strings"

	"a9s/internal/client"

	"github.com/aws/aws-sdk-go-v2/service/acm"
)

// ACMCertificate represents an ACM certificate
type ACMCertificate struct {
	DomainName         string
	CertificateArn     string
	Status             string
	Type               string
	InUseBy            string
	NotBefore          string
	NotAfter           string
	RenewalEligibility string
}

// ACMCertificates implements Resource for ACM certificates
type ACMCertificates struct {
	certificates []ACMCertificate
}

// NewACMCertificates creates a new ACMCertificates resource
func NewACMCertificates() *ACMCertificates {
	return &ACMCertificates{
		certificates: make([]ACMCertificate, 0),
	}
}

// Name returns the display name
func (a *ACMCertificates) Name() string {
	return "ACM Certificates"
}

// Columns returns the column definitions
func (a *ACMCertificates) Columns() []Column {
	return []Column{
		{Name: "Domain Name", Width: 40},
		{Name: "Status", Width: 15},
		{Name: "Type", Width: 15},
		{Name: "In Use", Width: 8},
		{Name: "Not Before", Width: 12},
		{Name: "Not After", Width: 12},
		{Name: "Renewal", Width: 15},
	}
}

// Fetch retrieves ACM certificates from AWS
func (a *ACMCertificates) Fetch(ctx context.Context, c *client.Client) error {
	a.certificates = make([]ACMCertificate, 0)

	paginator := acm.NewListCertificatesPaginator(c.ACM(), &acm.ListCertificatesInput{})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to list ACM certificates: %w", err)
		}

		for _, cert := range output.CertificateSummaryList {
			// Get detailed information for each certificate
			describeOutput, err := c.ACM().DescribeCertificate(ctx, &acm.DescribeCertificateInput{
				CertificateArn: cert.CertificateArn,
			})
			if err != nil {
				continue // Skip certificates we can't describe
			}

			certDetail := describeOutput.Certificate
			certificate := ACMCertificate{
				DomainName:     stringValue(certDetail.DomainName),
				CertificateArn: stringValue(certDetail.CertificateArn),
				Status:         string(certDetail.Status),
				Type:           string(certDetail.Type),
				InUseBy:        fmt.Sprintf("%d", len(certDetail.InUseBy)),
			}

			if certDetail.NotBefore != nil {
				certificate.NotBefore = certDetail.NotBefore.Format("2006-01-02")
			}

			if certDetail.NotAfter != nil {
				certificate.NotAfter = certDetail.NotAfter.Format("2006-01-02")
			}

			if certDetail.RenewalEligibility != "" {
				certificate.RenewalEligibility = string(certDetail.RenewalEligibility)
			}

			a.certificates = append(a.certificates, certificate)
		}
	}

	return nil
}

// Rows returns the table data
func (a *ACMCertificates) Rows() [][]string {
	rows := make([][]string, len(a.certificates))
	for i, cert := range a.certificates {
		rows[i] = []string{
			cert.DomainName,
			cert.Status,
			formatCertType(cert.Type),
			cert.InUseBy,
			cert.NotBefore,
			cert.NotAfter,
			cert.RenewalEligibility,
		}
	}
	return rows
}

// GetID returns the certificate ARN at the given index
func (a *ACMCertificates) GetID(index int) string {
	if index >= 0 && index < len(a.certificates) {
		return a.certificates[index].CertificateArn
	}
	return ""
}

// formatCertType formats the certificate type for display
func formatCertType(certType string) string {
	switch certType {
	case "AMAZON_ISSUED":
		return "Amazon"
	case "IMPORTED":
		return "Imported"
	case "PRIVATE":
		return "Private"
	default:
		return strings.ReplaceAll(certType, "_", " ")
	}
}
