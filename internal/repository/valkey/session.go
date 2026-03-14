// internal/repository/valkey/session.go
package valkey

import (
    "context"
    "encoding/json"
    "fmt"
    "strings"
    "time"
    
    "github.com/valkey-io/valkey-go"
    "fundivest/internal/domain"
)

type SessionRepo struct {
    client valkey.Client
    ttl    time.Duration
}

func NewSessionRepo(client valkey.Client, ttl time.Duration) *SessionRepo {
    return &SessionRepo{client: client, ttl: ttl}
}

func (r *SessionRepo) Create(ctx context.Context, sess *domain.Session) error {
    key := fmt.Sprintf("session:%s", sess.ID)
    data, _ := json.Marshal(sess)
    
    cmd := r.client.B().Set().
        Key(key).
        Value(string(data)).
        Ex(r.ttl).
        Build()
    
    return r.client.Do(ctx, cmd).Error()
}

func (r *SessionRepo) Get(ctx context.Context, id string) (*domain.Session, error) {
    key := fmt.Sprintf("session:%s", id)
    cmd := r.client.B().Get().Key(key).Build()
    
    result, err := r.client.Do(ctx, cmd).ToString()
    if err != nil {
        return nil, domain.ErrSessionNotFound
    }
    
    var sess domain.Session
    if err := json.Unmarshal([]byte(result), &sess); err != nil {
        return nil, err
    }
    
    return &sess, nil
}

func (r *SessionRepo) Update(ctx context.Context, sess *domain.Session) error {
    // Refresh TTL saat update
    return r.Create(ctx, sess)
}

// internal/repository/valkey/session.go
func (r *SessionRepo) CompleteChapter(ctx context.Context, sessionID string, chapter int) error {
    key := fmt.Sprintf("session:%s", sessionID)
    
    // Updated Lua script with proper return format
    script := `
local key = KEYS[1]
local chapter = tonumber(ARGV[1])
local ttl = tonumber(ARGV[2])
local total_chapters = tonumber(ARGV[3])

-- Get session data
local data = redis.call('GET', key)
if not data then
    return redis.error_reply('session_not_found')
end

-- Parse JSON
local sess = cjson.decode(data)

-- Validate chapter number
if chapter < 1 or chapter > total_chapters then
    return redis.error_reply('invalid_chapter')
end

-- Check if already completed
for _, c in ipairs(sess.ChaptersCompleted) do
    if c == chapter then
        return redis.error_reply('chapter_already_completed')
    end
end

-- Check previous chapter (if not chapter 1)
if chapter > 1 then
    local prev_found = false
    for _, c in ipairs(sess.ChaptersCompleted) do
        if c == (chapter - 1) then
            prev_found = true
            break
        end
    end
    if not prev_found then
        return redis.error_reply('previous_chapter_not_completed')
    end
end

-- Add chapter to completed list
table.insert(sess.ChaptersCompleted, chapter)

-- Update session
local updated = cjson.encode(sess)
redis.call('SET', key, updated, 'EX', ttl)

return 'OK'
`
    
    cmd := r.client.B().Eval().
        Script(script).
        Numkeys(1).
        Key(key).
        Arg(fmt.Sprintf("%d", chapter)).
        Arg(fmt.Sprintf("%d", int(r.ttl.Seconds()))).
        Arg(fmt.Sprintf("%d", domain.TotalChapters)).
        Build()
    
    result := r.client.Do(ctx, cmd)
    
    // Check for Redis errors (from redis.error_reply)
    if err := result.Error(); err != nil {
        errStr := err.Error()
        
        // Parse error messages
        if strings.Contains(errStr, "session_not_found") {
            return domain.ErrSessionNotFound
        }
        if strings.Contains(errStr, "invalid_chapter") {
            return domain.ErrInvalidChapter
        }
        if strings.Contains(errStr, "chapter_already_completed") {
            return domain.ErrChapterAlreadyCompleted
        }
        if strings.Contains(errStr, "previous_chapter_not_completed") {
            return domain.ErrPreviousChapterNotCompleted
        }
        
        return err
    }
    
    // Check result is 'OK'
    resultStr, err := result.ToString()
    if err != nil {
        return err
    }
    
    if resultStr != "OK" {
        return fmt.Errorf("unexpected result: %s", resultStr)
    }
    
    return nil
}