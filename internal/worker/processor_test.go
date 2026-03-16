// internal/worker/processor_test.go
package worker_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Ega-telkom/fundivest-backend/internal/domain"
	"github.com/Ega-telkom/fundivest-backend/internal/worker"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type MockCertRepo struct {
    mock.Mock
}

func (m *MockCertRepo) GetByID(ctx context.Context, id string) (*domain.Certificate, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*domain.Certificate), args.Error(1)
}

func (m *MockCertRepo) UpdatePDFPath(ctx context.Context, id, path string) error {
    args := m.Called(ctx, id, path)
    return args.Error(0)
}

func (m *MockCertRepo) UpdateStatus(ctx context.Context, id string, status domain.CertStatus) error {
    args := m.Called(ctx, id, status)
    return args.Error(0)
}

type MockPDFGen struct {
    mock.Mock
}

func (m *MockPDFGen) Generate(ctx context.Context, html string) ([]byte, error) {
    args := m.Called(ctx, html)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).([]byte), args.Error(1)
}

type MockStorage struct {
    mock.Mock
}

func (m *MockStorage) Save(ctx context.Context, filename string, data []byte) error {
    args := m.Called(ctx, filename, data)
    return args.Error(0)
}

type MockTemplate struct {
    mock.Mock
}

func (m *MockTemplate) Render(cert *domain.Certificate) (string, error) {
    args := m.Called(cert)
    return args.String(0), args.Error(1)
}

func TestProcessor_Process(t *testing.T) {
    t.Run("success", func(t *testing.T) {
        certRepo := new(MockCertRepo)
        pdfGen := new(MockPDFGen)
        storage := new(MockStorage)
        tmpl := new(MockTemplate)
        logger := zap.NewNop()
        
        cert := &domain.Certificate{
            ID:       "cert-123",
            Name:     "John Doe",
            CourseID: "course-1",
            IssuedAt: time.Now(),
        }
        
        certRepo.On("GetByID", mock.Anything, "cert-123").Return(cert, nil)
        tmpl.On("Render", cert).Return("<html>certificate</html>", nil)
        pdfGen.On("Generate", mock.Anything, "<html>certificate</html>").Return([]byte("pdf"), nil)
        storage.On("Save", mock.Anything, "cert-123.pdf", []byte("pdf")).Return(nil)
        certRepo.On("UpdatePDFPath", mock.Anything, "cert-123", "cert-123.pdf").Return(nil)
        
        processor := worker.NewProcessor(certRepo, pdfGen, storage, tmpl, logger)
        
        err := processor.Process(context.Background(), "cert-123")
        
        assert.NoError(t, err)
        certRepo.AssertExpectations(t)
        pdfGen.AssertExpectations(t)
        storage.AssertExpectations(t)
        tmpl.AssertExpectations(t)
    })
    
    t.Run("error - pdf generation failed", func(t *testing.T) {
        certRepo := new(MockCertRepo)
        pdfGen := new(MockPDFGen)
        storage := new(MockStorage)
        tmpl := new(MockTemplate)
        logger := zap.NewNop()
        
        cert := &domain.Certificate{ID: "cert-123"}
        
        certRepo.On("GetByID", mock.Anything, "cert-123").Return(cert, nil)
        tmpl.On("Render", cert).Return("<html>cert</html>", nil)
        pdfGen.On("Generate", mock.Anything, "<html>cert</html>").Return(nil, errors.New("gotenberg error"))
        certRepo.On("UpdateStatus", mock.Anything, "cert-123", domain.StatusFailed).Return(nil)
        
        processor := worker.NewProcessor(certRepo, pdfGen, storage, tmpl, logger)
        
        err := processor.Process(context.Background(), "cert-123")
        
        assert.Error(t, err)
        certRepo.AssertExpectations(t)
        pdfGen.AssertExpectations(t)
        tmpl.AssertExpectations(t)
    })
}