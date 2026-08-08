package domain_test

import (
	"coffe/internal/domain"
	"testing"
	"time"
)

func Test_NewWorker_accepts_HTTPS_registry_and_heartbeat(t *testing.T) {
	// Given
	name, err := domain.NewName("office-worker")
	if err != nil {
		t.Fatalf("NewName() error = %v", err)
	}
	now := time.Date(2026, 8, 7, 10, 0, 0, 0, time.UTC)

	// When
	worker, err := domain.NewWorker(domain.WorkerValues{
		ID: "worker-1", TargetID: "target-1", Name: name,
		BaseURL: "https://worker.example.com", SecretName: "worker-office-token",
		Enabled: true, CreatedAt: now,
	})
	if err != nil {
		t.Fatalf("NewWorker() error = %v", err)
	}
	heartbeat, err := worker.WithHeartbeat("v1.0.0", now.Add(time.Minute))
	// Then
	if err != nil {
		t.Fatalf("WithHeartbeat() error = %v", err)
	}
	if heartbeat.Version() != "v1.0.0" || heartbeat.LastSeen().IsZero() {
		t.Fatalf("heartbeat version/last_seen = %q/%v", heartbeat.Version(), heartbeat.LastSeen())
	}
}
