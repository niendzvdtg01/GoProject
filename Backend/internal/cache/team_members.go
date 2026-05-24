package cache

import (
	"backend/internal/model"
	"context"
	"errors"
	"time"
)

// teamMembersTTL is a soft upper bound: writes invalidate explicitly, the TTL
// is the safety net for missed events (broker outage, missed publish).
const teamMembersTTL = 10 * time.Minute

// TeamMembers wraps the raw Cache with type-safe helpers for the team member
// roster. Read-through pattern: services call Get → on miss, fall back to the
// DB and store the fresh result.
type TeamMembers struct {
	cache Cache
}

func NewTeamMembers(c Cache) *TeamMembers { return &TeamMembers{cache: c} }

func (t *TeamMembers) Get(ctx context.Context, teamID int) ([]model.TeamMember, bool) {
	var members []model.TeamMember
	err := t.cache.Get(ctx, TeamMembersKey(teamID), &members)
	if err != nil {
		return nil, false
	}
	return members, true
}

func (t *TeamMembers) Set(ctx context.Context, teamID int, members []model.TeamMember) error {
	return t.cache.Set(ctx, TeamMembersKey(teamID), members, teamMembersTTL)
}

// Invalidate is called on every team.activity write so the next read repopulates.
// Errors are returned but callers typically log-and-continue: cache staleness
// must never break a write that already committed to the DB.
func (t *TeamMembers) Invalidate(ctx context.Context, teamID int) error {
	err := t.cache.Delete(ctx, TeamMembersKey(teamID))
	if err != nil && !errors.Is(err, ErrCacheMiss) {
		return err
	}
	return nil
}
