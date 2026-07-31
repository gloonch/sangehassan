package usecase

import (
	"encoding/json"
	"testing"
)

func TestWorkflowValueContracts(t *testing.T) {
	tests := []struct {
		name, kind, value, unit, currency string
		valid                             bool
	}{
		{"text scalar", "SHORT_TEXT", `"سلام"`, "", "", true},
		{"text object rejected", "SHORT_TEXT", `{"value":"سلام"}`, "", "", false},
		{"measurement object", "WEIGHT", `{"value":12.5,"unit":"TON"}`, "TON", "", true},
		{"measurement scalar rejected", "WEIGHT", `12.5`, "TON", "", false},
		{"wrong unit rejected", "WEIGHT", `{"value":12.5,"unit":"KG"}`, "TON", "", false},
		{"money object", "MONEY", `{"amount":1000,"currency":"IRR"}`, "", "IRR", true},
		{"wrong currency rejected", "MONEY", `{"amount":1000,"currency":"USD"}`, "", "IRR", false},
		{"multi select", "MULTI_SELECT", `["a","b"]`, "", "", true},
	}
	options := []byte(`["a","b"]`)
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateRuntimeValue(test.kind, []byte(test.value), options, []byte(`{}`), true, test.unit, test.currency)
			if (err == nil) != test.valid {
				t.Fatalf("valid=%v error=%v", test.valid, err)
			}
		})
	}
}

func TestWorkflowNumericAndLengthValidation(t *testing.T) {
	if err := validateRuntimeValue("DECIMAL", []byte(`11`), nil, []byte(`{"max":10}`), true, "", ""); err == nil {
		t.Fatal("expected max validation error")
	}
	if err := validateRuntimeValue("SHORT_TEXT", []byte(`"طولانی"`), nil, []byte(`{"maxLength":3}`), true, "", ""); err == nil {
		t.Fatal("expected length validation error")
	}
}

func TestWorkflowStepStateMachine(t *testing.T) {
	allowed := [][2]string{{"NOT_STARTED", "WAITING_FOR_ASSIGNEE"}, {"WAITING_FOR_ASSIGNEE", "IN_PROGRESS"}, {"IN_PROGRESS", "WAITING_FOR_APPROVAL"}, {"IN_PROGRESS", "COMPLETED"}, {"WAITING_FOR_APPROVAL", "NEEDS_CORRECTION"}, {"WAITING_FOR_APPROVAL", "COMPLETED"}, {"COMPLETED", "IN_PROGRESS"}}
	for _, transition := range allowed {
		if !validWorkflowStepTransition(transition[0], transition[1]) {
			t.Errorf("expected %s -> %s", transition[0], transition[1])
		}
	}
	denied := [][2]string{{"NOT_STARTED", "COMPLETED"}, {"WAITING_FOR_ASSIGNEE", "COMPLETED"}, {"COMPLETED", "WAITING_FOR_APPROVAL"}, {"SKIPPED", "COMPLETED"}}
	for _, transition := range denied {
		if validWorkflowStepTransition(transition[0], transition[1]) {
			t.Errorf("unexpected %s -> %s", transition[0], transition[1])
		}
	}
}

func TestHandoffToleranceRules(t *testing.T) {
	abs := 0.2
	pct := 1.0
	if ok, _ := withinHandoffTolerance(10, 10.15, &abs, nil); !ok {
		t.Fatal("absolute tolerance should pass")
	}
	if ok, _ := withinHandoffTolerance(100, 100.8, nil, &pct); !ok {
		t.Fatal("percentage tolerance should pass")
	}
	if ok, percentage := withinHandoffTolerance(0, 0.1, nil, &pct); ok || percentage != nil {
		t.Fatal("zero expected must ignore percentage tolerance")
	}
	if ok, _ := withinHandoffTolerance(0, 0.1, &abs, &pct); !ok {
		t.Fatal("zero expected should use absolute tolerance")
	}
	if ok, _ := withinHandoffTolerance(10, 10, nil, nil); !ok {
		t.Fatal("exact match should pass without tolerance")
	}
}

func TestPublishValidationRejectsLateBlockingTrigger(t *testing.T) {
	role := int64(1)
	template := WorkflowTemplateVersion{Metrics: []HandoffMetricDefinition{}, Steps: []WorkflowTemplateStepV2{{StepCode: "TEST", SequenceNumber: 1, IsActive: true, ResponsibleRoleID: &role, RequiredPermissionCode: "workflow_steps.submit", Tasks: []WorkflowTaskTemplate{{TriggerType: "ON_STEP_COMPLETE", TitleFA: "late", BlocksStepCompletion: true}}}}}
	if err := validateTemplateForPublish(template); err == nil {
		t.Fatal("expected late blocking trigger rejection")
	}
}

func TestSelectDefinitionRequiresOptions(t *testing.T) {
	if err := validateFieldDefinition("choice", "SELECT", json.RawMessage(`[]`), json.RawMessage(`{}`), nil, nil, nil, nil, false, false, map[string]bool{}); err == nil {
		t.Fatal("expected select options error")
	}
}
