package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"goshen/backend/internal/model"

	"golang.org/x/crypto/bcrypt"
)

// mockAdminRepo implements adminRepo for testing.
type mockAdminRepo struct {
	findByEmail func(ctx context.Context, email string) (*model.Admin, error)
}

func (m *mockAdminRepo) FindByEmail(ctx context.Context, email string) (*model.Admin, error) {
	if m.findByEmail != nil {
		return m.findByEmail(ctx, email)
	}
	return nil, errors.New("not found")
}

func newAuthTestHandler(admins adminRepo) *Handler {
	return &Handler{
		products:     &mockProductRepo{},
		featured:     &stubFeaturedRepo{},
		homepageGrid: &stubHomepageGridRepo{},
		articles:     &stubArticleRepo{},
		slider:       &stubSliderRepo{},
		brands:       &stubBrandRepo{},
		customers:    &stubCustomerRepo{},
		admins:       admins,
		cfg:          testConfig(),
		storage:      nil,
	}
}

func TestLogin_MissingBody(t *testing.T) {
	h := newAuthTestHandler(&mockAdminRepo{})

	req := httptest.NewRequest(http.MethodPost, "/admin/login", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestLogin_EmptyCredentials(t *testing.T) {
	h := newAuthTestHandler(&mockAdminRepo{})

	body := `{"email":"","password":""}`
	req := httptest.NewRequest(http.MethodPost, "/admin/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestLogin_WrongCredentials(t *testing.T) {
	mock := &mockAdminRepo{
		findByEmail: func(_ context.Context, email string) (*model.Admin, error) {
			return nil, errors.New("not found")
		},
	}
	h := newAuthTestHandler(mock)

	body := `{"email":"wrong@example.com","password":"badpassword"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}

	var resp struct {
		Success bool   `json:"success"`
		Error   string `json:"error"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Success {
		t.Error("expected success:false")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	// Build a real bcrypt hash for "correct-password".
	hash, err := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("bcrypt error: %v", err)
	}

	mock := &mockAdminRepo{
		findByEmail: func(_ context.Context, email string) (*model.Admin, error) {
			return &model.Admin{ID: 1, Email: email, PasswordHash: string(hash)}, nil
		},
	}
	h := newAuthTestHandler(mock)

	body := `{"email":"admin@goshen.id","password":"wrong-password"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestLogin_ValidCredentials(t *testing.T) {
	const password = "admin123"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("bcrypt error: %v", err)
	}

	mock := &mockAdminRepo{
		findByEmail: func(_ context.Context, email string) (*model.Admin, error) {
			return &model.Admin{ID: 1, Email: email, PasswordHash: string(hash)}, nil
		},
	}
	h := newAuthTestHandler(mock)

	body, _ := json.Marshal(map[string]string{
		"email":    "admin@goshen.id",
		"password": password,
	})
	req := httptest.NewRequest(http.MethodPost, "/admin/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Success bool `json:"success"`
		Data    struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !resp.Success {
		t.Error("expected success:true")
	}
	if resp.Data.Token == "" {
		t.Error("expected non-empty token in response")
	}
}
