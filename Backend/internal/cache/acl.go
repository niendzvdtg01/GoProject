package cache

import (
	"context"
	"time"
)

// aclTTL is short because ACL changes (share, revoke, role change) flow through
// the event bus and a missed message could leave a stale grant in cache.
// Five minutes is the longest a revoked user keeps post-revocation access.
const aclTTL = 5 * time.Minute

// PermissionNone is the sentinel cached when the DB returns "no permission".
// Negative caching means denied probes don't fall through to the DB on every
// hit — important for the auth-check path which fires on most asset reads.
const PermissionNone = "none"

// ACL caches the (asset, user) → permission_type triple. The cached string is
// the permission_type directly ("read", "write", or PermissionNone for the
// negative case), so callers don't pay JSON-decode cost on every check.
type ACL struct {
	cache Cache
}

func NewACL(c Cache) *ACL { return &ACL{cache: c} }

// Get returns the cached permission and whether the entry was a hit.
// permission == PermissionNone means we know the user has no permission and
// the caller can short-circuit without touching the DB.
func (a *ACL) Get(ctx context.Context, assetType string, assetID int, userID string) (string, bool) {
	var perm string
	if err := a.cache.Get(ctx, ACLKey(assetType, assetID, userID), &perm); err != nil {
		return "", false
	}
	return perm, true
}

func (a *ACL) Set(ctx context.Context, assetType string, assetID int, userID, permission string) error {
	return a.cache.Set(ctx, ACLKey(assetType, assetID, userID), permission, aclTTL)
}

// InvalidateUser clears the single (asset, user) entry. Used when a share/revoke
// targets one user — no need to wipe the whole asset's ACL namespace.
func (a *ACL) InvalidateUser(ctx context.Context, assetType string, assetID int, userID string) error {
	return a.cache.Delete(ctx, ACLKey(assetType, assetID, userID))
}

// InvalidateAsset wipes every per-user ACL entry for an asset. Used for
// fan-out changes: folder delete (cascades to inherited note grants) or any
// case where we don't know the affected user IDs up front.
func (a *ACL) InvalidateAsset(ctx context.Context, assetType string, assetID int) error {
	return a.cache.DeletePattern(ctx, ACLPrefix(assetType, assetID))
}
