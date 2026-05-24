package event

import "context"

// NoopPublisher is the safe default used when the broker is disabled (e.g.
// unit tests). All methods are no-ops so services can call the publisher
// unconditionally without nil checks.
type NoopPublisher struct{}

func NewNoopPublisher() Publisher { return &NoopPublisher{} }

func (NoopPublisher) PublishTeamEvent(_ context.Context, _, _ string, _ map[string]any) {
}

func (NoopPublisher) PublishAssetEvent(_ context.Context, _, _ string, _ map[string]any) {
}
