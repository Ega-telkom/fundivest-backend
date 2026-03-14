// internal/service/cert_svc_test.go
package service_test

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "fundivest/internal/domain"
    "fundivest/internal/service"
)

func TestCertificateService_RequestCertificate(t *testing.T) {
    t.Run("success:all chapters completed", func(t *testing.T) {
        certRepo := new(MockCertRepo)
        sessionRepo := new(MockSessionRepo)
        queue := new(MockQueue)
        storage := new(MockStorage)
        
        sessionRepo.On("Get", mock.Anything, "sess-123").Return(&domain.Session{
            ID:              "sess-123",
            Name:            "John Doe",
            CourseID:        "course-1",
            ChaptersCompleted: []int{1, 2, 3},
            Completed:       false,
        }, nil)
        
        certRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Certificate")).Return(nil)
        sessionRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Session")).Return(nil)
        queue.On("Publish", mock.Anything, mock.AnythingOfType("string")).Return(nil)
        
        svc := service.NewCertificateService(certRepo, sessionRepo, queue, storage)
        
        certID, err := svc.RequestCertificate(context.Background(), "sess-123")
        
        assert.NoError(t, err)
        assert.NotEmpty(t, certID)
        
        certRepo.AssertExpectations(t)
        sessionRepo.AssertExpectations(t)
        queue.AssertExpectations(t)
    })
    
    t.Run("error:not all chapters completed", func(t *testing.T) {
        certRepo := new(MockCertRepo)
        sessionRepo := new(MockSessionRepo)
        queue := new(MockQueue)
        storage := new(MockStorage)
        
        sessionRepo.On("Get", mock.Anything, "sess-123").Return(&domain.Session{
            ChaptersCompleted: []int{1, 2},
            Completed:       false,
        }, nil)
        
        svc := service.NewCertificateService(certRepo, sessionRepo, queue, storage)
        
        _, err := svc.RequestCertificate(context.Background(), "sess-123")
        
        assert.ErrorIs(t, err, domain.ErrNotAllChaptersCompleted)
        sessionRepo.AssertExpectations(t)
    })
    
    t.Run("error:session not found", func(t *testing.T) {
        certRepo := new(MockCertRepo)
        sessionRepo := new(MockSessionRepo)
        queue := new(MockQueue)
        storage := new(MockStorage)
        
        sessionRepo.On("Get", mock.Anything, "invalid").Return(nil, domain.ErrSessionNotFound)
        
        svc := service.NewCertificateService(certRepo, sessionRepo, queue, storage)
        
        _, err := svc.RequestCertificate(context.Background(), "invalid")
        
        assert.ErrorIs(t, err, domain.ErrSessionNotFound)
        sessionRepo.AssertExpectations(t)
    })
}

func TestCertificateService_GetStatus(t *testing.T) {
    t.Run("success", func(t *testing.T) {
        certRepo := new(MockCertRepo)
        sessionRepo := new(MockSessionRepo)
        queue := new(MockQueue)
        storage := new(MockStorage)
        
        certRepo.On("GetByID", mock.Anything, "cert-123").Return(&domain.Certificate{
            ID:     "cert-123",
            Status: domain.StatusDone,
        }, nil)
        
        svc := service.NewCertificateService(certRepo, sessionRepo, queue, storage)
        
        status, err := svc.GetStatus(context.Background(), "cert-123")
        
        assert.NoError(t, err)
        assert.Equal(t, domain.StatusDone, status)
        certRepo.AssertExpectations(t)
    })
    
    t.Run("error:not found", func(t *testing.T) {
        certRepo := new(MockCertRepo)
        sessionRepo := new(MockSessionRepo)
        queue := new(MockQueue)
        storage := new(MockStorage)
        
        certRepo.On("GetByID", mock.Anything, "invalid").Return(nil, domain.ErrCertificateNotFound)
        
        svc := service.NewCertificateService(certRepo, sessionRepo, queue, storage)
        
        _, err := svc.GetStatus(context.Background(), "invalid")
        
        assert.ErrorIs(t, err, domain.ErrCertificateNotFound)
        certRepo.AssertExpectations(t)
    })
}

func TestCertificateService_DownloadCertificate(t *testing.T) {
    t.Run("success", func(t *testing.T) {
        certRepo := new(MockCertRepo)
        sessionRepo := new(MockSessionRepo)
        queue := new(MockQueue)
        storage := new(MockStorage)
        
        certRepo.On("GetByID", mock.Anything, "cert-123").Return(&domain.Certificate{
            ID:      "cert-123",
            Status:  domain.StatusDone,
            PDFPath: "cert-123.pdf",
        }, nil)
        
        storage.On("Get", mock.Anything, "cert-123.pdf").Return([]byte("pdf content"), nil)
        
        svc := service.NewCertificateService(certRepo, sessionRepo, queue, storage)
        
        data, err := svc.DownloadCertificate(context.Background(), "cert-123")
        
        assert.NoError(t, err)
        assert.Equal(t, []byte("pdf content"), data)
        certRepo.AssertExpectations(t)
        storage.AssertExpectations(t)
    })
    
    t.Run("error:certificate not ready", func(t *testing.T) {
        certRepo := new(MockCertRepo)
        sessionRepo := new(MockSessionRepo)
        queue := new(MockQueue)
        storage := new(MockStorage)
        
        certRepo.On("GetByID", mock.Anything, "cert-123").Return(&domain.Certificate{
            ID:     "cert-123",
            Status: domain.StatusPending,
        }, nil)
        
        svc := service.NewCertificateService(certRepo, sessionRepo, queue, storage)
        
        _, err := svc.DownloadCertificate(context.Background(), "cert-123")
        
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "not ready")
        certRepo.AssertExpectations(t)
    })
}