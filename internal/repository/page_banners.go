package repository

import (
	"context"

	"goshen/backend/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PageBannersRepo struct{ db *pgxpool.Pool }

func NewPageBannersRepo(db *pgxpool.Pool) *PageBannersRepo { return &PageBannersRepo{db: db} }

func (r *PageBannersRepo) ListBySlug(ctx context.Context, slug string) ([]model.Slider, error) {
	rows, err := r.db.Query(ctx,
		`SELECT s.id, s.title, s.image_url, s.order_num, s.created_at, s.updated_at
		 FROM page_banners pb
		 JOIN slider s ON s.id = pb.banner_id
		 WHERE pb.page_slug = $1
		 ORDER BY pb.order_num ASC, pb.id ASC`,
		slug,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Slider
	for rows.Next() {
		var s model.Slider
		if err := rows.Scan(&s.ID, &s.Title, &s.ImageURL, &s.OrderNum, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	if out == nil {
		out = []model.Slider{}
	}
	return out, nil
}

func (r *PageBannersRepo) Replace(ctx context.Context, slug string, bannerIDs []int64) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	if _, err := tx.Exec(ctx, `DELETE FROM page_banners WHERE page_slug = $1`, slug); err != nil {
		return err
	}
	for i, id := range bannerIDs {
		if _, err := tx.Exec(ctx,
			`INSERT INTO page_banners (page_slug, banner_id, order_num) VALUES ($1, $2, $3)
			 ON CONFLICT (page_slug, banner_id) DO UPDATE SET order_num = $3`,
			slug, id, i,
		); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
