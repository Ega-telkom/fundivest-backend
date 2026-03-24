// internal/worker/pdf_generator.go
package worker

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type GotenbergClient struct {
    baseURL string
    client  *http.Client
    logger *zap.Logger
}

func NewGotenbergClient(baseURL string, logger *zap.Logger) *GotenbergClient {
    return &GotenbergClient{
        baseURL: baseURL,
        client: &http.Client{
            Timeout: 30 * time.Second,
        },
        logger: logger,
    }
}

func (g *GotenbergClient) Generate(ctx context.Context, html string) ([]byte, error) {
	start := time.Now()
	
	g.logger.Info("Generating PDF",
		zap.Int("html_size", len(html)),
	)
	
    var buf bytes.Buffer
    writer := multipart.NewWriter(&buf)
    
    // Add HTML file
    part, err := writer.CreateFormFile("files", "index.html")
    if err != nil {
    	g.logger.Error("Failed to create form file",
        	zap.String("filename", "index.html"),
        	zap.Error(err),
       	)
    	return nil, fmt.Errorf("create form file: %w", err) 
    }
    
    if _, err := io.WriteString(part, html); err != nil {
    	g.logger.Error("Failed to write HTML to form",
        	zap.Int("html_size", len(html)),
         zap.Error(err),
     	)
     	return nil, fmt.Errorf("write html to form: %w", err)
    }
    
    _ = writer.WriteField("preferCssPageSize", "true")
    _ = writer.WriteField("printBackground", "true")
    
    _ = writer.Close()
    
    req, err := http.NewRequestWithContext(
        ctx,
        "POST",
        g.baseURL+"/forms/chromium/convert/html",
        &buf,
    )
    if err != nil {
    	g.logger.Error("Failed to create PDF request", zap.Error(err))
     	return nil, fmt.Errorf("create request: %w", err)
    }
    
    req.Header.Set("Content-Type", writer.FormDataContentType())
    g.logger.Debug("Sending request to Gotenberg",
        zap.String("url", g.baseURL),
    )
    
    resp, err := g.client.Do(req)
    if err != nil {
    	g.logger.Error("Failed to request to Gotenberg",
        	zap.String("url", g.baseURL),
        	zap.Error(err),
       	)
    	return nil, fmt.Errorf("gotenberg request: %w", err)
    }
    
    defer func() { _ = resp.Body.Close() } ()
    
    if resp.StatusCode != http.StatusOK {
    	g.logger.Error("Gotenberg returned error",
        	zap.Int("status_code", resp.StatusCode),
         	zap.String("status", resp.Status),
     	)
     	return nil, fmt.Errorf("gotenberg error: status %d", resp.StatusCode)
    }
    
    // Read response
    pdfData, err := io.ReadAll(resp.Body)
    if err != nil {
        g.logger.Error("Failed to read PDF response", zap.Error(err))
        return nil, fmt.Errorf("read response: %w", err)
    }
    
    duration := time.Since(start)
    
    g.logger.Info("PDF generated successfully",
        zap.Int("pdf_size", len(pdfData)),
        zap.Duration("duration", duration),
    )
    
    // Warn if slow
    if duration > 7*time.Second {
        g.logger.Warn("PDF generated but slow",
            zap.Duration("duration", duration),
            zap.Int("html_size", len(html)),
        )
    }
    
    return pdfData, nil
}