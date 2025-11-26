package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/noel-vega/hubble/docker"
)

type ContainersHandler struct {
	dockerService *docker.Service
}

func NewContainersHandler(dockerService *docker.Service) *ContainersHandler {
	return &ContainersHandler{
		dockerService: dockerService,
	}
}

func (h *ContainersHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	containers, err := h.dockerService.ListContainers(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"containers": containers,
		"count":      len(containers),
	})
}

func (h *ContainersHandler) Stop(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	containerID := chi.URLParam(r, "id")

	if containerID == "" {
		http.Error(w, "container ID is required", http.StatusBadRequest)
		return
	}

	if err := h.dockerService.StopContainer(ctx, containerID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "container stopped successfully",
		"id":      containerID,
	})
}

func (h *ContainersHandler) Start(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	containerID := chi.URLParam(r, "id")

	if containerID == "" {
		http.Error(w, "container ID is required", http.StatusBadRequest)
		return
	}

	if err := h.dockerService.StartContainer(ctx, containerID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "container started successfully",
		"id":      containerID,
	})
}

func (h *ContainersHandler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	containerID := chi.URLParam(r, "id")

	if containerID == "" {
		http.Error(w, "container ID is required", http.StatusBadRequest)
		return
	}

	container, err := h.dockerService.GetContainer(ctx, containerID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(container)
}
