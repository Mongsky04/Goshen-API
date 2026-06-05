package handler

import (
	"net/http"
	"strconv"

	"goshen/backend/internal/model"
	"goshen/backend/pkg/response"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) ListBanners(w http.ResponseWriter, r *http.Request) {
	banners, err := h.banners.List(r.Context())
	if err != nil {
		response.InternalError(w)
		return
	}
	if banners == nil {
		banners = []model.Banner{}
	}
	response.OK(w, banners)
}

func (h *Handler) CreateBanner(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		response.BadRequest(w, "invalid form")
		return
	}
	name := r.FormValue("name")
	imageURL := r.FormValue("image_url")

	if url, err := h.uploadFormImage(r, "image"); err != nil {
		response.InternalError(w)
		return
	} else if url != "" {
		imageURL = url
	}

	if imageURL == "" {
		response.BadRequest(w, "image required")
		return
	}

	b, err := h.banners.Create(r.Context(), name, imageURL)
	if err != nil {
		response.InternalError(w)
		return
	}
	response.Created(w, b)
}

func (h *Handler) DeleteBanner(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.BadRequest(w, "invalid id")
		return
	}
	if err := h.banners.Delete(r.Context(), id); err != nil {
		response.InternalError(w)
		return
	}
	response.OK(w, nil)
}
