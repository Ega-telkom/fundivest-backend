// internal/domain/session.go
package domain

import "time"

type Session struct {
    ID                string
    Name              string
    CourseID          string
    ChaptersCompleted []int
    Completed         bool
    CreatedAt         time.Time
    ExpiresAt         time.Time
}

func (s *Session) IsAllChaptersCompleted() bool {
    return len(s.ChaptersCompleted) == TotalChapters
}

func (s *Session) CanCompleteChapter(chapter int) error {
    if chapter < 1 || chapter > TotalChapters {
        return ErrInvalidChapter
    }
    
    // Chapter harus sequential
    if chapter > 1 && !s.HasCompletedChapter(chapter-1) {
        return ErrPreviousChapterNotCompleted
    }
    
    if s.HasCompletedChapter(chapter) {
        return ErrChapterAlreadyCompleted
    }
    
    return nil
}

func (s *Session) HasCompletedChapter(chapter int) bool {
    for _, completed := range s.ChaptersCompleted {
        if completed == chapter {
            return true
        }
    }
    return false
}
