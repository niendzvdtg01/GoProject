package service

import (
	"backend/internal/model"
	"backend/internal/respository"
	"backend/package/dtorequest"
	"database/sql"
	"errors"
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
		return model.Team{}, ErrNotTeamOwner
	}

	actorRole, err := s.teamMembers.GetTeamMemberRole(team.TeamID, actorUserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Team{}, ErrNotTeamOwner
		}
		return model.Team{}, err
	}
	if actorRole != "owner" {
		return model.Team{}, ErrNotTeamOwner
	}

	if role != "member" && role != "manager" {
		return model.Team{}, ErrInvalidTeamRole
	}

	user, err := s.users.GetUserByUsername(memberName)
	if err != nil {
		return model.Team{}, err
	}
	userID := user.UserID

	addErr := s.teamMembers.AddTeamMember(team.TeamID, userID, role)
	if addErr != nil {
		return model.Team{}, errors.New("Error: Adding team members failed")
	}

	return team, nil
}
