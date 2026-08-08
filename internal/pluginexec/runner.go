package pluginexec

import (
	"bytes"
	"coffe/internal/domain"
	"coffe/internal/redact"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

const (
	maxPluginBytes = 1024 * 1024 * 1024
	maxInputBytes  = 64 * 1024
	maxOutputBytes = 1024 * 1024
)

var (
	ErrHashMismatch = errors.New("pluginexec: executable hash mismatch")
	ErrProtocol     = errors.New("pluginexec: invalid plugin protocol")
)

type Runner struct{ timeout time.Duration }

func NewRunner(timeout time.Duration) *Runner { return &Runner{timeout: timeout} }

func (r *Runner) Verify(plugin domain.Plugin) error {
	hash, err := HashFile(plugin.ExecutablePath())
	if err != nil {
		return err
	}
	if hash != plugin.SHA256() {
		return ErrHashMismatch
	}
	return nil
}

type request struct {
	Method string          `json:"method"`
	Input  json.RawMessage `json:"input"`
}

type response struct {
	Result json.RawMessage `json:"result"`
	Error  string          `json:"error,omitempty"`
}

func (r *Runner) Invoke(
	ctx context.Context,
	plugin domain.Plugin,
	method string,
	input json.RawMessage,
) (json.RawMessage, error) {
	if !plugin.Enabled() || !plugin.Allows(method) || len(input) > maxInputBytes || !json.Valid(input) {
		return nil, ErrProtocol
	}
	if err := r.Verify(plugin); err != nil {
		return nil, err
	}
	payload, err := json.Marshal(request{Method: method, Input: input})
	if err != nil {
		return nil, fmt.Errorf("encode plugin request: %w", err)
	}
	invokeCtx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	command := exec.CommandContext(invokeCtx, plugin.ExecutablePath()) //nolint:gosec // executable is absolute and verified against its pinned SHA-256.
	command.Dir = filepath.Dir(plugin.ExecutablePath())
	command.Env = minimalEnvironment()
	command.Stdin = bytes.NewReader(append(payload, '\n'))
	var stdout, stderr cappedBuffer
	stdout.limit = maxOutputBytes
	stderr.limit = maxOutputBytes / 4
	command.Stdout = &stdout
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		return nil, fmt.Errorf("plugin process: %w: %s", err, redact.Line(stderr.String()))
	}
	var result response
	decoder := json.NewDecoder(bytes.NewReader(stdout.Bytes()))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&result); err != nil {
		return nil, fmt.Errorf("decode plugin response: %w", err)
	}
	var extra json.RawMessage
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return nil, ErrProtocol
	}
	if result.Error != "" {
		return nil, fmt.Errorf("plugin error: %s", redact.Line(result.Error))
	}
	if len(result.Result) == 0 || !json.Valid(result.Result) {
		return nil, ErrProtocol
	}
	return result.Result, nil
}

func HashFile(path string) (hash string, err error) {
	file, err := os.Open(path) //nolint:gosec // caller supplies an absolute plugin executable path.
	if err != nil {
		return "", fmt.Errorf("open plugin executable: %w", err)
	}
	defer func() { err = errors.Join(err, file.Close()) }()
	digest := sha256.New()
	written, err := io.Copy(digest, io.LimitReader(file, maxPluginBytes+1))
	if err != nil {
		return "", fmt.Errorf("hash plugin executable: %w", err)
	}
	if written > maxPluginBytes {
		return "", fmt.Errorf("plugin executable exceeds size limit")
	}
	return hex.EncodeToString(digest.Sum(nil)), nil
}

func minimalEnvironment() []string {
	environment := []string{"COFFE_PLUGIN_PROTOCOL=1"}
	for _, key := range []string{"PATH", "SystemRoot", "HOME", "USERPROFILE", "TMP", "TEMP"} {
		if value := os.Getenv(key); value != "" {
			environment = append(environment, key+"="+value)
		}
	}
	return environment
}

type cappedBuffer struct {
	buffer bytes.Buffer
	limit  int
}

func (b *cappedBuffer) Write(content []byte) (int, error) {
	length := len(content)
	remaining := b.limit - b.buffer.Len()
	if remaining <= 0 {
		return length, nil
	}
	if len(content) > remaining {
		content = content[:remaining]
	}
	_, err := b.buffer.Write(content)
	return length, err
}

func (b *cappedBuffer) Bytes() []byte  { return bytes.Clone(b.buffer.Bytes()) }
func (b *cappedBuffer) String() string { return b.buffer.String() }
