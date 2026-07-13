package usecase

import (
	"context"
	"errors"
	"net/mail"
	"strings"

	"sangehassan/back/internal/domain"
	"sangehassan/back/internal/ports"
)

var ErrContactSubmissionInvalid = errors.New("invalid contact submission")

type ContactSubmissionService struct {
	repo ports.ContactSubmissionRepository
}

func NewContactSubmissionService(repo ports.ContactSubmissionRepository) *ContactSubmissionService {
	return &ContactSubmissionService{repo: repo}
}

func (s *ContactSubmissionService) Create(ctx context.Context, submission domain.ContactSubmission) (domain.ContactSubmission, error) {
	normalized, err := normalizeContactSubmission(submission)
	if err != nil {
		return domain.ContactSubmission{}, err
	}
	return s.repo.Create(ctx, normalized)
}

func (s *ContactSubmissionService) List(ctx context.Context, limit, offset int) ([]domain.ContactSubmission, error) {
	return s.repo.List(ctx, limit, offset)
}

func (s *ContactSubmissionService) GetByID(ctx context.Context, id int64) (domain.ContactSubmission, error) {
	return s.repo.GetByID(ctx, id)
}

type countryPhoneRule struct {
	iso     string
	lengths map[int]bool
}

var countryPhoneRules = map[string]countryPhoneRule{
	"+1":   {iso: "US", lengths: map[int]bool{10: true}},
	"+7":   {iso: "RU", lengths: map[int]bool{10: true}},
	"+44":  {iso: "GB", lengths: map[int]bool{10: true}},
	"+49":  {iso: "DE", lengths: map[int]bool{10: true, 11: true}},
	"+86":  {iso: "CN", lengths: map[int]bool{11: true}},
	"+90":  {iso: "TR", lengths: map[int]bool{10: true}},
	"+91":  {iso: "IN", lengths: map[int]bool{10: true}},
	"+98":  {iso: "IR", lengths: map[int]bool{10: true}},
	"+964": {iso: "IQ", lengths: map[int]bool{10: true}},
	"+965": {iso: "KW", lengths: map[int]bool{8: true}},
	"+966": {iso: "SA", lengths: map[int]bool{9: true}},
	"+968": {iso: "OM", lengths: map[int]bool{8: true}},
	"+971": {iso: "AE", lengths: map[int]bool{9: true}},
	"+974": {iso: "QA", lengths: map[int]bool{8: true}},
}

func normalizeContactSubmission(submission domain.ContactSubmission) (domain.ContactSubmission, error) {
	submission.FullName = strings.TrimSpace(submission.FullName)
	submission.Email = strings.TrimSpace(submission.Email)
	submission.CountryCode = strings.TrimSpace(submission.CountryCode)
	submission.CountryISO = strings.ToUpper(strings.TrimSpace(submission.CountryISO))
	submission.PhoneNumber = strings.TrimSpace(submission.PhoneNumber)
	submission.Message = strings.TrimSpace(submission.Message)
	submission.Source = strings.TrimSpace(submission.Source)

	if submission.FullName == "" || len([]rune(submission.FullName)) > 160 {
		return domain.ContactSubmission{}, ErrContactSubmissionInvalid
	}
	if submission.Message == "" || len([]rune(submission.Message)) > 2000 {
		return domain.ContactSubmission{}, ErrContactSubmissionInvalid
	}
	if submission.Email != "" {
		if _, err := mail.ParseAddress(submission.Email); err != nil {
			return domain.ContactSubmission{}, ErrContactSubmissionInvalid
		}
	}

	rule, ok := countryPhoneRules[submission.CountryCode]
	if !ok {
		return domain.ContactSubmission{}, ErrContactSubmissionInvalid
	}
	if submission.CountryISO == "" {
		submission.CountryISO = rule.iso
	}
	if submission.CountryISO != rule.iso {
		return domain.ContactSubmission{}, ErrContactSubmissionInvalid
	}

	phone := strings.TrimSpace(submission.PhoneNumber)
	if phone == "" {
		return domain.ContactSubmission{}, ErrContactSubmissionInvalid
	}
	for _, char := range phone {
		if char < '0' || char > '9' {
			return domain.ContactSubmission{}, ErrContactSubmissionInvalid
		}
	}
	phone = strings.TrimLeft(phone, "0")
	if !rule.lengths[len(phone)] {
		return domain.ContactSubmission{}, ErrContactSubmissionInvalid
	}

	submission.PhoneNumber = phone
	submission.PhoneE164 = submission.CountryCode + phone
	if submission.Source == "" {
		submission.Source = "footer"
	}
	if len([]rune(submission.Source)) > 80 {
		return domain.ContactSubmission{}, ErrContactSubmissionInvalid
	}

	return submission, nil
}
