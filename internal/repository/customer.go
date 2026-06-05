package repository

import (
	"context"

	"goshen/backend/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type CustomerRepo struct{ db *pgxpool.Pool }

func NewCustomerRepo(db *pgxpool.Pool) *CustomerRepo { return &CustomerRepo{db: db} }

func (r *CustomerRepo) List(ctx context.Context) ([]model.Customer, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, image_url, alt_text, created_at, updated_at FROM customers ORDER BY id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Customer
	for rows.Next() {
		var c model.Customer
		if err := rows.Scan(&c.ID, &c.ImageURL, &c.AltText, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	if out == nil {
		out = []model.Customer{}
	}
	return out, nil
}

func (r *CustomerRepo) GetByID(ctx context.Context, id int64) (*model.Customer, error) {
	var c model.Customer
	err := r.db.QueryRow(ctx,
		`SELECT id, image_url, alt_text, created_at, updated_at FROM customers WHERE id=$1`,
		id,
	).Scan(&c.ID, &c.ImageURL, &c.AltText, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *CustomerRepo) Create(ctx context.Context, imageURL, altText string) (*model.Customer, error) {	var c model.Customer
	err := r.db.QueryRow(ctx,		`INSERT INTO customers (image_url, alt_text) VALUES ($1, $2)
		 RETURNING id, image_url, alt_text, created_at, updated_at`,
		imageURL, altText,
	).Scan(&c.ID, &c.ImageURL, &c.AltText, &c.CreatedAt, &c.UpdatedAt)
	return &c, err
}

func (r *CustomerRepo) Update(ctx context.Context, id int64, imageURL, altText string) (*model.Customer, error) {
	var c model.Customer
	err := r.db.QueryRow(ctx,
		`UPDATE customers SET image_url=$1, alt_text=$2, updated_at=NOW()
		 WHERE id=$3
		 RETURNING id, image_url, alt_text, created_at, updated_at`,
		imageURL, altText, id,
	).Scan(&c.ID, &c.ImageURL, &c.AltText, &c.CreatedAt, &c.UpdatedAt)
	return &c, err
}

func (r *CustomerRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM customers WHERE id=$1`, id)
	return err
}