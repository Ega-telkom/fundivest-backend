// internal/worker/template.go
package worker

import (
	"bytes"
	"fmt"
	"html/template"

	"fundivest/internal/domain"
	"fundivest/internal/pkg/qrcode"
)

type HTMLTemplateRenderer struct {
	tmpl        *template.Template
	frontendURL string
}

func NewHTMLTemplateRenderer(tmplPath string, frontendURL string) (*HTMLTemplateRenderer, error) {
    funcMap := template.FuncMap{
        "safeURL": func(s string) template.URL {
            return template.URL(s)
        },
    }
    
    tmpl, err := template.New("certificate.html").
        Funcs(funcMap).
        ParseFiles(tmplPath)
    
    if err != nil {
        return nil, fmt.Errorf("parse template: %w", err)
    }
    
    return &HTMLTemplateRenderer{
        tmpl:        tmpl,
        frontendURL: frontendURL,
    }, nil
}

func (t *HTMLTemplateRenderer) Render(cert *domain.Certificate) (string, error) {
	verifyURL := fmt.Sprintf("%s/verify/%s", t.frontendURL, cert.ID)
	qrData, err := qrcode.Generate(verifyURL)
	if err != nil {
		return "", fmt.Errorf("generate qr code: %w", err)
	}

	data := map[string]interface{}{
		"Name":      cert.Name,
		"CourseID":  cert.CourseID,
		"IssuedAt":  cert.IssuedAt.Format("02 January 2006"),
		"QRCode":    template.URL(qrData),
		"VerifyURL": verifyURL,
	}

	var buf bytes.Buffer
	if err := t.tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}

	return buf.String(), nil
}
