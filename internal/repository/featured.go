package repository

import (
	"context"

	"goshen/backend/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type FeaturedRepo struct{ db *pgxpool.Pool }

func NewFeaturedRepo(db *pgxpool.Pool) *FeaturedRepo { return &FeaturedRepo{db: db} }

const featuredJoin = `
	SELECT f.id, f.product_id, p.name, p.image_url, p.category, COALESCE(p.sub_category,''),
	       f.featured_categories, f.created_at, f.updated_at
	FROM featured f
	JOIN products p ON p.id = f.product_id`

func scanFeatured(row interface{ Scan(...any) error }) (model.Featured, error) {
	var f model.Featured
	err := row.Scan(&f.ID, &f.ProductID, &f.Name, &f.ImageURL, &f.Category, &f.SubCategory,
		&f.FeaturedCategories, &f.CreatedAt, &f.UpdatedAt)
	return f, err
}

func (r *FeaturedRepo) List(ctx context.Context) ([]model.Featured, error) {
	rows, err := r.db.Query(ctx, featuredJoin+` ORDER BY f.sort_order ASC, f.id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Featured
	for rows.Next() {
		f, err := scanFeatured(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	if out == nil {
		out = []model.Featured{}
	}
	return out, nil
}

func (r *FeaturedRepo) ListPaged(ctx context.Context, limit, offset int) ([]model.Featured, error) {
	rows, err := r.db.Query(ctx,
		featuredJoin+` ORDER BY f.sort_order ASC, f.id ASC LIMIT $1 OFFSET $2`,
		limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Featured
	for rows.Next() {
		f, err := scanFeatured(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	if out == nil {
		out = []model.Featured{}
	}
	return out, nil
}

func (r *FeaturedRepo) GetByID(ctx context.Context, id int64) (*model.Featured, error) {
	f, err := scanFeatured(r.db.QueryRow(ctx,
		featuredJoin+` WHERE f.id=$1`, id))
	return &f, err
}

func (r *FeaturedRepo) Create(ctx context.Context, productID int64, featuredCategories []string) (*model.Featured, error) {
	var id int64
	err := r.db.QueryRow(ctx,
		`INSERT INTO featured (product_id, featured_categories)
		 VALUES ($1, $2) RETURNING id`,
		productID, featuredCategories,
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	f, err := r.GetByID(ctx, id)
	return f, err
}

func (r *FeaturedRepo) Update(ctx context.Context, id int64, productID int64, featuredCategories []string) (*model.Featured, error) {
	_, err := r.db.Exec(ctx,
		`UPDATE featured SET product_id=$1, featured_categories=$2, updated_at=NOW() WHERE id=$3`,
		productID, featuredCategories, id)
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id)
}

func (r *FeaturedRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM featured WHERE id=$1`, id)
	return err
}

// ── Homepage grid products ─────────────────────────────────────

type HomepageGridRepo struct{ db *pgxpool.Pool }

func NewHomepageGridRepo(db *pgxpool.Pool) *HomepageGridRepo { return &HomepageGridRepo{db: db} }

func (r *HomepageGridRepo) List(ctx context.Context) ([]model.HomepageGridProduct, error) {
	rows, err := r.db.Query(ctx,
		`SELECT hg.id, hg.product_id, p.name, p.image_url, p.category, COALESCE(p.sub_category,'')
		 FROM homepage_grid_products hg
		 JOIN products p ON p.id = hg.product_id
		 ORDER BY hg.sort_order ASC, hg.id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.HomepageGridProduct
	for rows.Next() {
		var g model.HomepageGridProduct
		if err := rows.Scan(&g.ID, &g.ProductID, &g.Name, &g.ImageURL, &g.Category, &g.SubCategory); err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	if out == nil {
		out = []model.HomepageGridProduct{}
	}
	return out, nil
}

func (r *HomepageGridRepo) ReplaceAll(ctx context.Context, productIDs []int64) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	if _, err := tx.Exec(ctx, `DELETE FROM homepage_grid_products`); err != nil {
		return err
	}
	for i, pid := range productIDs {
		if _, err := tx.Exec(ctx,
			`INSERT INTO homepage_grid_products (product_id, sort_order) VALUES ($1, $2)`,
			pid, i); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
