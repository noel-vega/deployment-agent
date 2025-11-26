package platform

import (
	"context"
	"log"

	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

const (
	// HubbleNetworkName is the name of the shared network used by all Hubble services and projects
	HubbleNetworkName = "hubble"
)

// EnsureInfrastructure creates the necessary Docker infrastructure for Hubble to operate.
// This includes creating the shared 'hubble' network, and optionally starting Traefik and Registry.
func EnsureInfrastructure(dockerClient *client.Client) error {
	// Ensure Hubble network exists
	if err := ensureNetwork(dockerClient); err != nil {
		return err
	}

	// Check Traefik status (managed via docker-compose.yml)
	if err := EnsureTraefik(dockerClient); err != nil {
		log.Printf("Warning: %v", err)
		// Don't fail startup - Traefik is optional
	}

	// Check Registry status (managed via docker-compose.yml)
	if err := EnsureRegistry(dockerClient); err != nil {
		log.Printf("Warning: %v", err)
		// Don't fail startup - Registry is optional
	}

	return nil
}

// ensureNetwork creates the Hubble network if it doesn't exist
func ensureNetwork(dockerClient *client.Client) error {
	ctx := context.Background()

	// Check if the hubble network already exists
	networks, err := dockerClient.NetworkList(ctx, network.ListOptions{})
	if err != nil {
		return err
	}

	networkExists := false
	for _, net := range networks {
		if net.Name == HubbleNetworkName {
			networkExists = true
			log.Printf("✓ Hubble network '%s' already exists (ID: %s)", HubbleNetworkName, net.ID[:12])
			break
		}
	}

	// Create the network if it doesn't exist
	if !networkExists {
		log.Printf("Creating Hubble network '%s'...", HubbleNetworkName)
		resp, err := dockerClient.NetworkCreate(ctx, HubbleNetworkName, network.CreateOptions{
			Driver:     "bridge",
			Attachable: true,
			Labels: map[string]string{
				"com.hubble.managed": "true",
				"com.hubble.network": "platform",
			},
		})
		if err != nil {
			return err
		}
		log.Printf("✓ Created Hubble network '%s' (ID: %s)", HubbleNetworkName, resp.ID[:12])
	}

	return nil
}
