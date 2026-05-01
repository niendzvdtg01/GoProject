package service

import "backend/internal/respository"

type TeamManagementService struct {
	teams       *respository.TeamRepository
	teamMembers *respository.TeamMemberRepository
}
