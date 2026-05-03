package service

import (
	"backend/internal/model"
	"backend/internal/respository"
	"backend/package/dtorequest"
	"database/sql"
	"errors"
	"fmt"
)

var ErrInvalidTeamRole = errors.New("Error: role must be member or manager")
var ErrNotTeamOwner = errors.New("Error: only the team owner can assign roles")

type TeamManagementService struct {
	teams       *respository.TeamRepository
	teamMembers *respository.TeamMemberRepository
	users       *respository.UserRepository
}

func NewTeamManagementService(teams *respository.TeamRepository, teamMembers *respository.TeamMemberRepository, users *respository.UserRepository) *TeamManagementService {
	return &TeamManagementService{
		teams:       teams,
		teamMembers: teamMembers,
		users:       users,
	}
}

func (s *TeamManagementService) CreateTeam(teamName, creatorUserID string, members []dtorequest.MemberRequest) (int, error) {

	team, err := s.teams.GetTeamByName(teamName)
	if err == nil {
		return team.TeamID, errors.New("Error: Team already exists")
	}

	teamID, err := s.teams.CreateTeam(teamName)
	if err != nil {
		return 0, err
	}

	// Add the creator as a member of the team with the role "owner"
	err = s.teamMembers.AddTeamMember(teamID, creatorUserID, "owner")
	if err != nil {
		return 0, errors.New("Error: Adding team creator failed")
	}

	// Add additional members to the team with their specified roles
	for _, member := range members {
		if member.Role != "member" && member.Role != "manager" {
			return 0, ErrInvalidTeamRole
		}
		err = s.teamMembers.AddTeamMember(teamID, member.UserID, member.Role)
		if err != nil {
			return 0, errors.New("Error: Adding team members failed")
		}
	}

	return teamID, nil
}

func (s *TeamManagementService) AddMemberByName(teamName string, actorUserID string, memberName string, role string) (model.Team, error) {
	team, err := s.teams.GetTeamByName(teamName)
	if err != nil {
		return model.Team{}, err
	}

	if actorUserID == "" {
		return model.Team{}, errors.New("Error: Actor user ID is required")
	}

	actorRole, err := s.teamMembers.GetTeamMemberRole(team.TeamID, actorUserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Team{}, errors.New("Error: Actor is not a member of this team")
		}
		return model.Team{}, err
	}

	// Only managers and owners can add members
	if actorRole != "manager" && actorRole != "owner" {
		return model.Team{}, errors.New("Error: Only managers and owners can add members")
	}

	// Only owners can add managers
	if role == "manager" && actorRole != "owner" {
		return model.Team{}, errors.New("Error: Only the team owner can add managers")
	}

	if role != "member" && role != "manager" {
		return model.Team{}, ErrInvalidTeamRole
	}

	user, err := s.users.GetUserByUsername(memberName)
	if err != nil {
		return model.Team{}, err
	}
	userID := user.UserID

	// Check if the member is already in the team
	_, err = s.teamMembers.GetTeamMemberRole(team.TeamID, userID)
	if err == nil {
		return model.Team{}, errors.New("Error: User is already a member of this team")
	}

	addErr := s.teamMembers.AddTeamMember(team.TeamID, userID, role)
	if addErr != nil {
		return model.Team{}, errors.New("Error: Adding team members failed")
	}

	return team, nil
}

func (s *TeamManagementService) RemoveMemberByName(teamName string, actorUserID string, memberName string) error {
	team, err := s.teams.GetTeamByName(teamName)
	if err != nil {
		return err
	}

	if actorUserID == "" {
		return errors.New("Error: Actor user ID is required")
	}

	actorRole, err := s.teamMembers.GetTeamMemberRole(team.TeamID, actorUserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("Error: Actor is not a member of this team")
		}
		return err
	}

	// Only managers and owners can remove members
	if actorRole != "manager" && actorRole != "owner" {
		return errors.New("Error: Only managers and owners can remove members")
	}

	user, err := s.users.GetUserByUsername(memberName)
	if err != nil {
		return err
	}
	userID := user.UserID

	// Check if the member is in the team
	memberRole, err := s.teamMembers.GetTeamMemberRole(team.TeamID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("Error: User is not a member of this team")
		}
		return err
	}

	// Owners cannot be removed, or only owner can remove managers?
	// According to requirements, only main manager can add/remove other managers, but for removal, perhaps similar.
	// To keep simple, allow managers to remove members, but only owner can remove managers.
	if memberRole == "manager" && actorRole != "owner" {
		return errors.New("Error: Only the team owner can remove managers")
	}

	// Cannot remove owner
	if memberRole == "owner" {
		return errors.New("Error: Cannot remove the team owner")
	}

	err = s.teamMembers.RemoveTeamMember(team.TeamID, userID)
	if err != nil {
		return errors.New("Error: Removing team member failed")
	}

	return nil
}

func (tm *TeamManagementService) DeleteTeam(teamName string, actorUserID string) error {
	team, err := tm.teams.GetTeamByName(teamName)
	if err != nil {
		return fmt.Errorf("Error: Team not found: %w", err)
	}

	if actorUserID == "" {
		return errors.New("Error: Actor user ID is required")
	}

	actorRole, err := tm.teamMembers.GetTeamMemberRole(team.TeamID, actorUserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("Error: Actor is not a member of this team")
		}
		return err
	}

	if actorRole != "owner" {
		return errors.New("Error: Only the team owner can delete the team")
	}

	err = tm.teams.DeleteTeamByName(teamName)
	if err != nil {
		return fmt.Errorf("Failed to delete team: %w", err)
	}

	return nil
}
