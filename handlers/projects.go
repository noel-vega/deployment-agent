package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/noel-vega/hubble/projects"
)

type ProjectsHandler struct {
	projectsService *projects.Service
}

func NewProjectsHandler(projectsService *projects.Service) *ProjectsHandler {
	return &ProjectsHandler{
		projectsService: projectsService,
	}
}

func (h *ProjectsHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	projectsList, err := h.projectsService.ListProjects(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"projects": projectsList,
		"count":    len(projectsList),
	})
}

func (h *ProjectsHandler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	projectName := chi.URLParam(r, "name")

	if projectName == "" {
		http.Error(w, "project name is required", http.StatusBadRequest)
		return
	}

	project, err := h.projectsService.GetProject(ctx, projectName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(project)
}

func (h *ProjectsHandler) GetCompose(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	projectName := chi.URLParam(r, "name")

	if projectName == "" {
		http.Error(w, "project name is required", http.StatusBadRequest)
		return
	}

	composeContent, err := h.projectsService.GetProjectCompose(ctx, projectName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"content": composeContent,
	})
}

func (h *ProjectsHandler) GetContainers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	projectName := chi.URLParam(r, "name")

	if projectName == "" {
		http.Error(w, "project name is required", http.StatusBadRequest)
		return
	}

	containers, err := h.projectsService.GetProjectContainers(ctx, projectName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"containers": containers,
		"count":      len(containers),
	})
}

func (h *ProjectsHandler) GetVolumes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	projectName := chi.URLParam(r, "name")

	if projectName == "" {
		http.Error(w, "project name is required", http.StatusBadRequest)
		return
	}

	volumes, err := h.projectsService.GetProjectVolumes(ctx, projectName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"volumes": volumes,
		"count":   len(volumes),
	})
}

func (h *ProjectsHandler) GetEnvironment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	projectName := chi.URLParam(r, "name")

	if projectName == "" {
		http.Error(w, "project name is required", http.StatusBadRequest)
		return
	}

	environment, err := h.projectsService.GetProjectEnvironment(ctx, projectName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"environment": environment,
		"count":       len(environment),
	})
}

func (h *ProjectsHandler) GetNetworks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	projectName := chi.URLParam(r, "name")

	if projectName == "" {
		http.Error(w, "project name is required", http.StatusBadRequest)
		return
	}

	networks, err := h.projectsService.GetProjectNetworks(ctx, projectName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"networks": networks,
		"count":    len(networks),
	})
}

func (h *ProjectsHandler) GetServices(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	projectName := chi.URLParam(r, "name")

	if projectName == "" {
		http.Error(w, "project name is required", http.StatusBadRequest)
		return
	}

	services, err := h.projectsService.GetProjectServices(ctx, projectName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"services": services,
		"count":    len(services),
	})
}

func (h *ProjectsHandler) StartService(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	projectName := chi.URLParam(r, "name")
	serviceName := chi.URLParam(r, "service")

	if projectName == "" {
		http.Error(w, "project name is required", http.StatusBadRequest)
		return
	}

	if serviceName == "" {
		http.Error(w, "service name is required", http.StatusBadRequest)
		return
	}

	err := h.projectsService.StartService(ctx, projectName, serviceName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"message": "service started successfully",
		"project": projectName,
		"service": serviceName,
	})
}

func (h *ProjectsHandler) StopService(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	projectName := chi.URLParam(r, "name")
	serviceName := chi.URLParam(r, "service")

	if projectName == "" {
		http.Error(w, "project name is required", http.StatusBadRequest)
		return
	}

	if serviceName == "" {
		http.Error(w, "service name is required", http.StatusBadRequest)
		return
	}

	err := h.projectsService.StopService(ctx, projectName, serviceName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"message": "service stopped successfully",
		"project": projectName,
		"service": serviceName,
	})
}

func (h *ProjectsHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "project name is required", http.StatusBadRequest)
		return
	}

	err := h.projectsService.CreateProject(ctx, req.Name)
	if err != nil {
		if err.Error() == "project already exists: "+req.Name {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Get the created project info
	project, err := h.projectsService.GetProject(ctx, req.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{
		"message": "project created successfully",
		"project": project,
	})
}

func (h *ProjectsHandler) AddService(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	projectName := chi.URLParam(r, "name")

	if projectName == "" {
		http.Error(w, "project name is required", http.StatusBadRequest)
		return
	}

	var service projects.ComposeService
	if err := json.NewDecoder(r.Body).Decode(&service); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if service.Name == "" {
		http.Error(w, "service name is required", http.StatusBadRequest)
		return
	}

	err := h.projectsService.AddService(ctx, projectName, service)
	if err != nil {
		if err.Error() == "service already exists: "+service.Name {
			http.Error(w, err.Error(), http.StatusConflict)
		} else if err.Error() == "project not found: "+projectName {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{
		"message": "service added successfully",
		"project": projectName,
		"service": service.Name,
	})
}

func (h *ProjectsHandler) UpdateService(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	projectName := chi.URLParam(r, "name")
	serviceName := chi.URLParam(r, "service")

	if projectName == "" {
		http.Error(w, "project name is required", http.StatusBadRequest)
		return
	}

	if serviceName == "" {
		http.Error(w, "service name is required", http.StatusBadRequest)
		return
	}

	var service projects.ComposeService
	if err := json.NewDecoder(r.Body).Decode(&service); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Use service name from URL
	service.Name = serviceName

	err := h.projectsService.UpdateService(ctx, projectName, service)
	if err != nil {
		if err.Error() == "service not found: "+serviceName {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else if err.Error() == "project not found: "+projectName {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"message": "service updated successfully",
		"project": projectName,
		"service": serviceName,
	})
}

func (h *ProjectsHandler) DeleteService(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	projectName := chi.URLParam(r, "name")
	serviceName := chi.URLParam(r, "service")

	if projectName == "" {
		http.Error(w, "project name is required", http.StatusBadRequest)
		return
	}

	if serviceName == "" {
		http.Error(w, "service name is required", http.StatusBadRequest)
		return
	}

	err := h.projectsService.DeleteService(ctx, projectName, serviceName)
	if err != nil {
		if err.Error() == "service not found: "+serviceName {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else if err.Error() == "project not found: "+projectName {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"message": "service deleted successfully",
		"project": projectName,
		"service": serviceName,
	})
}

func (h *ProjectsHandler) AddNetwork(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	projectName := chi.URLParam(r, "name")

	if projectName == "" {
		http.Error(w, "project name is required", http.StatusBadRequest)
		return
	}

	var network projects.NetworkConfig
	if err := json.NewDecoder(r.Body).Decode(&network); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if network.Name == "" {
		http.Error(w, "network name is required", http.StatusBadRequest)
		return
	}

	err := h.projectsService.AddNetwork(ctx, projectName, network)
	if err != nil {
		if err.Error() == "network already exists: "+network.Name {
			http.Error(w, err.Error(), http.StatusConflict)
		} else if err.Error() == "project not found: "+projectName {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else if err.Error() == "external networks cannot specify a driver (driver is managed by the existing network)" {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{
		"message": "network added successfully",
		"project": projectName,
		"network": network.Name,
	})
}

func (h *ProjectsHandler) UpdateNetwork(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	projectName := chi.URLParam(r, "name")
	networkName := chi.URLParam(r, "network")

	if projectName == "" {
		http.Error(w, "project name is required", http.StatusBadRequest)
		return
	}

	if networkName == "" {
		http.Error(w, "network name is required", http.StatusBadRequest)
		return
	}

	var network projects.NetworkConfig
	if err := json.NewDecoder(r.Body).Decode(&network); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Use network name from URL
	network.Name = networkName

	err := h.projectsService.UpdateNetwork(ctx, projectName, network)
	if err != nil {
		if err.Error() == "network not found: "+networkName {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else if err.Error() == "project not found: "+projectName {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else if err.Error() == "external networks cannot specify a driver (driver is managed by the existing network)" {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"message": "network updated successfully",
		"project": projectName,
		"network": networkName,
	})
}

func (h *ProjectsHandler) DeleteNetwork(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	projectName := chi.URLParam(r, "name")
	networkName := chi.URLParam(r, "network")

	if projectName == "" {
		http.Error(w, "project name is required", http.StatusBadRequest)
		return
	}

	if networkName == "" {
		http.Error(w, "network name is required", http.StatusBadRequest)
		return
	}

	err := h.projectsService.DeleteNetwork(ctx, projectName, networkName)
	if err != nil {
		if err.Error() == "network not found: "+networkName {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else if err.Error() == "project not found: "+projectName {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"message": "network deleted successfully",
		"project": projectName,
		"network": networkName,
	})
}
