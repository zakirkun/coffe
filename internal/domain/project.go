package domain

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

var (
	ErrInvalidProject = errors.New("domain: invalid project")
	ErrInvalidTarget  = errors.New("domain: invalid target")
)

type (
	ProjectID string
	TargetID  string
)

type Project struct {
	createdAt time.Time
	id        ProjectID
	name      Name
	rootPath  string
}

type ProjectValues struct {
	ID        ProjectID
	Name      Name
	RootPath  string
	CreatedAt time.Time
}

func NewProject(values ProjectValues) (Project, error) {
	cleanRoot := filepath.Clean(strings.TrimSpace(values.RootPath))
	if values.ID == "" || !filepath.IsAbs(cleanRoot) || values.CreatedAt.IsZero() {
		return Project{}, fmt.Errorf("project: %w", ErrInvalidProject)
	}
	return Project{
		id: values.ID, name: values.Name, rootPath: cleanRoot, createdAt: values.CreatedAt.UTC(),
	}, nil
}

func (p Project) ID() ProjectID        { return p.id }
func (p Project) Name() Name           { return p.name }
func (p Project) RootPath() string     { return p.rootPath }
func (p Project) CreatedAt() time.Time { return p.createdAt }

type Target struct {
	createdAt  time.Time
	id         TargetID
	projectID  ProjectID
	name       Name
	host       Host
	port       Port
	user       SSHUser
	remotePath RemotePath
	branch     Branch
}

type TargetValues struct {
	ID         TargetID
	ProjectID  ProjectID
	Name       Name
	Host       Host
	Port       Port
	User       SSHUser
	RemotePath RemotePath
	Branch     Branch
	CreatedAt  time.Time
}

func NewTarget(values TargetValues) (Target, error) {
	if values.ID == "" || values.ProjectID == "" || values.CreatedAt.IsZero() {
		return Target{}, fmt.Errorf("target: %w", ErrInvalidTarget)
	}
	return Target{
		id:         values.ID,
		projectID:  values.ProjectID,
		name:       values.Name,
		host:       values.Host,
		port:       values.Port,
		user:       values.User,
		remotePath: values.RemotePath,
		branch:     values.Branch,
		createdAt:  values.CreatedAt.UTC(),
	}, nil
}

func (t Target) ID() TargetID           { return t.id }
func (t Target) ProjectID() ProjectID   { return t.projectID }
func (t Target) Name() Name             { return t.name }
func (t Target) Host() Host             { return t.host }
func (t Target) Port() Port             { return t.port }
func (t Target) User() SSHUser          { return t.user }
func (t Target) RemotePath() RemotePath { return t.remotePath }
func (t Target) Branch() Branch         { return t.branch }
func (t Target) CreatedAt() time.Time   { return t.createdAt }
