// internal/handler/certificate.go
package handler

import (
	"context"
	"fundivest/internal/domain"

	"github.com/gofiber/fiber/v3"
)

type CertificateService interface {
	RequestCertificate(ctx context.Context, sessionID string) (string, error)
	GetStatus(ctx context.Context, certID string) (domain.CertStatus, error)
	GetCertificate(ctx context.Context, certID string) (*domain.Certificate, error)
	DownloadCertificate(ctx context.Context, certID string) ([]byte, error)
}

type CertificateHandler struct {
	certSvc CertificateService
}

func NewCertificateHandler(certSvc CertificateService) *CertificateHandler {
	return &CertificateHandler{certSvc: certSvc}
}

// RequestCertificate godoc
// @Summary      Request sertifikat
// @Description  Request pembuatan sertifikat setelah semua chapter sudah selesai
// @Tags         sertifikat
// @Accept       json
// @Produce      json
// @Param        id path string true "Session ID"
// @Success      202 {object} SuccessResponse{data=RequestCertificateResponse}
// @Failure      400 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Router       /sessions/{id}/certificate [post]
func (h *CertificateHandler) RequestCertificate(c fiber.Ctx) error {
	sessionID := c.Params("id")

	certID, err := h.certSvc.RequestCertificate(c.Context(), sessionID)
	if err != nil {
		return Error(c, err)
	}

	return Success(c, fiber.StatusAccepted, RequestCertificateResponse{
		CertificateID: certID,
		Status:        string(domain.StatusPending),
	})
}

// GetStatus godoc
// @Summary      Cek status sertifikat
// @Description  Cek status proses pembuatan sertifikat
// @Tags         sertifikat
// @Produce      json
// @Param        certificate_id path string true "Certificate ID"
// @Success      200 {object} SuccessResponse{data=CertificateStatusResponse}
// @Failure      404 {object} ErrorResponse
// @Router       /certificates/{certificate_id}/status [get]
func (h *CertificateHandler) GetStatus(c fiber.Ctx) error {
	certID := c.Params("certificate_id")

	status, err := h.certSvc.GetStatus(c.Context(), certID)
	if err != nil {
		return Error(c, err)
	}

	return Success(c, fiber.StatusOK, CertificateStatusResponse{
		CertificateID: certID,
		Status:        string(status),
	})
}

// Download godoc
// @Summary      Unduh sertifikat
// @Description  Unduh sertifikat yang sudah dibuat
// @Tags         sertifikat
// @Produce      application/pdf
// @Param        certificate_id path string true "Certificate ID"
// @Success      200 {file} binary
// @Failure      400 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Router       /certificates/{certificate_id}/download [get]
func (h *CertificateHandler) Download(c fiber.Ctx) error {
	certID := c.Params("certificate_id")

	data, err := h.certSvc.DownloadCertificate(c.Context(), certID)
	if err != nil {
		return Error(c, err)
	}

	c.Set(fiber.HeaderContentType, "application/pdf")
	c.Set(fiber.HeaderContentDisposition, "attachment; filename=certificate.pdf")
	return c.Send(data)
}

// Verify godoc
// @Summary      Verifikasi sertifikat
// @Description  Validasi keasilan sertifikat
// @Tags         sertifikat
// @Produce      json
// @Param        certificate_id path string true "Certificate ID"
// @Success      200 {object} SuccessResponse{data=VerifyCertificateResponse}
// @Failure      400 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Router       /certificates/{certificate_id}/verify [get]
func (h *CertificateHandler) Verify(c fiber.Ctx) error {
	certID := c.Params("certificate_id")

	cert, err := h.certSvc.GetCertificate(c.Context(), certID)
	if err != nil {
		return Error(c, err)
	}

	return Success(c, fiber.StatusOK, VerifyCertificateResponse{
		ID:       cert.ID,
		Name:     cert.Name,
		CourseID: cert.CourseID,
		IssuedAt: cert.IssuedAt.Format("2006-01-02T15:04:05Z07:00"),
		Status:   string(cert.Status),
		Verified: cert.Status == domain.StatusDone,
	})
}
