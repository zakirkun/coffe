package config

import (
	"coffe/internal/domain"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const maxWorkflowBytes = 64 * 1024

type WorkflowLoader struct{}

type workflowFile struct {
	Version int        `yaml:"version"`
	Deploy  deployFile `yaml:"deploy"`
}

type deployFile struct {
	ComposeFile string `yaml:"compose_file"`
	Mode        string `yaml:"mode"`
	StackName   string `yaml:"stack_name"`
	Build       bool   `yaml:"build"`
}

func (WorkflowLoader) Load(
	ctx context.Context,
	project domain.Project,
) (workflow domain.Workflow, err error) {
	if err := ctx.Err(); err != nil {
		return domain.Workflow{}, fmt.Errorf("load workflow: %w", err)
	}
	path := filepath.Join(project.RootPath(), "coffe.yaml")
	file, err := os.Open(path) //nolint:gosec // Project root is user-selected; coffe.yaml is a fixed filename.
	if err != nil {
		return domain.Workflow{}, fmt.Errorf("open %s: %w", path, err)
	}
	defer func() { err = errors.Join(err, file.Close()) }()
	decoder := yaml.NewDecoder(io.LimitReader(file, maxWorkflowBytes))
	decoder.KnownFields(true)
	var raw workflowFile
	if err := decoder.Decode(&raw); err != nil {
		return domain.Workflow{}, fmt.Errorf("decode %s: %w", path, err)
	}
	var extra workflowFile
	if err := decoder.Decode(&extra); err == nil {
		return domain.Workflow{}, fmt.Errorf("decode %s: multiple YAML documents", path)
	} else if !errors.Is(err, io.EOF) {
		return domain.Workflow{}, fmt.Errorf("decode trailing document in %s: %w", path, err)
	}
	if raw.Version != 1 {
		return domain.Workflow{}, fmt.Errorf("workflow version %d: only version 1 is supported", raw.Version)
	}
	composeFile, err := domain.NewComposeFile(raw.Deploy.ComposeFile)
	if err != nil {
		return domain.Workflow{}, fmt.Errorf("workflow compose_file: %w", err)
	}
	if raw.Deploy.Mode == "swarm" {
		if raw.Deploy.Build {
			return domain.Workflow{}, fmt.Errorf("workflow build: swarm mode does not build images")
		}
		stackName, err := domain.NewStackName(raw.Deploy.StackName)
		if err != nil {
			return domain.Workflow{}, fmt.Errorf("workflow stack_name: %w", err)
		}
		return domain.NewSwarmWorkflow(composeFile, stackName), nil
	}
	if raw.Deploy.Mode != "" && raw.Deploy.Mode != "compose" {
		return domain.Workflow{}, fmt.Errorf("workflow mode %q: unsupported", raw.Deploy.Mode)
	}
	return domain.NewWorkflow(composeFile, raw.Deploy.Build), nil
}
