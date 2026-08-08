package service_test

import (
	"coffe/internal/service"
	"context"
	"testing"
)

func Test_App_registers_and_heartbeats_worker(t *testing.T) {
	// Given
	fixture := newFixture(t)
	registered, err := fixture.app.RegisterWorker(context.Background(), service.RegisterWorkerInput{
		TargetID: string(fixture.target.ID()), Name: "office-worker",
		BaseURL: "https://worker.example.com", SecretName: "worker-token", Enabled: true,
	})
	if err != nil {
		t.Fatalf("RegisterWorker() error = %v", err)
	}

	// When
	updated, err := fixture.app.HeartbeatWorker(context.Background(), string(registered.ID()))
	// Then
	if err != nil {
		t.Fatalf("HeartbeatWorker() error = %v", err)
	}
	if updated.Version() != "v1.0.0" || updated.LastSeen().IsZero() {
		t.Fatalf("worker version/last_seen = %q/%v", updated.Version(), updated.LastSeen())
	}
}
