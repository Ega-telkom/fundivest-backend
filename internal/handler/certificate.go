// internal/handler/certificate.go
package handler

import (
	"bufio"
	"context"
	"fmt"

	// "time"

	"github.com/Ega-telkom/fundivest-backend/internal/domain"
	"github.com/Ega-telkom/fundivest-backend/internal/pubsub"
	"go.uber.org/zap"

	"github.com/gofiber/fiber/v3"
)

type CertificateService interface {
	RequestCertificate(ctx context.Context, sessionID string) (string, error)
	GetStatus(ctx context.Context, certID string) (domain.CertStatus, error)
	// StreamStatus(ctx context.Context, certID string) (domain.CertStatus, error)
	GetCertificate(ctx context.Context, certID string) (*domain.Certificate, error)
	DownloadCertificate(ctx context.Context, certID string) ([]byte, error)
}

type CertificateHandler struct {
	certSvc CertificateService
	pubsub pubsub.PubSub
	logger *zap.Logger
}

func NewCertificateHandler(certSvc CertificateService, pubsub pubsub.PubSub, logger *zap.Logger) *CertificateHandler {
	return &CertificateHandler{certSvc: certSvc, pubsub: pubsub, logger: logger}
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
		Status:        string(domain.StatusProcessing),
	})
}

// GetStatus godoc
// @Deprecated
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

// StreamStatus godoc
// @Summary      Stream status sertifikat
// @Description  Stream status proses pembuatan sertifikat
// @Tags         sertifikat
// @Produce      json
// @Param        certificate_id path string true "Certificate ID"
// @Success      200 {object} string "text/event-stream"
// @Router       /certificates/{certificate_id}/stream [get]
func (h *CertificateHandler) StreamStatus(c fiber.Ctx) error {

	certID := c.Params("certificate_id")

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("X-Accel-Buffering", "no")

	return c.SendStreamWriter(func(w *bufio.Writer) {

		ctx := c.Context()

		// 1 — send snapshot (source of truth)
		status, err := h.certSvc.GetStatus(context.Background(), certID)
		if err != nil {
			return
		}

		fmt.Fprintf(w,"event: status\n")
		fmt.Fprintf(w,"data: %s\n\n",status)

		if err := w.Flush(); err != nil {
			return
		}

		// 2 — exit if finished
		if status == "done" || status == "failed" {
			h.logger.Info("Client disconnected", zap.String("cert_id", certID))
			return
		}

		// 3 — subscribe only if still active
		topic := "cert:" + certID
		ch := h.pubsub.Subscribe(topic)
		defer h.pubsub.Unsubscribe(topic, ch)

		// 4 — stream only transitions
		for {
			select {

				case status, ok := <-ch:

				if !ok {
					return
				}

				fmt.Fprintf(w,"event: status\n")
				fmt.Fprintf(w,"data: %s\n\n",status)

				if err := w.Flush(); err != nil {
					return
				}

				// close on terminal
				if status == "done" || status == "failed" {
					h.logger.Info("Client disconnected", zap.String("cert_id", certID))
					return
				}

				case <-ctx.Done():
				return
			}
		}
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
