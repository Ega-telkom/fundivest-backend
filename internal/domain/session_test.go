// internal/domain/session_test.go
package domain_test

import (
    "testing"
    
    "github.com/stretchr/testify/assert"
    "fundivest/internal/domain"
)

func TestSession_CanCompleteChapter(t *testing.T) {
    tests := []struct {
        name               string
        chaptersCompleted  []int
        chapterToComplete  int
        wantErr            error
    }{
        {"can complete chapter 1", []int{}, 1, nil},
        {"can complete chapter 2 after 1", []int{1}, 2, nil},
        {"cannot skip chapter", []int{}, 2, domain.ErrPreviousChapterNotCompleted},
        {"cannot repeat chapter", []int{1}, 1, domain.ErrChapterAlreadyCompleted},
        {"invalid chapter", []int{}, 5, domain.ErrInvalidChapter},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            sess := &domain.Session{ChaptersCompleted: tt.chaptersCompleted}
            err := sess.CanCompleteChapter(tt.chapterToComplete)
            assert.Equal(t, tt.wantErr, err)
        })
    }
}

func TestSession_IsAllChaptersCompleted(t *testing.T) {
    tests := []struct {
        name            string
        chapterCompleted []int
        want            bool
    }{
        {"all 3 done", []int{1, 2, 3}, true},
        {"only 2 done", []int{1, 2}, false},
        {"none done", []int{}, false},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            sess := &domain.Session{ChaptersCompleted: tt.chapterCompleted}
            assert.Equal(t, tt.want, sess.IsAllChaptersCompleted())
        })
    }
}