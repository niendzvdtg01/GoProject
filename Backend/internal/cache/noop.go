package cache

import (
	"context"
	"time"
)

// NoopCache is the safe default when Redis is unreachable: every Get is a miss
// and every Set/Delete is dropped. Lets services call the cache unconditionally
// without nil checks or fallback branches.
type noopCache struct{}

func NewNoopCache() Cache { return &noopCache{} }

func (noopCache) Get(_ context.Context, _ string, _ any) error { return ErrCacheMiss }

func (noopCache) Set(_ context.Context, _ string, _ any, _ time.Duration) error { return nil }

func (noopCache) Delete(_ context.Context, _ ...string) error { return nil }

func (noopCache) DeletePattern(_ context.Context, _ string) error { return nil }
