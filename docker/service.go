package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

type Service struct {
	client *client.Client
}

type ContainerInfo struct {
	ID     string            `json:"id"`
	Name   string            `json:"name"`
	Image  string            `json:"image"`
	State  string            `json:"state"`
	Status string            `json:"status"`
	Ports  []PortInfo        `json:"ports"`
	Labels map[string]string `json:"labels"`
}

type PortInfo struct {
	Private int    `json:"private"`
	Public  int    `json:"public"`
	Type    string `json:"type"`
}

func NewService() (*Service, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}
	return &Service{client: cli}, nil
}

func (s *Service) ListContainers(ctx context.Context) ([]ContainerInfo, error) {
	containers, err := s.client.ContainerList(ctx, container.ListOptions{
		All: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	result := make([]ContainerInfo, 0, len(containers))
	for _, c := range containers {
		ports := make([]PortInfo, 0, len(c.Ports))
		for _, p := range c.Ports {
			ports = append(ports, PortInfo{
				Private: int(p.PrivatePort),
				Public:  int(p.PublicPort),
				Type:    p.Type,
			})
		}

		name := c.Names[0]
		if len(name) > 0 && name[0] == '/' {
			name = name[1:]
		}

		result = append(result, ContainerInfo{
			ID:     c.ID[:12],
			Name:   name,
			Image:  c.Image,
			State:  c.State,
			Status: c.Status,
			Ports:  ports,
			Labels: c.Labels,
		})
	}

	return result, nil
}

func (s *Service) Close() error {
	if s.client != nil {
		return s.client.Close()
	}
	return nil
}

func (s *Service) Client() *client.Client {
	return s.client
}

func (s *Service) StopContainer(ctx context.Context, containerID string) error {
	if err := s.client.ContainerStop(ctx, containerID, container.StopOptions{}); err != nil {
		return fmt.Errorf("failed to stop container: %w", err)
	}
	return nil
}

func (s *Service) StartContainer(ctx context.Context, containerID string) error {
	if err := s.client.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}
	return nil
}

type DetailedContainerInfo struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Image         string                 `json:"image"`
	ImageID       string                 `json:"image_id"`
	State         ContainerState         `json:"state"`
	Created       string                 `json:"created"`
	Ports         []PortInfo             `json:"ports"`
	Labels        map[string]string      `json:"labels"`
	Mounts        []MountInfo            `json:"mounts"`
	Networks      map[string]NetworkInfo `json:"networks"`
	Env           []string               `json:"env"`
	RestartPolicy string                 `json:"restart_policy"`
	RestartCount  int                    `json:"restart_count"`
	Platform      string                 `json:"platform"`
	HostConfig    HostConfigInfo         `json:"host_config"`
}

type ContainerState struct {
	Status     string `json:"status"`
	Running    bool   `json:"running"`
	Paused     bool   `json:"paused"`
	Restarting bool   `json:"restarting"`
	OOMKilled  bool   `json:"oom_killed"`
	Dead       bool   `json:"dead"`
	Pid        int    `json:"pid"`
	ExitCode   int    `json:"exit_code"`
	Error      string `json:"error"`
	StartedAt  string `json:"started_at"`
	FinishedAt string `json:"finished_at"`
}

type MountInfo struct {
	Type        string `json:"type"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Mode        string `json:"mode"`
	RW          bool   `json:"rw"`
}

type NetworkInfo struct {
	NetworkID           string `json:"network_id"`
	EndpointID          string `json:"endpoint_id"`
	Gateway             string `json:"gateway"`
	IPAddress           string `json:"ip_address"`
	IPPrefixLen         int    `json:"ip_prefix_len"`
	IPv6Gateway         string `json:"ipv6_gateway"`
	GlobalIPv6Address   string `json:"global_ipv6_address"`
	GlobalIPv6PrefixLen int    `json:"global_ipv6_prefix_len"`
	MacAddress          string `json:"mac_address"`
}

type HostConfigInfo struct {
	CPUShares         int64                    `json:"cpu_shares"`
	Memory            int64                    `json:"memory"`
	MemoryReservation int64                    `json:"memory_reservation"`
	MemorySwap        int64                    `json:"memory_swap"`
	NanoCPUs          int64                    `json:"nano_cpus"`
	AutoRemove        bool                     `json:"auto_remove"`
	NetworkMode       string                   `json:"network_mode"`
	PortBindings      map[string][]PortBinding `json:"port_bindings"`
}

type PortBinding struct {
	HostIP   string `json:"host_ip"`
	HostPort string `json:"host_port"`
}

type ImageInfo struct {
	ID          string   `json:"id"`
	RepoTags    []string `json:"repo_tags"`
	RepoDigests []string `json:"repo_digests"`
	Size        int64    `json:"size"`
	Created     int64    `json:"created"`
}

func (s *Service) GetContainer(ctx context.Context, containerID string) (*DetailedContainerInfo, error) {
	inspect, err := s.client.ContainerInspect(ctx, containerID)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect container: %w", err)
	}

	// Extract container name (remove leading slash)
	name := inspect.Name
	if len(name) > 0 && name[0] == '/' {
		name = name[1:]
	}

	// Build ports info
	ports := make([]PortInfo, 0)
	if inspect.NetworkSettings != nil && inspect.NetworkSettings.Ports != nil {
		for port, bindings := range inspect.NetworkSettings.Ports {
			if len(bindings) > 0 {
				for _, binding := range bindings {
					privatePort := 0
					publicPort := 0

					// Parse private port from the port/protocol string
					fmt.Sscanf(string(port), "%d", &privatePort)
					// Parse public port
					fmt.Sscanf(binding.HostPort, "%d", &publicPort)

					ports = append(ports, PortInfo{
						Private: privatePort,
						Public:  publicPort,
						Type:    port.Proto(),
					})
				}
			}
		}
	}

	// Build mounts info
	mounts := make([]MountInfo, 0, len(inspect.Mounts))
	for _, m := range inspect.Mounts {
		mounts = append(mounts, MountInfo{
			Type:        string(m.Type),
			Source:      m.Source,
			Destination: m.Destination,
			Mode:        m.Mode,
			RW:          m.RW,
		})
	}

	// Build networks info
	networks := make(map[string]NetworkInfo)
	if inspect.NetworkSettings != nil && inspect.NetworkSettings.Networks != nil {
		for netName, netSettings := range inspect.NetworkSettings.Networks {
			networks[netName] = NetworkInfo{
				NetworkID:           netSettings.NetworkID,
				EndpointID:          netSettings.EndpointID,
				Gateway:             netSettings.Gateway,
				IPAddress:           netSettings.IPAddress,
				IPPrefixLen:         netSettings.IPPrefixLen,
				IPv6Gateway:         netSettings.IPv6Gateway,
				GlobalIPv6Address:   netSettings.GlobalIPv6Address,
				GlobalIPv6PrefixLen: netSettings.GlobalIPv6PrefixLen,
				MacAddress:          netSettings.MacAddress,
			}
		}
	}

	// Build port bindings for host config
	portBindings := make(map[string][]PortBinding)
	if inspect.HostConfig != nil && inspect.HostConfig.PortBindings != nil {
		for port, bindings := range inspect.HostConfig.PortBindings {
			bindingsList := make([]PortBinding, 0, len(bindings))
			for _, b := range bindings {
				bindingsList = append(bindingsList, PortBinding{
					HostIP:   b.HostIP,
					HostPort: b.HostPort,
				})
			}
			portBindings[string(port)] = bindingsList
		}
	}

	// Get restart policy
	restartPolicy := "no"
	if inspect.HostConfig != nil && inspect.HostConfig.RestartPolicy.Name != "" {
		restartPolicy = string(inspect.HostConfig.RestartPolicy.Name)
	}

	// Build detailed container info
	detailedInfo := &DetailedContainerInfo{
		ID:            inspect.ID[:12],
		Name:          name,
		Image:         inspect.Config.Image,
		ImageID:       inspect.Image,
		Created:       inspect.Created,
		Ports:         ports,
		Labels:        inspect.Config.Labels,
		Mounts:        mounts,
		Networks:      networks,
		Env:           inspect.Config.Env,
		RestartPolicy: restartPolicy,
		Platform:      inspect.Platform,
		State: ContainerState{
			Status:     inspect.State.Status,
			Running:    inspect.State.Running,
			Paused:     inspect.State.Paused,
			Restarting: inspect.State.Restarting,
			OOMKilled:  inspect.State.OOMKilled,
			Dead:       inspect.State.Dead,
			Pid:        inspect.State.Pid,
			ExitCode:   inspect.State.ExitCode,
			Error:      inspect.State.Error,
			StartedAt:  inspect.State.StartedAt,
			FinishedAt: inspect.State.FinishedAt,
		},
	}

	// Add host config if available
	if inspect.HostConfig != nil {
		detailedInfo.HostConfig = HostConfigInfo{
			CPUShares:         inspect.HostConfig.CPUShares,
			Memory:            inspect.HostConfig.Memory,
			MemoryReservation: inspect.HostConfig.MemoryReservation,
			MemorySwap:        inspect.HostConfig.MemorySwap,
			NanoCPUs:          inspect.HostConfig.NanoCPUs,
			AutoRemove:        inspect.HostConfig.AutoRemove,
			NetworkMode:       string(inspect.HostConfig.NetworkMode),
			PortBindings:      portBindings,
		}
		detailedInfo.RestartCount = inspect.RestartCount
	}

	return detailedInfo, nil
}

func (s *Service) ListImages(ctx context.Context) ([]ImageInfo, error) {
	images, err := s.client.ImageList(ctx, image.ListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("failed to list images: %w", err)
	}

	result := make([]ImageInfo, 0, len(images))
	for _, img := range images {
		id := img.ID
		if len(id) > 19 && id[:7] == "sha256:" {
			id = id[7:19]
		}

		result = append(result, ImageInfo{
			ID:          id,
			RepoTags:    img.RepoTags,
			RepoDigests: img.RepoDigests,
			Size:        img.Size,
			Created:     img.Created,
		})
	}

	return result, nil
}
