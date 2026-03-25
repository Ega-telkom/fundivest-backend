// internal/worker/processor.go
package worker

import (
	"context"
	// "time"

	"github.com/Ega-telkom/fundivest-backend/internal/domain"
	"github.com/Ega-telkom/fundivest-backend/internal/pubsub"

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
	pubsub		pubsub.PubSub
	logger   *zap.Logger
}

func NewProcessor(
	certRepo CertificateRepository,
	pdfGen PDFGenerator,
	storage FileStorage,
	tmpl TemplateRenderer,
	pubsub		pubsub.PubSub,
	logger *zap.Logger,
) *Processor {
	return &Processor{
		certRepo: certRepo,
		pdfGen:   pdfGen,
		storage:  storage,
		tmpl:     tmpl,
		pubsub: pubsub,
		logger:   logger,
	}
}

func (p *Processor) Process(ctx context.Context, certID string) error {

	p.logger.Info("Processing certificate", zap.String("cert_id", certID))

	cert, err := p.certRepo.GetByID(ctx, certID)
	if err != nil {
		return err
	}

	// Idempotency guard
	if cert.Status == domain.StatusDone {
		p.logger.Info("Certificate already processed", zap.String("cert_id", certID))
		return nil
	}

	// Move to processing state
	if err := p.certRepo.UpdateStatus(ctx, certID, domain.StatusProcessing); err != nil {
		return err
	}

	p.pubsub.Publish(ctx, "cert:"+certID, "processing")

	// time.Sleep(20 * time.Second)

	// Generate HTML
	html, err := p.tmpl.Render(cert)
	if err != nil {
		p.fail(ctx, certID, err, "render template")
		return err
	}

	// Generate PDF
	pdfData, err := p.pdfGen.Generate(ctx, html)
	if err != nil {
		p.fail(ctx, certID, err, "generate pdf")
		return err
	}

	// Save file (idempotent overwrite)
	filename := certID + ".pdf"

	if err := p.storage.Save(ctx, filename, pdfData); err != nil {
		p.fail(ctx, certID, err, "save file")
		return err
	}

	// Update DB
	if err := p.certRepo.UpdatePDFPath(ctx, certID, filename); err != nil {
		return err
	}

	if err := p.certRepo.UpdateStatus(ctx, certID, domain.StatusDone); err != nil {
		return err
	}

	p.pubsub.Publish(ctx, "cert:"+certID, "done")

	p.logger.Info("Certificate generated successfully", zap.String("cert_id", certID))

	return nil
}

func (p *Processor) fail(ctx context.Context, certID string, err error, step string) {

	p.logger.Error(
		"Certificate processing failed",
		zap.String("cert_id", certID),
		zap.String("step", step),
		zap.Error(err),
	)

	p.certRepo.UpdateStatus(ctx, certID, domain.StatusFailed)

	p.pubsub.Publish(ctx, "cert:"+certID, "failed")
}
