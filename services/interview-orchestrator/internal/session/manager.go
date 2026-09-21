package session

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/yourusername/ai-interview/services/interview-orchestrator/internal/model"
)

var (
	ErrLockFailed   = errors.New("failed to acquire interview lock, likely already in progress")
	ErrSessionNotFound = errors.New("session not found")
)

const (
	sessionPrefix = "interview:session:"
	lockPrefix    = "interview:lock:"
	sessionTTL    = 2 * time.Hour
	lockTTL       = 30 * time.Second
)

type Manager interface {
	AcquireLock(ctx context.Context, interviewID string, lockValue string) error
	RefreshLock(ctx context.Context, interviewID string, lockValue string) error
	ReleaseLock(ctx context.Context, interviewID string, lockValue string) error
	
	GetSession(ctx context.Context, interviewID string) (*model.Session, error)
	SaveSession(ctx context.Context, session *model.Session) error
	DeleteSession(ctx context.Context, interviewID string) error
}

type manager struct {
	client *redis.Client
}

func NewManager(client *redis.Client) Manager {
	return &manager{
		client: client,
	}
}

func (m *manager) lockKey(interviewID string) string {
	return lockPrefix + interviewID
}

func (m *manager) sessionKey(interviewID string) string {
	return sessionPrefix + interviewID
}

func (m *manager) AcquireLock(ctx context.Context, interviewID string, lockValue string) error {
	key := m.lockKey(interviewID)
	ok, err := m.client.SetNX(ctx, key, lockValue, lockTTL).Result()
	if err != nil {
		return fmt.Errorf("redis setnx error: %w", err)
	}
	if !ok {
		return ErrLockFailed
	}
	return nil
}

func (m *manager) RefreshLock(ctx context.Context, interviewID string, lockValue string) error {
	key := m.lockKey(interviewID)
	// Execute Lua script to only refresh if the lock value matches
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("expire", KEYS[1], ARGV[2])
		else
			return 0
		end
	`
	res, err := m.client.Eval(ctx, script, []string{key}, lockValue, int(lockTTL.Seconds())).Result()
	if err != nil {
		return fmt.Errorf("redis eval refresh error: %w", err)
	}
	if res.(int64) == 0 {
		return ErrLockFailed
	}
	return nil
}

func (m *manager) ReleaseLock(ctx context.Context, interviewID string, lockValue string) error {
	key := m.lockKey(interviewID)
	// Execute Lua script to only delete if the lock value matches
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`
	_, err := m.client.Eval(ctx, script, []string{key}, lockValue).Result()
	if err != nil {
		return fmt.Errorf("redis eval del error: %w", err)
	}
	return nil
}

func (m *manager) GetSession(ctx context.Context, interviewID string) (*model.Session, error) {
	key := m.sessionKey(interviewID)
	val, err := m.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrSessionNotFound
		}
		return nil, fmt.Errorf("redis get error: %w", err)
	}

	var sess model.Session
	if err := json.Unmarshal([]byte(val), &sess); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	return &sess, nil
}

func (m *manager) SaveSession(ctx context.Context, session *model.Session) error {
	key := m.sessionKey(session.InterviewID)
	b, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	err = m.client.Set(ctx, key, string(b), sessionTTL).Err()
	if err != nil {
		return fmt.Errorf("redis set error: %w", err)
	}

	return nil
}

func (m *manager) DeleteSession(ctx context.Context, interviewID string) error {
	key := m.sessionKey(interviewID)
	err := m.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("redis del error: %w", err)
	}
	return nil
}
