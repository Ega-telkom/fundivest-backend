// tests/integration/helpers_test.go
package integration

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/require"
    "github.com/valkey-io/valkey-go"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
    
    repoPostgres "github.com/Ega-telkom/fundivest-backend/internal/repository/postgres"
)

const (
    testPostgresDSN = "postgres://test:test@localhost:5433/certdb_test?sslmode=disable"
    testValkeyURL   = "localhost:6380"
)

func closeDB(t *testing.T, db *gorm.DB) {
    t.Helper()
    
    db.Exec("TRUNCATE TABLE certificates")
    
    sqlDB, _ := db.DB()
    if err := sqlDB.Close(); err != nil {
        t.Logf("cleanup warning: %v", err)
    }
}

func closeValkey(t *testing.T, client valkey.Client) {
    t.Helper()
    
    ctx := context.Background()
    client.Do(ctx, client.B().Flushdb().Build())
    
    client.Close()
}

func setupTestDB(t *testing.T) *gorm.DB {
    t.Helper()
    
    db, err := gorm.Open(postgres.Open(testPostgresDSN), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Silent),
    })
    require.NoError(t, err, "failed to connect to test database")
    
    // Auto migrate
    err = db.AutoMigrate(&repoPostgres.Certificate{})
    require.NoError(t, err, "failed to migrate")
    
    // Clean tables
    db.Exec("TRUNCATE TABLE certificates")
    
    t.Cleanup(func() { closeDB(t, db) })
    
    return db
}

func setupTestValkey(t *testing.T) valkey.Client {
    t.Helper()
    
    client, err := valkey.NewClient(valkey.ClientOption{
        InitAddress: []string{testValkeyURL},
    })
    require.NoError(t, err, "failed to connect to test valkey")
    
    // Flush test data
    ctx := context.Background()
    client.Do(ctx, client.B().Flushdb().Build())
    
    t.Cleanup(func() { closeValkey(t, client) })
    
    return client
}

// Mock queue for testing
type TestQueue struct {
    published []string
}

func NewTestQueue() *TestQueue {
    return &TestQueue{
        published: make([]string, 0),
    }
}

func (q *TestQueue) Publish(ctx context.Context, certID string) error {
    q.published = append(q.published, certID)
    return nil
}

func (q *TestQueue) GetPublished() []string {
    return q.published
}

// Mock storage for testing
type TestStorage struct {
    files map[string][]byte
}

func NewTestStorage() *TestStorage {
    return &TestStorage{
        files: make(map[string][]byte),
    }
}

func (s *TestStorage) Save(ctx context.Context, filename string, data []byte) error {
    s.files[filename] = data
    return nil
}

func (s *TestStorage) Get(ctx context.Context, filename string) ([]byte, error) {
    data, ok := s.files[filename]
    if !ok {
        return nil, nil
    }
    return data, nil
}

func (s *TestStorage) Delete(ctx context.Context, filename string) error {
    delete(s.files, filename)
    return nil
}

func (s *TestStorage) Exists(ctx context.Context, filename string) (bool, error) {
    _, ok := s.files[filename]
    return ok, nil
}