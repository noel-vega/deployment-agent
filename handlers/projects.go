package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/noel-vega/deployment-agent/projects"
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
