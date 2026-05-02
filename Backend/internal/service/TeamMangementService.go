package service

import (
	"backend/internal/model"
	"backend/internal/respository"
	"errors"
)

type TeamManagementService struct {
	teams       *respository.TeamRepository
	teamMembers *respository.TeamMemberRepository
	users       *respository.UserRepository
}

func NewTeamManagementService(teams *respository.TeamRepository, teamMembers *respository.TeamMemberRepository) *TeamManagementService {
	return &TeamManagementService{
		teams:       teams,
		teamMembers: teamMembers,
	}
}

func (s *TeamManagementService) CreateTeam(teamName string, creatorUserID string) (int, error) {

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
		return 0, errors.New("Error: Adding team members failed")
	}

	return teamID, nil
}

func (s *TeamManagementService) AddMemberByName(teamName string, memberName string, role string) (model.Team, error) {
	team, err := s.teams.GetTeamByName(teamName)
	if err != nil {
		return model.Team{}, err
	}
	user, err := s.users.GetUserByUsername(memberName)
	if err != nil {
		return model.Team{}, err
	}
	userID := user.UserID

	// Validate role

	if role != "member" && role != "manager" {
		return model.Team{}, errors.New("Error: role must be member or manager")
	}

	addErr := s.teamMembers.AddTeamMember(team.TeamID, userID, role)
	if addErr != nil {
		return model.Team{}, errors.New("Error: Adding team members failed")
	}

	return team, nil
}
