// internal/service/mocks_test.go
package service_test

import (
    "context"
    
    "github.com/stretchr/testify/mock"
    "fundivest/internal/domain"
)

// ===== SESSION REPO MOCK =====

type MockSessionRepo struct {
    mock.Mock
}

func (m *MockSessionRepo) Create(ctx context.Context, sess *domain.Session) error {
    args := m.Called(ctx, sess)
    return args.Error(0)
}

func (m *MockSessionRepo) Get(ctx context.Context, id string) (*domain.Session, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*domain.Session), args.Error(1)
}

func (m *MockSessionRepo) Update(ctx context.Context, sess *domain.Session) error {
    args := m.Called(ctx, sess)
    return args.Error(0)
}

func (m *MockSessionRepo) CompleteChapter(ctx context.Context, sessionID string, chapter int) error {
    args := m.Called(ctx, sessionID, chapter)
    return args.Error(0)
}

// ===== CERTIFICATE REPO MOCK =====

type MockCertRepo struct {
    mock.Mock
}

func (m *MockCertRepo) Create(ctx context.Context, cert *domain.Certificate) error {
    args := m.Called(ctx, cert)
    return args.Error(0)
}

func (m *MockCertRepo) GetByID(ctx context.Context, id string) (*domain.Certificate, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*domain.Certificate), args.Error(1)
}

func (m *MockCertRepo) GetBySessionID(ctx context.Context, sessionID string) (*domain.Certificate, error) {
    args := m.Called(ctx, sessionID)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*domain.Certificate), args.Error(1)
}

func (m *MockCertRepo) UpdatePDFPath(ctx context.Context, id, path string) error {
    args := m.Called(ctx, id, path)
    return args.Error(0)
}

func (m *MockCertRepo) UpdateStatus(ctx context.Context, id string, status domain.CertStatus) error {
    args := m.Called(ctx, id, status)
    return args.Error(0)
}

// ===== QUEUE MOCK =====

type MockQueue struct {
    mock.Mock
}

func (m *MockQueue) Publish(ctx context.Context, certID string) error {
    args := m.Called(ctx, certID)
    return args.Error(0)
}

// ===== STORAGE MOCK =====

type MockStorage struct {
    mock.Mock
}

func (m *MockStorage) Save(ctx context.Context, filename string, data []byte) error {
    args := m.Called(ctx, filename, data)
    return args.Error(0)
}

func (m *MockStorage) Get(ctx context.Context, filename string) ([]byte, error) {
    args := m.Called(ctx, filename)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).([]byte), args.Error(1)
}

func (m *MockStorage) Delete(ctx context.Context, filename string) error {
    args := m.Called(ctx, filename)
    return args.Error(0)
}

func (m *MockStorage) Exists(ctx context.Context, filename string) (bool, error) {
    args := m.Called(ctx, filename)
    return args.Bool(0), args.Error(1)
}