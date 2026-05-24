package event

import "time"

// Topic exchanges. Consumers bind queues here with routing-key patterns.
const (
	TopicTeamActivity = "team.activity"
	TopicAssetChanges = "asset.changes"
)

// Team event types and their topic-exchange routing keys.
const (
	TeamCreated    = "TEAM_CREATED"
	MemberAdded    = "MEMBER_ADDED"
	MemberRemoved  = "MEMBER_REMOVED"
	ManagerAdded   = "MANAGER_ADDED"
	ManagerRemoved = "MANAGER_REMOVED"
	TeamDeleted    = "TEAM_DELETED"
)

// Asset event types.
const (
	FolderCreated = "FOLDER_CREATED"
	FolderUpdated = "FOLDER_UPDATED"
	FolderDeleted = "FOLDER_DELETED"
	FolderShared  = "FOLDER_SHARED"
	NoteCreated   = "NOTE_CREATED"
	NoteUpdated   = "NOTE_UPDATED"
	NoteDeleted   = "NOTE_DELETED"
	NoteShared    = "NOTE_SHARED"
)

// teamRoutingKeys / assetRoutingKeys map an event type to its dotted routing
// key. Routing keys are hierarchical so consumers can subscribe broadly
// ("team.#", "asset.folder.*") or narrowly ("asset.note.shared").
var teamRoutingKeys = map[string]string{
	TeamCreated:    "team.created",
	MemberAdded:    "team.member.added",
	MemberRemoved:  "team.member.removed",
	ManagerAdded:   "team.manager.added",
	ManagerRemoved: "team.manager.removed",
	TeamDeleted:    "team.deleted",
}

var assetRoutingKeys = map[string]string{
	FolderCreated: "asset.folder.created",
	FolderUpdated: "asset.folder.updated",
	FolderDeleted: "asset.folder.deleted",
	FolderShared:  "asset.folder.shared",
	NoteCreated:   "asset.note.created",
	NoteUpdated:   "asset.note.updated",
	NoteDeleted:   "asset.note.deleted",
	NoteShared:    "asset.note.shared",
}

// RoutingKey returns the routing key for an event type. Falls back to a
// lowercased event-type so unknown future events still route under their topic.
func RoutingKey(topic, eventType string) string {
	switch topic {
	case TopicTeamActivity:
		if rk, ok := teamRoutingKeys[eventType]; ok {
			return rk
		}
	case TopicAssetChanges:
		if rk, ok := assetRoutingKeys[eventType]; ok {
			return rk
		}
	}
	return eventType
}

type Event struct {
	Topic       string         `json:"topic"`
	EventType   string         `json:"event_type"`
	RoutingKey  string         `json:"routing_key"`
	PerformedBy string         `json:"performed_by"`
	Timestamp   time.Time      `json:"timestamp"`
	Payload     map[string]any `json:"payload"`
}
