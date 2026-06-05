package repository

import (
	"context"

	"goshen/backend/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SliderRepo struct{ db *pgxpool.Pool }

func NewSliderRepo(db *pgxpool.Pool) *SliderRepo { return &SliderRepo{db: db} }

func (r *SliderRepo) List(ctx context.Context) ([]model.Slider, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, title, image_url, order_num, created_at, updated_at
		 FROM slider ORDER BY order_num ASC, id ASC`)
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

func (r *SliderRepo) GetByID(ctx context.Context, id int64) (*model.Slider, error) {
	var s model.Slider
	err := r.db.QueryRow(ctx,
		`SELECT id, title, image_url, order_num, created_at, updated_at FROM slider WHERE id=$1`,
		id,
	).Scan(&s.ID, &s.Title, &s.ImageURL, &s.OrderNum, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *SliderRepo) Create(ctx context.Context, title, imageURL string, orderNum int) (*model.Slider, error) {
	var s model.Slider
	err := r.db.QueryRow(ctx,
		`INSERT INTO slider (title, image_url, order_num) VALUES ($1, $2, $3)
		 RETURNING id, title, image_url, order_num, created_at, updated_at`,
		title, imageURL, orderNum,
	).Scan(&s.ID, &s.Title, &s.ImageURL, &s.OrderNum, &s.CreatedAt, &s.UpdatedAt)
	return &s, err
}

func (r *SliderRepo) Update(ctx context.Context, id int64, title, imageURL string, orderNum int) (*model.Slider, error) {
	var s model.Slider
	err := r.db.QueryRow(ctx,
		`UPDATE slider SET title=$1, image_url=$2, order_num=$3, updated_at=NOW()
		 WHERE id=$4
		 RETURNING id, title, image_url, order_num, created_at, updated_at`,
		title, imageURL, orderNum, id,
	).Scan(&s.ID, &s.Title, &s.ImageURL, &s.OrderNum, &s.CreatedAt, &s.UpdatedAt)
	return &s, err
}

func (r *SliderRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM slider WHERE id=$1`, id)
	return err
}
