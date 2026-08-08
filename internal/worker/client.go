package worker

import (
	"bytes"
	"coffe/internal/domain"
	"coffe/internal/service"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

type ClientConfig struct{ HTTPClient *http.Client }

type Client struct{ httpClient *http.Client }

func NewClient(config ClientConfig) *Client {
	client := config.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 2 * time.Minute}
	}
	return &Client{httpClient: client}
}

func (c *Client) Ping(
	ctx context.Context,
	worker domain.Worker,
	token string,
) (heartbeat service.WorkerHeartbeat, err error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, worker.BaseURL()+"/v1/health", nil)
	if err != nil {
		return service.WorkerHeartbeat{}, fmt.Errorf("create worker health request: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+token)
	response, err := c.httpClient.Do(request)
	if err != nil {
		return service.WorkerHeartbeat{}, fmt.Errorf("worker health request: %w", err)
	}
	defer func() { err = errors.Join(err, response.Body.Close()) }()
	if response.StatusCode != http.StatusOK {
		return service.WorkerHeartbeat{}, fmt.Errorf("worker health status %d", response.StatusCode)
	}
	var health HealthResponse
	if err := json.NewDecoder(io.LimitReader(response.Body, maxRequestBytes)).Decode(&health); err != nil {
		return service.WorkerHeartbeat{}, fmt.Errorf("decode worker health: %w", err)
	}
	if health.Status != "ready" || health.Version == "" {
		return service.WorkerHeartbeat{}, fmt.Errorf("worker health payload is invalid")
	}
	return service.WorkerHeartbeat{Version: health.Version}, nil
}

func (c *Client) Execute(
	ctx context.Context,
	worker domain.Worker,
	token string,
	execution service.ExecutionStep,
) (result service.StepResult, err error) {
	if execution.Plan.Status() != domain.PlanRunning {
		return service.StepResult{}, fmt.Errorf("worker task requires a running approved plan")
	}
	payload, err := json.Marshal(taskFromExecution(execution))
	if err != nil {
		return service.StepResult{}, fmt.Errorf("encode worker task: %w", err)
	}
	request, err := http.NewRequestWithContext(
		ctx, http.MethodPost, worker.BaseURL()+"/v1/tasks", bytes.NewReader(payload),
	)
	if err != nil {
		return service.StepResult{}, fmt.Errorf("create worker task request: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")
	response, err := c.httpClient.Do(request)
	if err != nil {
		return service.StepResult{}, fmt.Errorf("worker task request: %w", err)
	}
	defer func() { err = errors.Join(err, response.Body.Close()) }()
	var task TaskResponse
	if err := json.NewDecoder(io.LimitReader(response.Body, maxOutputBytes+maxRequestBytes)).Decode(&task); err != nil {
		return service.StepResult{}, fmt.Errorf("decode worker task: %w", err)
	}
	for _, line := range task.Logs {
		if execution.Output != nil {
			if err := execution.Output(line); err != nil {
				return service.StepResult{}, err
			}
		}
	}
	if response.StatusCode != http.StatusOK {
		return service.StepResult{}, fmt.Errorf("worker task status %d", response.StatusCode)
	}
	if task.Revision != "" {
		revision, err := domain.NewGitRevision(task.Revision)
		if err != nil {
			return service.StepResult{}, err
		}
		result.Revision = revision
	}
	return result, nil
}

func taskFromExecution(execution service.ExecutionStep) TaskRequest {
	revision := ""
	if value, ok := execution.Plan.Revision(); ok {
		revision = value.String()
	}
	return TaskRequest{
		PlanID: string(execution.Plan.ID()), ProjectID: string(execution.Plan.ProjectID()),
		TargetID: string(execution.Target.ID()), PlanKind: execution.Plan.Kind(), Step: execution.Step,
		Root: execution.Target.RemotePath().String(), Branch: execution.Target.Branch().String(),
		ComposeFile: execution.Workflow.ComposeFile().String(), Mode: execution.Workflow.Mode(),
		StackName: execution.Workflow.StackName().String(), Revision: revision, Approved: true,
	}
}
