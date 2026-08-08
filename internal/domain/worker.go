package domain

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
)

var ErrInvalidWorker = errors.New("domain: invalid worker")

type WorkerID string

type Worker struct {
	createdAt  time.Time
	lastSeen   time.Time
	id         WorkerID
	targetID   TargetID
	name       Name
	baseURL    string
	secretName string
	version    string
	enabled    bool
}

type WorkerValues struct {
	ID         WorkerID
	TargetID   TargetID
	Name       Name
	BaseURL    string
	SecretName string
	Version    string
	Enabled    bool
	CreatedAt  time.Time
	LastSeen   time.Time
}

func NewWorker(values WorkerValues) (Worker, error) {
	baseURL, err := parseWorkerURL(values.BaseURL)
	secretName := strings.TrimSpace(values.SecretName)
	version := strings.TrimSpace(values.Version)
	if values.ID == "" || values.TargetID == "" || values.Name.String() == "" || err != nil ||
		secretName == "" || values.CreatedAt.IsZero() || hasControl(secretName) || hasControl(version) ||
		(!values.LastSeen.IsZero() && version == "") {
		return Worker{}, fmt.Errorf("worker: %w", ErrInvalidWorker)
	}
	return Worker{
		id: values.ID, targetID: values.TargetID, name: values.Name,
		baseURL: baseURL, secretName: secretName, version: version, enabled: values.Enabled,
		createdAt: values.CreatedAt.UTC(), lastSeen: values.LastSeen.UTC(),
	}, nil
}

func parseWorkerURL(raw string) (string, error) {
	parsed, err := url.Parse(strings.TrimRight(strings.TrimSpace(raw), "/"))
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.User != nil ||
		parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", ErrInvalidWorker
	}
	return parsed.String(), nil
}

func (w Worker) WithHeartbeat(version string, seenAt time.Time) (Worker, error) {
	version = strings.TrimSpace(version)
	if version == "" || seenAt.IsZero() || hasControl(version) {
		return Worker{}, fmt.Errorf("worker heartbeat: %w", ErrInvalidWorker)
	}
	w.version = version
	w.lastSeen = seenAt.UTC()
	return w, nil
}

func (w Worker) ID() WorkerID         { return w.id }
func (w Worker) TargetID() TargetID   { return w.targetID }
func (w Worker) Name() Name           { return w.name }
func (w Worker) BaseURL() string      { return w.baseURL }
func (w Worker) SecretName() string   { return w.secretName }
func (w Worker) Version() string      { return w.version }
func (w Worker) Enabled() bool        { return w.enabled }
func (w Worker) CreatedAt() time.Time { return w.createdAt }
func (w Worker) LastSeen() time.Time  { return w.lastSeen }
