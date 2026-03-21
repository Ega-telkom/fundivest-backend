// internal/worker/processor.go
package worker

import (
	"context"
	"fmt"

	"github.com/Ega-telkom/fundivest-backend/internal/domain"

	"go.uber.org/zap"
)

type CertificateRepository interface {
	GetByID(ctx context.Context, id string) (*domain.Certificate, error)
	UpdatePDFPath(ctx context.Context, id, path string) error
	UpdateStatus(ctx context.Context, id string, status domain.CertStatus) error
}

type PDFGenerator interface {
	Generate(ctx context.Context, html string) ([]byte, error)
}

type FileStorage interface {
	Save(ctx context.Context, filename string, data []byte) error
}

type TemplateRenderer interface {
	Render(cert *domain.Certificate) (string, error)
}

type Processor struct {
	certRepo CertificateRepository
	pdfGen   PDFGenerator
	storage  FileStorage
	tmpl     TemplateRenderer
	logger   *zap.Logger
}

func NewProcessor(
	certRepo CertificateRepository,
	pdfGen PDFGenerator,
	storage FileStorage,
	tmpl TemplateRenderer,
	logger *zap.Logger,
) *Processor {
	return &Processor{
		certRepo: certRepo,
		pdfGen:   pdfGen,
		storage:  storage,
		tmpl:     tmpl,
		logger:   logger,
	}
}

func (p *Processor) Process(ctx context.Context, certID string) error {
	p.logger.Info("Processing certificate", zap.String("cert_id", certID))
	
	// 1. Get certificate
	cert, err := p.certRepo.GetByID(ctx, certID)
	if err != nil {
		p.logger.Error("Failed to get certificate", zap.String("cert_id", certID), zap.Error(err))
		return fmt.Errorf("get certificate: %w", err)
	}

	// 2. Generate HTML
	html, err := p.tmpl.Render(cert)
	if err != nil {
		p.logger.Error("Failed to render template", zap.String("cert_id", certID), zap.Error(err))
		if err := p.certRepo.UpdateStatus(ctx, certID, domain.StatusFailed); err != nil {
			p.logger.Error("Failed to update certificate status", zap.Error(err))
		}
		return fmt.Errorf("render template: %w", err)
	}

	// 3. Convert to PDF
	pdfData, err := p.pdfGen.Generate(ctx, html)
	if err != nil {
		p.logger.Error("Failed to generate PDF", zap.String("cert_id", certID), zap.Error(err))
		if err := p.certRepo.UpdateStatus(ctx, certID, domain.StatusFailed); err != nil {
			p.logger.Error("Failed to update certificate status", zap.Error(err))
		}
		return fmt.Errorf("generate pdf: %w", err)
	}

	// 4. Save file
	filename := fmt.Sprintf("%s.pdf", certID)
	if err := p.storage.Save(ctx, filename, pdfData); err != nil {
		p.logger.Error("Failed to save PDF", zap.String("cert_id", certID), zap.Error(err))
		if err := p.certRepo.UpdateStatus(ctx, certID, domain.StatusFailed); err != nil {
			p.logger.Error("Failed to update certificate status", zap.Error(err))
		}
		return fmt.Errorf("save file: %w", err)
	}

	// 5. Update database
	if err := p.certRepo.UpdatePDFPath(ctx, certID, filename); err != nil {
		p.logger.Error("Failed to update PDF path", zap.String("cert_id", certID), zap.Error(err))
		return fmt.Errorf("update pdf path: %w", err)
	}
	
    p.logger.Info("Certificate generated successfully", zap.String("cert_id", certID))
	return nil
}
