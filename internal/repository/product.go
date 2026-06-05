package repository

import (
	"context"

	"goshen/backend/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductRepo struct{ db *pgxpool.Pool }

func NewProductRepo(db *pgxpool.Pool) *ProductRepo { return &ProductRepo{db: db} }

const productCols = `id, name, image_url, category, sub_category, created_at, updated_at`

func scanProduct(row interface{ Scan(...any) error }) (model.Product, error) {
	var p model.Product
	err := row.Scan(&p.ID, &p.Name, &p.ImageURL, &p.Category, &p.SubCategory, &p.CreatedAt, &p.UpdatedAt)
	return p, err
}

func (r *ProductRepo) List(ctx context.Context) ([]model.Product, error) {
	rows, err := r.db.Query(ctx,
		`SELECT `+productCols+` FROM products ORDER BY sort_order ASC, id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Product
	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	if out == nil {
		out = []model.Product{}
	}
	return out, nil
}

func (r *ProductRepo) ListPaged(ctx context.Context, limit, offset int) ([]model.Product, error) {
	rows, err := r.db.Query(ctx,
		`SELECT `+productCols+` FROM products ORDER BY sort_order ASC, id ASC LIMIT $1 OFFSET $2`,
		limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Product
	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	if out == nil {
		out = []model.Product{}
	}
	return out, nil
}

func (r *ProductRepo) GetByID(ctx context.Context, id int64) (*model.Product, error) {
	p, err := scanProduct(r.db.QueryRow(ctx,
		`SELECT `+productCols+` FROM products WHERE id=$1`, id))
	return &p, err
}

func (r *ProductRepo) Create(ctx context.Context, name, imageURL, category, subCategory string) (*model.Product, error) {
	p, err := scanProduct(r.db.QueryRow(ctx,
		`INSERT INTO products (name, image_url, category, sub_category)
		 VALUES ($1, $2, $3, $4)
		 RETURNING `+productCols,
		name, imageURL, category, subCategory))
	return &p, err
}

func (r *ProductRepo) Update(ctx context.Context, id int64, name, imageURL, category, subCategory string) (*model.Product, error) {
	p, err := scanProduct(r.db.QueryRow(ctx,
		`UPDATE products SET name=$1, image_url=$2, category=$3, sub_category=$4, updated_at=NOW()
		 WHERE id=$5
		 RETURNING `+productCols,
		name, imageURL, category, subCategory, id))
	return &p, err
}

func (r *ProductRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM products WHERE id=$1`, id)
	return err
}
