package deploy

import (
	"coffe/internal/domain"
	"coffe/internal/service"
	"errors"
	"fmt"
	"strings"
)

var ErrUnsupportedStep = errors.New("deploy: unsupported step")

type CommandRenderer struct{}

func (CommandRenderer) Render(
	execution service.ExecutionStep,
) (string, error) {
	return Render(execution)
}

func Render(execution service.ExecutionStep) (string, error) {
	root := shellQuote(execution.Target.RemotePath().String())
	branch := shellQuote(execution.Target.Branch().String())
	composeFile := shellQuote(execution.Workflow.ComposeFile().String())
	prefix := "cd -- " + root + " && "
	switch execution.Step {
	case domain.StepInspectRepository:
		return prefix + "test \"$(git branch --show-current)\" = " + branch +
			" && test -z \"$(git status --porcelain)\" && git rev-parse HEAD", nil
	case domain.StepGitPull:
		return prefix + "git pull --ff-only origin " + branch, nil
	case domain.StepCaptureRevision:
		return prefix + "git rev-parse HEAD", nil
	case domain.StepComposeConfig:
		return prefix + "docker compose -f " + composeFile + " config --quiet", nil
	case domain.StepComposeBuild:
		return prefix + "docker compose -f " + composeFile + " build", nil
	case domain.StepComposeUp:
		return prefix + "docker compose -f " + composeFile + " up -d", nil
	case domain.StepComposeHealth:
		return prefix + "test -z \"$(docker compose -f " + composeFile +
			" ps --status exited --quiet)\" && test -n \"$(docker compose -f " + composeFile +
			" ps --status running --quiet)\"", nil
	case domain.StepGitReset:
		revision, ok := execution.Plan.Revision()
		if !ok {
			return "", fmt.Errorf("rollback revision: %w", ErrUnsupportedStep)
		}
		return prefix + "git reset --hard " + shellQuote(revision.String()), nil
	case domain.StepSwarmConfig:
		return prefix + "docker stack config --compose-file " + composeFile + " >/dev/null", nil
	case domain.StepSwarmDeploy:
		return prefix + "docker stack deploy --compose-file " + composeFile + " --prune " +
			shellQuote(execution.Workflow.StackName().String()), nil
	case domain.StepSwarmHealth:
		return prefix + "test -n \"$(docker stack services --format '{{json .}}' " +
			shellQuote(execution.Workflow.StackName().String()) + ")\"", nil
	default:
		return "", fmt.Errorf("step %q: %w", execution.Step, ErrUnsupportedStep)
	}
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'"'"'`) + "'"
}
