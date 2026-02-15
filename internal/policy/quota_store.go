package policy

import "context"

// QuotaStore is pluggable storage for quota reservation/consumption/release.
type QuotaStore interface {
	Reserve(ctx context.Context, req QuotaRequest) (QuotaResult, error)
	Consume(ctx context.Context, req QuotaRequest) (QuotaResult, error)
	Release(ctx context.Context, req QuotaRequest) error
	Peek(ctx context.Context, req QuotaRequest) (QuotaResult, error)
	Close() error
}
