package policy

import (
	"context"
	"testing"
	"time"
)

func TestMemoryQuotaStore_RateScopeIsolation(t *testing.T) {
	store := NewMemoryQuotaStore()
	t.Cleanup(func() { _ = store.Close() })

	ctx := context.Background()
	window := time.Minute

	firstKeyFirst, err := store.Consume(ctx, QuotaRequest{
		Kind:   QuotaKindRate,
		Key:    "scope=key-a",
		Limit:  1,
		Amount: 1,
		Window: window,
	})
	if err != nil {
		t.Fatalf("first key first consume error = %v", err)
	}
	if !firstKeyFirst.Allowed {
		t.Fatal("first key first consume should be allowed")
	}

	secondKeyFirst, err := store.Consume(ctx, QuotaRequest{
		Kind:   QuotaKindRate,
		Key:    "scope=key-b",
		Limit:  1,
		Amount: 1,
		Window: window,
	})
	if err != nil {
		t.Fatalf("second key first consume error = %v", err)
	}
	if !secondKeyFirst.Allowed {
		t.Fatal("second key first consume should be allowed")
	}

	firstKeySecond, err := store.Consume(ctx, QuotaRequest{
		Kind:   QuotaKindRate,
		Key:    "scope=key-a",
		Limit:  1,
		Amount: 1,
		Window: window,
	})
	if err != nil {
		t.Fatalf("first key second consume error = %v", err)
	}
	if firstKeySecond.Allowed {
		t.Fatal("first key second consume should be rejected by limit")
	}
}

func TestMemoryQuotaStore_ConcurrencyReserveRelease(t *testing.T) {
	store := NewMemoryQuotaStore()
	t.Cleanup(func() { _ = store.Close() })

	ctx := context.Background()
	req := QuotaRequest{
		Kind:   QuotaKindConcurrency,
		Key:    "rule=concurrency|tenant=t1",
		Limit:  1,
		Amount: 1,
	}

	first, err := store.Reserve(ctx, req)
	if err != nil {
		t.Fatalf("first reserve error = %v", err)
	}
	if !first.Allowed {
		t.Fatal("first reserve should be allowed")
	}

	second, err := store.Reserve(ctx, req)
	if err != nil {
		t.Fatalf("second reserve error = %v", err)
	}
	if second.Allowed {
		t.Fatal("second reserve should be blocked by limit")
	}

	if err := store.Release(ctx, QuotaRequest{
		Kind:   QuotaKindConcurrency,
		Key:    req.Key,
		Amount: 1,
	}); err != nil {
		t.Fatalf("release error = %v", err)
	}

	third, err := store.Reserve(ctx, req)
	if err != nil {
		t.Fatalf("third reserve error = %v", err)
	}
	if !third.Allowed {
		t.Fatal("third reserve should be allowed after release")
	}
}
