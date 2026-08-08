package deploy

import (
	"bufio"
	"bytes"
	"coffe/internal/domain"
	"context"
	"encoding/json"
	"fmt"
)

type SwarmInspector struct{ runner Runner }

func NewSwarmInspector(runner Runner) *SwarmInspector {
	return &SwarmInspector{runner: runner}
}

func (i *SwarmInspector) ListNodes(ctx context.Context, target domain.Target) ([]domain.SwarmNode, error) {
	var output bytes.Buffer
	if err := i.runner.Run(ctx, RunRequest{
		Target: target, Command: "docker node ls --format '{{json .}}'", Output: &output,
	}); err != nil {
		return nil, fmt.Errorf("list swarm nodes: %w", err)
	}
	var nodes []domain.SwarmNode
	if err := scanJSONLines(&output, func(line []byte) error {
		var raw struct {
			ID            string `json:"ID"`
			Hostname      string `json:"Hostname"`
			Status        string `json:"Status"`
			Availability  string `json:"Availability"`
			ManagerStatus string `json:"ManagerStatus"`
		}
		if err := json.Unmarshal(line, &raw); err != nil {
			return fmt.Errorf("decode swarm node: %w", err)
		}
		node, err := domain.NewSwarmNode(domain.SwarmNodeValues(raw))
		if err != nil {
			return err
		}
		nodes = append(nodes, node)
		return nil
	}); err != nil {
		return nil, err
	}
	return nodes, nil
}

func (i *SwarmInspector) ListServices(
	ctx context.Context,
	target domain.Target,
	stackName domain.StackName,
) ([]domain.SwarmService, error) {
	var output bytes.Buffer
	command := "docker stack services --format '{{json .}}' " + shellQuote(stackName.String())
	if err := i.runner.Run(ctx, RunRequest{Target: target, Command: command, Output: &output}); err != nil {
		return nil, fmt.Errorf("list swarm services: %w", err)
	}
	var services []domain.SwarmService
	if err := scanJSONLines(&output, func(line []byte) error {
		var raw struct {
			ID       string `json:"ID"`
			Name     string `json:"Name"`
			Mode     string `json:"Mode"`
			Replicas string `json:"Replicas"`
			Image    string `json:"Image"`
		}
		if err := json.Unmarshal(line, &raw); err != nil {
			return fmt.Errorf("decode swarm service: %w", err)
		}
		service, err := domain.NewSwarmService(domain.SwarmServiceValues(raw))
		if err != nil {
			return err
		}
		services = append(services, service)
		return nil
	}); err != nil {
		return nil, err
	}
	return services, nil
}

func scanJSONLines(input *bytes.Buffer, consume func([]byte) error) error {
	scanner := bufio.NewScanner(input)
	scanner.Buffer(make([]byte, 4*1024), 1024*1024)
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		if err := consume(line); err != nil {
			return err
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan docker output: %w", err)
	}
	return nil
}
