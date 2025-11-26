package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/noel-vega/deployment-agent/registry"
)

type RegistryHandler struct {
	registryClient *registry.Client
}

func NewRegistryHandler(registryClient *registry.Client) *RegistryHandler {
	return &RegistryHandler{
		registryClient: registryClient,
	}
}

// ListRepositories returns all repositories in the registry
func (h *RegistryHandler) ListRepositories(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	repositories, err := h.registryClient.ListRepositories(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"registry":     h.registryClient.GetRegistryURL(),
		"repositories": repositories,
		"count":        len(repositories),
	})
}

// ListTags returns all tags for a specific repository
func (h *RegistryHandler) ListTags(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	repoName := chi.URLParam(r, "name")

	if repoName == "" {
		http.Error(w, "repository name is required", http.StatusBadRequest)
		return
	}

	tags, err := h.registryClient.ListTags(ctx, repoName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"registry":   h.registryClient.GetRegistryURL(),
		"repository": repoName,
		"tags":       tags,
		"count":      len(tags),
	})
}

// ListRepositoriesWithTags returns all repositories with their tags
func (h *RegistryHandler) ListRepositoriesWithTags(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	repositories, err := h.registryClient.ListRepositoriesWithTags(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"registry":     h.registryClient.GetRegistryURL(),
		"repositories": repositories,
		"count":        len(repositories),
	})
}
