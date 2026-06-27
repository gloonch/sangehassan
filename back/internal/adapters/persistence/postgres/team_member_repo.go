package postgres

import (
	"context"
	"database/sql"

	"sangehassan/back/internal/domain"
)

type TeamMemberRepository struct {
	db *sql.DB
}

func NewTeamMemberRepository(db *sql.DB) *TeamMemberRepository {
	return &TeamMemberRepository{db: db}
}

func (r *TeamMemberRepository) List(ctx context.Context, activeOnly bool) ([]domain.TeamMember, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id,
		       name_en, name_fa, name_ar,
		       role_en, role_fa, role_ar,
		       COALESCE(bio_en, ''), COALESCE(bio_fa, ''), COALESCE(bio_ar, ''),
		       COALESCE(photo_url, ''), COALESCE(linkedin_url, ''),
		       order_index, is_active,
		       created_at, COALESCE(updated_at, created_at)
		FROM team_members
		WHERE ($1 = FALSE OR is_active = TRUE)
		ORDER BY order_index ASC, id ASC
	`, activeOnly)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	members := make([]domain.TeamMember, 0)
	for rows.Next() {
		var member domain.TeamMember
		if err := rows.Scan(
			&member.ID,
			&member.NameEN,
			&member.NameFA,
			&member.NameAR,
			&member.RoleEN,
			&member.RoleFA,
			&member.RoleAR,
			&member.BioEN,
			&member.BioFA,
			&member.BioAR,
			&member.PhotoURL,
			&member.LinkedInURL,
			&member.OrderIndex,
			&member.IsActive,
			&member.CreatedAt,
			&member.UpdatedAt,
		); err != nil {
			return nil, err
		}
		members = append(members, member)
	}
	return members, rows.Err()
}

func (r *TeamMemberRepository) GetByID(ctx context.Context, id int64) (domain.TeamMember, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id,
		       name_en, name_fa, name_ar,
		       role_en, role_fa, role_ar,
		       COALESCE(bio_en, ''), COALESCE(bio_fa, ''), COALESCE(bio_ar, ''),
		       COALESCE(photo_url, ''), COALESCE(linkedin_url, ''),
		       order_index, is_active,
		       created_at, COALESCE(updated_at, created_at)
		FROM team_members
		WHERE id = $1
	`, id)

	var member domain.TeamMember
	if err := row.Scan(
		&member.ID,
		&member.NameEN,
		&member.NameFA,
		&member.NameAR,
		&member.RoleEN,
		&member.RoleFA,
		&member.RoleAR,
		&member.BioEN,
		&member.BioFA,
		&member.BioAR,
		&member.PhotoURL,
		&member.LinkedInURL,
		&member.OrderIndex,
		&member.IsActive,
		&member.CreatedAt,
		&member.UpdatedAt,
	); err != nil {
		return domain.TeamMember{}, err
	}
	return member, nil
}

func (r *TeamMemberRepository) Create(ctx context.Context, member domain.TeamMember) (domain.TeamMember, error) {
	row := r.db.QueryRowContext(ctx, `
		INSERT INTO team_members (
			name_en, name_fa, name_ar,
			role_en, role_fa, role_ar,
			bio_en, bio_fa, bio_ar,
			photo_url, linkedin_url,
			order_index, is_active
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, created_at, COALESCE(updated_at, created_at)
	`,
		member.NameEN,
		member.NameFA,
		member.NameAR,
		member.RoleEN,
		member.RoleFA,
		member.RoleAR,
		nullableString(member.BioEN),
		nullableString(member.BioFA),
		nullableString(member.BioAR),
		nullableString(member.PhotoURL),
		nullableString(member.LinkedInURL),
		member.OrderIndex,
		member.IsActive,
	)
	if err := row.Scan(&member.ID, &member.CreatedAt, &member.UpdatedAt); err != nil {
		return domain.TeamMember{}, err
	}
	return member, nil
}

func (r *TeamMemberRepository) Update(ctx context.Context, member domain.TeamMember) (domain.TeamMember, error) {
	row := r.db.QueryRowContext(ctx, `
		UPDATE team_members
		SET name_en = $1,
		    name_fa = $2,
		    name_ar = $3,
		    role_en = $4,
		    role_fa = $5,
		    role_ar = $6,
		    bio_en = $7,
		    bio_fa = $8,
		    bio_ar = $9,
		    photo_url = $10,
		    linkedin_url = $11,
		    order_index = $12,
		    is_active = $13,
		    updated_at = NOW()
		WHERE id = $14
		RETURNING created_at, COALESCE(updated_at, created_at)
	`,
		member.NameEN,
		member.NameFA,
		member.NameAR,
		member.RoleEN,
		member.RoleFA,
		member.RoleAR,
		nullableString(member.BioEN),
		nullableString(member.BioFA),
		nullableString(member.BioAR),
		nullableString(member.PhotoURL),
		nullableString(member.LinkedInURL),
		member.OrderIndex,
		member.IsActive,
		member.ID,
	)
	if err := row.Scan(&member.CreatedAt, &member.UpdatedAt); err != nil {
		return domain.TeamMember{}, err
	}
	return member, nil
}

func (r *TeamMemberRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM team_members WHERE id = $1`, id)
	return err
}
