package repository

import (
	"context"

	"goshen/backend/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type BrandRepo struct{ db *pgxpool.Pool }

func NewBrandRepo(db *pgxpool.Pool) *BrandRepo { return &BrandRepo{db: db} }

func (r *BrandRepo) List(ctx context.Context) ([]model.Brand, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, name, image_url, created_at, updated_at FROM brands ORDER BY id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Brand
	for rows.Next() {
		var b model.Brand
		if err := rows.Scan(&b.ID, &b.Name, &b.ImageURL, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	if out == nil {
		out = []model.Brand{}
	}
	return out, nil
}

func (r *BrandRepo) GetByID(ctx context.Context, id int64) (*model.Brand, error) {
	var b model.Brand
	err := r.db.QueryRow(ctx,
		`SELECT id, name, image_url, created_at, updated_at FROM brands WHERE id=$1`,
		id,
	).Scan(&b.ID, &b.Name, &b.ImageURL, &b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *BrandRepo) Create(ctx context.Context, name, imageURL string) (*model.Brand, error) {	var b model.Brand
	err := r.db.QueryRow(ctx,		`INSERT INTO brands (name, image_url) VALUES ($1, $2)
		 RETURNING id, name, image_url, created_at, updated_at`,
		name, imageURL,
	).Scan(&b.ID, &b.Name, &b.ImageURL, &b.CreatedAt, &b.UpdatedAt)
	return &b, err
}

func (r *BrandRepo) Update(ctx context.Context, id int64, name, imageURL string) (*model.Brand, error) {
	var b model.Brand
	err := r.db.QueryRow(ctx,
		`UPDATE brands SET name=$1, image_url=$2, updated_at=NOW()
		 WHERE id=$3
		 RETURNING id, name, image_url, created_at, updated_at`,
		name, imageURL, id,
	).Scan(&b.ID, &b.Name, &b.ImageURL, &b.CreatedAt, &b.UpdatedAt)
	return &b, err
}

func (r *BrandRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM brands WHERE id=$1`, id)
	return err
}
