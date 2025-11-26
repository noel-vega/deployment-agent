package platform

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

const (
	// RegistryContainerName is the name of the Registry container
	RegistryContainerName = "hubble-registry"
	// RegistryImage is the Docker image used for the Registry
	RegistryImage = "registry:2"
	// RegistryDataPath is where Registry stores images
	RegistryDataPath = "/var/lib/hubble/registry"
	// RegistryAuthPath is where htpasswd file is stored
	RegistryAuthPath = "/var/lib/hubble/registry-auth"
)

// RegistryConfig holds the configuration for the Docker Registry
type RegistryConfig struct {
	Enabled       bool
	Domain        string
	DeleteEnabled bool
	StoragePath   string
	AuthPath      string
}

// GetRegistryConfig reads Registry configuration from environment variables
func GetRegistryConfig() RegistryConfig {
	enabled := os.Getenv("HUBBLE_REGISTRY_ENABLED")
	// Default to true (core feature) if not explicitly set to false
	isEnabled := enabled != "false"
	if enabled == "" {
		isEnabled = true // Default: enabled
	}

	storagePath := getRegistryStoragePath()
	authPath := getRegistryAuthPath(storagePath)

	return RegistryConfig{
		Enabled:       isEnabled,
		Domain:        os.Getenv("HUBBLE_DOMAIN"),
		DeleteEnabled: os.Getenv("HUBBLE_REGISTRY_DELETE_ENABLED") != "false", // Default: true
		StoragePath:   storagePath,
		AuthPath:      authPath,
	}
}

func getRegistryStoragePath() string {
	path := os.Getenv("HUBBLE_REGISTRY_STORAGE")
	if path == "" {
		return RegistryDataPath
	}
	return path
}

func getRegistryAuthPath(storagePath string) string {
	// If using default path, use default auth path
	if storagePath == RegistryDataPath {
		return RegistryAuthPath
	}
	// Otherwise, derive auth path from storage path
	// e.g. /home/user/hubble/tmp/registry -> /home/user/hubble/tmp/registry-auth
	return filepath.Dir(storagePath) + "/registry-auth"
}

// EnsureRegistry ensures that the Docker Registry is running if enabled
func EnsureRegistry(dockerClient *client.Client, config RegistryConfig) error {
	// If disabled, skip
	if !config.Enabled {
		log.Println("Registry is disabled (set HUBBLE_REGISTRY_ENABLED=true to enable)")
		return nil
	}

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

	// If container exists
	if len(containers) > 0 {
		c := containers[0]
		registryID := c.ID

		// Check if it's running
		if c.State == "running" {
			log.Printf("✓ Registry already running (ID: %s)", registryID[:12])
			if config.Domain != "" {
				log.Printf("  Registry URL: https://registry.%s", config.Domain)
			}
			return nil
		}

		// If stopped, start it
		log.Printf("Starting Registry container...")
		if err := dockerClient.ContainerStart(ctx, registryID, container.StartOptions{}); err != nil {
			return fmt.Errorf("failed to start Registry: %w", err)
		}
		log.Printf("✓ Registry started (ID: %s)", registryID[:12])
		if config.Domain != "" {
			log.Printf("  Registry URL: https://registry.%s", config.Domain)
		}
		return nil
	}

	// Container doesn't exist, create it
	log.Println("Creating Registry container...")
	return createRegistryContainer(dockerClient, config)
}

func createRegistryContainer(dockerClient *client.Client, config RegistryConfig) error {
	ctx := context.Background()

	// Ensure Registry image is available
	if err := ensureImageAvailable(dockerClient, RegistryImage); err != nil {
		return fmt.Errorf("failed to ensure Registry image is available: %w", err)
	}

	// Create htpasswd file in the Docker volume
	log.Println("Creating htpasswd file in Docker volume...")
	if err := createHtpasswdFileInVolume(); err != nil {
		return fmt.Errorf("failed to create htpasswd file: %w", err)
	}

	// Build environment variables
	env := []string{
		"REGISTRY_AUTH=htpasswd",
		"REGISTRY_AUTH_HTPASSWD_REALM=Hubble Registry",
		"REGISTRY_AUTH_HTPASSWD_PATH=/auth/htpasswd",
	}

	if config.DeleteEnabled {
		env = append(env, "REGISTRY_STORAGE_DELETE_ENABLED=true")
	}

	// Build labels
	labels := map[string]string{
		"com.hubble.managed": "true",
		"com.hubble.service": "registry",
	}

	// Add Traefik labels if domain is configured
	if config.Domain != "" {
		labels["traefik.enable"] = "true"
		labels["traefik.http.routers.registry.rule"] = fmt.Sprintf("Host(`registry.%s`)", config.Domain)
		labels["traefik.http.routers.registry.entrypoints"] = "websecure"
		labels["traefik.http.routers.registry.tls.certresolver"] = "letsencrypt"
		labels["traefik.http.services.registry.loadbalancer.server.port"] = "5000"
		log.Printf("Registry configured for Traefik: https://registry.%s", config.Domain)
	} else {
		log.Println("Warning: HUBBLE_DOMAIN not set - Registry will not be accessible via Traefik")
		log.Println("         Set HUBBLE_DOMAIN=yourdomain.com to enable HTTPS access")
	}

	// Container configuration
	containerConfig := &container.Config{
		Image:  RegistryImage,
		Env:    env,
		Labels: labels,
	}

	// Host configuration with port binding for local access (when Traefik is disabled)
	portBindings := nat.PortMap{}
	if config.Domain == "" || config.Domain == "localhost" {
		// Expose registry on localhost:5000 for local testing
		portBindings["5000/tcp"] = []nat.PortBinding{
			{
				HostIP:   "127.0.0.1",
				HostPort: "5000",
			},
		}
	}

	hostConfig := &container.HostConfig{
		RestartPolicy: container.RestartPolicy{
			Name: "unless-stopped",
		},
		PortBindings: portBindings,
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeVolume,
				Source: "hubble-registry-data",
				Target: "/var/lib/registry",
			},
			{
				Type:     mount.TypeVolume,
				Source:   "hubble-registry-auth",
				Target:   "/auth",
				ReadOnly: true,
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
		RegistryContainerName,
	)
	if err != nil {
		return fmt.Errorf("failed to create Registry container: %w", err)
	}

	// Start the container
	if err := dockerClient.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return fmt.Errorf("failed to start Registry container: %w", err)
	}

	log.Printf("✓ Created and started Registry (ID: %s)", resp.ID[:12])
	if config.Domain != "" {
		log.Printf("  Registry URL: https://registry.%s", config.Domain)
		log.Printf("  Login: docker login registry.%s", config.Domain)
	}
	log.Printf("  Storage: %s", config.StoragePath)

	return nil
}

// createHtpasswdFileInVolume creates an htpasswd file in the Docker volume
func createHtpasswdFileInVolume() error {
	username := os.Getenv("ADMIN_USERNAME")
	password := os.Getenv("ADMIN_PASSWORD")

	if username == "" || password == "" {
		return fmt.Errorf("ADMIN_USERNAME and ADMIN_PASSWORD must be set")
	}

	// Use Docker to generate htpasswd and write directly to the volume
	// Mount the volume and run htpasswd command to create the file
	cmd := exec.Command("docker", "run", "--rm",
		"-v", "hubble-registry-auth:/auth",
		"httpd:alpine",
		"sh", "-c",
		fmt.Sprintf("htpasswd -Bbn %s %s > /auth/htpasswd", username, password),
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create htpasswd in volume: %w, output: %s", err, string(output))
	}

	log.Println("✓ htpasswd file created in Docker volume")
	return nil
}
