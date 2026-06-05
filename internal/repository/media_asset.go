package repository

import (
	"context"
	"fmt"

	"goshen/backend/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type MediaAssetRepo struct {
	db *pgxpool.Pool
}

func NewMediaAssetRepo(db *pgxpool.Pool) *MediaAssetRepo {
	return &MediaAssetRepo{db: db}
}

func (r *MediaAssetRepo) List(ctx context.Context) ([]model.MediaAsset, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, filename, url, size, mime_type, created_at
		 FROM media_assets
		 ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("media_asset.List: %w", err)
	}
	defer rows.Close()

	var assets []model.MediaAsset
	for rows.Next() {
		var a model.MediaAsset
		if err := rows.Scan(&a.ID, &a.Filename, &a.URL, &a.Size, &a.MimeType, &a.CreatedAt); err != nil {
			return nil, fmt.Errorf("media_asset.List scan: %w", err)
		}
		assets = append(assets, a)
	}
	return assets, nil
}

func (r *MediaAssetRepo) Create(ctx context.Context, filename, url string, size int64, mimeType string) (*model.MediaAsset, error) {
	var a model.MediaAsset
	err := r.db.QueryRow(ctx,
		`INSERT INTO media_assets (filename, url, size, mime_type)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, filename, url, size, mime_type, created_at`,
		filename, url, size, mimeType,
	).Scan(&a.ID, &a.Filename, &a.URL, &a.Size, &a.MimeType, &a.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("media_asset.Create: %w", err)
	}
	return &a, nil
}

func (r *MediaAssetRepo) GetByID(ctx context.Context, id int64) (*model.MediaAsset, error) {
	var a model.MediaAsset
	err := r.db.QueryRow(ctx,
		`SELECT id, filename, url, size, mime_type, created_at
		 FROM media_assets WHERE id = $1`,
		id,
	).Scan(&a.ID, &a.Filename, &a.URL, &a.Size, &a.MimeType, &a.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("media_asset.GetByID: %w", err)
	}
	return &a, nil
}

func (r *MediaAssetRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM media_assets WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("media_asset.Delete: %w", err)
	}
	return nil
}
