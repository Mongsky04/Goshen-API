package repository

import (
	"context"

	"goshen/backend/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AdminRepo struct{ db *pgxpool.Pool }

func NewAdminRepo(db *pgxpool.Pool) *AdminRepo { return &AdminRepo{db: db} }

func (r *AdminRepo) FindByEmail(ctx context.Context, email string) (*model.Admin, error) {
	var a model.Admin
	err := r.db.QueryRow(ctx,
		`SELECT id, email, password_hash FROM admins WHERE email=$1`, email,
	).Scan(&a.ID, &a.Email, &a.PasswordHash)
	return &a, err
}
