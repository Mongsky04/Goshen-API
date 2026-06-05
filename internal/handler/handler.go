package handler

import (
	"context"
	"time"

	"goshen/backend/internal/config"
	"goshen/backend/internal/model"
	"goshen/backend/pkg/storage"

	"github.com/jackc/pgx/v5/pgxpool"

	"goshen/backend/internal/repository"
)

// productRepo is the interface the handler needs from ProductRepo.
type productRepo interface {
	ListPaged(ctx context.Context, limit, offset int) ([]model.Product, error)
	GetByID(ctx context.Context, id int64) (*model.Product, error)
	Create(ctx context.Context, name, imageURL, category, subCategory string) (*model.Product, error)
	Update(ctx context.Context, id int64, name, imageURL, category, subCategory string) (*model.Product, error)
	Delete(ctx context.Context, id int64) error
}

// featuredRepo is the interface the handler needs from FeaturedRepo.
type featuredRepo interface {
	ListPaged(ctx context.Context, limit, offset int) ([]model.Featured, error)
	GetByID(ctx context.Context, id int64) (*model.Featured, error)
	Create(ctx context.Context, productID int64, featuredCategories []string) (*model.Featured, error)
	Update(ctx context.Context, id int64, productID int64, featuredCategories []string) (*model.Featured, error)
	Delete(ctx context.Context, id int64) error
}

// homepageGridRepo is the interface for homepage grid products.
type homepageGridRepo interface {
	List(ctx context.Context) ([]model.HomepageGridProduct, error)
	ReplaceAll(ctx context.Context, productIDs []int64) error
}

// bannerRepo is the interface the handler needs from BannerRepo.
type bannerRepo interface {
	List(ctx context.Context) ([]model.Banner, error)
	Create(ctx context.Context, name, imageURL string) (*model.Banner, error)
	Delete(ctx context.Context, id int64) error
}

// articleRepo is the interface the handler needs from ArticleRepo.
type articleRepo interface {
	ListPaged(ctx context.Context, limit, offset int) ([]model.Article, error)
	GetByID(ctx context.Context, id int64) (*model.Article, error)
	Create(ctx context.Context, title, description, imageURL string, publishedAt *time.Time) (*model.Article, error)
	Update(ctx context.Context, id int64, title, description, imageURL string, publishedAt *time.Time) (*model.Article, error)
	Delete(ctx context.Context, id int64) error
}

// sliderRepo is the interface the handler needs from SliderRepo.
type sliderRepo interface {
	List(ctx context.Context) ([]model.Slider, error)
	GetByID(ctx context.Context, id int64) (*model.Slider, error)
	Create(ctx context.Context, title, imageURL string, orderNum int) (*model.Slider, error)
	Update(ctx context.Context, id int64, title, imageURL string, orderNum int) (*model.Slider, error)
	Delete(ctx context.Context, id int64) error
}

// brandRepo is the interface the handler needs from BrandRepo.
type brandRepo interface {
	List(ctx context.Context) ([]model.Brand, error)
	GetByID(ctx context.Context, id int64) (*model.Brand, error)
	Create(ctx context.Context, name, imageURL string) (*model.Brand, error)
	Update(ctx context.Context, id int64, name, imageURL string) (*model.Brand, error)
	Delete(ctx context.Context, id int64) error
}

// customerRepo is the interface the handler needs from CustomerRepo.
type customerRepo interface {
	List(ctx context.Context) ([]model.Customer, error)
	GetByID(ctx context.Context, id int64) (*model.Customer, error)
	Create(ctx context.Context, imageURL, altText string) (*model.Customer, error)
	Update(ctx context.Context, id int64, imageURL, altText string) (*model.Customer, error)
	Delete(ctx context.Context, id int64) error
}

// supportCardRepo is the interface the handler needs from SupportCardRepo.
type supportCardRepo interface {
	List(ctx context.Context) ([]model.SupportCard, error)
	Create(ctx context.Context, title, description, ctaLabel, ctaHref string, sortOrder int) (*model.SupportCard, error)
	Delete(ctx context.Context, id string) error
}

// adminRepo is the interface the handler needs from AdminRepo.
type adminRepo interface {
	FindByEmail(ctx context.Context, email string) (*model.Admin, error)
}

// conferenceRepo is the interface the handler needs from ConferenceRepo.
type conferenceRepo interface {
	Get(ctx context.Context, slug string) (*model.ConferenceAdminData, error)
	Save(ctx context.Context, data *model.ConferenceAdminData) error
	ListUnpublishedSlugs(ctx context.Context) ([]string, error)
}

// performerRepo is the interface the handler needs from PerformerRepo.
type performerRepo interface {
	Get(ctx context.Context, slug string) (*model.PerformerAdminData, error)
	Save(ctx context.Context, data *model.PerformerAdminData) error
	ListUnpublishedSlugs(ctx context.Context) ([]string, error)
}

// mediaAssetRepo is the interface the handler needs from MediaAssetRepo.
type mediaAssetRepo interface {
	List(ctx context.Context) ([]model.MediaAsset, error)
	Create(ctx context.Context, filename, url string, size int64, mimeType string) (*model.MediaAsset, error)
	GetByID(ctx context.Context, id int64) (*model.MediaAsset, error)
	Delete(ctx context.Context, id int64) error
}

type Handler struct {
	products       productRepo
	featured       featuredRepo
	homepageGrid   homepageGridRepo
	banners        bannerRepo
	articles       articleRepo
	slider         sliderRepo
	brands         brandRepo
	customers      customerRepo
	supportCards   supportCardRepo
	admins         adminRepo
	conference     conferenceRepo
	performer      performerRepo
	mediaRepo      mediaAssetRepo
	cfg            *config.Config
	storage        *storage.Client
}

func New(db *pgxpool.Pool, cfg *config.Config) *Handler {
	return &Handler{
		products:     repository.NewProductRepo(db),
		featured:     repository.NewFeaturedRepo(db),
		homepageGrid: repository.NewHomepageGridRepo(db),
		banners:      repository.NewBannerRepo(db),
		articles:     repository.NewArticleRepo(db),
		slider:       repository.NewSliderRepo(db),
		brands:       repository.NewBrandRepo(db),
		customers:    repository.NewCustomerRepo(db),
		supportCards: repository.NewSupportCardRepo(db),
		admins:       repository.NewAdminRepo(db),
		conference:   repository.NewConferenceRepo(db),
		performer:    repository.NewPerformerRepo(db),
		mediaRepo:    repository.NewMediaAssetRepo(db),
		cfg:          cfg,
		storage:      storage.New(cfg.UploadDir, cfg.BackendURL),
	}
}
