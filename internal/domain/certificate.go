// internal/domain/certificate.go
package domain

import "time"

type Certificate struct {
	ID        string
	Name      string
	CourseID  string
	SessionID string
	IssuedAt  time.Time
	Status    CertStatus
	PDFPath   string
}

type CertStatus string

const (
	StatusPending CertStatus = "pending"
	StatusDone    CertStatus = "done"
	StatusFailed  CertStatus = "failed"
)
