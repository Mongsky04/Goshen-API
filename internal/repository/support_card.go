package repository

import (
	"context"
	"fmt"

	"goshen/backend/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SupportCardRepo struct{ db *pgxpool.Pool }

func NewSupportCardRepo(db *pgxpool.Pool) *SupportCardRepo {
	return &SupportCardRepo{db: db}
}

func (r *SupportCardRepo) List(ctx context.Context) ([]model.SupportCard, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, title, description, cta_label, cta_href, sort_order
		 FROM homepage_support_cards ORDER BY sort_order ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.SupportCard
	for rows.Next() {
		var c model.SupportCard
		if err := rows.Scan(&c.ID, &c.Title, &c.Description, &c.CtaLabel, &c.CtaHref, &c.SortOrder); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	if out == nil {
		out = []model.SupportCard{}
	}
	return out, nil
}

func (r *SupportCardRepo) Create(ctx context.Context, title, description, ctaLabel, ctaHref string, sortOrder int) (*model.SupportCard, error) {
	var c model.SupportCard
	err := r.db.QueryRow(ctx,
		`INSERT INTO homepage_support_cards (title, description, cta_label, cta_href, sort_order)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, title, description, cta_label, cta_href, sort_order`,
		title, description, ctaLabel, ctaHref, sortOrder,
	).Scan(&c.ID, &c.Title, &c.Description, &c.CtaLabel, &c.CtaHref, &c.SortOrder)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *SupportCardRepo) Delete(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx,
		`DELETE FROM homepage_support_cards WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("support card not found")
	}
	return nil
}
