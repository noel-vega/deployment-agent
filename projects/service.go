package projects

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"gopkg.in/yaml.v3"
)

type Service struct {
	rootPath     string
	dockerClient *client.Client
}

type ProjectInfo struct {
	Name              string `json:"name"`
	Path              string `json:"path"`
	ServiceCount      int    `json:"service_count"`
	ContainersRunning int    `json:"containers_running"`
	ContainersStopped int    `json:"containers_stopped"`
}

type ProjectContainerInfo struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Service string `json:"service"`
	State   string `json:"state"`
	Status  string `json:"status"`
}

type ProjectVolume struct {
	Service string `json:"service"`
	Volume  string `json:"volume"`
}

type ProjectEnvironment struct {
	Service string            `json:"service"`
	Env     map[string]string `json:"env"`
}

type ProjectNetwork struct {
	Name   string                 `json:"name"`
	Driver string                 `json:"driver"`
	Config map[string]interface{} `json:"config"`
}

type ComposeService struct {
	Name        string            `json:"name"`
	Image       string            `json:"image"`
	Build       string            `json:"build"`
	Ports       []string          `json:"ports"`
	Environment map[string]string `json:"environment"`
	Volumes     []string          `json:"volumes"`
	DependsOn   []string          `json:"depends_on"`
	Networks    []string          `json:"networks"`
	Restart     string            `json:"restart"`
	Command     string            `json:"command"`
	Status      string            `json:"status"`
}

type ComposeFile struct {
	Version  string                 `yaml:"version,omitempty"`
	Services map[string]interface{} `yaml:"services"`
	Networks map[string]interface{} `yaml:"networks,omitempty"`
	Volumes  map[string]interface{} `yaml:"volumes,omitempty"`
	Secrets  map[string]interface{} `yaml:"secrets,omitempty"`
	Configs  map[string]interface{} `yaml:"configs,omitempty"`
}

func NewService(dockerClient *client.Client) (*Service, error) {
	rootPath := os.Getenv("PROJECTS_ROOT_PATH")
	if rootPath == "" {
		return nil, fmt.Errorf("PROJECTS_ROOT_PATH environment variable is not set")
	}

	// Check if the root path exists
	if _, err := os.Stat(rootPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("projects root path does not exist: %s", rootPath)
	}

	return &Service{
		rootPath:     rootPath,
		dockerClient: dockerClient,
	}, nil
}

func (s *Service) ListProjects(ctx context.Context) ([]ProjectInfo, error) {
	entries, err := os.ReadDir(s.rootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read projects directory: %w", err)
	}

	projects := make([]ProjectInfo, 0)

	for _, entry := range entries {
		// Skip non-directories
		if !entry.IsDir() {
			continue
		}

		projectName := entry.Name()
		projectPath := filepath.Join(s.rootPath, projectName)

		// Check for docker-compose.yml or docker-compose.yaml
		composeFile := ""
		for _, filename := range []string{"docker-compose.yml", "docker-compose.yaml"} {
			composePath := filepath.Join(projectPath, filename)
			if _, err := os.Stat(composePath); err == nil {
				composeFile = composePath
				break
			}
		}

		// Only include directories that have a docker-compose file
		if composeFile != "" {
			// Read and parse the compose file to count services
			serviceCount := 0
			content, err := os.ReadFile(composeFile)
			if err == nil {
				var compose ComposeFile
				if err := yaml.Unmarshal(content, &compose); err == nil {
					serviceCount = len(compose.Services)
				}
			}

			// Get container counts for this project
			running, stopped := s.getContainerCounts(ctx, projectName)

			projects = append(projects, ProjectInfo{
				Name:              projectName,
				Path:              projectPath,
				ServiceCount:      serviceCount,
				ContainersRunning: running,
				ContainersStopped: stopped,
			})
		}
	}

	return projects, nil
}

func (s *Service) GetProject(ctx context.Context, projectName string) (*ProjectInfo, error) {
	projectPath := filepath.Join(s.rootPath, projectName)

	// Check if project directory exists
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("project not found: %s", projectName)
	}

	// Find the compose file
	var composeFilePath string
	for _, filename := range []string{"docker-compose.yml", "docker-compose.yaml"} {
		path := filepath.Join(projectPath, filename)
		if _, err := os.Stat(path); err == nil {
			composeFilePath = path
			break
		}
	}

	if composeFilePath == "" {
		return nil, fmt.Errorf("no docker-compose file found in project: %s", projectName)
	}

	// Read and parse the compose file to count services
	serviceCount := 0
	content, err := os.ReadFile(composeFilePath)
	if err == nil {
		var compose ComposeFile
		if err := yaml.Unmarshal(content, &compose); err == nil {
			serviceCount = len(compose.Services)
		}
	}

	// Get container counts for this project
	running, stopped := s.getContainerCounts(ctx, projectName)

	return &ProjectInfo{
		Name:              projectName,
		Path:              projectPath,
		ServiceCount:      serviceCount,
		ContainersRunning: running,
		ContainersStopped: stopped,
	}, nil
}

func (s *Service) GetProjectCompose(ctx context.Context, projectName string) (string, error) {
	projectPath := filepath.Join(s.rootPath, projectName)

	// Find the compose file
	var composeFilePath string
	for _, filename := range []string{"docker-compose.yml", "docker-compose.yaml"} {
		path := filepath.Join(projectPath, filename)
		if _, err := os.Stat(path); err == nil {
			composeFilePath = path
			break
		}
	}

	if composeFilePath == "" {
		return "", fmt.Errorf("no docker-compose file found in project: %s", projectName)
	}

	content, err := os.ReadFile(composeFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to read compose file: %w", err)
	}

	return string(content), nil
}

func (s *Service) GetProjectContainers(ctx context.Context, projectName string) ([]ProjectContainerInfo, error) {
	if s.dockerClient == nil {
		return []ProjectContainerInfo{}, nil
	}

	filterArgs := filters.NewArgs()
	filterArgs.Add("label", fmt.Sprintf("com.docker.compose.project=%s", projectName))

	containers, err := s.dockerClient.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: filterArgs,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	result := make([]ProjectContainerInfo, 0, len(containers))
	for _, c := range containers {
		name := ""
		if len(c.Names) > 0 {
			name = c.Names[0]
			if len(name) > 0 && name[0] == '/' {
				name = name[1:]
			}
		}

		serviceName := c.Labels["com.docker.compose.service"]

		result = append(result, ProjectContainerInfo{
			ID:      c.ID[:12],
			Name:    name,
			Service: serviceName,
			State:   c.State,
			Status:  c.Status,
		})
	}

	return result, nil
}

func (s *Service) GetProjectVolumes(ctx context.Context, projectName string) ([]ProjectVolume, error) {
	projectPath := filepath.Join(s.rootPath, projectName)

	// Find and read compose file
	var composeFilePath string
	for _, filename := range []string{"docker-compose.yml", "docker-compose.yaml"} {
		path := filepath.Join(projectPath, filename)
		if _, err := os.Stat(path); err == nil {
			composeFilePath = path
			break
		}
	}

	if composeFilePath == "" {
		return nil, fmt.Errorf("no docker-compose file found in project: %s", projectName)
	}

	content, err := os.ReadFile(composeFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read compose file: %w", err)
	}

	var compose ComposeFile
	volumes := []ProjectVolume{}

	if err := yaml.Unmarshal(content, &compose); err == nil {
		for serviceName, serviceData := range compose.Services {
			if svcMap, ok := serviceData.(map[string]interface{}); ok {
				if vols, ok := svcMap["volumes"].([]interface{}); ok {
					for _, vol := range vols {
						if volStr, ok := vol.(string); ok {
							volumes = append(volumes, ProjectVolume{
								Service: serviceName,
								Volume:  volStr,
							})
						}
					}
				}
			}
		}
	}

	return volumes, nil
}

func (s *Service) GetProjectEnvironment(ctx context.Context, projectName string) ([]ProjectEnvironment, error) {
	projectPath := filepath.Join(s.rootPath, projectName)

	// Find and read compose file
	var composeFilePath string
	for _, filename := range []string{"docker-compose.yml", "docker-compose.yaml"} {
		path := filepath.Join(projectPath, filename)
		if _, err := os.Stat(path); err == nil {
			composeFilePath = path
			break
		}
	}

	if composeFilePath == "" {
		return nil, fmt.Errorf("no docker-compose file found in project: %s", projectName)
	}

	content, err := os.ReadFile(composeFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read compose file: %w", err)
	}

	var compose ComposeFile
	environments := []ProjectEnvironment{}

	if err := yaml.Unmarshal(content, &compose); err == nil {
		for serviceName, serviceData := range compose.Services {
			if svcMap, ok := serviceData.(map[string]interface{}); ok {
				envVars := make(map[string]string)

				// Handle environment as map
				if env, ok := svcMap["environment"].(map[string]interface{}); ok {
					for k, v := range env {
						if vStr, ok := v.(string); ok {
							envVars[k] = vStr
						}
					}
				}

				// Handle environment as array
				if envArray, ok := svcMap["environment"].([]interface{}); ok {
					for _, envItem := range envArray {
						if envStr, ok := envItem.(string); ok {
							// Parse KEY=VALUE format
							if idx := filepath.Base(envStr); idx != "" {
								envVars[envStr] = ""
							}
						}
					}
				}

				if len(envVars) > 0 {
					environments = append(environments, ProjectEnvironment{
						Service: serviceName,
						Env:     envVars,
					})
				}
			}
		}
	}

	return environments, nil
}

func (s *Service) GetProjectNetworks(ctx context.Context, projectName string) ([]ProjectNetwork, error) {
	projectPath := filepath.Join(s.rootPath, projectName)

	// Find and read compose file
	var composeFilePath string
	for _, filename := range []string{"docker-compose.yml", "docker-compose.yaml"} {
		path := filepath.Join(projectPath, filename)
		if _, err := os.Stat(path); err == nil {
			composeFilePath = path
			break
		}
	}

	if composeFilePath == "" {
		return nil, fmt.Errorf("no docker-compose file found in project: %s", projectName)
	}

	content, err := os.ReadFile(composeFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read compose file: %w", err)
	}

	var composeData map[string]interface{}
	networks := []ProjectNetwork{}

	if err := yaml.Unmarshal(content, &composeData); err == nil {
		if networksData, ok := composeData["networks"].(map[string]interface{}); ok {
			for networkName, networkConfig := range networksData {
				network := ProjectNetwork{
					Name:   networkName,
					Config: make(map[string]interface{}),
				}

				if netMap, ok := networkConfig.(map[string]interface{}); ok {
					if driver, ok := netMap["driver"].(string); ok {
						network.Driver = driver
					}
					network.Config = netMap
				}

				networks = append(networks, network)
			}
		}
	}

	return networks, nil
}

func (s *Service) GetProjectServices(ctx context.Context, projectName string) ([]ComposeService, error) {
	projectPath := filepath.Join(s.rootPath, projectName)

	// Find and read compose file
	var composeFilePath string
	for _, filename := range []string{"docker-compose.yml", "docker-compose.yaml"} {
		path := filepath.Join(projectPath, filename)
		if _, err := os.Stat(path); err == nil {
			composeFilePath = path
			break
		}
	}

	if composeFilePath == "" {
		return nil, fmt.Errorf("no docker-compose file found in project: %s", projectName)
	}

	content, err := os.ReadFile(composeFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read compose file: %w", err)
	}

	var compose ComposeFile
	services := []ComposeService{}

	// Get status for all services in this project
	serviceStatuses := s.getServiceStatuses(ctx, projectName)

	if err := yaml.Unmarshal(content, &compose); err == nil {
		for serviceName, serviceData := range compose.Services {
			// Initialize with empty slices and maps
			service := ComposeService{
				Name:        serviceName,
				Ports:       []string{},
				Environment: map[string]string{},
				Volumes:     []string{},
				DependsOn:   []string{},
				Networks:    []string{},
				Status:      "not_created", // Default status
			}

			// Set actual status if container exists
			if status, exists := serviceStatuses[serviceName]; exists {
				service.Status = status
			}

			if svcMap, ok := serviceData.(map[string]interface{}); ok {
				// Image
				if image, ok := svcMap["image"].(string); ok {
					service.Image = image
				}

				// Build (can be string or map)
				if build, ok := svcMap["build"].(string); ok {
					service.Build = build
				} else if buildMap, ok := svcMap["build"].(map[string]interface{}); ok {
					if context, ok := buildMap["context"].(string); ok {
						service.Build = context
					}
				}

				// Ports
				if ports, ok := svcMap["ports"].([]interface{}); ok {
					portsList := []string{}
					for _, port := range ports {
						if portStr, ok := port.(string); ok {
							portsList = append(portsList, portStr)
						}
					}
					if len(portsList) > 0 {
						service.Ports = portsList
					}
				}

				// Environment (map or array format)
				if env, ok := svcMap["environment"].(map[string]interface{}); ok {
					envMap := map[string]string{}
					for k, v := range env {
						if vStr, ok := v.(string); ok {
							envMap[k] = vStr
						}
					}
					if len(envMap) > 0 {
						service.Environment = envMap
					}
				} else if envArray, ok := svcMap["environment"].([]interface{}); ok {
					envMap := map[string]string{}
					for _, envItem := range envArray {
						if envStr, ok := envItem.(string); ok {
							envMap[envStr] = ""
						}
					}
					if len(envMap) > 0 {
						service.Environment = envMap
					}
				}

				// Volumes
				if volumes, ok := svcMap["volumes"].([]interface{}); ok {
					volumesList := []string{}
					for _, vol := range volumes {
						if volStr, ok := vol.(string); ok {
							volumesList = append(volumesList, volStr)
						}
					}
					if len(volumesList) > 0 {
						service.Volumes = volumesList
					}
				}

				// DependsOn (array or map format)
				if dependsOn, ok := svcMap["depends_on"].([]interface{}); ok {
					depsList := []string{}
					for _, dep := range dependsOn {
						if depStr, ok := dep.(string); ok {
							depsList = append(depsList, depStr)
						}
					}
					if len(depsList) > 0 {
						service.DependsOn = depsList
					}
				} else if dependsOnMap, ok := svcMap["depends_on"].(map[string]interface{}); ok {
					depsList := []string{}
					for dep := range dependsOnMap {
						depsList = append(depsList, dep)
					}
					if len(depsList) > 0 {
						service.DependsOn = depsList
					}
				}

				// Networks
				if networks, ok := svcMap["networks"].([]interface{}); ok {
					netsList := []string{}
					for _, net := range networks {
						if netStr, ok := net.(string); ok {
							netsList = append(netsList, netStr)
						}
					}
					if len(netsList) > 0 {
						service.Networks = netsList
					}
				} else if networksMap, ok := svcMap["networks"].(map[string]interface{}); ok {
					netsList := []string{}
					for net := range networksMap {
						netsList = append(netsList, net)
					}
					if len(netsList) > 0 {
						service.Networks = netsList
					}
				}

				// Restart
				if restart, ok := svcMap["restart"].(string); ok {
					service.Restart = restart
				}

				// Command (string or array)
				if command, ok := svcMap["command"].(string); ok {
					service.Command = command
				} else if commandArray, ok := svcMap["command"].([]interface{}); ok {
					cmdParts := []string{}
					for _, cmd := range commandArray {
						if cmdStr, ok := cmd.(string); ok {
							cmdParts = append(cmdParts, cmdStr)
						}
					}
					if len(cmdParts) > 0 {
						service.Command = fmt.Sprintf("%v", cmdParts)
					}
				}
			}

			services = append(services, service)
		}
	}

	return services, nil
}

func (s *Service) getContainerCounts(ctx context.Context, projectName string) (running, stopped int) {
	// If docker client is not available, return zeros
	if s.dockerClient == nil {
		return 0, 0
	}

	// Filter containers by project label (docker-compose project label)
	filterArgs := filters.NewArgs()
	filterArgs.Add("label", fmt.Sprintf("com.docker.compose.project=%s", projectName))

	containers, err := s.dockerClient.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: filterArgs,
	})
	if err != nil {
		return 0, 0
	}

	for _, c := range containers {
		if c.State == "running" {
			running++
		} else {
			stopped++
		}
	}

	return running, stopped
}

func (s *Service) getServiceStatuses(ctx context.Context, projectName string) map[string]string {
	statuses := make(map[string]string)

	if s.dockerClient == nil {
		return statuses
	}

	filterArgs := filters.NewArgs()
	filterArgs.Add("label", fmt.Sprintf("com.docker.compose.project=%s", projectName))

	containers, err := s.dockerClient.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: filterArgs,
	})
	if err != nil {
		return statuses
	}

	// Map service name to highest priority status
	// Priority: running > stopped > not_created
	for _, c := range containers {
		serviceName := c.Labels["com.docker.compose.service"]
		if serviceName == "" {
			continue
		}

		currentStatus := statuses[serviceName]

		// If any container is running, mark service as running
		if c.State == "running" {
			statuses[serviceName] = "running"
		} else if currentStatus != "running" {
			// Only set to stopped if not already marked as running
			statuses[serviceName] = "stopped"
		}
	}

	return statuses
}

func (s *Service) StartService(ctx context.Context, projectName, serviceName string) error {
	if s.dockerClient == nil {
		return fmt.Errorf("docker client not available")
	}

	// Find containers for this service
	filterArgs := filters.NewArgs()
	filterArgs.Add("label", fmt.Sprintf("com.docker.compose.project=%s", projectName))
	filterArgs.Add("label", fmt.Sprintf("com.docker.compose.service=%s", serviceName))

	containers, err := s.dockerClient.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: filterArgs,
	})
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	if len(containers) == 0 {
		// No containers found - create and start using docker-compose
		return s.dockerComposeUp(ctx, projectName, serviceName)
	}

	// Containers exist - start them
	var startErrors []string
	for _, c := range containers {
		if c.State != "running" {
			if err := s.dockerClient.ContainerStart(ctx, c.ID, container.StartOptions{}); err != nil {
				startErrors = append(startErrors, fmt.Sprintf("container %s: %v", c.ID[:12], err))
			}
		}
	}

	if len(startErrors) > 0 {
		return fmt.Errorf("failed to start some containers: %v", startErrors)
	}

	return nil
}

func (s *Service) dockerComposeUp(ctx context.Context, projectName, serviceName string) error {
	projectPath := filepath.Join(s.rootPath, projectName)

	// Verify project directory exists
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		return fmt.Errorf("project directory not found: %s", projectPath)
	}

	// Verify docker-compose file exists
	composeFiles := []string{"docker-compose.yml", "docker-compose.yaml"}
	var composeFile string
	for _, filename := range composeFiles {
		path := filepath.Join(projectPath, filename)
		if _, err := os.Stat(path); err == nil {
			composeFile = filename
			break
		}
	}

	if composeFile == "" {
		return fmt.Errorf("no docker-compose file found in project: %s", projectName)
	}

	// Run docker compose up -d <service>
	// Using "docker compose" (modern plugin) instead of "docker-compose" (legacy)
	cmd := exec.CommandContext(ctx, "docker", "compose", "-f", composeFile, "-p", projectName, "up", "-d", serviceName)
	cmd.Dir = projectPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to start service with docker compose: %w (output: %s)", err, string(output))
	}

	return nil
}

func (s *Service) StopService(ctx context.Context, projectName, serviceName string) error {
	if s.dockerClient == nil {
		return fmt.Errorf("docker client not available")
	}

	// Find containers for this service
	filterArgs := filters.NewArgs()
	filterArgs.Add("label", fmt.Sprintf("com.docker.compose.project=%s", projectName))
	filterArgs.Add("label", fmt.Sprintf("com.docker.compose.service=%s", serviceName))

	containers, err := s.dockerClient.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: filterArgs,
	})
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	if len(containers) == 0 {
		return fmt.Errorf("no containers found for service %s in project %s", serviceName, projectName)
	}

	// Stop all containers for this service
	var stopErrors []string
	for _, c := range containers {
		if c.State == "running" {
			if err := s.dockerClient.ContainerStop(ctx, c.ID, container.StopOptions{}); err != nil {
				stopErrors = append(stopErrors, fmt.Sprintf("container %s: %v", c.ID[:12], err))
			}
		}
	}

	if len(stopErrors) > 0 {
		return fmt.Errorf("failed to stop some containers: %v", stopErrors)
	}

	return nil
}

// CreateProject creates a new project with an empty docker-compose.yml
func (s *Service) CreateProject(ctx context.Context, name string) error {
	// Validate project name
	if name == "" {
		return fmt.Errorf("project name cannot be empty")
	}

	// Check if project already exists
	projectPath := filepath.Join(s.rootPath, name)
	if _, err := os.Stat(projectPath); err == nil {
		return fmt.Errorf("project already exists: %s", name)
	}

	// Create project directory
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	// Create minimal docker-compose.yml
	// Note: version field is obsolete in Docker Compose v2 and should be omitted
	compose := ComposeFile{
		Services: make(map[string]interface{}),
	}

	composeFilePath := filepath.Join(projectPath, "docker-compose.yml")
	output, err := yaml.Marshal(&compose)
	if err != nil {
		return fmt.Errorf("failed to marshal compose file: %w", err)
	}

	if err := os.WriteFile(composeFilePath, output, 0644); err != nil {
		return fmt.Errorf("failed to write compose file: %w", err)
	}

	return nil
}

// AddService adds a new service to an existing project
func (s *Service) AddService(ctx context.Context, projectName string, service ComposeService) error {
	// Validate service name
	if service.Name == "" {
		return fmt.Errorf("service name cannot be empty")
	}

	return s.updateComposeFile(projectName, func(compose *ComposeFile) error {
		// Check if service already exists
		if _, exists := compose.Services[service.Name]; exists {
			return fmt.Errorf("service already exists: %s", service.Name)
		}

		// Build service configuration
		serviceConfig := make(map[string]interface{})

		if service.Image != "" {
			serviceConfig["image"] = service.Image
		}

		if service.Build != "" {
			serviceConfig["build"] = service.Build
		}

		if len(service.Ports) > 0 {
			serviceConfig["ports"] = service.Ports
		}

		if len(service.Environment) > 0 {
			serviceConfig["environment"] = service.Environment
		}

		if len(service.Volumes) > 0 {
			serviceConfig["volumes"] = service.Volumes
		}

		if len(service.DependsOn) > 0 {
			serviceConfig["depends_on"] = service.DependsOn
		}

		if len(service.Networks) > 0 {
			serviceConfig["networks"] = service.Networks
		}

		if service.Restart != "" {
			serviceConfig["restart"] = service.Restart
		}

		if service.Command != "" {
			serviceConfig["command"] = service.Command
		}

		compose.Services[service.Name] = serviceConfig
		return nil
	})
}

// UpdateService updates an existing service in a project
func (s *Service) UpdateService(ctx context.Context, projectName string, service ComposeService) error {
	// Validate service name
	if service.Name == "" {
		return fmt.Errorf("service name cannot be empty")
	}

	return s.updateComposeFile(projectName, func(compose *ComposeFile) error {
		// Check if service exists
		if _, exists := compose.Services[service.Name]; !exists {
			return fmt.Errorf("service not found: %s", service.Name)
		}

		// Build updated service configuration
		serviceConfig := make(map[string]interface{})

		if service.Image != "" {
			serviceConfig["image"] = service.Image
		}

		if service.Build != "" {
			serviceConfig["build"] = service.Build
		}

		if len(service.Ports) > 0 {
			serviceConfig["ports"] = service.Ports
		}

		if len(service.Environment) > 0 {
			serviceConfig["environment"] = service.Environment
		}

		if len(service.Volumes) > 0 {
			serviceConfig["volumes"] = service.Volumes
		}

		if len(service.DependsOn) > 0 {
			serviceConfig["depends_on"] = service.DependsOn
		}

		if len(service.Networks) > 0 {
			serviceConfig["networks"] = service.Networks
		}

		if service.Restart != "" {
			serviceConfig["restart"] = service.Restart
		}

		if service.Command != "" {
			serviceConfig["command"] = service.Command
		}

		compose.Services[service.Name] = serviceConfig
		return nil
	})
}

// DeleteService removes a service from a project
func (s *Service) DeleteService(ctx context.Context, projectName, serviceName string) error {
	if serviceName == "" {
		return fmt.Errorf("service name cannot be empty")
	}

	return s.updateComposeFile(projectName, func(compose *ComposeFile) error {
		// Check if service exists
		if _, exists := compose.Services[serviceName]; !exists {
			return fmt.Errorf("service not found: %s", serviceName)
		}

		delete(compose.Services, serviceName)
		return nil
	})
}

// updateComposeFile is a helper function to safely update docker-compose.yml
func (s *Service) updateComposeFile(projectName string, updateFn func(*ComposeFile) error) error {
	projectPath := filepath.Join(s.rootPath, projectName)

	// Check if project exists
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		return fmt.Errorf("project not found: %s", projectName)
	}

	// Find compose file
	var composeFilePath string
	for _, filename := range []string{"docker-compose.yml", "docker-compose.yaml"} {
		path := filepath.Join(projectPath, filename)
		if _, err := os.Stat(path); err == nil {
			composeFilePath = path
			break
		}
	}

	if composeFilePath == "" {
		return fmt.Errorf("no docker-compose file found in project: %s", projectName)
	}

	// Read existing compose file
	content, err := os.ReadFile(composeFilePath)
	if err != nil {
		return fmt.Errorf("failed to read compose file: %w", err)
	}

	// Parse compose file
	var compose ComposeFile
	if err := yaml.Unmarshal(content, &compose); err != nil {
		return fmt.Errorf("failed to parse compose file: %w", err)
	}

	// Ensure services map is initialized
	if compose.Services == nil {
		compose.Services = make(map[string]interface{})
	}

	// Apply update function
	if err := updateFn(&compose); err != nil {
		return err
	}

	// Marshal back to YAML
	output, err := yaml.Marshal(&compose)
	if err != nil {
		return fmt.Errorf("failed to marshal compose file: %w", err)
	}

	// Write back to file
	if err := os.WriteFile(composeFilePath, output, 0644); err != nil {
		return fmt.Errorf("failed to write compose file: %w", err)
	}

	return nil
}
