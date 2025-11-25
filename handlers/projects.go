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
