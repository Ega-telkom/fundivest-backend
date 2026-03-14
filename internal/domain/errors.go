// internal/domain/errors.go
package domain

import "errors"

var (
	ErrSessionNotFound             = errors.New("session not found")
	ErrSessionExpired              = errors.New("session expired")
	ErrInvalidChapter              = errors.New("invalid chapter number")
	ErrPreviousChapterNotCompleted = errors.New("previous chapter not completed")
	ErrChapterAlreadyCompleted     = errors.New("chapter already completed")
	ErrNotAllChaptersCompleted     = errors.New("not all chapters completed")
	ErrCertificateAlreadyRequested = errors.New("certificate already requested")
	ErrCertificateNotFound         = errors.New("certificate not found")
)
