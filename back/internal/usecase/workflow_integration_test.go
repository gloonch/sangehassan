package usecase

import (
	"context"
	"database/sql"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

func TestWorkflowSnapshotAndIdempotencyIntegration(t *testing.T) {
	dsn := os.Getenv("WORKFLOW_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("WORKFLOW_TEST_DATABASE_URL is not set")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()
	actor := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	_, err = db.ExecContext(ctx, `INSERT INTO users(id,email,password_hash,phone,phone_normalized,user_type,status,is_active) VALUES($1,'phase2-operator@example.test','x','+989121111111','+989121111111','INTERNAL','ACTIVE',TRUE) ON CONFLICT(id) DO NOTHING`, actor)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.ExecContext(ctx, `INSERT INTO user_roles(user_id,role_id) SELECT $1,id FROM roles WHERE code='OPERATOR' ON CONFLICT DO NOTHING`, actor)
	if err != nil {
		t.Fatal(err)
	}
	var legacyTemplateID int64
	if err = db.QueryRowContext(ctx, `SELECT id FROM workflow_templates WHERE template_group_code='factory_to_project' AND version_number=1`).Scan(&legacyTemplateID); err != nil {
		t.Fatal(err)
	}
	service := NewOperationsService(db)
	idempotencyKey := "phase2-integration-" + randomDigits(8)
	started, err := service.StartWorkflow(ctx, actor, legacyTemplateID, "09123334444", "مشتری تست", idempotencyKey, []string{"INSTALLATION"})
	if err != nil {
		t.Fatal(err)
	}
	retried, err := service.StartWorkflow(ctx, actor, legacyTemplateID, "09123334444", "مشتری تست", idempotencyKey, []string{"INSTALLATION"})
	if err != nil {
		t.Fatal(err)
	}
	if retried.WorkflowInstanceID != started.WorkflowInstanceID {
		t.Fatalf("idempotent retry created another workflow: %s != %s", retried.WorkflowInstanceID, started.WorkflowInstanceID)
	}
	var version, stepCount, installationCount int
	if err = db.QueryRowContext(ctx, `SELECT wi.template_version_number,COUNT(si.id),COUNT(si.id) FILTER(WHERE si.step_code='INSTALLATION') FROM workflow_instances wi JOIN workflow_step_instances si ON si.workflow_instance_id=wi.id WHERE wi.id=$1 GROUP BY wi.template_version_number`, started.WorkflowInstanceID).Scan(&version, &stepCount, &installationCount); err != nil {
		t.Fatal(err)
	}
	if version != 2 || stepCount != 11 || installationCount != 0 {
		t.Fatalf("unexpected snapshot version=%d steps=%d installation=%d", version, stepCount, installationCount)
	}
	var firstActionCount int
	if err = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM action_items WHERE workflow_instance_id=$1 AND source_trigger_type='MAIN_STEP'`, started.WorkflowInstanceID).Scan(&firstActionCount); err != nil {
		t.Fatal(err)
	}
	if firstActionCount != 1 {
		t.Fatalf("expected exactly one initial action, got %d", firstActionCount)
	}
	var templateStepID int64
	var originalTitle, snapshotTitle string
	if err = db.QueryRowContext(ctx, `SELECT si.workflow_template_step_id,ts.internal_title_fa FROM workflow_step_instances si JOIN workflow_template_steps ts ON ts.id=si.workflow_template_step_id WHERE si.workflow_instance_id=$1 AND si.step_code='ISSUE_PROFORMA'`, started.WorkflowInstanceID).Scan(&templateStepID, &originalTitle); err != nil {
		t.Fatal(err)
	}
	if _, err = db.ExecContext(ctx, `UPDATE workflow_template_steps SET internal_title_fa=internal_title_fa||' تغییر' WHERE id=$1`, templateStepID); err != nil {
		t.Fatal(err)
	}
	if err = db.QueryRowContext(ctx, `SELECT internal_title_fa FROM workflow_step_instances WHERE workflow_instance_id=$1 AND step_code='ISSUE_PROFORMA'`, started.WorkflowInstanceID).Scan(&snapshotTitle); err != nil {
		t.Fatal(err)
	}
	if snapshotTitle != originalTitle {
		t.Fatalf("snapshot changed with template: %q != %q", snapshotTitle, originalTitle)
	}
	_, _ = db.ExecContext(ctx, `UPDATE workflow_template_steps SET internal_title_fa=$2 WHERE id=$1`, templateStepID, originalTitle)
	proforma, err := service.CreateProforma(ctx, actor, started.OrderID, "IRR", "100000", "0", "integration")
	if err != nil {
		t.Fatal(err)
	}
	if err = service.IssueProforma(ctx, actor, proforma.ID); err != nil {
		t.Fatal(err)
	}
	var completed, waiting int
	if err = db.QueryRowContext(ctx, `SELECT COUNT(*) FILTER(WHERE step_code='ISSUE_PROFORMA' AND status='COMPLETED'),COUNT(*) FILTER(WHERE sequence_number=2 AND status='WAITING_FOR_ASSIGNEE') FROM workflow_step_instances WHERE workflow_instance_id=$1`, started.WorkflowInstanceID).Scan(&completed, &waiting); err != nil {
		t.Fatal(err)
	}
	if completed != 1 || waiting != 1 {
		t.Fatalf("proforma transition did not activate exactly one next step: completed=%d waiting=%d", completed, waiting)
	}
}
