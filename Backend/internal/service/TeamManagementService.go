package service

import (
	"backend/internal/model"
	"backend/internal/repository"
	"backend/package/dtorequest"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

var ErrInvalidTeamRole = errors.New("error: role must be member or manager")
var ErrNotTeamOwner = errors.New("error: only the team owner can assign roles")

const (
	roleOwner   = "OWNER"
	roleManager = "MANAGER"
	roleMember  = "MEMBER"
)

type TeamManagementService struct {
	teams       *repository.TeamRepository
	teamMembers *repository.TeamMemberRepository
	users       *repository.UserRepository
}

func NewTeamManagementService(teams *repository.TeamRepository, teamMembers *repository.TeamMemberRepository, users *repository.UserRepository) *TeamManagementService {
	return &TeamManagementService{
		teams:       teams,
		teamMembers: teamMembers,
		users:       users,
	}
}

// CreateTeam creates a team, assigns the caller as OWNER, and optionally adds initial members resolved by username.
func (s *TeamManagementService) CreateTeam(teamName, creatorUserID string, members []dtorequest.MemberRequest) (int, error) {
	_, err := s.teams.GetTeamByName(teamName)
	if err == nil {
		return 0, errors.New("error: team already exists")
	}

	teamID, err := s.teams.CreateTeam(teamName)
	if err != nil {
		return 0, err
	}

	if err = s.teamMembers.AddTeamMember(teamID, creatorUserID, roleOwner); err != nil {
		return 0, fmt.Errorf("error adding team creator as owner: %w", err)
	}

	for _, m := range members {
		user, err := s.users.GetUserByUsername(m.MemberName)
		if err != nil {
			return 0, fmt.Errorf("error user not found: %w", err)
		}
		err = s.teamMembers.AddTeamMember(teamID, user.UserID, m.Role)
		if err != nil {
			return 0, fmt.Errorf("error adding team member %s: %w", m.MemberName, err)
		}
	}

	return teamID, nil
}

// AddMemberByName adds a user to a team; requires the actor to be OWNER or MANAGER, re-checked from DB to avoid stale JWT claims.
func (s *TeamManagementService) AddMemberByName(teamName string, actorUserID string, memberName string, role string) (model.Team, error) {
	team, err := s.teams.GetTeamByName(teamName)
	if err != nil {
		return model.Team{}, err
	}

	actorRole, err := s.teamMembers.GetTeamMemberRole(team.TeamID, actorUserID)
	if err != nil {
		return model.Team{}, fmt.Errorf("error user is not a member of the team: %w", err)
	}

	if strings.EqualFold(actorRole, roleMember) {
		return model.Team{}, errors.New("error: only the team owner and managers can add members")
	}

	user, err := s.users.GetUserByUsername(memberName)
	if err != nil {
		fmt.Print(err)
		return model.Team{}, err
	}

	err = s.teamMembers.AddTeamMember(team.TeamID, user.UserID, role)
	if err != nil {
		return model.Team{}, fmt.Errorf("error failed to add member: %w", err)
	}

	return team, nil
}

// RemoveMemberByName removes a team member with a three-tier hierarchy: OWNER > MANAGER > MEMBER; the OWNER cannot be removed via this path.
func (s *TeamManagementService) RemoveMemberByName(teamName string, actorUserID string, memberName string) error {
	team, err := s.teams.GetTeamByName(teamName)
	if err != nil {
		return fmt.Errorf("error team not found: %w", err)
	}

	actorRole, err := s.teamMembers.GetTeamMemberRole(team.TeamID, actorUserID)
	if err != nil {
		return fmt.Errorf("error user is not a member of the team: %w", err)
	}

	if strings.EqualFold(actorRole, roleMember) {
		return errors.New("error: only the team owner and managers can remove members")
	}

	user, err := s.users.GetUserByUsername(memberName)
	if err != nil {
		return err
	}

	memberRole, err := s.teamMembers.GetTeamMemberRole(team.TeamID, user.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("error: user is not a member of this team")
		}
		return err
	}

	// Only OWNER can remove a MANAGER; prevents privilege-escalation by collusion.
	if strings.EqualFold(memberRole, roleManager) && !strings.EqualFold(actorRole, roleOwner) {
		return errors.New("error: only the team owner can remove a manager")
	}

	if strings.EqualFold(memberRole, roleOwner) {
		return errors.New("error: team owner cannot be removed")
	}

	err = s.teamMembers.RemoveTeamMember(team.TeamID, user.UserID)
	if err != nil {
		return fmt.Errorf("error failed to remove member: %w", err)
	}

	return nil
}

// DeleteTeam permanently removes a team; only the OWNER may do so.
func (s *TeamManagementService) DeleteTeam(teamName string, actorUserID string) error {
	team, teamErr := s.teams.GetTeamByName(teamName)
	if teamErr != nil {
		return fmt.Errorf("Error: can not find the team: %w", teamErr)
	}

	teamMember, err := s.teamMembers.GetTeamMemberRole(team.TeamID, actorUserID)
	if err != nil {
		return fmt.Errorf("error team not found: %w", err)
	}

	if teamMember != roleOwner {
		fmt.Println(teamMember)
		return errors.New("error: only the team owner can delete the team")
	}

	err = s.teams.DeleteTeamByName(teamName)
	if err != nil {
		return fmt.Errorf("failed to delete team: %w", err)
	}

	return nil
}

func (s *TeamManagementService) ListTeamsForUser(userID string) ([]model.Team, error) {
	return s.teams.GetTeamsByUserID(userID)
}