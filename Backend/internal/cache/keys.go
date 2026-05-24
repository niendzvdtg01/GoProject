package cache

import "fmt"

// Centralised key builders so the schema lives in one place. Whenever a new
// cache namespace appears, add the constructor here rather than scattering
// fmt.Sprintf calls across services.
func TeamMembersKey(teamID int) string {
	return fmt.Sprintf("team:%d:members", teamID)
}

func AssetKey(assetType string, assetID int) string {
	return fmt.Sprintf("asset:%s:%d", assetType, assetID)
}

func ACLKey(assetType string, assetID int, userID string) string {
	return fmt.Sprintf("asset:%s:%d:acl:%s", assetType, assetID, userID)
}

// ACLPrefix matches every ACL entry for an asset; used when a cascading
// permission change (folder share/revoke, asset delete) makes per-user ACL
// entries stale and we don't know the affected users up front.
func ACLPrefix(assetType string, assetID int) string {
	return fmt.Sprintf("asset:%s:%d:acl:*", assetType, assetID)
}
