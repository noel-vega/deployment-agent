package platform

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

const (
	// TraefikContainerName is the name of the Traefik container
	TraefikContainerName = "hubble-traefik"
	// TraefikImage is the Docker image used for Traefik
	TraefikImage = "traefik:v3.0"
	// TraefikDataPath is where Traefik stores its data (acme.json, etc.)
	TraefikDataPath = "/var/lib/hubble/traefik"
)

// TraefikConfig holds the configuration for Traefik
type TraefikConfig struct {
	Enabled          bool
	Email            string
	DashboardEnabled bool
	DashboardAuth    string
}

// GetTraefikConfig reads Traefik configuration from environment variables
func GetTraefikConfig() TraefikConfig {
	return TraefikConfig{
		Enabled:          os.Getenv("HUBBLE_TRAEFIK_ENABLED") == "true",
		Email:            os.Getenv("HUBBLE_TRAEFIK_EMAIL"),
		DashboardEnabled: os.Getenv("HUBBLE_TRAEFIK_DASHBOARD") == "true",
		DashboardAuth:    os.Getenv("HUBBLE_TRAEFIK_DASHBOARD_AUTH"),
	}
}

// EnsureTraefik ensures that Traefik is running if enabled
func EnsureTraefik(dockerClient *client.Client, config TraefikConfig) error {
	// If disabled, skip
	if !config.Enabled {
		log.Println("Traefik is disabled (set HUBBLE_TRAEFIK_ENABLED=true to enable)")
		return nil
	}

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

	// If container exists
	if len(containers) > 0 {
		c := containers[0]
		traefikID := c.ID

		// Check if it's running
		if c.State == "running" {
			log.Printf("✓ Traefik already running (ID: %s)", traefikID[:12])
			return nil
		}

		// If stopped, start it
		log.Printf("Starting Traefik container...")
		if err := dockerClient.ContainerStart(ctx, traefikID, container.StartOptions{}); err != nil {
			return fmt.Errorf("failed to start Traefik: %w", err)
		}
		log.Printf("✓ Traefik started (ID: %s)", traefikID[:12])
		return nil
	}

	// Container doesn't exist - log deprecation warning
	log.Println("⚠️  WARNING: Traefik container not found!")
	log.Println("⚠️  DEPRECATED: Programmatic Traefik creation is deprecated.")
	log.Println("⚠️  Please add hubble-traefik service to your docker-compose.yml")
	log.Println("⚠️  See: https://github.com/noel-vega/hubble#infrastructure-services")

	// Check if fallback creation is disabled
	if os.Getenv("HUBBLE_DISABLE_PLATFORM_FALLBACK") == "true" {
		log.Println("⚠️  Fallback creation disabled. Please start Traefik via docker-compose.")
		return fmt.Errorf("Traefik container not found and fallback creation is disabled")
	}

	log.Println("Creating Traefik container via fallback (deprecated)...")
	return createTraefikContainer(dockerClient, config)
}

func createTraefikContainer(dockerClient *client.Client, config TraefikConfig) error {
	ctx := context.Background()

	// Ensure Traefik image is available
	if err := ensureImageAvailable(dockerClient, TraefikImage); err != nil {
		return fmt.Errorf("failed to ensure Traefik image is available: %w", err)
	}

	// Docker volumes handle storage automatically - no manual directory creation needed
	log.Println("Using Docker volume for Traefik data storage")

	// Build Traefik command arguments
	cmd := []string{
		"--providers.docker=true",
		"--providers.docker.network=" + HubbleNetworkName,
		"--providers.docker.exposedbydefault=false",
		"--entrypoints.web.address=:80",
		"--entrypoints.websecure.address=:443",
		// HTTP to HTTPS redirect
		"--entrypoints.web.http.redirections.entrypoint.to=websecure",
		"--entrypoints.web.http.redirections.entrypoint.scheme=https",
	}

	// Add Let's Encrypt configuration if email is provided
	if config.Email != "" {
		cmd = append(cmd,
			"--certificatesresolvers.letsencrypt.acme.email="+config.Email,
			"--certificatesresolvers.letsencrypt.acme.storage=/data/acme.json",
			"--certificatesresolvers.letsencrypt.acme.httpchallenge.entrypoint=web",
		)
		log.Printf("Let's Encrypt configured with email: %s", config.Email)
	} else {
		log.Println("Warning: HUBBLE_TRAEFIK_EMAIL not set - HTTPS certificates will not be issued")
	}

	// Add dashboard configuration if enabled
	if config.DashboardEnabled {
		cmd = append(cmd, "--api.dashboard=true")
		if config.DashboardAuth != "" {
			cmd = append(cmd, "--api.dashboard.middlewares=auth")
			log.Println("Traefik dashboard enabled with authentication")
		} else {
			log.Println("Warning: Traefik dashboard enabled without authentication")
		}
	}

	// Define port bindings
	exposedPorts := nat.PortSet{
		"80/tcp":  struct{}{},
		"443/tcp": struct{}{},
	}
	portBindings := nat.PortMap{
		"80/tcp": []nat.PortBinding{
			{HostIP: "0.0.0.0", HostPort: "80"},
		},
		"443/tcp": []nat.PortBinding{
			{HostIP: "0.0.0.0", HostPort: "443"},
		},
	}

	// Add dashboard port if enabled
	if config.DashboardEnabled {
		exposedPorts["8080/tcp"] = struct{}{}
		portBindings["8080/tcp"] = []nat.PortBinding{
			{HostIP: "127.0.0.1", HostPort: "8080"}, // Localhost only for security
		}
	}

	// Container configuration
	containerConfig := &container.Config{
		Image:        TraefikImage,
		Cmd:          cmd,
		ExposedPorts: exposedPorts,
		Labels: map[string]string{
			"com.hubble.managed": "true",
			"com.hubble.service": "traefik",
		},
	}

	// Host configuration
	hostConfig := &container.HostConfig{
		RestartPolicy: container.RestartPolicy{
			Name: "unless-stopped",
		},
		PortBindings: portBindings,
		Mounts: []mount.Mount{
			{
				Type:     mount.TypeBind,
				Source:   "/var/run/docker.sock",
				Target:   "/var/run/docker.sock",
				ReadOnly: true,
			},
			{
				Type:   mount.TypeVolume,
				Source: "hubble-traefik-data",
				Target: "/data",
			},
		},
	}

	// Network configuration
	networkConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			HubbleNetworkName: {},
		},
	}

	// Create the container
	resp, err := dockerClient.ContainerCreate(
		ctx,
		containerConfig,
		hostConfig,
		networkConfig,
		nil,
		TraefikContainerName,
	)
	if err != nil {
		return fmt.Errorf("failed to create Traefik container: %w", err)
	}

	// Start the container
	if err := dockerClient.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return fmt.Errorf("failed to start Traefik container: %w", err)
	}

	log.Printf("✓ Created and started Traefik (ID: %s)", resp.ID[:12])
	if config.DashboardEnabled {
		log.Println("  Dashboard available at: http://localhost:8080")
	}
	log.Println("  HTTP: http://0.0.0.0:80")
	log.Println("  HTTPS: https://0.0.0.0:443")

	return nil
}

// ensureImageAvailable checks if a Docker image exists locally, and pulls it if not
func ensureImageAvailable(dockerClient *client.Client, imageName string) error {
	ctx := context.Background()

	// Check if image exists locally
	images, err := dockerClient.ImageList(ctx, image.ListOptions{
		Filters: filters.NewArgs(filters.Arg("reference", imageName)),
	})
	if err != nil {
		return fmt.Errorf("failed to list images: %w", err)
	}

	// Image already exists
	if len(images) > 0 {
		log.Printf("✓ Image %s already available locally", imageName)
		return nil
	}

	// Image doesn't exist, pull it
	log.Printf("Pulling image %s...", imageName)
	reader, err := dockerClient.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image %s: %w", imageName, err)
	}
	defer reader.Close()

	// Read the output to ensure pull completes
	if _, err := io.Copy(io.Discard, reader); err != nil {
		return fmt.Errorf("failed to read pull output: %w", err)
	}

	log.Printf("✓ Successfully pulled image %s", imageName)
	return nil
}
