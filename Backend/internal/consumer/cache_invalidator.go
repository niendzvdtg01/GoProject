package consumer

import (
	"context"

	"backend/internal/cache"
	"backend/internal/rabbitmq"
	"backend/internal/repository"
	"backend/package/event"

	"github.com/rs/zerolog"
)

// cacheInvalidatorQueues split team and asset events into separate queues so
// either side can be scaled / acked independently.
const (
	cacheTeamQueue  = "cache.invalidate.team"
	cacheAssetQueue = "cache.invalidate.asset"
)

// NewCacheInvalidator wires two consumers that listen to every team.* and
// asset.* event and drop the corresponding cache entries. This is what gives
// us cross-instance consistency: the publishing instance has already
// invalidated locally, every other instance picks up the event here.
//
// Cache writes from this consumer use Delete (not Set). Repopulation happens
// lazily on the next read so we never marshal full DB rows in the broker path.
func NewCacheInvalidator(broker rabbitmq.RabbitMQService, teamCache *cache.TeamMembers, assetCache *cache.AssetMetadata, aclCache *cache.ACL, logger *zerolog.Logger) []*Consumer {
	return []*Consumer{
		New(broker, logger, "cache/team", event.TopicTeamActivity,
			rabbitmq.TopicBinding{Queue: cacheTeamQueue, RoutingKeys: []string{"team.#"}},
			teamCacheHandler(teamCache, logger),
		),
		New(broker, logger, "cache/asset", event.TopicAssetChanges,
			rabbitmq.TopicBinding{Queue: cacheAssetQueue, RoutingKeys: []string{"asset.#"}},
			assetCacheHandler(assetCache, aclCache, logger),
		),
	}
}

func teamCacheHandler(teamCache *cache.TeamMembers, logger *zerolog.Logger) func(context.Context, event.Event) error {
	return func(ctx context.Context, evt event.Event) error {
		teamID, ok := intFromPayload(evt.Payload, "team_id")
		if !ok {
			return nil // event missing team_id — nothing to invalidate
		}
		if err := teamCache.Invalidate(ctx, teamID); err != nil {
			logger.Error().Err(err).Int("team_id", teamID).Msg("cache: team invalidate failed")
			return err
		}
		return nil
	}
}

func assetCacheHandler(assetCache *cache.AssetMetadata, aclCache *cache.ACL, logger *zerolog.Logger) func(context.Context, event.Event) error {
	return func(ctx context.Context, evt event.Event) error {
		assetType, assetID, ok := assetFromEvent(evt)
		if !ok {
			return nil
		}
		// Metadata invalidation also clears the ACL namespace for that asset
		// (AssetMetadata.Invalidate handles both), so we only call the granular
		// ACL invalidator on share/revoke where the payload names a single user.
		if grantee, hasUser := evt.Payload["granted_to"].(string); hasUser {
			if err := aclCache.InvalidateUser(ctx, assetType, assetID, grantee); err != nil {
				logger.Error().Err(err).Str("asset_type", assetType).Int("asset_id", assetID).Msg("cache: acl invalidate failed")
				return err
			}
			return nil
		}
		if err := assetCache.Invalidate(ctx, assetType, assetID); err != nil {
			logger.Error().Err(err).Str("asset_type", assetType).Int("asset_id", assetID).Msg("cache: asset invalidate failed")
			return err
		}
		return nil
	}
}

// assetFromEvent extracts (assetType, assetID) from an asset event payload.
// Note events expose note_id, folder events expose folder_id; share events
// expose asset_type + asset_id explicitly. The order of the checks reflects
// the most-specific source first.
func assetFromEvent(evt event.Event) (string, int, bool) {
	if t, ok := evt.Payload["asset_type"].(string); ok {
		if id, ok := intFromPayload(evt.Payload, "asset_id"); ok {
			return t, id, true
		}
	}
	if id, ok := intFromPayload(evt.Payload, "note_id"); ok {
		return repository.AssetTypeNote, id, true
	}
	if id, ok := intFromPayload(evt.Payload, "folder_id"); ok {
		return repository.AssetTypeFolder, id, true
	}
	return "", 0, false
}

// intFromPayload handles the JSON round-trip: numeric payload fields come
// back as float64 once the event has gone through the broker.
func intFromPayload(payload map[string]any, key string) (int, bool) {
	v, ok := payload[key]
	if !ok {
		return 0, false
	}
	switch n := v.(type) {
	case int:
		return n, true
	case int64:
		return int(n), true
	case float64:
		return int(n), true
	}
	return 0, false
}
