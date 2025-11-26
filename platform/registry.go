package platform

import (
	"context"
	"fmt"
	"log"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

const (
	// RegistryContainerName is the name of the Registry container
	RegistryContainerName = "hubble-registry"
)

// EnsureRegistry checks that Registry is running (started via docker-compose.yml)
func EnsureRegistry(dockerClient *client.Client) error {
	ctx := context.Background()

	// Check if Registry container exists
	listFilters := filters.NewArgs()
	listFilters.Add("name", RegistryContainerName)
	containers, err := dockerClient.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: listFilters,
	})
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	// Container doesn't exist
	if len(containers) == 0 {
		log.Println("⚠️  Registry container not found!")
		log.Println("⚠️  Please ensure hubble-registry service is running via docker-compose.yml")
		log.Println("⚠️  Run: docker compose up -d")
		return fmt.Errorf("Registry container not found - start via docker-compose")
	}

	c := containers[0]
	registryID := c.ID

	// Check if it's running
	if c.State == "running" {
		log.Printf("✓ Registry running (ID: %s)", registryID[:12])
		return nil
	}

	// If stopped, try to start it
	log.Printf("Starting Registry container...")
	if err := dockerClient.ContainerStart(ctx, registryID, container.StartOptions{}); err != nil {
		return fmt.Errorf("failed to start Registry: %w", err)
	}
	log.Printf("✓ Registry started (ID: %s)", registryID[:12])
	return nil
}
