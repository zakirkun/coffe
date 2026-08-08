package worker

import (
	"coffe/internal/domain"
	"coffe/internal/service"
	"fmt"
	"time"
)

func parseExecution(input TaskRequest) (service.ExecutionStep, error) {
	if !input.Approved || input.PlanID == "" || input.ProjectID == "" || input.TargetID == "" ||
		!input.PlanKind.Valid() || !input.Step.Valid() {
		return service.ExecutionStep{}, fmt.Errorf("worker task identity is invalid")
	}
	remotePath, err := domain.NewRemotePath(input.Root)
	if err != nil {
		return service.ExecutionStep{}, err
	}
	branch, err := domain.NewBranch(input.Branch)
	if err != nil {
		return service.ExecutionStep{}, err
	}
	composeFile, err := domain.NewComposeFile(input.ComposeFile)
	if err != nil {
		return service.ExecutionStep{}, err
	}
	target, err := workerTarget(input, remotePath, branch)
	if err != nil {
		return service.ExecutionStep{}, err
	}
	workflow, err := workerWorkflow(input, composeFile)
	if err != nil {
		return service.ExecutionStep{}, err
	}
	plan, err := workerPlan(input)
	if err != nil {
		return service.ExecutionStep{}, err
	}
	return service.ExecutionStep{Plan: plan, Target: target, Workflow: workflow, Step: input.Step}, nil
}

func workerTarget(
	input TaskRequest,
	remotePath domain.RemotePath,
	branch domain.Branch,
) (domain.Target, error) {
	name, err := domain.NewName("worker-target")
	if err != nil {
		return domain.Target{}, err
	}
	host, err := domain.NewHost("localhost")
	if err != nil {
		return domain.Target{}, err
	}
	port, err := domain.NewPort(22)
	if err != nil {
		return domain.Target{}, err
	}
	user, err := domain.NewSSHUser("worker")
	if err != nil {
		return domain.Target{}, err
	}
	return domain.NewTarget(domain.TargetValues{
		ID: domain.TargetID(input.TargetID), ProjectID: domain.ProjectID(input.ProjectID),
		Name: name, Host: host, Port: port, User: user,
		RemotePath: remotePath, Branch: branch, CreatedAt: time.Now().UTC(),
	})
}

func workerWorkflow(input TaskRequest, composeFile domain.ComposeFile) (domain.Workflow, error) {
	if input.Mode == domain.DeploymentCompose {
		return domain.NewWorkflow(composeFile, input.Step == domain.StepComposeBuild), nil
	}
	if input.Mode != domain.DeploymentSwarm {
		return domain.Workflow{}, fmt.Errorf("worker deployment mode is invalid")
	}
	stackName, err := domain.NewStackName(input.StackName)
	if err != nil {
		return domain.Workflow{}, err
	}
	return domain.NewSwarmWorkflow(composeFile, stackName), nil
}

func workerPlan(input TaskRequest) (domain.ActionPlan, error) {
	var revision domain.GitRevision
	var err error
	if input.Revision != "" {
		revision, err = domain.NewGitRevision(input.Revision)
		if err != nil {
			return domain.ActionPlan{}, err
		}
	}
	now := time.Now().UTC()
	return domain.RestoreActionPlan(domain.PlanValues{
		ID: domain.PlanID(input.PlanID), ProjectID: domain.ProjectID(input.ProjectID),
		TargetID: domain.TargetID(input.TargetID), Kind: input.PlanKind,
		Status: domain.PlanRunning, Revision: revision, Steps: []domain.StepKind{input.Step},
		CreatedAt: now, ApprovedAt: now,
	})
}
