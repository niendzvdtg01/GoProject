package cache

import (
	"context"
	"time"
)

// Cache is the minimal contract every backend (Redis, noop, in-memory) must
// satisfy. Get returns ErrCacheMiss on miss so callers can distinguish "not
// cached" from a real error without inspecting a typed nil.
type Cache interface {
	Get(ctx context.Context, key string, dest any) error
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Delete(ctx context.Context, keys ...string) error
	// DeletePattern removes every key matching the given glob (e.g.
	// "asset:folder:42:acl:*"). Used when a write fans out to an unknown
	// number of cached entries — folder shared/revoked invalidates per-user
	// ACL entries we don't enumerate up front.
	DeletePattern(ctx context.Context, pattern string) error
}
