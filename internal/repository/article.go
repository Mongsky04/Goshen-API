package repository

import (
	"context"
	"time"

	"goshen/backend/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ArticleRepo struct{ db *pgxpool.Pool }

func NewArticleRepo(db *pgxpool.Pool) *ArticleRepo { return &ArticleRepo{db: db} }

func (r *ArticleRepo) List(ctx context.Context) ([]model.Article, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, title, description, image_url, published_at, created_at, updated_at
		 FROM articles ORDER BY published_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Article
	for rows.Next() {
		var a model.Article
		if err := rows.Scan(&a.ID, &a.Title, &a.Description, &a.ImageURL,
			&a.PublishedAt, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	if out == nil {
		out = []model.Article{}
	}
	return out, nil
}

func (r *ArticleRepo) ListPaged(ctx context.Context, limit, offset int) ([]model.Article, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, title, description, image_url, published_at, created_at, updated_at
		 FROM articles ORDER BY published_at DESC LIMIT $1 OFFSET $2`,
		limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Article
	for rows.Next() {
		var a model.Article
		if err := rows.Scan(&a.ID, &a.Title, &a.Description, &a.ImageURL,
			&a.PublishedAt, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	if out == nil {
		out = []model.Article{}
	}
	return out, nil
}

func (r *ArticleRepo) GetByID(ctx context.Context, id int64) (*model.Article, error) {
	var a model.Article
	err := r.db.QueryRow(ctx,
		`SELECT id, title, description, image_url, published_at, created_at, updated_at
		 FROM articles WHERE id=$1`, id,
	).Scan(&a.ID, &a.Title, &a.Description, &a.ImageURL, &a.PublishedAt, &a.CreatedAt, &a.UpdatedAt)
	return &a, err
}

func (r *ArticleRepo) Create(ctx context.Context, title, description, imageURL string, publishedAt *time.Time) (*model.Article, error) {
	var a model.Article
	pub := time.Now()
	if publishedAt != nil {
		pub = *publishedAt
	}
	err := r.db.QueryRow(ctx,
		`INSERT INTO articles (title, description, image_url, published_at)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, title, description, image_url, published_at, created_at, updated_at`,
		title, description, imageURL, pub,
	).Scan(&a.ID, &a.Title, &a.Description, &a.ImageURL, &a.PublishedAt, &a.CreatedAt, &a.UpdatedAt)
	return &a, err
}

func (r *ArticleRepo) Update(ctx context.Context, id int64, title, description, imageURL string, publishedAt *time.Time) (*model.Article, error) {
	var a model.Article
	pub := time.Now()
	if publishedAt != nil {
		pub = *publishedAt
	}
	err := r.db.QueryRow(ctx,
		`UPDATE articles
		 SET title=$1, description=$2, image_url=$3, published_at=$4, updated_at=NOW()
		 WHERE id=$5
		 RETURNING id, title, description, image_url, published_at, created_at, updated_at`,
		title, description, imageURL, pub, id,
	).Scan(&a.ID, &a.Title, &a.Description, &a.ImageURL, &a.PublishedAt, &a.CreatedAt, &a.UpdatedAt)
	return &a, err
}

func (r *ArticleRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM articles WHERE id=$1`, id)
	return err
}
