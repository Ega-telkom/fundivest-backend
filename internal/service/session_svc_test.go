// internal/service/session_svc_test.go
package service_test

import (
    "context"
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/Ega-telkom/fundivest-backend/internal/domain"
    "github.com/Ega-telkom/fundivest-backend/internal/service"
)

func TestSessionService_CreateSession(t *testing.T) {
    sessionRepo := new(MockSessionRepo)
    
    sessionRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Session")).Return(nil)
    
    svc := service.NewSessionService(sessionRepo, 24*time.Hour)
    
    sessionID, err := svc.CreateSession(context.Background(), "John Doe", "course-1")
    
    assert.NoError(t, err)
    assert.NotEmpty(t, sessionID)
    sessionRepo.AssertExpectations(t)
}

func TestSessionService_CompleteChapter(t *testing.T) {
    t.Run("success:complete chapter 1", func(t *testing.T) {
        sessionRepo := new(MockSessionRepo)
        
        // Mock Get untuk check expiry
        sessionRepo.On("Get", mock.Anything, "sess-123").Return(&domain.Session{
            ID:                "sess-123",
            ChaptersCompleted: []int{},
            ExpiresAt:         time.Now().Add(1 * time.Hour),
        }, nil)
        
        // Mock CompleteChapter (atomic operation)
        sessionRepo.On("CompleteChapter", mock.Anything, "sess-123", 1).Return(nil)
        
        svc := service.NewSessionService(sessionRepo, 24*time.Hour)
        
        err := svc.CompleteChapter(context.Background(), "sess-123", 1)
        
        assert.NoError(t, err)
        sessionRepo.AssertExpectations(t)
    })
    
    t.Run("error:session expired", func(t *testing.T) {
        sessionRepo := new(MockSessionRepo)
        
        sessionRepo.On("Get", mock.Anything, "sess-123").Return(&domain.Session{
            ID:                "sess-123",
            ChaptersCompleted: []int{},
            ExpiresAt:         time.Now().Add(-1 * time.Hour),  // Expired
        }, nil)
        
        // CompleteChapter tidak dipanggil karena sudah fail di expiry check
        
        svc := service.NewSessionService(sessionRepo, 24*time.Hour)
        
        err := svc.CompleteChapter(context.Background(), "sess-123", 1)
        
        assert.ErrorIs(t, err, domain.ErrSessionExpired)
        sessionRepo.AssertExpectations(t)
    })
    
    t.Run("error:invalid chapter", func(t *testing.T) {
        sessionRepo := new(MockSessionRepo)
        
        // Get tidak dipanggil karena validasi chapter number duluan
        
        svc := service.NewSessionService(sessionRepo, 24*time.Hour)
        
        err := svc.CompleteChapter(context.Background(), "sess-123", 0)
        
        assert.ErrorIs(t, err, domain.ErrInvalidChapter)
        sessionRepo.AssertExpectations(t)
    })
    
    t.Run("error:session not found", func(t *testing.T) {
        sessionRepo := new(MockSessionRepo)
        
        sessionRepo.On("Get", mock.Anything, "invalid-id").Return(nil, domain.ErrSessionNotFound)
        
        svc := service.NewSessionService(sessionRepo, 24*time.Hour)
        
        err := svc.CompleteChapter(context.Background(), "invalid-id", 1)
        
        assert.ErrorIs(t, err, domain.ErrSessionNotFound)
        sessionRepo.AssertExpectations(t)
    })
    
    t.Run("error:chapter already completed from lua", func(t *testing.T) {
        sessionRepo := new(MockSessionRepo)
        
        sessionRepo.On("Get", mock.Anything, "sess-123").Return(&domain.Session{
            ID:                "sess-123",
            ChaptersCompleted: []int{1},  // Chapter 1 already done
            ExpiresAt:         time.Now().Add(1 * time.Hour),
        }, nil)
        
        // Lua script returns error
        sessionRepo.On("CompleteChapter", mock.Anything, "sess-123", 1).
            Return(domain.ErrChapterAlreadyCompleted)
        
        svc := service.NewSessionService(sessionRepo, 24*time.Hour)
        
        err := svc.CompleteChapter(context.Background(), "sess-123", 1)
        
        assert.ErrorIs(t, err, domain.ErrChapterAlreadyCompleted)
        sessionRepo.AssertExpectations(t)
    })
    
    t.Run("error:previous chapter not completed from lua", func(t *testing.T) {
        sessionRepo := new(MockSessionRepo)
        
        sessionRepo.On("Get", mock.Anything, "sess-123").Return(&domain.Session{
            ID:                "sess-123",
            ChaptersCompleted: []int{},  // No chapters done
            ExpiresAt:         time.Now().Add(1 * time.Hour),
        }, nil)
        
        // Try to complete chapter 2 without chapter 1
        sessionRepo.On("CompleteChapter", mock.Anything, "sess-123", 2).
            Return(domain.ErrPreviousChapterNotCompleted)
        
        svc := service.NewSessionService(sessionRepo, 24*time.Hour)
        
        err := svc.CompleteChapter(context.Background(), "sess-123", 2)
        
        assert.ErrorIs(t, err, domain.ErrPreviousChapterNotCompleted)
        sessionRepo.AssertExpectations(t)
    })
}