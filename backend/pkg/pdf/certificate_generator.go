package pdf

import (
	"fmt"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/code"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/consts/orientation"
	"github.com/johnfercher/maroto/v2/pkg/consts/pagesize"
	"github.com/johnfercher/maroto/v2/pkg/props"
	"github.com/secusense/backend/internal/domain"
)

type CertificateGenerator struct {
	verificationBaseURL string
}

func NewCertificateGenerator(verificationBaseURL string) *CertificateGenerator {
	return &CertificateGenerator{
		verificationBaseURL: verificationBaseURL,
	}
}

func (g *CertificateGenerator) Generate(cert *domain.Certificate) ([]byte, error) {
	cfg := config.NewBuilder().
		WithOrientation(orientation.Horizontal).
		WithPageSize(pagesize.A4).
		WithLeftMargin(20).
		WithTopMargin(20).
		WithRightMargin(20).
		Build()

	m := maroto.New(cfg)

	// Header
	m.AddRow(20,
		col.New(12).Add(
			text.New("CERTIFICATE OF COMPLETION", props.Text{
				Top:   5,
				Style: fontstyle.Bold,
				Size:  28,
				Align: align.Center,
				Color: &props.Color{Red: 0, Green: 51, Blue: 102},
			}),
		),
	)

	m.AddRow(10,
		col.New(12).Add(
			text.New("SecuSense Training Portal", props.Text{
				Top:   2,
				Style: fontstyle.Normal,
				Size:  14,
				Align: align.Center,
				Color: &props.Color{Red: 100, Green: 100, Blue: 100},
			}),
		),
	)

	// Divider
	m.AddRow(15)

	// This is to certify
	m.AddRow(10,
		col.New(12).Add(
			text.New("This is to certify that", props.Text{
				Top:   2,
				Size:  12,
				Align: align.Center,
			}),
		),
	)

	// Name
	holderName := fmt.Sprintf("%s %s", cert.UserFirstName, cert.UserLastName)
	m.AddRow(15,
		col.New(12).Add(
			text.New(holderName, props.Text{
				Top:   2,
				Style: fontstyle.Bold,
				Size:  24,
				Align: align.Center,
				Color: &props.Color{Red: 0, Green: 51, Blue: 102},
			}),
		),
	)

	// Has successfully completed
	m.AddRow(10,
		col.New(12).Add(
			text.New("has successfully completed the course", props.Text{
				Top:   2,
				Size:  12,
				Align: align.Center,
			}),
		),
	)

	// Course title
	m.AddRow(15,
		col.New(12).Add(
			text.New(cert.CourseTitle, props.Text{
				Top:   2,
				Style: fontstyle.BoldItalic,
				Size:  18,
				Align: align.Center,
			}),
		),
	)

	// Score
	scoreText := fmt.Sprintf("with a score of %d out of %d", cert.Score, cert.MaxScore)
	m.AddRow(10,
		col.New(12).Add(
			text.New(scoreText, props.Text{
				Top:   2,
				Size:  12,
				Align: align.Center,
			}),
		),
	)

	// Date
	m.AddRow(15,
		col.New(12).Add(
			text.New(fmt.Sprintf("Issued on %s", cert.IssuedAt.Format("January 2, 2006")), props.Text{
				Top:   5,
				Size:  11,
				Align: align.Center,
				Color: &props.Color{Red: 80, Green: 80, Blue: 80},
			}),
		),
	)

	// Certificate number
	m.AddRow(10,
		col.New(12).Add(
			text.New(fmt.Sprintf("Certificate Number: %s", cert.CertificateNumber), props.Text{
				Top:   2,
				Size:  10,
				Align: align.Center,
				Color: &props.Color{Red: 100, Green: 100, Blue: 100},
			}),
		),
	)

	// Verification QR code and link
	verificationURL := fmt.Sprintf("%s/verify/%s", g.verificationBaseURL, cert.VerificationHash)
	m.AddRow(35,
		col.New(4),
		col.New(4).Add(
			code.NewQr(verificationURL, props.Rect{
				Center:  true,
				Percent: 100,
			}),
		),
		col.New(4),
	)

	m.AddRow(8,
		col.New(12).Add(
			text.New("Scan to verify or visit:", props.Text{
				Top:   1,
				Size:  8,
				Align: align.Center,
				Color: &props.Color{Red: 120, Green: 120, Blue: 120},
			}),
		),
	)

	m.AddRow(6,
		col.New(12).Add(
			text.New(verificationURL, props.Text{
				Top:   0,
				Size:  8,
				Align: align.Center,
				Color: &props.Color{Red: 0, Green: 102, Blue: 204},
			}),
		),
	)

	// Generate PDF
	doc, err := m.Generate()
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return doc.GetBytes(), nil
}
