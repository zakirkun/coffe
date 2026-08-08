package worker

import (
	"testing"
)

func Test_ParseExecution_rejects_unapproved_task(t *testing.T) {
	// Given
	input := TaskRequest{PlanID: "plan-1", ProjectID: "project-1", TargetID: "target-1"}

	// When
	_, err := parseExecution(input)

	// Then
	if err == nil {
		t.Fatal("parseExecution() error = nil, want rejection")
	}
}
