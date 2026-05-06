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

	teamID, err := s.teams.CreateTeam(teamName, creatorUserID)
	if err != nil {
		return 0, err
	}

	// Add the creator as a member of the team
	err = s.teamMembers.AddTeamMember(teamID, creatorUserID, "owner")
	if err != nil {
		return 0, fmt.Errorf("Error: Adding team creator as owner failed: %w", err)
	}

	// Add additional members to the team
	for _, member := range members {
		user, err := s.users.GetUserByUsername(member.MemberName)
		if err != nil {
			return 0, fmt.Errorf("Error: User not found: %w", err)
		}
		err = s.teamMembers.AddTeamMember(teamID, user.UserID, "member")
		if err != nil {
			return 0, fmt.Errorf("Error: error when adding new team members: %w", err)
		}
	}

	return teamID, nil
}

func (s *TeamManagementService) AddMemberByName(teamName string, actorUserID string, memberName string) (model.Team, error) {
	team, err := s.teams.GetTeamByName(teamName)
	if err != nil {
		return model.Team{}, err
	}

	// Check ownership: only the owner can manage the team
	if team.OwnerID != actorUserID {
		return model.Team{}, errors.New("Error: Only the team owner can add members")
	}

	user, err := s.users.GetUserByUsername(memberName)
	if err != nil {
		return model.Team{}, err
	}

	err = s.teamMembers.AddTeamMember(team.TeamID, user.UserID, "member")
	if err != nil {
		return model.Team{}, fmt.Errorf("Error: Failed to add member: %w", err)
	}

	return team, nil
}

func (s *TeamManagementService) RemoveMemberByName(teamName string, actorUserID string, memberName string) error {
	team, err := s.teams.GetTeamByName(teamName)
	if err != nil {
		return err
	}

	// Check ownership: only the owner can manage the team
	if team.OwnerID != actorUserID {
		return errors.New("Error: Only the team owner can remove members")
	}

	user, err := s.users.GetUserByUsername(memberName)
	if err != nil {
		return err
	}

	// Check if the member is in the team
	_, err = s.teamMembers.GetTeamMemberRole(team.TeamID, user.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("Error: User is not a member of this team")
		}
		return err
	}

	err = s.teamMembers.RemoveTeamMember(team.TeamID, user.UserID)
	if err != nil {
		return fmt.Errorf("Error: Failed to remove member: %w", err)
	}

	return nil
}

func (tm *TeamManagementService) DeleteTeam(teamName string, actorUserID string) error {
	team, err := tm.teams.GetTeamByName(teamName)
	if err != nil {
		return fmt.Errorf("Error: Team not found: %w", err)
	}

	// Check ownership: only the owner can delete the team
	if team.OwnerID != actorUserID {
		return errors.New("Error: Only the team owner can delete the team")
	}

	err = tm.teams.DeleteTeamByName(teamName)
	if err != nil {
		return fmt.Errorf("Failed to delete team: %w", err)
	}

	return nil
}
