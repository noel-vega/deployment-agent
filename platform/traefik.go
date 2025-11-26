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
	// TraefikContainerName is the name of the Traefik container
	TraefikContainerName = "hubble-traefik"
)

// EnsureTraefik checks that Traefik is running (started via docker-compose.yml)
func EnsureTraefik(dockerClient *client.Client) error {
	ctx := context.Background()

	// Check if Traefik container exists
	listFilters := filters.NewArgs()
	listFilters.Add("name", TraefikContainerName)
	containers, err := dockerClient.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: listFilters,
	})
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	// Container doesn't exist
	if len(containers) == 0 {
		log.Println("⚠️  Traefik container not found!")
		log.Println("⚠️  Please ensure hubble-traefik service is running via docker-compose.yml")
		log.Println("⚠️  Run: docker compose up -d")
		return fmt.Errorf("Traefik container not found - start via docker-compose")
	}

	c := containers[0]
	traefikID := c.ID

	// Check if it's running
	if c.State == "running" {
		log.Printf("✓ Traefik running (ID: %s)", traefikID[:12])
		return nil
	}

	// If stopped, try to start it
	log.Printf("Starting Traefik container...")
	if err := dockerClient.ContainerStart(ctx, traefikID, container.StartOptions{}); err != nil {
		return fmt.Errorf("failed to start Traefik: %w", err)
	}
	log.Printf("✓ Traefik started (ID: %s)", traefikID[:12])
	return nil
}
