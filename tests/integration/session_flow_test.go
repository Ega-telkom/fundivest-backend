// tests/integration/session_flow_test.go
package integration

import (
    "context"
    "testing"
    "errors"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    
    "github.com/Ega-telkom/fundivest-backend/internal/domain"
    repoValkey "github.com/Ega-telkom/fundivest-backend/internal/repository/valkey"
    "github.com/Ega-telkom/fundivest-backend/internal/service"
)

func TestSessionFlow_ChapterProgression(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }
    
    valkeyClient := setupTestValkey(t)
    sessionRepo := repoValkey.NewSessionRepo(valkeyClient, 1*time.Hour)
    sessionSvc := service.NewSessionService(sessionRepo, 1*time.Hour)
    
    ctx := context.Background()
    
    sessionID, err := sessionSvc.CreateSession(ctx, "Test User", "test-course")
    require.NoError(t, err)
    
    // Test: Cannot skip to chapter 2
    err = sessionSvc.CompleteChapter(ctx, sessionID, 2)
    assert.ErrorIs(t, err, domain.ErrPreviousChapterNotCompleted)
    
    // Test: Complete chapter 1
    err = sessionSvc.CompleteChapter(ctx, sessionID, 1)
    assert.NoError(t, err)
    
    // Test: Cannot repeat chapter 1
    err = sessionSvc.CompleteChapter(ctx, sessionID, 1)
    assert.ErrorIs(t, err, domain.ErrChapterAlreadyCompleted)
    
    // Test: Can complete chapter 2 after chapter 1
    err = sessionSvc.CompleteChapter(ctx, sessionID, 2)
    assert.NoError(t, err)
    
    // Test: Can complete chapter 3 after chapter 2
    err = sessionSvc.CompleteChapter(ctx, sessionID, 3)
    assert.NoError(t, err)
    
    // Verify session state
    sess, err := sessionRepo.Get(ctx, sessionID)
    require.NoError(t, err)
    assert.Equal(t, []int{1, 2, 3}, sess.ChaptersCompleted)
    
    // Verify chapters completed
    assert.True(t, sess.IsAllChaptersCompleted())
}

func TestSessionFlow_InvalidChapter(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }
    
    valkeyClient := setupTestValkey(t)
    sessionRepo := repoValkey.NewSessionRepo(valkeyClient, 1*time.Hour)
    sessionSvc := service.NewSessionService(sessionRepo, 1*time.Hour)
    
    ctx := context.Background()
    
    sessionID, _ := sessionSvc.CreateSession(ctx, "Test User", "test-course")
    
    // Test: Invalid chapter numbers
    err := sessionSvc.CompleteChapter(ctx, sessionID, 0)
    assert.ErrorIs(t, err, domain.ErrInvalidChapter)
    
    err = sessionSvc.CompleteChapter(ctx, sessionID, 4)
    assert.ErrorIs(t, err, domain.ErrInvalidChapter)
    
    err = sessionSvc.CompleteChapter(ctx, sessionID, -1)
    assert.ErrorIs(t, err, domain.ErrInvalidChapter)
}

func TestSessionFlow_SessionExpiry(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }
    
    valkeyClient := setupTestValkey(t)
    
    // Create repo with very short TTL
    sessionRepo := repoValkey.NewSessionRepo(valkeyClient, 500*time.Millisecond)
    sessionSvc := service.NewSessionService(sessionRepo, 500*time.Millisecond)
    
    ctx := context.Background()
    
    sessionID, _ := sessionSvc.CreateSession(ctx, "Test User", "test-course")
    
    // Wait for session to expire
    time.Sleep(1 * time.Second)
    
    // Try to complete chapter - should fail
    err := sessionSvc.CompleteChapter(ctx, sessionID, 1)
    if err != domain.ErrSessionExpired && err != domain.ErrSessionNotFound {
        t.Errorf("expected expired or not found, got: %v", err)
    }
}

func TestSessionFlow_ConcurrentRequests(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }
    
    valkeyClient := setupTestValkey(t)
    sessionRepo := repoValkey.NewSessionRepo(valkeyClient, 1*time.Hour)
    sessionSvc := service.NewSessionService(sessionRepo, 1*time.Hour)
    
    ctx := context.Background()
    
    sessionID, _ := sessionSvc.CreateSession(ctx, "Test User", "test-course")
    
    // Launch 10 concurrent requests
    const goroutines = 100
    done := make(chan error, goroutines)
    
    for i := 0; i < goroutines; i++ {
        go func() {
            done <- sessionSvc.CompleteChapter(ctx, sessionID, 1)
        }()
    }
    
    // Collect results
    var successCount int
    var failCount int
    for i := 0; i < goroutines; i++ {
        err := <-done
        switch {
        case err == nil:
            successCount++
        case errors.Is(err, domain.ErrChapterAlreadyCompleted):
            failCount++
        default:
            t.Fatalf("unexpected error: %v", err)
        }
    }
    
    // Exactly ONE should succeed
    assert.Equal(t, 1, successCount, "Exactly one request should succeed")
    assert.Equal(t, goroutines-1, failCount, "All others should fail with already completed")
    
    // Verify no duplicate in session
    sess, _ := sessionRepo.Get(ctx, sessionID)
    assert.Equal(t, []int{1}, sess.ChaptersCompleted, "Chapter should be recorded only once")
}