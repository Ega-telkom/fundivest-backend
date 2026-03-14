// internal/handler/response.go
package handler

import (
    "github.com/gofiber/fiber/v3"
    "fundivest/internal/domain"
)

// Success response
func Success(c fiber.Ctx, statusCode int, data any) error {
    return c.Status(statusCode).JSON(SuccessResponse{
	   	Success: true,
	    Data: data,
    })
}

// Error response with domain error mapping
// internal/handler/response.go
func Error(c fiber.Ctx, err error) error {
    statusCode := fiber.StatusInternalServerError
    
    switch err {
    case domain.ErrSessionNotFound, domain.ErrCertificateNotFound:
        statusCode = fiber.StatusNotFound
    case domain.ErrSessionExpired,
        domain.ErrPreviousChapterNotCompleted,
        domain.ErrChapterAlreadyCompleted,
        domain.ErrInvalidChapter,
        domain.ErrNotAllChaptersCompleted,
        domain.ErrCertificateAlreadyRequested:
        statusCode = fiber.StatusBadRequest
    }
    
    return c.Status(statusCode).JSON(ErrorResponse{
        Success: false,
        Error: err.Error(),
    })
}