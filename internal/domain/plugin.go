package domain

import (
	"encoding/hex"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"
)

var ErrInvalidPlugin = errors.New("domain: invalid plugin")

var pluginMethodPattern = regexp.MustCompile(`^[a-z][a-z0-9_-]*(?:\.[a-z][a-z0-9_-]*)+$`)

type PluginID string

type Plugin struct {
	createdAt      time.Time
	id             PluginID
	name           Name
	executablePath string
	sha256         string
	methods        []string
	enabled        bool
}

type PluginValues struct {
	ID             PluginID
	Name           Name
	ExecutablePath string
	SHA256         string
	Methods        []string
	Enabled        bool
	CreatedAt      time.Time
}

func NewPlugin(values PluginValues) (Plugin, error) {
	path := filepath.Clean(strings.TrimSpace(values.ExecutablePath))
	hash := strings.ToLower(strings.TrimSpace(values.SHA256))
	decoded, hashErr := hex.DecodeString(hash)
	methods := slices.Clone(values.Methods)
	slices.Sort(methods)
	methods = slices.Compact(methods)
	if values.ID == "" || values.Name.String() == "" || !filepath.IsAbs(path) ||
		hashErr != nil || len(decoded) != 32 || len(methods) == 0 || values.CreatedAt.IsZero() {
		return Plugin{}, fmt.Errorf("plugin: %w", ErrInvalidPlugin)
	}
	for _, method := range methods {
		if !pluginMethodPattern.MatchString(method) {
			return Plugin{}, fmt.Errorf("plugin method %q: %w", method, ErrInvalidPlugin)
		}
	}
	return Plugin{
		id: values.ID, name: values.Name, executablePath: path,
		sha256: hash, methods: methods, enabled: values.Enabled, createdAt: values.CreatedAt.UTC(),
	}, nil
}

func (p Plugin) ID() PluginID              { return p.id }
func (p Plugin) Name() Name                { return p.name }
func (p Plugin) ExecutablePath() string    { return p.executablePath }
func (p Plugin) SHA256() string            { return p.sha256 }
func (p Plugin) Methods() []string         { return slices.Clone(p.methods) }
func (p Plugin) Enabled() bool             { return p.enabled }
func (p Plugin) CreatedAt() time.Time      { return p.createdAt }
func (p Plugin) Allows(method string) bool { return slices.Contains(p.methods, method) }
