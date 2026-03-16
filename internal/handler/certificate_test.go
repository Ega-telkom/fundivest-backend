// internal/handler/certificate_test.go
package handler_test

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Ega-telkom/fundivest-backend/internal/domain"
	"github.com/Ega-telkom/fundivest-backend/internal/handler"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockCertService struct {
    mock.Mock
}

func (m *MockCertService) RequestCertificate(ctx context.Context, sessionID string) (string, error) {
    args := m.Called(ctx, sessionID)
    return args.String(0), args.Error(1)
}

// FIX: Return domain.CertStatus, bukan string
func (m *MockCertService) GetStatus(ctx context.Context, certID string) (domain.CertStatus, error) {
    args := m.Called(ctx, certID)
    // Cast ke domain.CertStatus
    if args.Get(0) == nil {
        return "", args.Error(1)
    }
    return args.Get(0).(domain.CertStatus), args.Error(1)
}

func (m *MockCertService) GetCertificate(ctx context.Context, certID string) (*domain.Certificate, error) {
    args := m.Called(ctx, certID)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*domain.Certificate), args.Error(1)
}

func (m *MockCertService) DownloadCertificate(ctx context.Context, certID string) ([]byte, error) {
    args := m.Called(ctx, certID)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).([]byte), args.Error(1)
}

func TestCertificateHandler_RequestCertificate(t *testing.T) {
    t.Run("success", func(t *testing.T) {
        svc := new(MockCertService)
        h := handler.NewCertificateHandler(svc)
        
        app := fiber.New()
        app.Post("/api/sessions/:id/certificate", h.RequestCertificate)
        
        svc.On("RequestCertificate", mock.Anything, "sess-123").Return("cert-456", nil)
        
        req := httptest.NewRequest("POST", "/api/sessions/sess-123/certificate", nil)
        
        resp, _ := app.Test(req)
        
        assert.Equal(t, fiber.StatusAccepted, resp.StatusCode)
        
        var body map[string]interface{}
        err := json.NewDecoder(resp.Body).Decode(&body)
        require.NoError(t, err, "failed to decode response")
        
        assert.Equal(t, true, body["success"])
        data := body["data"].(map[string]interface{})
        assert.Equal(t, "cert-456", data["certificate_id"])
        assert.Equal(t, string(domain.StatusPending), data["status"])
        
        svc.AssertExpectations(t)
    })
    
    t.Run("error:not all chapters completed", func(t *testing.T) {
        svc := new(MockCertService)
        h := handler.NewCertificateHandler(svc)
        
        app := fiber.New()
        app.Post("/api/sessions/:id/certificate", h.RequestCertificate)
        
        svc.On("RequestCertificate", mock.Anything, "sess-123").
            Return("", domain.ErrNotAllChaptersCompleted)
        
        req := httptest.NewRequest("POST", "/api/sessions/sess-123/certificate", nil)
        
        resp, _ := app.Test(req)
        
        assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
        
        var body map[string]interface{}
        err := json.NewDecoder(resp.Body).Decode(&body)
        require.NoError(t, err, "failed to decode response")
        
        assert.Equal(t, false, body["success"])
        assert.Contains(t, body["error"], "not all chapters completed")
        
        svc.AssertExpectations(t)
    })
}

func TestCertificateHandler_GetStatus(t *testing.T) {
    t.Run("success:pending", func(t *testing.T) {
        svc := new(MockCertService)
        h := handler.NewCertificateHandler(svc)
        
        app := fiber.New()
        app.Get("/api/certificates/:certificate_id/status", h.GetStatus)
        
        // FIX: Pass domain.CertStatus type
        svc.On("GetStatus", mock.Anything, "cert-123").
            Return(domain.StatusPending, nil)
        
        req := httptest.NewRequest("GET", "/api/certificates/cert-123/status", nil)
        
        resp, _ := app.Test(req)
        
        assert.Equal(t, fiber.StatusOK, resp.StatusCode)
        
        var body map[string]interface{}
        err := json.NewDecoder(resp.Body).Decode(&body)
        require.NoError(t, err, "failed to decode response")
        
        assert.Equal(t, true, body["success"])
        data := body["data"].(map[string]interface{})
        assert.Equal(t, string(domain.StatusPending), data["status"])
        
        svc.AssertExpectations(t)
    })
    
    t.Run("error:not found", func(t *testing.T) {
        svc := new(MockCertService)
        h := handler.NewCertificateHandler(svc)
        
        app := fiber.New()
        app.Get("/api/certificates/:certificate_id/status", h.GetStatus)
        
        // FIX: Return empty CertStatus on error
        svc.On("GetStatus", mock.Anything, "invalid").
            Return(domain.CertStatus(""), domain.ErrCertificateNotFound)
        
        req := httptest.NewRequest("GET", "/api/certificates/invalid/status", nil)
        
        resp, _ := app.Test(req)
        
        assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
        
        svc.AssertExpectations(t)
    })
}

func TestCertificateHandler_Download(t *testing.T) {
    t.Run("success", func(t *testing.T) {
        svc := new(MockCertService)
        h := handler.NewCertificateHandler(svc)
        
        app := fiber.New()
        app.Get("/api/certificates/:certificate_id/download", h.Download)
        
        svc.On("DownloadCertificate", mock.Anything, "cert-123").
            Return([]byte("pdf content"), nil)
        
        req := httptest.NewRequest("GET", "/api/certificates/cert-123/download", nil)
        
        resp, _ := app.Test(req)
        
        assert.Equal(t, fiber.StatusOK, resp.StatusCode)
        assert.Equal(t, "application/pdf", resp.Header.Get("Content-Type"))
        
        svc.AssertExpectations(t)
    })
}

func TestCertificateHandler_Verify(t *testing.T) {
    t.Run("success", func(t *testing.T) {
        svc := new(MockCertService)
        h := handler.NewCertificateHandler(svc)
        
        app := fiber.New()
        app.Get("/verify/:certificate_id", h.Verify)
        
        cert := &domain.Certificate{
            ID:       "cert-123",
            Name:     "John Doe",
            CourseID: "course-1",
            IssuedAt: time.Now(),
            Status:   domain.StatusDone,
        }
        
        svc.On("GetCertificate", mock.Anything, "cert-123").Return(cert, nil)
        
        req := httptest.NewRequest("GET", "/verify/cert-123", nil)
        
        resp, _ := app.Test(req)
        
        assert.Equal(t, fiber.StatusOK, resp.StatusCode)
        
        var body map[string]interface{}
        err := json.NewDecoder(resp.Body).Decode(&body)
        require.NoError(t, err, "failed to decode response")
        
        assert.Equal(t, true, body["success"])
        data := body["data"].(map[string]interface{})
        assert.Equal(t, "cert-123", data["id"])
        assert.Equal(t, "John Doe", data["name"])
        assert.Equal(t, true, data["verified"])
        
        svc.AssertExpectations(t)
    })
}