package deploy

import (
	"bytes"
	"coffe/internal/domain"
	"coffe/internal/service"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

type Runner interface {
	Run(context.Context, RunRequest) error
}

type RunRequest struct {
	Target  domain.Target
	Command string
	Output  io.Writer
}

type Executor struct{ runner Runner }

func NewExecutor(runner Runner) *Executor { return &Executor{runner: runner} }

func (e *Executor) Execute(
	ctx context.Context,
	execution service.ExecutionStep,
) (service.StepResult, error) {
	command, err := Render(execution)
	if err != nil {
		return service.StepResult{}, err
	}
	writer := newLineWriter(execution.Output)
	if err := e.runner.Run(ctx, RunRequest{
		Target: execution.Target, Command: command, Output: writer,
	}); err != nil {
		return service.StepResult{}, fmt.Errorf("run %s: %w", execution.Step, err)
	}
	if err := writer.Flush(); err != nil {
		return service.StepResult{}, fmt.Errorf("flush %s output: %w", execution.Step, err)
	}
	if execution.Step != domain.StepInspectRepository && execution.Step != domain.StepCaptureRevision {
		return service.StepResult{}, nil
	}
	revision, err := domain.NewGitRevision(writer.LastLine())
	if err != nil {
		return service.StepResult{}, fmt.Errorf("parse %s revision: %w", execution.Step, err)
	}
	return service.StepResult{Revision: revision}, nil
}

type BinaryRunner struct{}

func (BinaryRunner) Run(ctx context.Context, request RunRequest) error {
	command := exec.CommandContext(ctx, "ssh", SSHArguments(request.Target, request.Command)...) //nolint:gosec // ssh is fixed; args come from typed values and allowlisted renderer.
	command.Stdout = request.Output
	command.Stderr = request.Output
	if err := command.Run(); err != nil {
		return fmt.Errorf("ssh process: %w", err)
	}
	return nil
}

func SSHArguments(target domain.Target, command string) []string {
	destination := target.User().String() + "@" + target.Host().String()
	return []string{
		"-o", "BatchMode=yes",
		"-o", "StrictHostKeyChecking=yes",
		"-o", "ConnectTimeout=10",
		"-p", strconv.Itoa(target.Port().Int()),
		destination,
		command,
	}
}

type lineWriter struct {
	mu       sync.Mutex
	buffer   bytes.Buffer
	lastLine string
	output   func(string) error
}

func newLineWriter(output func(string) error) *lineWriter {
	return &lineWriter{output: output}
}

func (w *lineWriter) Write(data []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	length := len(data)
	if _, err := w.buffer.Write(data); err != nil {
		return 0, fmt.Errorf("buffer output: %w", err)
	}
	contents := w.buffer.String()
	for {
		separator := strings.IndexByte(contents, '\n')
		if separator < 0 {
			break
		}
		if err := w.emit(strings.TrimRight(contents[:separator], "\r")); err != nil {
			return 0, err
		}
		contents = contents[separator+1:]
	}
	w.buffer = *bytes.NewBufferString(contents)
	return length, nil
}

func (w *lineWriter) Flush() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.buffer.Len() == 0 {
		return nil
	}
	line := w.buffer.String()
	w.buffer.Reset()
	return w.emit(strings.TrimRight(line, "\r\n"))
}

func (w *lineWriter) LastLine() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.lastLine
}

func (w *lineWriter) emit(line string) error {
	if line == "" {
		return nil
	}
	w.lastLine = line
	if w.output == nil {
		return nil
	}
	return w.output(line)
}
