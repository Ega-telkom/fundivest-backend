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
)

type GotenbergClient struct {
    baseURL string
    client  *http.Client
}

func NewGotenbergClient(baseURL string) *GotenbergClient {
    return &GotenbergClient{
        baseURL: baseURL,
        client: &http.Client{
            Timeout: 30 * time.Second,
        },
    }
}

func (g *GotenbergClient) Generate(ctx context.Context, html string) ([]byte, error) {
    var buf bytes.Buffer
    writer := multipart.NewWriter(&buf)
    
    // Add HTML file
    part, err := writer.CreateFormFile("files", "index.html")
    if err != nil {
        return nil, err
    }
    
    if _, err := io.WriteString(part, html); err != nil {
        return nil, err
    }
    
    // Optional: Add custom options
    _ = writer.WriteField("marginTop", "0.5")
    _ = writer.WriteField("marginBottom", "0.5")
    _ = writer.WriteField("marginLeft", "0.5")
    _ = writer.WriteField("marginRight", "0.5")
    
    _ = writer.Close()
    
    req, err := http.NewRequestWithContext(
        ctx,
        "POST",
        g.baseURL+"/forms/chromium/convert/html",
        &buf,
    )
    if err != nil {
        return nil, err
    }
    
    req.Header.Set("Content-Type", writer.FormDataContentType())
    
    resp, err := g.client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("gotenberg request failed: %w", err)
    }
    
    defer func() { _ = resp.Body.Close() } ()
    
    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("gotenberg returned %d: %s", resp.StatusCode, string(body))
    }
    
    return io.ReadAll(resp.Body)
}