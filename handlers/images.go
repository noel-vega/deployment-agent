package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/noel-vega/deployment-agent/docker"
)

type ImagesHandler struct {
	dockerService *docker.Service
}

func NewImagesHandler(dockerService *docker.Service) *ImagesHandler {
	return &ImagesHandler{
		dockerService: dockerService,
	}
}

func (h *ImagesHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	images, err := h.dockerService.ListImages(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"images": images,
		"count":  len(images),
	})
}
