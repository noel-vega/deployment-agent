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

type ProjectDetail struct {
	Name           string                   `json:"name"`
	Path           string                   `json:"path"`
	ComposeContent string                   `json:"compose_content"`
	Services       map[string]ServiceDetail `json:"services"`
	Containers     []ProjectContainerInfo   `json:"containers"`
}

type ServiceDetail struct {
	Image       string            `json:"image"`
	Ports       []string          `json:"ports"`
	Environment map[string]string `json:"environment"`
	Volumes     []string          `json:"volumes"`
}

type ProjectContainerInfo struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Service string `json:"service"`
	State   string `json:"state"`
	Status  string `json:"status"`
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

func (s *Service) GetProject(ctx context.Context, projectName string) (*ProjectDetail, error) {
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

	// Read compose file content
	content, err := os.ReadFile(composeFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read compose file: %w", err)
	}

	// Parse compose file to extract services
	var compose ComposeFile
	services := make(map[string]ServiceDetail)
	if err := yaml.Unmarshal(content, &compose); err == nil {
		for serviceName, serviceData := range compose.Services {
			// Initialize with empty slices and maps to avoid null in JSON
			serviceDetail := ServiceDetail{
				Ports:       []string{},
				Environment: map[string]string{},
				Volumes:     []string{},
			}

			// Type assert the service data to map
			if svcMap, ok := serviceData.(map[string]interface{}); ok {
				if image, ok := svcMap["image"].(string); ok {
					serviceDetail.Image = image
				}
				if ports, ok := svcMap["ports"].([]interface{}); ok {
					portsList := []string{}
					for _, port := range ports {
						if portStr, ok := port.(string); ok {
							portsList = append(portsList, portStr)
						}
					}
					if len(portsList) > 0 {
						serviceDetail.Ports = portsList
					}
				}
				if volumes, ok := svcMap["volumes"].([]interface{}); ok {
					volumesList := []string{}
					for _, vol := range volumes {
						if volStr, ok := vol.(string); ok {
							volumesList = append(volumesList, volStr)
						}
					}
					if len(volumesList) > 0 {
						serviceDetail.Volumes = volumesList
					}
				}
				if env, ok := svcMap["environment"].(map[string]interface{}); ok {
					envMap := make(map[string]string)
					for k, v := range env {
						if vStr, ok := v.(string); ok {
							envMap[k] = vStr
						}
					}
					if len(envMap) > 0 {
						serviceDetail.Environment = envMap
					}
				}
			}

			services[serviceName] = serviceDetail
		}
	}

	// Get containers for this project
	projectContainers := s.getProjectContainers(ctx, projectName)

	return &ProjectDetail{
		Name:           projectName,
		Path:           projectPath,
		ComposeContent: string(content),
		Services:       services,
		Containers:     projectContainers,
	}, nil
}

func (s *Service) getProjectContainers(ctx context.Context, projectName string) []ProjectContainerInfo {
	if s.dockerClient == nil {
		return []ProjectContainerInfo{}
	}

	filterArgs := filters.NewArgs()
	filterArgs.Add("label", fmt.Sprintf("com.docker.compose.project=%s", projectName))

	containers, err := s.dockerClient.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: filterArgs,
	})
	if err != nil {
		return []ProjectContainerInfo{}
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

		// Extract service name from label
		serviceName := c.Labels["com.docker.compose.service"]

		result = append(result, ProjectContainerInfo{
			ID:      c.ID[:12],
			Name:    name,
			Service: serviceName,
			State:   c.State,
			Status:  c.Status,
		})
	}

	return result
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
