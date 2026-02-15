// Package requestctx carries request-scoped metadata shared across middleware layers.
package requestctx

import "context"

type contextKey string

const (
	identityKey contextKey = "auth_identity"
)

// Identity contains normalized authenticated subject metadata for downstream policy checks.
type Identity struct {
	SubjectKeyID  string
	SubjectSource string
	TenantID      string
	Authenticated bool
	SubjectTags   []string
}

// WithIdentity stores normalized identity metadata in context.
func WithIdentity(ctx context.Context, identity *Identity) context.Context {
	if identity == nil {
		return context.WithValue(ctx, identityKey, Identity{})
	}
	return context.WithValue(ctx, identityKey, *identity)
}

// IdentityFromContext returns normalized identity metadata if present.
func IdentityFromContext(ctx context.Context) Identity {
	if ctx == nil {
		return Identity{}
	}
	identity, ok := ctx.Value(identityKey).(Identity)
	if !ok {
		return Identity{}
	}
	return identity
}
