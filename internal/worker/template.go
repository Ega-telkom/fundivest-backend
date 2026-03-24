// internal/worker/template.go
package worker

import (
	"bytes"
	"fmt"
	"html/template"
	"time"

	"github.com/Ega-telkom/fundivest-backend/internal/domain"
	"github.com/Ega-telkom/fundivest-backend/internal/pkg/qrcode"
	"go.uber.org/zap"
)

type HTMLTemplateRenderer struct {
	tmpl        *template.Template
	frontendURL string
	logger		*zap.Logger
}

func NewHTMLTemplateRenderer(tmplPath string, frontendURL string, logger *zap.Logger) (*HTMLTemplateRenderer, error) {
	logger.Info("Initializing template renderer",
		zap.String("template_path", tmplPath),
		zap.String("frontend_url", frontendURL),
	)
    funcMap := template.FuncMap{
        "safeURL": func(s string) template.URL {
            return template.URL(s)
        },
    }
    
    tmpl, err := template.New("certificate.html").
        Funcs(funcMap).
        ParseFiles(tmplPath)
    
    if err != nil {
   		logger.Error("Failed to parse template", 
     		zap.String("template_path", tmplPath),
     		zap.Error(err),
     	)
        return nil, fmt.Errorf("parse template: %w", err)
    }
    
    logger.Info("Template renderer initialized successfully")
    
    return &HTMLTemplateRenderer{
        tmpl:        tmpl,
        frontendURL: frontendURL,
        logger: logger,
    }, nil
}

func formatTanggalID(t time.Time) string {
    bulan := []string{
        "", "Januari", "Februari", "Maret", "April", "Mei", "Juni",
        "Juli", "Agustus", "September", "Oktober", "November", "Desember",
    }
    return fmt.Sprintf("%02d %s %d", t.Day(), bulan[t.Month()], t.Year())
}

func (t *HTMLTemplateRenderer) Render(cert *domain.Certificate) (string, error) {
	t.logger.Debug("Rendering certificate template",
		zap.String("cert_id", cert.ID),
		zap.String("name", cert.Name),
	)
	
	verifyURL := fmt.Sprintf("%s/verify/%s", t.frontendURL, cert.ID)
	qrData, err := qrcode.Generate(verifyURL)
	if err != nil {
		t.logger.Error("Failed to generate qr code", 
			zap.String("cert_id", cert.ID),
			zap.String("verify_url", verifyURL),
			zap.Error(err),
		)
		return "", fmt.Errorf("generate qr code: %w", err)
	}

	data := map[string]interface{}{
		"Name":      cert.Name,
		"CourseID":  cert.CourseID,
		"IssuedAt":  formatTanggalID(cert.IssuedAt),
		"QRCode":    template.URL(qrData),
		"VerifyURL": verifyURL,
	}

	// Execute template
	var buf bytes.Buffer
	if err := t.tmpl.Execute(&buf, data); err != nil {
		t.logger.Error("Failed to generate from template", 
			zap.String("cert_id", cert.ID),
			zap.String("name", cert.Name),
			zap.Error(err),
		)
		return "", fmt.Errorf("execute template: %w", err)
	}
	
	htmlSize := buf.Len()
	
	t.logger.Info("Template rendered successfully",
		zap.String("cert_id", cert.ID),
		zap.String("verify_url", verifyURL),
		zap.Int("html_size", htmlSize),
	)

	return buf.String(), nil
}
