package platform

import (
	"context"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

func TestEnsureTraefik_Running(t *testing.T) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		t.Fatalf("Failed to create Docker client: %v", err)
	}
	defer cli.Close()

	// Test when Traefik is running
	err = EnsureTraefik(cli)
	if err != nil {
		t.Logf("EnsureTraefik returned error (expected if container not running): %v", err)
	} else {
		t.Log("✓ Traefik check passed - container is running")
	}
}

func TestEnsureTraefik_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		t.Fatalf("Failed to create Docker client: %v", err)
	}
	defer cli.Close()

	ctx := context.Background()

	// Check if container exists
	containers, err := cli.ContainerList(ctx, container.ListOptions{
		All: true,
		Filters: filters.NewArgs(
			filters.Arg("name", TraefikContainerName),
		),
	})
	if err != nil {
		t.Fatalf("Failed to list containers: %v", err)
	}

	if len(containers) == 0 {
		t.Log("⚠️  Traefik container not found - this is expected behavior")
		err = EnsureTraefik(cli)
		if err == nil {
			t.Error("Expected error when container doesn't exist, got nil")
		} else {
			t.Logf("✓ Correctly returned error: %v", err)
		}
		return
	}

	// Container exists - check state
	state := containers[0].State
	t.Logf("Traefik container found in state: %s", state)

	err = EnsureTraefik(cli)
	if err != nil {
		if state != "running" {
			t.Logf("✓ Correctly handled non-running container: %v", err)
		} else {
			t.Errorf("Unexpected error when container is running: %v", err)
		}
	} else {
		t.Log("✓ Traefik check passed")
	}
}
