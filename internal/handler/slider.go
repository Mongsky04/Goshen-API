package handler

import (
"errors"
"log/slog"
"net/http"
"strconv"

"goshen/backend/pkg/response"

"github.com/go-chi/chi/v5"
)

func (h *Handler) ListSlider(w http.ResponseWriter, r *http.Request) {
items, err := h.slider.List(r.Context())
if err != nil {
response.InternalError(w)
return
}
response.OK(w, items)
}

func (h *Handler) GetSlider(w http.ResponseWriter, r *http.Request) {
id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
if err != nil {
response.BadRequest(w, "invalid id")
return
}
s, err := h.slider.GetByID(r.Context(), id)
if err != nil {
response.NotFound(w, "slider not found")
return
}
response.OK(w, s)
}

func (h *Handler) CreateSlider(w http.ResponseWriter, r *http.Request) {
if err := r.ParseMultipartForm(maxUploadSize); err != nil {
response.BadRequest(w, "invalid multipart form")
return
}

imageURL, err := h.uploadFormImage(r, "image")
if errors.Is(err, errUnsupportedFileType) {
response.BadRequest(w, "unsupported file type: only JPEG, PNG, WebP, GIF, SVG allowed")
return
}
if err != nil {
response.InternalError(w)
return
}
if imageURL == "" {
imageURL = r.FormValue("image_url")
}
if imageURL == "" {
response.BadRequest(w, "image is required")
return
}

orderNum, _ := strconv.Atoi(r.FormValue("order_num"))
title := r.FormValue("title")

s, err := h.slider.Create(r.Context(), title, imageURL, orderNum)
if err != nil {
response.InternalError(w)
return
}
response.Created(w, s)
}

func (h *Handler) UpdateSlider(w http.ResponseWriter, r *http.Request) {
id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
if err != nil {
response.BadRequest(w, "invalid id")
return
}
if err := r.ParseMultipartForm(maxUploadSize); err != nil {
response.BadRequest(w, "invalid multipart form")
return
}

existing, err := h.slider.GetByID(r.Context(), id)
if err != nil {
response.NotFound(w, "slider item not found")
return
}

title := existing.Title
if t := r.FormValue("title"); t != "" {
title = t
}

imageURL := existing.ImageURL
if _, _, err := r.FormFile("image"); err == nil {
newURL, uploadErr := h.uploadFormImage(r, "image")
if errors.Is(uploadErr, errUnsupportedFileType) {
response.BadRequest(w, "unsupported file type: only JPEG, PNG, WebP, GIF, SVG allowed")
return
}
if uploadErr != nil {
response.InternalError(w)
return
}
go func() {
if err := h.storage.Delete(existing.ImageURL); err != nil {
slog.Error("storage: failed to delete image", "err", err)
}
}()
imageURL = newURL
}

orderNum := existing.OrderNum
if s := r.FormValue("order_num"); s != "" {
orderNum, _ = strconv.Atoi(s)
}

sl, err := h.slider.Update(r.Context(), id, title, imageURL, orderNum)
if err != nil {
response.NotFound(w, "slider item not found")
return
}
response.OK(w, sl)
}

func (h *Handler) DeleteSlider(w http.ResponseWriter, r *http.Request) {
id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
if err != nil {
response.BadRequest(w, "invalid id")
return
}
existing, err := h.slider.GetByID(r.Context(), id)
if err != nil {
response.NotFound(w, "slider item not found")
return
}
if err := h.slider.Delete(r.Context(), id); err != nil {
response.InternalError(w)
return
}
go func() {
if err := h.storage.Delete(existing.ImageURL); err != nil {
slog.Error("storage: failed to delete image", "err", err)
}
}()
response.OK(w, map[string]bool{"deleted": true})
}
