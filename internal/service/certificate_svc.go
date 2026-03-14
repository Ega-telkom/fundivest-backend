// internal/service/cert_svc.go
package service

import (
    "context"
    "fmt"
    "time"
    
    "github.com/google/uuid"
    "fundivest/internal/domain"
)

type CertificateRepository interface {
    Create(ctx context.Context, cert *domain.Certificate) error
    GetByID(ctx context.Context, id string) (*domain.Certificate, error)
    GetBySessionID(ctx context.Context, sessionID string) (*domain.Certificate, error)
    UpdatePDFPath(ctx context.Context, id, path string) error
    UpdateStatus(ctx context.Context, id string, status domain.CertStatus) error
}

type QueuePublisher interface {
    Publish(ctx context.Context, certID string) error
}

type FileStorage interface {
    Save(ctx context.Context, filename string, data []byte) error
    Get(ctx context.Context, filename string) ([]byte, error)
}

type CertificateService struct {
    certRepo    CertificateRepository
    sessionRepo SessionRepository
    queue       QueuePublisher
    storage     FileStorage  // Tambah ini
}

func NewCertificateService(
    certRepo CertificateRepository,
    sessionRepo SessionRepository,
    queue QueuePublisher,
    storage FileStorage,  // Tambah parameter
) *CertificateService {
    return &CertificateService{
        certRepo:    certRepo,
        sessionRepo: sessionRepo,
        queue:       queue,
        storage:     storage,  // Assign
    }
}

// internal/service/cert_svc.go
func (s *CertificateService) RequestCertificate(ctx context.Context, sessionID string) (string, error) {
    sess, err := s.sessionRepo.Get(ctx, sessionID)
    if err != nil {
        return "", err
    }
    
    if !sess.IsAllChaptersCompleted() {
        return "", domain.ErrNotAllChaptersCompleted
    }
    
    // Kalau sudah completed, cari certificate yang sudah ada
    if sess.Completed {
        existingCert, err := s.certRepo.GetBySessionID(ctx, sessionID)
        if err == nil && existingCert != nil {
            // Return existing certificate (idempotent)
            return existingCert.ID, nil
        }
        // Kalau gak ketemu (edge case), return error
        return "", domain.ErrCertificateAlreadyRequested
    }
    
    // Create new certificate
    cert := &domain.Certificate{
        ID:        uuid.NewString(),
        Name:      sess.Name,
        CourseID:  sess.CourseID,
        SessionID: sessionID,
        IssuedAt:  time.Now(),
        Status:    domain.StatusPending,
    }
    
    if err := s.certRepo.Create(ctx, cert); err != nil {
        return "", err
    }
    
    sess.Completed = true
    if err := s.sessionRepo.Update(ctx, sess); err != nil {
        return "", err
    }
    
    if err := s.queue.Publish(ctx, cert.ID); err != nil {
        return "", err
    }
    
    return cert.ID, nil
}

func (s *CertificateService) GetStatus(ctx context.Context, certID string) (domain.CertStatus, error) {
    cert, err := s.certRepo.GetByID(ctx, certID)
    if err != nil {
        return "", err
    }
    return cert.Status, nil
}

func (s *CertificateService) GetCertificate(ctx context.Context, certID string) (*domain.Certificate, error) {
    return s.certRepo.GetByID(ctx, certID)
}

func (s *CertificateService) DownloadCertificate(ctx context.Context, certID string) ([]byte, error) {
    cert, err := s.certRepo.GetByID(ctx, certID)
    if err != nil {
        return nil, err
    }
    
    if cert.Status != domain.StatusDone {
        return nil, fmt.Errorf("certificate not ready")
    }
    
    return s.storage.Get(ctx, cert.PDFPath)
}