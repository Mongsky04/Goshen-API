package handler

import (
	"errors"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"

	"goshen/backend/internal/model"
	"goshen/backend/pkg/response"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

func (h *Handler) ListMedia(w http.ResponseWriter, r *http.Request) {
	assets, err := h.mediaRepo.List(r.Context())
	if err != nil {
		response.InternalError(w)
		return
	}
	if assets == nil {
		assets = []model.MediaAsset{}
	}
	response.OK(w, assets)
}

func (h *Handler) UploadMedia(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		response.BadRequest(w, "invalid multipart form")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		response.BadRequest(w, "no file provided")
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = mime.TypeByExtension(filepath.Ext(header.Filename))
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	if !allowedImageMIME[contentType] {
		response.BadRequest(w, "unsupported file type")
		return
	}

	url, err := h.storage.Upload(file, header.Filename, contentType)
	if err != nil {
		response.InternalError(w)
		return
	}

	asset, err := h.mediaRepo.Create(r.Context(), header.Filename, url, header.Size, contentType)
	if err != nil {
		_ = h.storage.Delete(url)
		response.InternalError(w)
		return
	}

	response.Created(w, asset)
}

func (h *Handler) DeleteMedia(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.BadRequest(w, "invalid id")
		return
	}

	asset, err := h.mediaRepo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			response.NotFound(w, "asset not found")
			return
		}
		response.InternalError(w)
		return
	}

	if err := h.storage.Delete(asset.URL); err != nil {
		response.InternalError(w)
		return
	}

	if err := h.mediaRepo.Delete(r.Context(), id); err != nil {
		response.InternalError(w)
		return
	}

	response.OK(w, nil)
}
