package repository

import (
	"context"
	"errors"
	"fmt"

	"goshen/backend/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const performerRows = 3
const performerCols = 4

type PerformerRepo struct{ db *pgxpool.Pool }

func (r *PerformerRepo) ListUnpublishedSlugs(ctx context.Context) ([]string, error) {
	rows, err := r.db.Query(ctx, `SELECT slug FROM performer_pages WHERE is_published=false`)
	if err != nil {
		return nil, fmt.Errorf("list unpublished performer slugs: %w", err)
	}
	defer rows.Close()
	var slugs []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		slugs = append(slugs, s)
	}
	return slugs, rows.Err()
}

func NewPerformerRepo(db *pgxpool.Pool) *PerformerRepo { return &PerformerRepo{db: db} }

func (r *PerformerRepo) Get(ctx context.Context, slug string) (*model.PerformerAdminData, error) {
	var data model.PerformerAdminData
	var pageID string

	err := r.db.QueryRow(ctx,
		`SELECT id::text, slug, label, is_published,
		        hero_image_url, product_grid_title, videos_section_title
		 FROM performer_pages WHERE slug=$1`, slug,
	).Scan(&pageID, &data.Slug, &data.Label, &data.IsPublished,
		&data.HeroImageURL, &data.ProductGridTitle, &data.VideosSectionTitle)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get performer page: %w", err)
	}
	data.ID = pageID

	prodRows, err := r.db.Query(ctx,
		`SELECT p.id, p.name, p.category, COALESCE(p.sub_category,''), COALESCE(p.image_url,''), pp.is_hidden, pp.sort_order
		 FROM performer_products pp
		 JOIN products p ON p.id = pp.product_id
		 WHERE pp.page_id=$1 ORDER BY pp.sort_order ASC`, pageID)
	if err != nil {
		return nil, fmt.Errorf("get performer products: %w", err)
	}
	var flatProducts []model.PerformerCellData
	for prodRows.Next() {
		var cell model.PerformerCellData
		if err := prodRows.Scan(&cell.ProductID, &cell.Name, &cell.Category, &cell.SubCategory, &cell.ImageURL, &cell.IsHidden, &cell.SortOrder); err != nil {
			prodRows.Close()
			return nil, err
		}
		flatProducts = append(flatProducts, cell)
	}
	prodRows.Close()
	if flatProducts == nil {
		flatProducts = []model.PerformerCellData{}
	}
	data.Products = flatProducts

	vidRows, err := r.db.Query(ctx,
		`SELECT is_main, title, subtitle, thumbnail_url, video_url, sort_order
		 FROM performer_videos WHERE page_id=$1 ORDER BY sort_order ASC`, pageID)
	if err != nil {
		return nil, fmt.Errorf("get performer videos: %w", err)
	}
	var relatedVideos []model.PerformerVideoData
	mainFound := false
	for vidRows.Next() {
		var v model.PerformerVideoData
		if err := vidRows.Scan(&v.IsMain, &v.Title, &v.Subtitle, &v.ThumbnailURL, &v.VideoURL, &v.SortOrder); err != nil {
			vidRows.Close()
			return nil, err
		}
		if v.IsMain && !mainFound {
			data.MainVideo = v
			mainFound = true
		} else {
			relatedVideos = append(relatedVideos, v)
		}
	}
	vidRows.Close()
	if relatedVideos == nil {
		relatedVideos = []model.PerformerVideoData{}
	}
	data.RelatedVideos = relatedVideos

	return &data, nil
}

func (r *PerformerRepo) Save(ctx context.Context, data *model.PerformerAdminData) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	var pageID string
	err = tx.QueryRow(ctx,
		`INSERT INTO performer_pages
		   (slug, label, is_published, hero_image_url, product_grid_title, videos_section_title)
		 VALUES ($1,$2,$3,$4,$5,$6)
		 ON CONFLICT (slug) DO UPDATE SET
		   label=EXCLUDED.label, is_published=EXCLUDED.is_published,
		   hero_image_url=EXCLUDED.hero_image_url,
		   product_grid_title=EXCLUDED.product_grid_title,
		   videos_section_title=EXCLUDED.videos_section_title
		 RETURNING id::text`,
		data.Slug, data.Label, data.IsPublished,
		data.HeroImageURL, data.ProductGridTitle, data.VideosSectionTitle,
	).Scan(&pageID)
	if err != nil {
		return fmt.Errorf("upsert performer page: %w", err)
	}

	_, err = tx.Exec(ctx, `DELETE FROM performer_products WHERE page_id=$1`, pageID)
	if err != nil {
		return fmt.Errorf("delete performer products: %w", err)
	}
	for i, cell := range data.Products {
		_, err = tx.Exec(ctx,
			`INSERT INTO performer_products (page_id, product_id, is_hidden, sort_order)
			 VALUES ($1,$2,$3,$4)`,
			pageID, cell.ProductID, cell.IsHidden, i,
		)
		if err != nil {
			return fmt.Errorf("insert performer product %d: %w", i, err)
		}
	}

	_, err = tx.Exec(ctx, `DELETE FROM performer_videos WHERE page_id=$1`, pageID)
	if err != nil {
		return fmt.Errorf("delete performer videos: %w", err)
	}

	allVideos := append([]model.PerformerVideoData{data.MainVideo}, data.RelatedVideos...)
	for idx, v := range allVideos {
		_, err = tx.Exec(ctx,
			`INSERT INTO performer_videos
			   (page_id, is_main, title, subtitle, thumbnail_url, video_url, sort_order)
			 VALUES ($1,$2,$3,$4,$5,$6,$7)`,
			pageID, v.IsMain, v.Title, v.Subtitle, v.ThumbnailURL, v.VideoURL, idx,
		)
		if err != nil {
			return fmt.Errorf("insert performer video %d: %w", idx, err)
		}
	}

	return tx.Commit(ctx)
}
