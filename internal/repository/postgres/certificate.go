// internal/repository/postgres/certificate.go
package postgres

import (
	"context"
	"errors"
	"time"

	"fundivest/internal/domain"

	"gorm.io/gorm"
)

type CertificateRepo struct {
	db *gorm.DB
}

func NewCertificateRepo(db *gorm.DB) *CertificateRepo {
	return &CertificateRepo{db: db}
}

// Model untuk GORM
type Certificate struct {
	ID        string `gorm:"primaryKey"`
	Name      string
	CourseID  string
	SessionID string `gorm:"index"`
	IssuedAt  int64
	Status    string
	PDFPath   string
}

func (CertificateRepo) TableName() string {
	return "certificates"
}

func (r *CertificateRepo) Create(ctx context.Context, cert *domain.Certificate) error {
	model := &Certificate{
		ID:        cert.ID,
		Name:      cert.Name,
		CourseID:  cert.CourseID,
		SessionID: cert.SessionID,
		IssuedAt:  cert.IssuedAt.Unix(),
		Status:    string(cert.Status),
		PDFPath:   cert.PDFPath,
	}

	return r.db.WithContext(ctx).Create(model).Error
}

func (r *CertificateRepo) GetByID(ctx context.Context, id string) (*domain.Certificate, error) {
	var model Certificate

	err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrCertificateNotFound
		}
		return nil, err
	}

	return &domain.Certificate{
		ID:        model.ID,
		Name:      model.Name,
		CourseID:  model.CourseID,
		SessionID: model.SessionID,
		IssuedAt:  time.Unix(model.IssuedAt, 0),
		Status:    domain.CertStatus(model.Status),
		PDFPath:   model.PDFPath,
	}, nil
}

// internal/repository/postgres/certificate.go
func (r *CertificateRepo) GetBySessionID(ctx context.Context, sessionID string) (*domain.Certificate, error) {
	query := `SELECT id, name, course_id, session_id, issued_at, status, pdf_path
              FROM certificates WHERE session_id = $1 LIMIT 1`

	var cert Certificate
	err := r.db.WithContext(ctx).Raw(query, sessionID).Scan(&cert).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil, bukan error
		}
		return nil, err
	}

	return &domain.Certificate{
		ID:        cert.ID,
		Name:      cert.Name,
		CourseID:  cert.CourseID,
		SessionID: cert.SessionID,
		IssuedAt:  time.Unix(cert.IssuedAt, 0),
		Status:    domain.CertStatus(cert.Status),
		PDFPath:   cert.PDFPath,
	}, nil
}

func (r *CertificateRepo) UpdatePDFPath(ctx context.Context, id, path string) error {
	return r.db.WithContext(ctx).
		Model(&Certificate{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"pdf_path": path,
			"status":   domain.StatusDone,
		}).Error
}

func (r *CertificateRepo) UpdateStatus(ctx context.Context, id string, status domain.CertStatus) error {
	return r.db.WithContext(ctx).
		Model(&Certificate{}).
		Where("id = ?", id).
		Update("status", string(status)).Error
}
