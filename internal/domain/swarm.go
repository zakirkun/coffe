package domain

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrInvalidSwarmNode    = errors.New("domain: invalid swarm node")
	ErrInvalidSwarmService = errors.New("domain: invalid swarm service")
)

type SwarmNode struct {
	id           string
	hostname     string
	status       string
	availability string
	manager      bool
}

type SwarmNodeValues struct {
	ID            string
	Hostname      string
	Status        string
	Availability  string
	ManagerStatus string
}

func NewSwarmNode(values SwarmNodeValues) (SwarmNode, error) {
	if strings.TrimSpace(values.ID) == "" || strings.TrimSpace(values.Hostname) == "" ||
		strings.TrimSpace(values.Status) == "" || strings.TrimSpace(values.Availability) == "" {
		return SwarmNode{}, fmt.Errorf("swarm node: %w", ErrInvalidSwarmNode)
	}
	return SwarmNode{
		id: values.ID, hostname: values.Hostname, status: values.Status,
		availability: values.Availability, manager: strings.TrimSpace(values.ManagerStatus) != "",
	}, nil
}

func (n SwarmNode) ID() string           { return n.id }
func (n SwarmNode) Hostname() string     { return n.hostname }
func (n SwarmNode) Status() string       { return n.status }
func (n SwarmNode) Availability() string { return n.availability }
func (n SwarmNode) Manager() bool        { return n.manager }

type SwarmService struct {
	id       string
	name     string
	mode     string
	replicas string
	image    string
}

type SwarmServiceValues struct {
	ID       string
	Name     string
	Mode     string
	Replicas string
	Image    string
}

func NewSwarmService(values SwarmServiceValues) (SwarmService, error) {
	if strings.TrimSpace(values.ID) == "" || strings.TrimSpace(values.Name) == "" ||
		strings.TrimSpace(values.Mode) == "" || strings.TrimSpace(values.Replicas) == "" ||
		strings.TrimSpace(values.Image) == "" {
		return SwarmService{}, fmt.Errorf("swarm service: %w", ErrInvalidSwarmService)
	}
	return SwarmService{
		id: values.ID, name: values.Name, mode: values.Mode,
		replicas: values.Replicas, image: values.Image,
	}, nil
}

func (s SwarmService) ID() string       { return s.id }
func (s SwarmService) Name() string     { return s.name }
func (s SwarmService) Mode() string     { return s.mode }
func (s SwarmService) Replicas() string { return s.replicas }
func (s SwarmService) Image() string    { return s.image }
