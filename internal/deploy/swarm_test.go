package deploy_test

import (
	"coffe/internal/deploy"
	"coffe/internal/domain"
	"context"
	"fmt"
	"testing"
)

func Test_SwarmInspector_parses_nodes_and_services(t *testing.T) {
	// Given
	target := newTarget(t, "/srv/payments")
	runner := &payloadRunner{payloads: map[string]string{
		"docker node ls":        `{"ID":"node-1","Hostname":"manager-1","Status":"Ready","Availability":"Active","ManagerStatus":"Leader"}` + "\n",
		"docker stack services": `{"ID":"service-1","Name":"payments_api","Mode":"replicated","Replicas":"2/2","Image":"payments:1"}` + "\n",
	}}
	inspector := deploy.NewSwarmInspector(runner)
	stackName, err := domain.NewStackName("payments")
	if err != nil {
		t.Fatalf("NewStackName() error = %v", err)
	}

	// When
	nodes, err := inspector.ListNodes(context.Background(), target)
	if err != nil {
		t.Fatalf("ListNodes() error = %v", err)
	}
	services, err := inspector.ListServices(context.Background(), target, stackName)
	// Then
	if err != nil {
		t.Fatalf("ListServices() error = %v", err)
	}
	if len(nodes) != 1 || nodes[0].Hostname() != "manager-1" || !nodes[0].Manager() {
		t.Fatalf("nodes = %v", nodes)
	}
	if len(services) != 1 || services[0].Replicas() != "2/2" {
		t.Fatalf("services = %v", services)
	}
}

type payloadRunner struct{ payloads map[string]string }

func (r *payloadRunner) Run(_ context.Context, request deploy.RunRequest) error {
	for prefix, payload := range r.payloads {
		if len(request.Command) >= len(prefix) && request.Command[:len(prefix)] == prefix {
			if _, err := fmt.Fprint(request.Output, payload); err != nil {
				return err
			}
			return nil
		}
	}
	return fmt.Errorf("unexpected command %q", request.Command)
}
