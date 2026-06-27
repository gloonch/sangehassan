package usecase

import (
	"context"
	"strings"

	"sangehassan/back/internal/domain"
	"sangehassan/back/internal/ports"
)

type TeamMemberService struct {
	repo ports.TeamMemberRepository
}

func NewTeamMemberService(repo ports.TeamMemberRepository) *TeamMemberService {
	return &TeamMemberService{repo: repo}
}

func (s *TeamMemberService) ListPublic(ctx context.Context) ([]domain.TeamMember, error) {
	return s.repo.List(ctx, true)
}

func (s *TeamMemberService) List(ctx context.Context) ([]domain.TeamMember, error) {
	return s.repo.List(ctx, false)
}

func (s *TeamMemberService) GetByID(ctx context.Context, id int64) (domain.TeamMember, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *TeamMemberService) Create(ctx context.Context, member domain.TeamMember) (domain.TeamMember, error) {
	member = normalizeTeamMember(member)
	return s.repo.Create(ctx, member)
}

func (s *TeamMemberService) Update(ctx context.Context, member domain.TeamMember) (domain.TeamMember, error) {
	member = normalizeTeamMember(member)
	return s.repo.Update(ctx, member)
}

func (s *TeamMemberService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func normalizeTeamMember(member domain.TeamMember) domain.TeamMember {
	member.NameEN = strings.TrimSpace(member.NameEN)
	member.NameFA = strings.TrimSpace(member.NameFA)
	member.NameAR = strings.TrimSpace(member.NameAR)
	member.RoleEN = strings.TrimSpace(member.RoleEN)
	member.RoleFA = strings.TrimSpace(member.RoleFA)
	member.RoleAR = strings.TrimSpace(member.RoleAR)
	member.BioEN = strings.TrimSpace(member.BioEN)
	member.BioFA = strings.TrimSpace(member.BioFA)
	member.BioAR = strings.TrimSpace(member.BioAR)
	member.PhotoURL = strings.TrimSpace(member.PhotoURL)
	member.LinkedInURL = strings.TrimSpace(member.LinkedInURL)

	if member.NameEN == "" {
		if member.NameFA != "" {
			member.NameEN = member.NameFA
		} else {
			member.NameEN = member.NameAR
		}
	}
	if member.NameFA == "" {
		member.NameFA = member.NameEN
	}
	if member.NameAR == "" {
		member.NameAR = member.NameEN
	}

	if member.RoleEN == "" {
		if member.RoleFA != "" {
			member.RoleEN = member.RoleFA
		} else {
			member.RoleEN = member.RoleAR
		}
	}
	if member.RoleFA == "" {
		member.RoleFA = member.RoleEN
	}
	if member.RoleAR == "" {
		member.RoleAR = member.RoleEN
	}

	if member.BioEN == "" {
		if member.BioFA != "" {
			member.BioEN = member.BioFA
		} else {
			member.BioEN = member.BioAR
		}
	}
	if member.BioFA == "" {
		member.BioFA = member.BioEN
	}
	if member.BioAR == "" {
		member.BioAR = member.BioEN
	}

	return member
}
