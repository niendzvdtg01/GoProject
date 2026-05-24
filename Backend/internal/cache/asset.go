package cache

import (
	"backend/internal/model"
	"context"
	"time"
)

// assetMetadataTTL covers the window between an out-of-band DB write (manual
// SQL, migration) and the next event-driven invalidation. Most writes happen
// through the service layer and refresh the cache write-through, so this TTL
// rarely matters.
const assetMetadataTTL = 15 * time.Minute

// AssetMetadata is the write-through cache for folder/note metadata. The two
// asset types share a key namespace (asset:{type}:{id}) so the same helper can
// invalidate either side, e.g. when a folder delete cascades to notes.
type AssetMetadata struct {
	cache Cache
}

func NewAssetMetadata(c Cache) *AssetMetadata { return &AssetMetadata{cache: c} }

// GetFolder / SetFolder / GetNote / SetNote are intentionally specialised
// (instead of a generic any) so callers don't need to thread the asset-type
// string everywhere — the typed call sites are easier to grep and less error
// prone.

func (a *AssetMetadata) GetFolder(ctx context.Context, folderID int) (model.Folder, bool) {
	var f model.Folder
	if err := a.cache.Get(ctx, AssetKey("folder", folderID), &f); err != nil {
		return model.Folder{}, false
	}
	return f, true
}

func (a *AssetMetadata) SetFolder(ctx context.Context, f model.Folder) error {
	return a.cache.Set(ctx, AssetKey("folder", f.ID), f, assetMetadataTTL)
}

func (a *AssetMetadata) GetNote(ctx context.Context, noteID int) (model.Note, bool) {
	var n model.Note
	if err := a.cache.Get(ctx, AssetKey("note", noteID), &n); err != nil {
		return model.Note{}, false
	}
	return n, true
}

func (a *AssetMetadata) SetNote(ctx context.Context, n model.Note) error {
	return a.cache.Set(ctx, AssetKey("note", n.ID), n, assetMetadataTTL)
}

// Invalidate removes the metadata entry AND every per-user ACL entry that
// hangs off it — a deleted folder/note can no longer authorise anyone, and a
// renamed/shared one needs fresh ACL lookups.
func (a *AssetMetadata) Invalidate(ctx context.Context, assetType string, assetID int) error {
	if err := a.cache.Delete(ctx, AssetKey(assetType, assetID)); err != nil {
		return err
	}
	return a.cache.DeletePattern(ctx, ACLPrefix(assetType, assetID))
}
