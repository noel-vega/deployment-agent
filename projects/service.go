package projects

import (
	"context"
	"fmt"
	"os"
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
}

type ComposeFile struct {
	Services map[string]interface{} `yaml:"services"`
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
