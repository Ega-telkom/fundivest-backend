// tests/integration/certificate_flow_test.go
package integration

import (
    "context"
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    
    "fundivest/internal/domain"
    repoPostgres "fundivest/internal/repository/postgres"
    repoValkey "fundivest/internal/repository/valkey"
    "fundivest/internal/service"
)

func TestCertificateFlow_HappyPath(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }
    
    db := setupTestDB(t)
    valkeyClient := setupTestValkey(t)
    testQueue := NewTestQueue()
    testStorage := NewTestStorage()
    
    certRepo := repoPostgres.NewCertificateRepo(db)
    sessionRepo := repoValkey.NewSessionRepo(valkeyClient, 1*time.Hour)
    
    certSvc := service.NewCertificateService(certRepo, sessionRepo, testQueue, testStorage)
    sessionSvc := service.NewSessionService(sessionRepo, 1*time.Hour)
    
    ctx := context.Background()
    
    // === TEST FLOW ===
    
    // Step 1: Create session
    sessionID, err := sessionSvc.CreateSession(ctx, "John Doe", "golang-basics")
    require.NoError(t, err)
    assert.NotEmpty(t, sessionID)
    
    // Step 2: Complete chapter 1
    err = sessionSvc.CompleteChapter(ctx, sessionID, 1)
    require.NoError(t, err)
    
    // Step 3: Complete chapter 2
    err = sessionSvc.CompleteChapter(ctx, sessionID, 2)
    require.NoError(t, err)
    
    // Step 4: Complete chapter 3
    err = sessionSvc.CompleteChapter(ctx, sessionID, 3)
    require.NoError(t, err)
    
    // Step 5: Request certificate
    certID, err := certSvc.RequestCertificate(ctx, sessionID)
    require.NoError(t, err)
    assert.NotEmpty(t, certID)
    
    // === ASSERTIONS ===
    
    // Verify certificate created in database
    cert, err := certRepo.GetByID(ctx, certID)
    require.NoError(t, err)
    assert.Equal(t, "John Doe", cert.Name)
    assert.Equal(t, "golang-basics", cert.CourseID)
    assert.Equal(t, domain.StatusPending, cert.Status)
    assert.Empty(t, cert.PDFPath)
    
    // Verify session marked as completed
    sess, err := sessionRepo.Get(ctx, sessionID)
    require.NoError(t, err)
    assert.True(t, sess.Completed)
    assert.Equal(t, []int{1, 2, 3}, sess.ChaptersCompleted)
    
    // Verify job published to queue
    published := testQueue.GetPublished()
    require.Len(t, published, 1)
    assert.Equal(t, certID, published[0])
    
    // Step 6: Get certificate status
    status, err := certSvc.GetStatus(ctx, certID)
    require.NoError(t, err)
    assert.Equal(t, domain.StatusPending, status)
}

func TestCertificateFlow_CannotRequestTwice(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }
    
    db := setupTestDB(t)
    valkeyClient := setupTestValkey(t)
    testQueue := NewTestQueue()
    testStorage := NewTestStorage()
    
    certRepo := repoPostgres.NewCertificateRepo(db)
    sessionRepo := repoValkey.NewSessionRepo(valkeyClient, 1*time.Hour)
    
    certSvc := service.NewCertificateService(certRepo, sessionRepo, testQueue, testStorage)
    sessionSvc := service.NewSessionService(sessionRepo, 1*time.Hour)
    
    ctx := context.Background()
    
    // Create session and complete all chapters
    sessionID, _ := sessionSvc.CreateSession(ctx, "Jane Doe", "course-1")
    sessionSvc.CompleteChapter(ctx, sessionID, 1)
    sessionSvc.CompleteChapter(ctx, sessionID, 2)
    sessionSvc.CompleteChapter(ctx, sessionID, 3)
    
    // First request succeeds
    certID1, err := certSvc.RequestCertificate(ctx, sessionID)
    require.NoError(t, err)
    assert.NotEmpty(t, certID1)
    
    // Second request should succeed aswell
    // Respon harus idempoten
    certID2, err := certSvc.RequestCertificate(ctx, sessionID)
    require.NoError(t, err)
    assert.NotEmpty(t, certID2)
}

func TestCertificateFlow_CannotRequestWithIncompleteChapters(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }
    
    db := setupTestDB(t)
    valkeyClient := setupTestValkey(t)
    testQueue := NewTestQueue()
    testStorage := NewTestStorage()
    
    certRepo := repoPostgres.NewCertificateRepo(db)
    sessionRepo := repoValkey.NewSessionRepo(valkeyClient, 1*time.Hour)
    
    certSvc := service.NewCertificateService(certRepo, sessionRepo, testQueue, testStorage)
    sessionSvc := service.NewSessionService(sessionRepo, 1*time.Hour)
    
    ctx := context.Background()
    
    // Create session and complete only 2 chapters
    sessionID, _ := sessionSvc.CreateSession(ctx, "Bob Smith", "course-1")
    sessionSvc.CompleteChapter(ctx, sessionID, 1)
    sessionSvc.CompleteChapter(ctx, sessionID, 2)
    
    // Request certificate should fail
    _, err := certSvc.RequestCertificate(ctx, sessionID)
    assert.ErrorIs(t, err, domain.ErrNotAllChaptersCompleted)
}