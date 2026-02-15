package policy

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type rateBucket struct {
	WindowStart time.Time
	Used        int64
}

// MemoryQuotaStore is internal-first quota storage with per-process state.
type MemoryQuotaStore struct {
	mu          sync.Mutex
	rate        map[string]rateBucket
	tokens      map[string]rateBucket
	concurrency map[string]int64
}

// NewMemoryQuotaStore creates an in-memory quota store.
func NewMemoryQuotaStore() *MemoryQuotaStore {
	return &MemoryQuotaStore{
		rate:        make(map[string]rateBucket),
		tokens:      make(map[string]rateBucket),
		concurrency: make(map[string]int64),
	}
}

// Reserve reserves concurrency quota units for a specific key.
func (s *MemoryQuotaStore) Reserve(_ context.Context, req QuotaRequest) (QuotaResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if req.Kind != QuotaKindConcurrency {
		return QuotaResult{}, fmt.Errorf("reserve only supports concurrency kind")
	}
	if req.Limit <= 0 {
		return QuotaResult{Allowed: true}, nil
	}

	current := s.concurrency[req.Key]
	if current >= req.Limit {
		return QuotaResult{Allowed: false, Remaining: 0}, nil
	}

	s.concurrency[req.Key] = current + req.Amount
	remaining := req.Limit - s.concurrency[req.Key]
	if remaining < 0 {
		remaining = 0
	}
	return QuotaResult{Allowed: true, Remaining: remaining}, nil
}

// Consume consumes rate or token quota units for a specific key.
func (s *MemoryQuotaStore) Consume(_ context.Context, req QuotaRequest) (QuotaResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if req.Limit <= 0 {
		return QuotaResult{Allowed: true}, nil
	}
	if req.Window <= 0 {
		req.Window = time.Minute
	}
	if req.Amount <= 0 {
		req.Amount = 1
	}

	now := time.Now().UTC()
	switch req.Kind {
	case QuotaKindRate:
		bucket := s.rate[req.Key]
		bucket = rotateWindow(now, req.Window, bucket)
		if bucket.Used+req.Amount > req.Limit {
			return QuotaResult{Allowed: false, Remaining: nonNegative(req.Limit - bucket.Used), ResetAt: bucket.WindowStart.Add(req.Window)}, nil
		}
		bucket.Used += req.Amount
		s.rate[req.Key] = bucket
		return QuotaResult{Allowed: true, Remaining: nonNegative(req.Limit - bucket.Used), ResetAt: bucket.WindowStart.Add(req.Window)}, nil
	case QuotaKindToken:
		bucket := s.tokens[req.Key]
		bucket = rotateWindow(now, req.Window, bucket)
		if bucket.Used+req.Amount > req.Limit {
			return QuotaResult{Allowed: false, Remaining: nonNegative(req.Limit - bucket.Used), ResetAt: bucket.WindowStart.Add(req.Window)}, nil
		}
		bucket.Used += req.Amount
		s.tokens[req.Key] = bucket
		return QuotaResult{Allowed: true, Remaining: nonNegative(req.Limit - bucket.Used), ResetAt: bucket.WindowStart.Add(req.Window)}, nil
	case QuotaKindConcurrency:
		return QuotaResult{}, fmt.Errorf("consume does not support concurrency kind")
	default:
		return QuotaResult{}, fmt.Errorf("unsupported consume kind %q", req.Kind)
	}
}

// Release releases previously reserved concurrency quota units.
func (s *MemoryQuotaStore) Release(_ context.Context, req QuotaRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if req.Kind != QuotaKindConcurrency {
		return nil
	}
	if req.Amount <= 0 {
		req.Amount = 1
	}

	current := s.concurrency[req.Key]
	current -= req.Amount
	if current <= 0 {
		delete(s.concurrency, req.Key)
		return nil
	}
	s.concurrency[req.Key] = current
	return nil
}

// Peek returns current quota state for the requested key and kind.
func (s *MemoryQuotaStore) Peek(_ context.Context, req QuotaRequest) (QuotaResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	switch req.Kind {
	case QuotaKindConcurrency:
		current := s.concurrency[req.Key]
		remaining := req.Limit - current
		if remaining < 0 {
			remaining = 0
		}
		return QuotaResult{Allowed: current < req.Limit, Remaining: remaining}, nil
	case QuotaKindRate:
		bucket := rotateWindow(now, req.Window, s.rate[req.Key])
		remaining := req.Limit - bucket.Used
		if remaining < 0 {
			remaining = 0
		}
		return QuotaResult{Allowed: bucket.Used < req.Limit, Remaining: remaining, ResetAt: bucket.WindowStart.Add(req.Window)}, nil
	case QuotaKindToken:
		bucket := rotateWindow(now, req.Window, s.tokens[req.Key])
		remaining := req.Limit - bucket.Used
		if remaining < 0 {
			remaining = 0
		}
		return QuotaResult{Allowed: bucket.Used < req.Limit, Remaining: remaining, ResetAt: bucket.WindowStart.Add(req.Window)}, nil
	default:
		return QuotaResult{}, fmt.Errorf("unsupported peek kind %q", req.Kind)
	}
}

// Close clears in-memory quota state.
func (s *MemoryQuotaStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rate = map[string]rateBucket{}
	s.tokens = map[string]rateBucket{}
	s.concurrency = map[string]int64{}
	return nil
}

func rotateWindow(now time.Time, window time.Duration, bucket rateBucket) rateBucket {
	if window <= 0 {
		window = time.Minute
	}
	if bucket.WindowStart.IsZero() {
		return rateBucket{WindowStart: now, Used: 0}
	}
	if now.Sub(bucket.WindowStart) >= window {
		return rateBucket{WindowStart: now, Used: 0}
	}
	return bucket
}

func nonNegative(v int64) int64 {
	if v < 0 {
		return 0
	}
	return v
}
