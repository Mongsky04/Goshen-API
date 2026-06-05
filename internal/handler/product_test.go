package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"goshen/backend/internal/config"
	"goshen/backend/internal/model"

	"github.com/go-chi/chi/v5"
)

// mockProductRepo implements productRepo for testing.
type mockProductRepo struct {
	listPaged func(ctx context.Context, limit, offset int) ([]model.Product, error)
	getByID   func(ctx context.Context, id int64) (*model.Product, error)
}

func (m *mockProductRepo) ListPaged(ctx context.Context, limit, offset int) ([]model.Product, error) {
	if m.listPaged != nil {
		return m.listPaged(ctx, limit, offset)
	}
	return []model.Product{}, nil
}

func (m *mockProductRepo) GetByID(ctx context.Context, id int64) (*model.Product, error) {
	if m.getByID != nil {
		return m.getByID(ctx, id)
	}
	return nil, errors.New("not found")
}

func (m *mockProductRepo) Create(ctx context.Context, name, imageURL, category, subCategory string) (*model.Product, error) {
	return nil, errors.New("not implemented")
}

func (m *mockProductRepo) Update(ctx context.Context, id int64, name, imageURL, category, subCategory string) (*model.Product, error) {
	return nil, errors.New("not implemented")
}

func (m *mockProductRepo) Delete(ctx context.Context, id int64) error {
	return errors.New("not implemented")
}

// newTestHandler returns a Handler wired with the given product mock.
// All other repos are set to a no-op stub to satisfy the struct.
func newTestHandler(products productRepo) *Handler {
	return &Handler{
		products:     products,
		featured:     &stubFeaturedRepo{},
		homepageGrid: &stubHomepageGridRepo{},
		articles:     &stubArticleRepo{},
		slider:       &stubSliderRepo{},
		brands:       &stubBrandRepo{},
		customers:    &stubCustomerRepo{},
		admins:       &stubAdminRepo{},
		cfg:          testConfig(),
		storage:      nil,
	}
}

func testConfig() *config.Config {
	return &config.Config{
		JWTSecret: "test-secret",
	}
}

func TestListProducts_OK(t *testing.T) {
	now := time.Now()
	mock := &mockProductRepo{
		listPaged: func(_ context.Context, limit, offset int) ([]model.Product, error) {
			return []model.Product{
				{ID: 1, Name: "Speaker A", ImageURL: "http://example.com/a.jpg", CreatedAt: now, UpdatedAt: now},
				{ID: 2, Name: "Speaker B", ImageURL: "http://example.com/b.jpg", CreatedAt: now, UpdatedAt: now},
			}, nil
		},
	}
	h := newTestHandler(mock)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/products", nil)
	rec := httptest.NewRecorder()

	h.ListProducts(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body struct {
		Success bool `json:"success"`
		Data    struct {
			Data  []model.Product `json:"data"`
			Page  int             `json:"page"`
			Limit int             `json:"limit"`
		} `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !body.Success {
		t.Error("expected success:true")
	}
	if len(body.Data.Data) != 2 {
		t.Errorf("expected 2 products, got %d", len(body.Data.Data))
	}
}

func TestListProducts_Pagination(t *testing.T) {
	var capturedLimit, capturedOffset int
	mock := &mockProductRepo{
		listPaged: func(_ context.Context, limit, offset int) ([]model.Product, error) {
			capturedLimit = limit
			capturedOffset = offset
			return []model.Product{}, nil
		},
	}
	h := newTestHandler(mock)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/products?page=2&limit=5", nil)
	rec := httptest.NewRecorder()

	h.ListProducts(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body struct {
		Success bool `json:"success"`
		Data    struct {
			Page  int `json:"page"`
			Limit int `json:"limit"`
		} `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body.Data.Page != 2 {
		t.Errorf("expected page=2, got %d", body.Data.Page)
	}
	if body.Data.Limit != 5 {
		t.Errorf("expected limit=5, got %d", body.Data.Limit)
	}
	if capturedLimit != 5 {
		t.Errorf("expected repo called with limit=5, got %d", capturedLimit)
	}
	if capturedOffset != 5 {
		t.Errorf("expected repo called with offset=5 (page=2, limit=5), got %d", capturedOffset)
	}
}

func TestGetProduct_NotFound(t *testing.T) {
	mock := &mockProductRepo{
		getByID: func(_ context.Context, id int64) (*model.Product, error) {
			return nil, errors.New("no rows")
		},
	}
	h := newTestHandler(mock)

	// Build a chi router so URLParam works.
	r := chi.NewRouter()
	r.Get("/api/v1/products/{id}", h.GetProduct)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/products/999", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}

	var body struct {
		Success bool   `json:"success"`
		Error   string `json:"error"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body.Success {
		t.Error("expected success:false")
	}
}

func TestGetProduct_InvalidID(t *testing.T) {
	h := newTestHandler(&mockProductRepo{})

	r := chi.NewRouter()
	r.Get("/api/v1/products/{id}", h.GetProduct)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/products/notanumber", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestGetProduct_OK(t *testing.T) {
	now := time.Now()
	mock := &mockProductRepo{
		getByID: func(_ context.Context, id int64) (*model.Product, error) {
			return &model.Product{ID: id, Name: "Mixer X", ImageURL: "http://example.com/x.jpg", CreatedAt: now, UpdatedAt: now}, nil
		},
	}
	h := newTestHandler(mock)

	r := chi.NewRouter()
	r.Get("/api/v1/products/{id}", h.GetProduct)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/products/42", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body struct {
		Success bool          `json:"success"`
		Data    model.Product `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !body.Success {
		t.Error("expected success:true")
	}
	if body.Data.ID != 42 {
		t.Errorf("expected product ID 42, got %d", body.Data.ID)
	}
}

// ---- no-op stubs for repos not under test ----

type stubFeaturedRepo struct{}

func (s *stubFeaturedRepo) ListPaged(_ context.Context, _, _ int) ([]model.Featured, error) {
	return []model.Featured{}, nil
}
func (s *stubFeaturedRepo) GetByID(_ context.Context, _ int64) (*model.Featured, error) {
	return nil, errors.New("stub")
}
func (s *stubFeaturedRepo) Create(_ context.Context, _ int64, _ []string) (*model.Featured, error) {
	return nil, errors.New("stub")
}
func (s *stubFeaturedRepo) Update(_ context.Context, _ int64, _ int64, _ []string) (*model.Featured, error) {
	return nil, errors.New("stub")
}
func (s *stubFeaturedRepo) Delete(_ context.Context, _ int64) error { return errors.New("stub") }

type stubHomepageGridRepo struct{}

func (s *stubHomepageGridRepo) List(_ context.Context) ([]model.HomepageGridProduct, error) {
	return []model.HomepageGridProduct{}, nil
}
func (s *stubHomepageGridRepo) ReplaceAll(_ context.Context, _ []int64) error { return nil }

type stubArticleRepo struct{}

func (s *stubArticleRepo) ListPaged(_ context.Context, _, _ int) ([]model.Article, error) {
	return []model.Article{}, nil
}
func (s *stubArticleRepo) GetByID(_ context.Context, _ int64) (*model.Article, error) {
	return nil, errors.New("stub")
}
func (s *stubArticleRepo) Create(_ context.Context, _, _, _ string, _ *time.Time) (*model.Article, error) {
	return nil, errors.New("stub")
}
func (s *stubArticleRepo) Update(_ context.Context, _ int64, _, _, _ string, _ *time.Time) (*model.Article, error) {
	return nil, errors.New("stub")
}
func (s *stubArticleRepo) Delete(_ context.Context, _ int64) error { return errors.New("stub") }

type stubSliderRepo struct{}

func (s *stubSliderRepo) List(_ context.Context) ([]model.Slider, error) {
	return []model.Slider{}, nil
}
func (s *stubSliderRepo) GetByID(_ context.Context, _ int64) (*model.Slider, error) {
	return nil, errors.New("stub")
}
func (s *stubSliderRepo) Create(_ context.Context, _, _ string, _ int) (*model.Slider, error) {
	return nil, errors.New("stub")
}
func (s *stubSliderRepo) Update(_ context.Context, _ int64, _, _ string, _ int) (*model.Slider, error) {
	return nil, errors.New("stub")
}
func (s *stubSliderRepo) Delete(_ context.Context, _ int64) error { return errors.New("stub") }

type stubBrandRepo struct{}

func (s *stubBrandRepo) List(_ context.Context) ([]model.Brand, error) {
	return []model.Brand{}, nil
}
func (s *stubBrandRepo) GetByID(_ context.Context, _ int64) (*model.Brand, error) {
	return nil, errors.New("stub")
}
func (s *stubBrandRepo) Create(_ context.Context, _, _ string) (*model.Brand, error) {
	return nil, errors.New("stub")
}
func (s *stubBrandRepo) Update(_ context.Context, _ int64, _, _ string) (*model.Brand, error) {
	return nil, errors.New("stub")
}
func (s *stubBrandRepo) Delete(_ context.Context, _ int64) error { return errors.New("stub") }

type stubCustomerRepo struct{}

func (s *stubCustomerRepo) List(_ context.Context) ([]model.Customer, error) {
	return []model.Customer{}, nil
}
func (s *stubCustomerRepo) GetByID(_ context.Context, _ int64) (*model.Customer, error) {
	return nil, errors.New("stub")
}
func (s *stubCustomerRepo) Create(_ context.Context, _, _ string) (*model.Customer, error) {
	return nil, errors.New("stub")
}
func (s *stubCustomerRepo) Update(_ context.Context, _ int64, _, _ string) (*model.Customer, error) {
	return nil, errors.New("stub")
}
func (s *stubCustomerRepo) Delete(_ context.Context, _ int64) error { return errors.New("stub") }

type stubAdminRepo struct{}

func (s *stubAdminRepo) FindByEmail(_ context.Context, _ string) (*model.Admin, error) {
	return nil, errors.New("stub")
}
