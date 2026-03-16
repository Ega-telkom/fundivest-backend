// internal/handler/session_test.go
package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/Ega-telkom/fundivest-backend/internal/domain"
	"github.com/Ega-telkom/fundivest-backend/internal/handler"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockSessionService struct {
    mock.Mock
}

func (m *MockSessionService) CreateSession(ctx context.Context, name, courseID string) (string, error) {
    args := m.Called(ctx, name, courseID)
    return args.String(0), args.Error(1)
}

func (m *MockSessionService) CompleteChapter(ctx context.Context, sessionID string, chapter int) error {
    args := m.Called(ctx, sessionID, chapter)
    return args.Error(0)
}

func (m *MockSessionService) GetSession(ctx context.Context, sessionID string) (*domain.Session, error) {
	args := m.Called(ctx, sessionID)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*domain.Session), args.Error(1)
}

func TestSessionHandler_CreateSession(t *testing.T) {
    t.Run("success", func(t *testing.T) {
        svc := new(MockSessionService)
        h := handler.NewSessionHandler(svc)
        
        app := fiber.New()
        app.Post("/api/sessions", h.CreateSession)
        
        svc.On("CreateSession", mock.Anything, "John Doe", "course-1").
            Return("sess-123", nil)
        
        body := map[string]string{
            "name":      "John Doe",
            "course_id": "course-1",
        }
        jsonBody, _ := json.Marshal(body)
        
        req := httptest.NewRequest("POST", "/api/sessions", bytes.NewReader(jsonBody))
        req.Header.Set("Content-Type", "application/json")
        
        resp, _ := app.Test(req)
        
        assert.Equal(t, fiber.StatusCreated, resp.StatusCode)
        
        var respBody map[string]interface{}
        err := json.NewDecoder(resp.Body).Decode(&respBody)
        require.NoError(t, err, "failed to decode response")
        
        // FIX: Parse nested data
        assert.Equal(t, true, respBody["success"])
        data := respBody["data"].(map[string]interface{})
        assert.Equal(t, "sess-123", data["session_id"])
        
        svc.AssertExpectations(t)
    })
}

func TestSessionHandler_CompleteChapter(t *testing.T) {
    t.Run("success", func(t *testing.T) {
        svc := new(MockSessionService)
        h := handler.NewSessionHandler(svc)
        
        app := fiber.New()
        app.Post("/api/sessions/:id/chapters/:chapter/complete", h.CompleteChapter)
        
        // Mock CompleteChapter
        svc.On("CompleteChapter", mock.Anything, "sess-123", 1).Return(nil)
        
        // Mock GetSession
        svc.On("GetSession", mock.Anything, "sess-123").Return(&domain.Session{
            ID:                "sess-123",
            ChaptersCompleted: []int{1},
            Completed:         false,
        }, nil)
        
        req := httptest.NewRequest("POST", "/api/sessions/sess-123/chapters/1/complete", nil)
        
        resp, _ := app.Test(req)
        
        assert.Equal(t, fiber.StatusOK, resp.StatusCode)
        
        var respBody map[string]interface{}
        err := json.NewDecoder(resp.Body).Decode(&respBody)
        require.NoError(t, err, "failed to decode response")
        
        // FIX: Parse nested data
        assert.Equal(t, true, respBody["success"])
        data := respBody["data"].(map[string]interface{})
        assert.Equal(t, float64(1), data["chapter"])
        assert.Equal(t, true, data["completed"])
        assert.Equal(t, false, data["completed_all"])
        
        svc.AssertExpectations(t)
    })
    
    t.Run("error:chapter already completed", func(t *testing.T) {
        svc := new(MockSessionService)
        h := handler.NewSessionHandler(svc)
        
        app := fiber.New()
        app.Post("/api/sessions/:id/chapters/:chapter/complete", h.CompleteChapter)
        
        svc.On("CompleteChapter", mock.Anything, "sess-123", 1).
            Return(domain.ErrChapterAlreadyCompleted)
        
        req := httptest.NewRequest("POST", "/api/sessions/sess-123/chapters/1/complete", nil)
        
        resp, _ := app.Test(req)
        
        assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
        
        var respBody map[string]interface{}
        err := json.NewDecoder(resp.Body).Decode(&respBody)
        require.NoError(t, err, "failed to decode response")
        
        assert.Equal(t, false, respBody["success"])
        assert.Contains(t, respBody["error"], "chapter already completed")
        
        svc.AssertExpectations(t)
    })
}