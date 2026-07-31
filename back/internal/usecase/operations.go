package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrForbidden         = errors.New("forbidden")
	ErrActivationInvalid = errors.New("invalid or expired activation code")
	ErrProtectedRole     = errors.New("protected role")
)

type OperationsService struct{ db *sql.DB }

func NewOperationsService(db *sql.DB) *OperationsService { return &OperationsService{db: db} }

type OperationalUser struct {
	ID                 string     `json:"id"`
	Phone              string     `json:"phone"`
	FirstName          string     `json:"first_name"`
	LastName           string     `json:"last_name"`
	UserType           string     `json:"user_type"`
	Status             string     `json:"status"`
	MustChangePassword bool       `json:"must_change_password"`
	LastLoginAt        *time.Time `json:"last_login_at,omitempty"`
	Roles              []string   `json:"roles"`
	Permissions        []string   `json:"permissions"`
}
type Permission struct {
	ID            int64  `json:"id"`
	Code          string `json:"code"`
	NameFA        string `json:"name_fa"`
	DescriptionFA string `json:"description_fa"`
	GroupCode     string `json:"group_code"`
}
type Role struct {
	ID              int64    `json:"id"`
	Code            string   `json:"code"`
	NameFA          string   `json:"name_fa"`
	DescriptionFA   string   `json:"description_fa"`
	IsSystem        bool     `json:"is_system"`
	IsProtected     bool     `json:"is_protected"`
	IsActive        bool     `json:"is_active"`
	UserCount       int      `json:"user_count"`
	PermissionCodes []string `json:"permission_codes,omitempty"`
}
type ActionItem struct {
	ID                 string     `json:"id"`
	WorkflowInstanceID string     `json:"workflow_instance_id"`
	TitleFA            string     `json:"title_fa"`
	DescriptionFA      string     `json:"description_fa"`
	Status             string     `json:"status"`
	Priority           string     `json:"priority"`
	DueAt              *time.Time `json:"due_at,omitempty"`
	OrderID            string     `json:"order_id"`
	OrderNumber        string     `json:"order_number"`
	Customer           string     `json:"customer"`
	WorkflowName       string     `json:"workflow_name"`
	StepName           string     `json:"step_name"`
}
type WorkflowTemplate struct {
	ID            int64                  `json:"id"`
	Code          string                 `json:"code"`
	NameFA        string                 `json:"name_fa"`
	DescriptionFA string                 `json:"description_fa"`
	IconKey       string                 `json:"icon_key"`
	VersionNumber int                    `json:"version_number"`
	OptionalSteps []WorkflowOptionalStep `json:"optional_steps"`
}
type WorkflowOptionalStep struct {
	Code  string `json:"code"`
	Title string `json:"title_fa"`
}
type WorkflowStartResult struct {
	WorkflowInstanceID string `json:"workflow_instance_id"`
	OrderID            string `json:"order_id"`
	OrderNumber        string `json:"order_number"`
	CustomerUserID     string `json:"customer_user_id"`
	CustomerCreated    bool   `json:"customer_created"`
	ActivationCode     string `json:"activation_code,omitempty"`
}
type CustomerMatch struct {
	ID          string `json:"id"`
	Phone       string `json:"phone"`
	DisplayName string `json:"display_name"`
	Status      string `json:"status"`
}
type AccountOrder struct {
	ID                 string         `json:"id"`
	WorkflowInstanceID string         `json:"workflow_instance_id"`
	OrderNumber        string         `json:"order_number"`
	Status             string         `json:"status"`
	WorkflowName       string         `json:"workflow_name"`
	CreatedAt          time.Time      `json:"created_at"`
	ProformaIssuedAt   *time.Time     `json:"proforma_issued_at,omitempty"`
	Timeline           []TimelineItem `json:"timeline,omitempty"`
}
type TimelineItem struct {
	TitleFA    string    `json:"title_fa"`
	StatusCode string    `json:"status_code"`
	OccurredAt time.Time `json:"occurred_at"`
}
type Proforma struct {
	ID             string     `json:"id"`
	ProformaNumber string     `json:"proforma_number"`
	OrderID        string     `json:"order_id"`
	Status         string     `json:"status"`
	Currency       string     `json:"currency"`
	Subtotal       string     `json:"subtotal"`
	DiscountAmount string     `json:"discount_amount"`
	TotalAmount    string     `json:"total_amount"`
	Notes          *string    `json:"notes,omitempty"`
	IssuedAt       *time.Time `json:"issued_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

func (s *OperationsService) BootstrapSuperAdmin(ctx context.Context, phone, password, first, last string) error {
	phone = NormalizePhone(phone)
	if phone == "" || password == "" {
		return nil
	}
	var exists bool
	if err := s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM user_roles ur JOIN roles r ON r.id=ur.role_id JOIN users u ON u.id=ur.user_id WHERE r.code='SUPER_ADMIN' AND u.status='ACTIVE')`).Scan(&exists); err != nil {
		return err
	}
	if exists {
		return nil
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var id string
	err = tx.QueryRowContext(ctx, `INSERT INTO users(email,password_hash,full_name,phone,phone_normalized,role,is_active,first_name,last_name,user_type,status,must_change_password) VALUES($1,$2,NULLIF(TRIM($3||' '||$4),''),$5,$5,'admin',TRUE,NULLIF($3,''),NULLIF($4,''),'INTERNAL','ACTIVE',TRUE) ON CONFLICT(phone_normalized) DO UPDATE SET password_hash=EXCLUDED.password_hash,first_name=EXCLUDED.first_name,last_name=EXCLUDED.last_name,user_type='INTERNAL',status='ACTIVE',is_active=TRUE,must_change_password=TRUE RETURNING id`, phoneAliasEmail(phone), string(hash), first, last, phone).Scan(&id)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO user_roles(user_id,role_id) SELECT $1,id FROM roles WHERE code='SUPER_ADMIN' ON CONFLICT DO NOTHING`, id)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (s *OperationsService) GetOperationalUser(ctx context.Context, id string) (OperationalUser, error) {
	var u OperationalUser
	var first, last, phone sql.NullString
	var lastLogin sql.NullTime
	err := s.db.QueryRowContext(ctx, `SELECT id,phone_normalized,first_name,last_name,user_type,status,must_change_password,last_login_at FROM users WHERE id=$1`, id).Scan(&u.ID, &phone, &first, &last, &u.UserType, &u.Status, &u.MustChangePassword, &lastLogin)
	if err != nil {
		return u, err
	}
	u.Phone = phone.String
	u.FirstName = first.String
	u.LastName = last.String
	if lastLogin.Valid {
		t := lastLogin.Time
		u.LastLoginAt = &t
	}
	u.Roles, u.Permissions, err = s.authorization(ctx, id)
	return u, err
}

func (s *OperationsService) authorization(ctx context.Context, id string) ([]string, []string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT DISTINCT r.code,COALESCE(p.code,'') FROM user_roles ur JOIN roles r ON r.id=ur.role_id AND r.is_active LEFT JOIN role_permissions rp ON rp.role_id=r.id LEFT JOIN permissions p ON p.id=rp.permission_id AND p.is_active WHERE ur.user_id=$1`, id)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	rs, ps := map[string]bool{}, map[string]bool{}
	for rows.Next() {
		var r, p string
		if err := rows.Scan(&r, &p); err != nil {
			return nil, nil, err
		}
		rs[r] = true
		if p != "" {
			ps[p] = true
		}
	}
	if rs["SUPER_ADMIN"] {
		all, err := s.db.QueryContext(ctx, `SELECT code FROM permissions WHERE is_active`)
		if err != nil {
			return nil, nil, err
		}
		defer all.Close()
		for all.Next() {
			var p string
			if err := all.Scan(&p); err != nil {
				return nil, nil, err
			}
			ps[p] = true
		}
	}
	roles, perms := make([]string, 0, len(rs)), make([]string, 0, len(ps))
	for r := range rs {
		roles = append(roles, r)
	}
	for p := range ps {
		perms = append(perms, p)
	}
	return roles, perms, rows.Err()
}
func (s *OperationsService) HasPermission(ctx context.Context, id, code string) bool {
	_, ps, err := s.authorization(ctx, id)
	if err != nil {
		return false
	}
	for _, p := range ps {
		if p == code {
			return true
		}
	}
	return false
}
func (s *OperationsService) IsInternal(ctx context.Context, id string) bool {
	var ok bool
	_ = s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE id=$1 AND user_type='INTERNAL' AND status='ACTIVE')`, id).Scan(&ok)
	return ok
}

func (s *OperationsService) roleIDsContain(ctx context.Context, roleIDs []int64, codes ...string) bool {
	if len(roleIDs) == 0 {
		return false
	}
	var ok bool
	_ = s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM roles WHERE id=ANY($1) AND code=ANY($2))`, pq.Array(roleIDs), pq.Array(codes)).Scan(&ok)
	return ok
}
func (s *OperationsService) userHasRole(ctx context.Context, userID string, codes ...string) bool {
	var ok bool
	_ = s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM user_roles ur JOIN roles r ON r.id=ur.role_id WHERE ur.user_id=$1 AND r.code=ANY($2))`, userID, pq.Array(codes)).Scan(&ok)
	return ok
}

func (s *OperationsService) ListPermissions(ctx context.Context) ([]Permission, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,code,name_fa,description_fa,group_code FROM permissions WHERE is_active ORDER BY group_code,name_fa`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Permission{}
	for rows.Next() {
		var p Permission
		if err := rows.Scan(&p.ID, &p.Code, &p.NameFA, &p.DescriptionFA, &p.GroupCode); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}
func (s *OperationsService) ListRoles(ctx context.Context) ([]Role, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT r.id,r.code,r.name_fa,r.description_fa,r.is_system,r.is_protected,r.is_active,COUNT(DISTINCT ur.user_id) FROM roles r LEFT JOIN user_roles ur ON ur.role_id=r.id GROUP BY r.id ORDER BY r.is_system DESC,r.name_fa`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Role{}
	for rows.Next() {
		var r Role
		if err := rows.Scan(&r.ID, &r.Code, &r.NameFA, &r.DescriptionFA, &r.IsSystem, &r.IsProtected, &r.IsActive, &r.UserCount); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
func (s *OperationsService) GetRole(ctx context.Context, id int64) (Role, error) {
	var r Role
	err := s.db.QueryRowContext(ctx, `SELECT r.id,r.code,r.name_fa,r.description_fa,r.is_system,r.is_protected,r.is_active,COUNT(DISTINCT ur.user_id) FROM roles r LEFT JOIN user_roles ur ON ur.role_id=r.id WHERE r.id=$1 GROUP BY r.id`, id).Scan(&r.ID, &r.Code, &r.NameFA, &r.DescriptionFA, &r.IsSystem, &r.IsProtected, &r.IsActive, &r.UserCount)
	if err != nil {
		return r, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT p.code FROM role_permissions rp JOIN permissions p ON p.id=rp.permission_id WHERE rp.role_id=$1 ORDER BY p.code`, id)
	if err != nil {
		return r, err
	}
	defer rows.Close()
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return r, err
		}
		r.PermissionCodes = append(r.PermissionCodes, p)
	}
	return r, rows.Err()
}
func (s *OperationsService) CreateRole(ctx context.Context, actor, code, name, description string, cloneID *int64) (Role, error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	if code == "" || name == "" {
		return Role{}, errors.New("code and name required")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Role{}, err
	}
	defer tx.Rollback()
	var id int64
	err = tx.QueryRowContext(ctx, `INSERT INTO roles(code,name_fa,description_fa,created_by_user_id) VALUES($1,$2,$3,$4) RETURNING id`, code, name, description, actor).Scan(&id)
	if err != nil {
		return Role{}, err
	}
	if cloneID != nil {
		_, err = tx.ExecContext(ctx, `INSERT INTO role_permissions(role_id,permission_id,created_by_user_id) SELECT $1,permission_id,$2 FROM role_permissions WHERE role_id=$3`, id, actor, *cloneID)
		if err != nil {
			return Role{}, err
		}
	}
	s.auditTx(ctx, tx, actor, "roles.create", "role", fmt.Sprint(id), nil, map[string]any{"code": code})
	if err = tx.Commit(); err != nil {
		return Role{}, err
	}
	return s.GetRole(ctx, id)
}
func (s *OperationsService) UpdateRolePermissions(ctx context.Context, actor string, id int64, codes []string) error {
	var protected bool
	if err := s.db.QueryRowContext(ctx, `SELECT is_protected FROM roles WHERE id=$1`, id).Scan(&protected); err != nil {
		return err
	}
	if protected {
		return ErrProtectedRole
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `DELETE FROM role_permissions WHERE role_id=$1`, id); err != nil {
		return err
	}
	if len(codes) > 0 {
		if _, err = tx.ExecContext(ctx, `INSERT INTO role_permissions(role_id,permission_id,created_by_user_id) SELECT $1,id,$2 FROM permissions WHERE code=ANY($3)`, id, actor, pq.Array(codes)); err != nil {
			return err
		}
	}
	s.auditTx(ctx, tx, actor, "roles.assign_permissions", "role", fmt.Sprint(id), nil, map[string]any{"permission_codes": codes})
	return tx.Commit()
}
func (s *OperationsService) UpdateRole(ctx context.Context, actor string, id int64, name, description string, active bool) error {
	var protected bool
	if err := s.db.QueryRowContext(ctx, `SELECT is_protected FROM roles WHERE id=$1`, id).Scan(&protected); err != nil {
		return err
	}
	if protected && !active {
		return ErrProtectedRole
	}
	_, err := s.db.ExecContext(ctx, `UPDATE roles SET name_fa=$2,description_fa=$3,is_active=$4,updated_at=NOW() WHERE id=$1`, id, name, description, active)
	if err == nil {
		s.audit(ctx, actor, "roles.update", "role", fmt.Sprint(id), map[string]any{"is_active": active})
	}
	return err
}

func (s *OperationsService) ListUsers(ctx context.Context, search, status, role string) ([]OperationalUser, error) {
	args := []any{"%" + strings.TrimSpace(search) + "%", status, role}
	rows, err := s.db.QueryContext(ctx, `SELECT DISTINCT u.id,COALESCE(u.phone_normalized,u.phone,''),COALESCE(u.first_name,''),COALESCE(u.last_name,''),u.user_type,u.status,u.must_change_password,u.last_login_at FROM users u LEFT JOIN user_roles ur ON ur.user_id=u.id LEFT JOIN roles r ON r.id=ur.role_id WHERE u.user_type='INTERNAL' AND ($1='%%' OR CONCAT_WS(' ',u.first_name,u.last_name,u.phone_normalized) ILIKE $1) AND ($2='' OR u.status=$2) AND ($3='' OR r.code=$3) ORDER BY u.last_login_at DESC NULLS LAST`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []OperationalUser{}
	for rows.Next() {
		var u OperationalUser
		var ll sql.NullTime
		if err := rows.Scan(&u.ID, &u.Phone, &u.FirstName, &u.LastName, &u.UserType, &u.Status, &u.MustChangePassword, &ll); err != nil {
			return nil, err
		}
		if ll.Valid {
			t := ll.Time
			u.LastLoginAt = &t
		}
		u.Roles, u.Permissions, _ = s.authorization(ctx, u.ID)
		out = append(out, u)
	}
	return out, rows.Err()
}
func (s *OperationsService) CreateInternalUser(ctx context.Context, actor, phone, first, last, password string, roleIDs []int64) (OperationalUser, error) {
	if s.roleIDsContain(ctx, roleIDs, "SUPER_ADMIN", "ADMIN") && !s.HasPermission(ctx, actor, "administrators.create") {
		return OperationalUser{}, ErrForbidden
	}
	phone = NormalizePhone(phone)
	if phone == "" || len(password) < 8 {
		return OperationalUser{}, errors.New("valid phone and password required")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return OperationalUser{}, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return OperationalUser{}, err
	}
	defer tx.Rollback()
	var id string
	err = tx.QueryRowContext(ctx, `INSERT INTO users(email,password_hash,full_name,phone,phone_normalized,role,is_active,first_name,last_name,user_type,status,must_change_password,created_by_user_id) VALUES($1,$2,NULLIF(TRIM($3||' '||$4),''),$5,$5,'internal',TRUE,NULLIF($3,''),NULLIF($4,''),'INTERNAL','ACTIVE',TRUE,$6) RETURNING id`, phoneAliasEmail(phone), string(hash), first, last, phone, actor).Scan(&id)
	if err != nil {
		return OperationalUser{}, err
	}
	if len(roleIDs) > 0 {
		_, err = tx.ExecContext(ctx, `INSERT INTO user_roles(user_id,role_id,assigned_by_user_id) SELECT $1,id,$2 FROM roles WHERE id=ANY($3) AND is_active`, id, actor, pq.Array(roleIDs))
		if err != nil {
			return OperationalUser{}, err
		}
	}
	s.auditTx(ctx, tx, actor, "users.create", "user", id, nil, map[string]any{"phone": phone})
	if err = tx.Commit(); err != nil {
		return OperationalUser{}, err
	}
	return s.GetOperationalUser(ctx, id)
}
func (s *OperationsService) AssignRoles(ctx context.Context, actor, userID string, roleIDs []int64) error {
	var targetSuper bool
	_ = s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM user_roles ur JOIN roles r ON r.id=ur.role_id WHERE ur.user_id=$1 AND r.code='SUPER_ADMIN')`, userID).Scan(&targetSuper)
	if targetSuper {
		return ErrProtectedRole
	}
	if s.roleIDsContain(ctx, roleIDs, "SUPER_ADMIN", "ADMIN") && !s.HasPermission(ctx, actor, "administrators.update") {
		return ErrForbidden
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `DELETE FROM user_roles WHERE user_id=$1`, userID); err != nil {
		return err
	}
	if len(roleIDs) > 0 {
		_, err = tx.ExecContext(ctx, `INSERT INTO user_roles(user_id,role_id,assigned_by_user_id) SELECT $1,id,$2 FROM roles WHERE id=ANY($3) AND is_active`, userID, actor, pq.Array(roleIDs))
		if err != nil {
			return err
		}
	}
	s.auditTx(ctx, tx, actor, "users.assign_roles", "user", userID, nil, map[string]any{"role_ids": roleIDs})
	return tx.Commit()
}
func (s *OperationsService) SetUserStatus(ctx context.Context, actor, userID, status string) error {
	if status != "ACTIVE" && status != "DISABLED" {
		return errors.New("invalid status")
	}
	var targetSuper bool
	_ = s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM user_roles ur JOIN roles r ON r.id=ur.role_id WHERE ur.user_id=$1 AND r.code='SUPER_ADMIN')`, userID).Scan(&targetSuper)
	if targetSuper && status != "ACTIVE" {
		return ErrProtectedRole
	}
	if s.userHasRole(ctx, userID, "ADMIN", "SUPER_ADMIN") && !s.HasPermission(ctx, actor, "administrators.disable") {
		return ErrForbidden
	}
	_, err := s.db.ExecContext(ctx, `UPDATE users SET status=$2,is_active=($2='ACTIVE'),disabled_at=CASE WHEN $2='DISABLED' THEN NOW() ELSE NULL END,updated_at=NOW() WHERE id=$1`, userID, status)
	if err == nil {
		s.audit(ctx, actor, "users.status", "user", userID, map[string]any{"status": status})
	}
	return err
}
func (s *OperationsService) ResetPassword(ctx context.Context, actor, userID, password string) error {
	if s.userHasRole(ctx, userID, "SUPER_ADMIN") {
		return ErrProtectedRole
	}
	if s.userHasRole(ctx, userID, "ADMIN") && !s.HasPermission(ctx, actor, "administrators.update") {
		return ErrForbidden
	}
	if len(password) < 8 {
		return errors.New("password too short")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `UPDATE users SET password_hash=$2,must_change_password=TRUE,updated_at=NOW() WHERE id=$1`, userID, string(hash))
	if err == nil {
		s.audit(ctx, actor, "users.reset_password", "user", userID, nil)
	}
	return err
}

func (s *OperationsService) AvailableWorkflows(ctx context.Context, userID string) ([]WorkflowTemplate, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT wt.id,wt.code,wt.name_fa,wt.description_fa,wt.icon_key,wt.version_number FROM workflow_templates wt WHERE wt.is_active AND wt.status='PUBLISHED' AND wt.scope_type='ORDER' AND wt.version_number=(SELECT MAX(v.version_number) FROM workflow_templates v WHERE v.template_group_code=wt.template_group_code AND v.status='PUBLISHED' AND v.is_active AND v.scope_type='ORDER') AND (EXISTS(SELECT 1 FROM user_roles ur JOIN roles r ON r.id=ur.role_id WHERE ur.user_id=$1 AND r.code='SUPER_ADMIN') OR EXISTS(SELECT 1 FROM user_roles ur JOIN role_permissions rp ON rp.role_id=ur.role_id JOIN permissions p ON p.id=rp.permission_id WHERE ur.user_id=$1 AND p.code=wt.start_permission_code AND p.is_active)) ORDER BY wt.sort_order`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []WorkflowTemplate{}
	for rows.Next() {
		var w WorkflowTemplate
		if err := rows.Scan(&w.ID, &w.Code, &w.NameFA, &w.DescriptionFA, &w.IconKey, &w.VersionNumber); err != nil {
			return nil, err
		}
		optionalRows, queryErr := s.db.QueryContext(ctx, `SELECT step_code,internal_title_fa FROM workflow_template_steps WHERE workflow_template_id=$1 AND is_active AND is_optional ORDER BY sequence_number`, w.ID)
		if queryErr != nil {
			return nil, queryErr
		}
		for optionalRows.Next() {
			var step WorkflowOptionalStep
			if scanErr := optionalRows.Scan(&step.Code, &step.Title); scanErr != nil {
				optionalRows.Close()
				return nil, scanErr
			}
			w.OptionalSteps = append(w.OptionalSteps, step)
		}
		if queryErr = optionalRows.Close(); queryErr != nil {
			return nil, queryErr
		}
		out = append(out, w)
	}
	return out, rows.Err()
}
func (s *OperationsService) SearchCustomer(ctx context.Context, phone string) (*CustomerMatch, error) {
	phone = NormalizePhone(phone)
	if phone == "" {
		return nil, ErrPhoneRequired
	}
	var c CustomerMatch
	err := s.db.QueryRowContext(ctx, `SELECT u.id,u.phone_normalized,COALESCE(cp.display_name,NULLIF(TRIM(CONCAT_WS(' ',u.first_name,u.last_name)),''),u.phone_normalized),u.status FROM users u LEFT JOIN customer_profiles cp ON cp.user_id=u.id WHERE u.phone_normalized=$1`, phone).Scan(&c.ID, &c.Phone, &c.DisplayName, &c.Status)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &c, err
}
func (s *OperationsService) ListActionItems(ctx context.Context, userID string, all bool) ([]ActionItem, error) {
	if all && !s.HasPermission(ctx, userID, "action_items.view_all") {
		return nil, ErrForbidden
	}
	rows, err := s.db.QueryContext(ctx, `SELECT a.id,a.workflow_instance_id,a.title_fa,a.description_fa,a.status,a.priority,a.due_at,o.id,o.order_number,COALESCE(NULLIF(TRIM(CONCAT_WS(' ',u.first_name,u.last_name)),''),u.phone_normalized),wt.name_fa,COALESCE(wsi.internal_title_fa,wts.internal_title_fa,'اقدام فرایند') FROM action_items a JOIN orders o ON o.id=a.order_id JOIN users u ON u.id=a.customer_user_id JOIN workflow_instances wi ON wi.id=a.workflow_instance_id JOIN workflow_templates wt ON wt.id=wi.workflow_template_id LEFT JOIN workflow_step_instances wsi ON wsi.id=a.workflow_step_instance_id LEFT JOIN workflow_template_steps wts ON wts.id=wsi.workflow_template_step_id WHERE a.status NOT IN ('COMPLETED','CANCELLED') AND ($2 OR a.assigned_user_id=$1 OR (a.assigned_user_id IS NULL AND EXISTS(SELECT 1 FROM user_roles ur WHERE ur.user_id=$1 AND ur.role_id=a.assigned_role_id))) AND (a.required_permission_code IS NULL OR EXISTS(SELECT 1 FROM user_roles ur JOIN roles r ON r.id=ur.role_id JOIN role_permissions rp ON rp.role_id=r.id JOIN permissions p ON p.id=rp.permission_id WHERE ur.user_id=$1 AND (p.code=a.required_permission_code OR r.code='SUPER_ADMIN'))) ORDER BY (a.due_at<NOW()) DESC,(a.priority='URGENT') DESC,a.due_at NULLS LAST,a.created_at DESC`, userID, all)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []ActionItem{}
	for rows.Next() {
		var a ActionItem
		var due sql.NullTime
		if err := rows.Scan(&a.ID, &a.WorkflowInstanceID, &a.TitleFA, &a.DescriptionFA, &a.Status, &a.Priority, &due, &a.OrderID, &a.OrderNumber, &a.Customer, &a.WorkflowName, &a.StepName); err != nil {
			return nil, err
		}
		if due.Valid {
			t := due.Time
			a.DueAt = &t
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (s *OperationsService) StartWorkflow(ctx context.Context, actor string, templateID int64, phone, name, idempotency string, excludedOptionalStepCodes []string) (WorkflowStartResult, error) {
	var out WorkflowStartResult
	if idempotency == "" {
		return out, errors.New("Idempotency-Key required")
	}
	phone = NormalizePhone(phone)
	if phone == "" {
		return out, ErrPhoneRequired
	}
	var startPermission string
	if err := s.db.QueryRowContext(ctx, `SELECT latest.id,latest.start_permission_code FROM workflow_templates requested JOIN LATERAL (SELECT id,start_permission_code FROM workflow_templates candidate WHERE candidate.template_group_code=requested.template_group_code AND candidate.status='PUBLISHED' AND candidate.is_active AND candidate.scope_type='ORDER' ORDER BY candidate.version_number DESC LIMIT 1) latest ON TRUE WHERE requested.id=$1 AND requested.scope_type='ORDER'`, templateID).Scan(&templateID, &startPermission); err != nil {
		return out, err
	}
	if !s.HasPermission(ctx, actor, "workflow_instances.start") || !s.HasPermission(ctx, actor, startPermission) {
		return out, ErrForbidden
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return out, err
	}
	defer tx.Rollback()
	var existingOrder, existingWorkflow string
	err = tx.QueryRowContext(ctx, `SELECT order_id,id FROM workflow_instances WHERE started_by_user_id=$1 AND idempotency_key=$2`, actor, idempotency).Scan(&existingOrder, &existingWorkflow)
	if err == nil {
		out.OrderID = existingOrder
		out.WorkflowInstanceID = existingWorkflow
		_ = tx.QueryRowContext(ctx, `SELECT order_number,customer_user_id FROM orders WHERE id=$1`, existingOrder).Scan(&out.OrderNumber, &out.CustomerUserID)
		return out, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return out, err
	}
	var templateAvailable bool
	if err = tx.QueryRowContext(ctx, `SELECT status='PUBLISHED' AND is_active FROM workflow_templates WHERE id=$1 FOR SHARE`, templateID).Scan(&templateAvailable); err != nil {
		return out, err
	}
	if !templateAvailable {
		return out, errors.New("workflow template is not available")
	}
	var customerID string
	err = tx.QueryRowContext(ctx, `SELECT id FROM users WHERE phone_normalized=$1 FOR UPDATE`, phone).Scan(&customerID)
	activation := ""
	if errors.Is(err, sql.ErrNoRows) {
		out.CustomerCreated = true
		hash, _ := bcrypt.GenerateFromPassword([]byte(randomPassword()), bcrypt.DefaultCost)
		err = tx.QueryRowContext(ctx, `INSERT INTO users(email,password_hash,full_name,phone,phone_normalized,role,is_active,first_name,user_type,status,must_change_password,created_by_user_id) VALUES($1,$2,NULLIF($3,''),$4,$4,'customer',FALSE,NULLIF($3,''),'CUSTOMER','INVITED',TRUE,$5) RETURNING id`, phoneAliasEmail(phone), string(hash), name, phone, actor).Scan(&customerID)
		if err != nil {
			return out, err
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO user_roles(user_id,role_id,assigned_by_user_id) SELECT $1,id,$2 FROM roles WHERE code='CUSTOMER'`, customerID, actor)
		if err != nil {
			return out, err
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO customer_profiles(user_id,customer_code,display_name) VALUES($1::uuid,'CU-'||UPPER(SUBSTRING(REPLACE(($1::uuid)::text,'-',''),1,10)),NULLIF($2,''))`, customerID, name)
		if err != nil {
			return out, err
		}
		activation, err = s.createActivationTx(ctx, tx, customerID, actor)
		if err != nil {
			return out, err
		}
	} else if err != nil {
		return out, err
	} else {
		var customerRole bool
		_ = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM user_roles ur JOIN roles r ON r.id=ur.role_id WHERE ur.user_id=$1 AND r.code='CUSTOMER')`, customerID).Scan(&customerRole)
		if !customerRole {
			_, err = tx.ExecContext(ctx, `INSERT INTO user_roles(user_id,role_id,assigned_by_user_id) SELECT $1,id,$2 FROM roles WHERE code='CUSTOMER' ON CONFLICT DO NOTHING`, customerID, actor)
			if err != nil {
				return out, err
			}
			_, _ = tx.ExecContext(ctx, `INSERT INTO customer_profiles(user_id,customer_code,display_name) VALUES($1::uuid,'CU-'||UPPER(SUBSTRING(REPLACE(($1::uuid)::text,'-',''),1,10)),NULLIF($2,'')) ON CONFLICT(user_id) DO NOTHING`, customerID, name)
		}
	}
	orderNumber := "ORD-" + time.Now().UTC().Format("20060102-150405") + "-" + randomDigits(4)
	err = tx.QueryRowContext(ctx, `INSERT INTO orders(order_number,customer_user_id,created_by_user_id,workflow_template_id) VALUES($1,$2,$3,$4) RETURNING id`, orderNumber, customerID, actor, templateID).Scan(&out.OrderID)
	if err != nil {
		return out, err
	}
	err = tx.QueryRowContext(ctx, `INSERT INTO workflow_instances(workflow_template_id,order_id,customer_user_id,started_by_user_id,idempotency_key,template_group_code,template_version_number,scope_type,scope_id) SELECT wt.id,$2,$3,$4,$5,wt.template_group_code,wt.version_number,'ORDER',$2 FROM workflow_templates wt WHERE wt.id=$1 RETURNING id`, templateID, out.OrderID, customerID, actor, idempotency).Scan(&out.WorkflowInstanceID)
	if err != nil {
		return out, err
	}
	if err = s.snapshotWorkflowTx(ctx, tx, out.WorkflowInstanceID, templateID, actor, excludedOptionalStepCodes); err != nil {
		return out, err
	}
	s.auditTx(ctx, tx, actor, "workflow_instances.start", "workflow_instance", out.WorkflowInstanceID, nil, map[string]any{"order_id": out.OrderID, "customer_user_id": customerID})
	if err = tx.Commit(); err != nil {
		return out, err
	}
	out.OrderNumber = orderNumber
	out.CustomerUserID = customerID
	out.ActivationCode = activation
	return out, nil
}

func (s *OperationsService) createActivationTx(ctx context.Context, tx *sql.Tx, userID, actor string) (string, error) {
	code := randomDigits(6)
	sum := sha256.Sum256([]byte(code))
	_, err := tx.ExecContext(ctx, `UPDATE customer_activation_tokens SET used_at=NOW() WHERE user_id=$1 AND used_at IS NULL`, userID)
	if err != nil {
		return "", err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO customer_activation_tokens(user_id,token_hash,expires_at,created_by_user_id) VALUES($1,$2,NOW()+INTERVAL '7 days',$3)`, userID, hex.EncodeToString(sum[:]), actor)
	return code, err
}
func (s *OperationsService) RegenerateActivation(ctx context.Context, actor, userID string) (string, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	code, err := s.createActivationTx(ctx, tx, userID, actor)
	if err != nil {
		return "", err
	}
	s.auditTx(ctx, tx, actor, "customers.regenerate_activation", "user", userID, nil, nil)
	return code, tx.Commit()
}
func (s *OperationsService) ActivateCustomer(ctx context.Context, phone, code, password string) error {
	phone = NormalizePhone(phone)
	if phone == "" || len(password) < 8 {
		return ErrActivationInvalid
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var tokenID int64
	var userID, hash string
	var attempts int
	err = tx.QueryRowContext(ctx, `SELECT t.id,t.user_id,t.token_hash,t.attempt_count FROM customer_activation_tokens t JOIN users u ON u.id=t.user_id WHERE u.phone_normalized=$1 AND t.used_at IS NULL AND t.expires_at>NOW() ORDER BY t.created_at DESC LIMIT 1 FOR UPDATE`, phone).Scan(&tokenID, &userID, &hash, &attempts)
	if err != nil {
		return ErrActivationInvalid
	}
	if attempts >= 3 {
		return ErrActivationInvalid
	}
	sum := sha256.Sum256([]byte(code))
	if subtle.ConstantTimeCompare([]byte(hex.EncodeToString(sum[:])), []byte(hash)) != 1 {
		_, _ = tx.ExecContext(ctx, `UPDATE customer_activation_tokens SET attempt_count=attempt_count+1 WHERE id=$1`, tokenID)
		_ = tx.Commit()
		return ErrActivationInvalid
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `UPDATE users SET password_hash=$2,status='ACTIVE',is_active=TRUE,must_change_password=FALSE,updated_at=NOW() WHERE id=$1`, userID, string(passwordHash))
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `UPDATE customer_activation_tokens SET used_at=NOW() WHERE id=$1`, tokenID)
	if err != nil {
		return err
	}
	s.auditTx(ctx, tx, userID, "customers.activate", "user", userID, nil, nil)
	return tx.Commit()
}

func (s *OperationsService) CreateProforma(ctx context.Context, actor, orderID, currency, subtotal, discount, notes string) (Proforma, error) {
	var p Proforma
	if currency == "" {
		currency = "IRR"
	}
	number := "PF-" + time.Now().UTC().Format("20060102-150405") + "-" + randomDigits(4)
	err := s.db.QueryRowContext(ctx, `INSERT INTO proformas(proforma_number,order_id,customer_user_id,currency,subtotal,discount_amount,total_amount,notes) SELECT $2,o.id,o.customer_user_id,$3,$4::numeric,$5::numeric,GREATEST($4::numeric-$5::numeric,0),NULLIF($6,'') FROM orders o WHERE o.id=$1 RETURNING id,proforma_number,order_id,status,currency,subtotal::text,discount_amount::text,total_amount::text,notes,issued_at,created_at`, orderID, number, currency, subtotal, discount, notes).Scan(&p.ID, &p.ProformaNumber, &p.OrderID, &p.Status, &p.Currency, &p.Subtotal, &p.DiscountAmount, &p.TotalAmount, &p.Notes, &p.IssuedAt, &p.CreatedAt)
	if err == nil {
		s.audit(ctx, actor, "proformas.create", "proforma", p.ID, map[string]any{"order_id": orderID})
	}
	return p, err
}
func (s *OperationsService) IssueProforma(ctx context.Context, actor, proformaID string) error {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var orderID string
	err = tx.QueryRowContext(ctx, `UPDATE proformas SET status='ISSUED',issued_by_user_id=$2,issued_at=NOW(),updated_at=NOW() WHERE id=$1 AND status='DRAFT' RETURNING order_id`, proformaID, actor).Scan(&orderID)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `UPDATE orders SET status='PROFORMA_ISSUED',proforma_issued_at=NOW(),updated_at=NOW() WHERE id=$1`, orderID)
	if err != nil {
		return err
	}
	var workflowID, stepID, stepStatus string
	err = tx.QueryRowContext(ctx, `SELECT wi.id,si.id,si.status FROM workflow_instances wi JOIN workflow_step_instances si ON si.workflow_instance_id=wi.id WHERE wi.order_id=$1 AND si.step_code='ISSUE_PROFORMA' FOR UPDATE OF si`, orderID).Scan(&workflowID, &stepID, &stepStatus)
	if err != nil {
		return err
	}
	if stepStatus != "WAITING_FOR_ASSIGNEE" && stepStatus != "IN_PROGRESS" && stepStatus != "NEEDS_CORRECTION" {
		return ErrInvalidTransition
	}
	if stepStatus != "IN_PROGRESS" {
		if _, err = tx.ExecContext(ctx, `UPDATE workflow_step_instances SET status='IN_PROGRESS',assigned_user_id=COALESCE(assigned_user_id,$2),actual_start_at=COALESCE(actual_start_at,NOW()),updated_at=NOW() WHERE id=$1`, stepID, actor); err != nil {
			return err
		}
		if err = s.runStepTriggersTx(ctx, tx, workflowID, stepID, "ON_STEP_START"); err != nil {
			return err
		}
	}
	if err = s.runStepTriggersTx(ctx, tx, workflowID, stepID, "ON_STEP_SUBMIT"); err != nil {
		return err
	}
	if err = s.completeStepTx(ctx, tx, actor, workflowID, stepID, "COMPLETED"); err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO order_timeline(order_id,title_fa,status_code) VALUES($1,'پیش‌فاکتور صادر شد','PROFORMA_ISSUED')`, orderID)
	if err != nil {
		return err
	}
	s.auditTx(ctx, tx, actor, "proformas.issue", "proforma", proformaID, nil, map[string]any{"order_id": orderID})
	return tx.Commit()
}

func (s *OperationsService) AccountOrders(ctx context.Context, userID string) ([]AccountOrder, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT o.id,wi.id,o.order_number,o.status,wt.name_fa,o.created_at,o.proforma_issued_at FROM orders o JOIN workflow_instances wi ON wi.order_id=o.id JOIN workflow_templates wt ON wt.id=o.workflow_template_id WHERE o.customer_user_id=$1 ORDER BY o.created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []AccountOrder{}
	for rows.Next() {
		var o AccountOrder
		var issued sql.NullTime
		if err := rows.Scan(&o.ID, &o.WorkflowInstanceID, &o.OrderNumber, &o.Status, &o.WorkflowName, &o.CreatedAt, &issued); err != nil {
			return nil, err
		}
		if issued.Valid {
			t := issued.Time
			o.ProformaIssuedAt = &t
		}
		timeline, err := s.db.QueryContext(ctx, `SELECT title_fa,status_code,occurred_at FROM order_timeline WHERE order_id=$1 ORDER BY occurred_at`, o.ID)
		if err != nil {
			return nil, err
		}
		for timeline.Next() {
			var t TimelineItem
			if err := timeline.Scan(&t.TitleFA, &t.StatusCode, &t.OccurredAt); err != nil {
				timeline.Close()
				return nil, err
			}
			o.Timeline = append(o.Timeline, t)
		}
		timeline.Close()
		out = append(out, o)
	}
	return out, rows.Err()
}
func (s *OperationsService) AccountProformas(ctx context.Context, userID string) ([]Proforma, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,proforma_number,order_id,status,currency,subtotal::text,discount_amount::text,total_amount::text,notes,issued_at,created_at FROM proformas WHERE customer_user_id=$1 AND status<>'DRAFT' ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Proforma{}
	for rows.Next() {
		var p Proforma
		if err := rows.Scan(&p.ID, &p.ProformaNumber, &p.OrderID, &p.Status, &p.Currency, &p.Subtotal, &p.DiscountAmount, &p.TotalAmount, &p.Notes, &p.IssuedAt, &p.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}
func (s *OperationsService) AuditLogs(ctx context.Context, search string) ([]map[string]any, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT a.id,COALESCE(CONCAT_WS(' ',u.first_name,u.last_name),u.phone_normalized,'سیستم'),a.action_code,a.entity_type,COALESCE(a.entity_id,''),a.metadata,a.created_at FROM audit_logs a LEFT JOIN users u ON u.id=a.actor_user_id WHERE $1='' OR a.action_code ILIKE '%'||$1||'%' OR a.entity_type ILIKE '%'||$1||'%' ORDER BY a.created_at DESC LIMIT 200`, search)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id int64
		var actor, action, entity, entityID string
		var metadata []byte
		var created time.Time
		if err := rows.Scan(&id, &actor, &action, &entity, &entityID, &metadata, &created); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{"id": id, "actor": actor, "action": action, "entity": entity, "entity_id": entityID, "metadata": string(metadata), "created_at": created})
	}
	return out, rows.Err()
}

func (s *OperationsService) audit(ctx context.Context, actor, action, entity, entityID string, metadata any) {
	_, _ = s.db.ExecContext(ctx, `INSERT INTO audit_logs(actor_user_id,action_code,entity_type,entity_id,metadata) VALUES(NULLIF($1,'')::uuid,$2,$3,$4,COALESCE($5::jsonb,'{}'))`, actor, action, entity, entityID, jsonText(metadata))
}
func (s *OperationsService) auditTx(ctx context.Context, tx *sql.Tx, actor, action, entity, entityID string, before, after any) {
	_, _ = tx.ExecContext(ctx, `INSERT INTO audit_logs(actor_user_id,action_code,entity_type,entity_id,before_data,after_data) VALUES(NULLIF($1,'')::uuid,$2,$3,$4,$5::jsonb,$6::jsonb)`, actor, action, entity, entityID, jsonText(before), jsonText(after))
}
func jsonText(v any) string {
	if v == nil {
		return "null"
	}
	encoded, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(encoded)
}
func randomDigits(n int) string {
	var b strings.Builder
	for i := 0; i < n; i++ {
		v, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			b.WriteByte('0')
		} else {
			b.WriteByte(byte('0' + v.Int64()))
		}
	}
	return b.String()
}
func randomPassword() string { return randomDigits(12) + "Aa!" }
