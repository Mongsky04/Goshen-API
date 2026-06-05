package handler

import (
"errors"
"log/slog"
"net/http"
"strconv"
"time"

"goshen/backend/pkg/response"

"github.com/go-chi/chi/v5"
)

func (h *Handler) ListArticles(w http.ResponseWriter, r *http.Request) {
page, limit, offset := parsePage(r)
items, err := h.articles.ListPaged(r.Context(), limit, offset)
if err != nil {
response.InternalError(w)
return
}
response.OK(w, pagedResponse{Data: items, Page: page, Limit: limit, Total: nil})
}

func (h *Handler) GetArticle(w http.ResponseWriter, r *http.Request) {
id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
if err != nil {
response.BadRequest(w, "invalid id")
return
}
a, err := h.articles.GetByID(r.Context(), id)
if err != nil {
response.NotFound(w, "article not found")
return
}
response.OK(w, a)
}

func (h *Handler) CreateArticle(w http.ResponseWriter, r *http.Request) {
if err := r.ParseMultipartForm(maxUploadSize); err != nil {
response.BadRequest(w, "invalid multipart form")
return
}
title := r.FormValue("title")
if title == "" {
response.BadRequest(w, "title is required")
return
}
description := r.FormValue("description")

var publishedAt *time.Time
if s := r.FormValue("published_at"); s != "" {
t, err := time.Parse(time.RFC3339, s)
if err != nil {
response.BadRequest(w, "invalid published_at, use RFC3339 format")
return
}
publishedAt = &t
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

a, err := h.articles.Create(r.Context(), title, description, imageURL, publishedAt)
if err != nil {
response.InternalError(w)
return
}
response.Created(w, a)
}

func (h *Handler) UpdateArticle(w http.ResponseWriter, r *http.Request) {
id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
if err != nil {
response.BadRequest(w, "invalid id")
return
}
if err := r.ParseMultipartForm(maxUploadSize); err != nil {
response.BadRequest(w, "invalid multipart form")
return
}

existing, err := h.articles.GetByID(r.Context(), id)
if err != nil {
response.NotFound(w, "article not found")
return
}

title := r.FormValue("title")
if title == "" {
title = existing.Title
}
description := r.FormValue("description")
if description == "" {
description = existing.Description
}

var publishedAt *time.Time
if s := r.FormValue("published_at"); s != "" {
t, err := time.Parse(time.RFC3339, s)
if err != nil {
response.BadRequest(w, "invalid published_at, use RFC3339 format")
return
}
publishedAt = &t
} else {
publishedAt = &existing.PublishedAt
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

a, err := h.articles.Update(r.Context(), id, title, description, imageURL, publishedAt)
if err != nil {
response.NotFound(w, "article not found")
return
}
response.OK(w, a)
}

func (h *Handler) DeleteArticle(w http.ResponseWriter, r *http.Request) {
id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
if err != nil {
response.BadRequest(w, "invalid id")
return
}
existing, err := h.articles.GetByID(r.Context(), id)
if err != nil {
response.NotFound(w, "article not found")
return
}
if err := h.articles.Delete(r.Context(), id); err != nil {
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
