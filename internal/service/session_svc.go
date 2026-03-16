// internal/service/session_svc.go
package service

import (
    "context"
    "time"
    
    "github.com/google/uuid"
    "github.com/Ega-telkom/fundivest-backend/internal/domain"
)

type SessionRepository interface {
    Create(ctx context.Context, sess *domain.Session) error
    Get(ctx context.Context, id string) (*domain.Session, error)
    Update(ctx context.Context, sess *domain.Session) error
    CompleteChapter(ctx context.Context, sessionID string, chapter int) error
}

type SessionService struct {
    repo       SessionRepository
    sessionTTL time.Duration
}

func NewSessionService(repo SessionRepository, ttl time.Duration) *SessionService {
    return &SessionService{repo: repo, sessionTTL: ttl}
}

func (s *SessionService) GetSession(ctx context.Context, sessionID string) (*domain.Session, error) {
    return s.repo.Get(ctx, sessionID)
}

func (s *SessionService) CreateSession(ctx context.Context, name, courseID string) (string, error) {
    sess := &domain.Session{
        ID:                uuid.NewString(),
        Name:              name,
        CourseID:          courseID,
        ChaptersCompleted: []int{},
        Completed:         false,
        CreatedAt:         time.Now(),
        ExpiresAt:         time.Now().Add(s.sessionTTL),
    }
    
    if err := s.repo.Create(ctx, sess); err != nil {
        return "", err
    }
    
    return sess.ID, nil
}

func (s *SessionService) CompleteChapter(ctx context.Context, sessionID string, chapter int) error {
    // Validate chapter number early
    if chapter < 1 || chapter > domain.TotalChapters {
        return domain.ErrInvalidChapter
    }
    
    // Check session expiry
    sess, err := s.repo.Get(ctx, sessionID)
    if err != nil {
        return err
    }
    
    if time.Now().After(sess.ExpiresAt) {
        return domain.ErrSessionExpired
    }
    
    // Atomic complete via Lua script
    return s.repo.CompleteChapter(ctx, sessionID, chapter)
}