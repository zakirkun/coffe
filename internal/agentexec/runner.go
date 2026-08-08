package agentexec

import (
	"bytes"
	"coffe/internal/domain"
	"coffe/internal/service"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"sync"
)

var ErrUnsupportedTool = errors.New("agentexec: unsupported tool")

type Spec struct {
	Binary string
	Args   []string
}

func CommandSpec(tool domain.AgentTool, prompt string) (Spec, error) {
	switch tool {
	case domain.AgentOpenCode:
		return Spec{Binary: "opencode", Args: []string{"run", prompt}}, nil
	case domain.AgentClaude:
		return Spec{Binary: "claude", Args: []string{"-p", prompt}}, nil
	case domain.AgentCodex:
		return Spec{Binary: "codex", Args: []string{"exec", prompt}}, nil
	case domain.AgentGemini:
		return Spec{Binary: "gemini", Args: []string{"-p", prompt}}, nil
	case domain.AgentAider:
		return Spec{Binary: "aider", Args: []string{"--message", prompt}}, nil
	default:
		return Spec{}, fmt.Errorf("tool %q: %w", tool, ErrUnsupportedTool)
	}
}

type Runner struct{}

func (Runner) Run(ctx context.Context, run service.AgentRun) error {
	spec, err := CommandSpec(run.Tool, run.Prompt)
	if err != nil {
		return err
	}
	writer := &outputWriter{output: run.Output}
	command := exec.CommandContext(ctx, spec.Binary, spec.Args...) //nolint:gosec // binary and argument shape are selected from a closed AgentTool set.
	command.Dir = run.WorkingDir
	command.Stdout = writer
	command.Stderr = writer
	if err := command.Run(); err != nil {
		return fmt.Errorf("run %s: %w", run.Tool, err)
	}
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("flush %s output: %w", run.Tool, err)
	}
	return nil
}

type outputWriter struct {
	mu     sync.Mutex
	buffer bytes.Buffer
	output func(string) error
}

func (w *outputWriter) Write(data []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	length := len(data)
	if _, err := w.buffer.Write(data); err != nil {
		return 0, fmt.Errorf("buffer agent output: %w", err)
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

func (w *outputWriter) Flush() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.buffer.Len() == 0 {
		return nil
	}
	line := strings.TrimRight(w.buffer.String(), "\r\n")
	w.buffer.Reset()
	return w.emit(line)
}

func (w *outputWriter) emit(line string) error {
	if line == "" || w.output == nil {
		return nil
	}
	return w.output(line)
}
