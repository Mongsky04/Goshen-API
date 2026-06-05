package repository

import (
	"context"

	"goshen/backend/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type BannerRepo struct {
	db *pgxpool.Pool
}

func NewBannerRepo(db *pgxpool.Pool) *BannerRepo {
	return &BannerRepo{db: db}
}

func (r *BannerRepo) List(ctx context.Context) ([]model.Banner, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, name, image_url, sort_order, created_at, updated_at
		 FROM banners ORDER BY sort_order ASC, created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Banner
	for rows.Next() {
		var b model.Banner
		if err := rows.Scan(&b.ID, &b.Name, &b.ImageURL, &b.SortOrder, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (r *BannerRepo) Create(ctx context.Context, name, imageURL string) (*model.Banner, error) {
	var b model.Banner
	err := r.db.QueryRow(ctx,
		`INSERT INTO banners (name, image_url)
		 VALUES ($1, $2)
		 RETURNING id, name, image_url, sort_order, created_at, updated_at`,
		name, imageURL,
	).Scan(&b.ID, &b.Name, &b.ImageURL, &b.SortOrder, &b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *BannerRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM banners WHERE id = $1`, id)
	return err
}
