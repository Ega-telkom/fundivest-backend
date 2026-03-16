// internal/handler/session.go
package handler

import (
	"context"
	"fmt"
	"github.com/Ega-telkom/fundivest-backend/internal/domain"

	"github.com/gofiber/fiber/v3"
)

type SessionService interface {
	CreateSession(ctx context.Context, name, courseID string) (string, error)
	CompleteChapter(ctx context.Context, sessionID string, chapter int) error
	GetSession(ctx context.Context, sessionID string) (*domain.Session, error)
}

type SessionHandler struct {
	svc SessionService
}

func NewSessionHandler(svc SessionService) *SessionHandler {
	return &SessionHandler{svc: svc}
}

// CreateSession godoc
// @Summary      Buat sesi baru
// @Description  Buat sesi permainan baru bagi pengguna
// @Tags         sesi
// @Accept       json
// @Produce      json
// @Param        request body CreateSessionRequest true "Session details"
// @Success      201 {object} SuccessResponse{data=CreateSessionResponse}
// @Failure      400 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /sessions [post]
func (h *SessionHandler) CreateSession(c fiber.Ctx) error {
	var req CreateSessionRequest
	if err := c.Bind().JSON(&req); err != nil {
		return Error(c, err)
	}

	sessionID, err := h.svc.CreateSession(c.Context(), req.Name, req.CourseID)
	if err != nil {
		return Error(c, err)
	}

	return Success(c, fiber.StatusCreated, CreateSessionResponse{
		SessionID: sessionID,
	})
}

// CompleteChapter godoc
// @Summary      Selesaikan chapter
// @Description  Tandai chapter sebagai selesai. Misal chapter 1 sudah selesai—maka tandai sebagai selesai.
// @Tags         sesi
// @Accept       json
// @Produce      json
// @Param        id path string true "Session ID"
// @Param        chapter path int true "Chapter number"
// @Success      200 {object} SuccessResponse{data=CompleteChapterResponse}
// @Failure      400 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Router       /sessions/{id}/chapters/{chapter}/complete [post]
func (h *SessionHandler) CompleteChapter(c fiber.Ctx) error {
	sessionID := c.Params("id")
	chapterStr := c.Params("chapter")

	// Parse chapter number
	var chapter int
	if _, err := fmt.Sscanf(chapterStr, "%d", &chapter); err != nil {
		return Error(c, domain.ErrInvalidChapter)
	}

	if err := h.svc.CompleteChapter(c.Context(), sessionID, chapter); err != nil {
		return Error(c, err)
	}

	sess, err := h.svc.GetSession(c.Context(), sessionID)
	if err != nil {
		return Error(c, err)
	}

	return Success(c, fiber.StatusOK, CompleteChapterResponse{
		Chapter: chapter,
		Completed: true,
		CompletedAll: sess.IsAllChaptersCompleted(),
	})
}
