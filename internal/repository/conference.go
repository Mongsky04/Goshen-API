package repository

import (
	"context"
	"errors"
	"fmt"

	"goshen/backend/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ConferenceRepo struct{ db *pgxpool.Pool }

func NewConferenceRepo(db *pgxpool.Pool) *ConferenceRepo { return &ConferenceRepo{db: db} }

func (r *ConferenceRepo) ListUnpublishedSlugs(ctx context.Context) ([]string, error) {
	rows, err := r.db.Query(ctx, `SELECT slug FROM conference_pages WHERE is_published=false`)
	if err != nil {
		return nil, fmt.Errorf("list unpublished conference slugs: %w", err)
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

func (r *ConferenceRepo) Get(ctx context.Context, slug string) (*model.ConferenceAdminData, error) {
	var data model.ConferenceAdminData
	var pageID string

	err := r.db.QueryRow(ctx,
		`SELECT id::text, slug, label, is_published FROM conference_pages WHERE slug=$1`, slug,
	).Scan(&pageID, &data.Slug, &data.Label, &data.IsPublished)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get conference page: %w", err)
	}
	data.ID = pageID

	_ = r.db.QueryRow(ctx,
		`SELECT hero_image_url, badge_text, headline, sub_text
		 FROM conference_hero WHERE page_id=$1`, pageID,
	).Scan(&data.Hero.HeroImageURL, &data.Hero.BadgeText, &data.Hero.Headline, &data.Hero.SubText)

	titleRows, err := r.db.Query(ctx,
		`SELECT section_key, title FROM conference_section_titles WHERE page_id=$1`, pageID)
	if err != nil {
		return nil, fmt.Errorf("get section titles: %w", err)
	}
	for titleRows.Next() {
		var key, title string
		if err := titleRows.Scan(&key, &title); err != nil {
			titleRows.Close()
			return nil, err
		}
		switch key {
		case "product_grid":
			data.Titles.ProductGrid = title
		case "workspace":
			data.Titles.Workspace = title
		case "solutions":
			data.Titles.Solutions = title
		case "contact":
			data.Titles.Contact = title
		}
	}
	titleRows.Close()

	_ = r.db.QueryRow(ctx,
		`SELECT description FROM conference_workspace WHERE page_id=$1`, pageID,
	).Scan(&data.WorkspaceDescription)

	solRows, err := r.db.Query(ctx,
		`SELECT id::text, room_size, title, description, kit_label,
		        COALESCE(image_url,''), COALESCE(image_url_2,''),
		        COALESCE(card1_name,''), COALESCE(card1_category,''), COALESCE(card1_sub_category,''),
		        COALESCE(card2_name,''), COALESCE(card2_category,''), COALESCE(card2_sub_category,''),
		        is_hidden, sort_order
		 FROM conference_room_solutions WHERE page_id=$1 ORDER BY sort_order ASC`, pageID)
	if err != nil {
		return nil, fmt.Errorf("get room solutions: %w", err)
	}
	var solutions []model.ConferenceSolutionData
	for solRows.Next() {
		var s model.ConferenceSolutionData
		if err := solRows.Scan(&s.ID, &s.RoomSize, &s.Title, &s.Description, &s.KitLabel,
			&s.ImageURL, &s.ImageURL2,
			&s.Card1Name, &s.Card1Category, &s.Card1SubCategory,
			&s.Card2Name, &s.Card2Category, &s.Card2SubCategory,
			&s.IsHidden, &s.SortOrder); err != nil {
			solRows.Close()
			return nil, err
		}
		solutions = append(solutions, s)
	}
	solRows.Close()

	for i := range solutions {
		itemRows, err := r.db.Query(ctx,
			`SELECT item FROM conference_room_kit_items WHERE room_solution_id=$1 ORDER BY sort_order ASC`,
			solutions[i].ID)
		if err != nil {
			return nil, fmt.Errorf("get kit items: %w", err)
		}
		var items []string
		for itemRows.Next() {
			var item string
			if err := itemRows.Scan(&item); err != nil {
				itemRows.Close()
				return nil, err
			}
			items = append(items, item)
		}
		itemRows.Close()
		if items == nil {
			items = []string{}
		}
		solutions[i].KitItems = items
	}
	if solutions == nil {
		solutions = []model.ConferenceSolutionData{}
	}
	data.Solutions = solutions

	prodRows, err := r.db.Query(ctx,
		`SELECT cp.id::text, p.id, p.name, p.category, COALESCE(p.sub_category,''), COALESCE(p.image_url,''), cp.is_hidden
		 FROM conference_products cp
		 JOIN products p ON p.id = cp.product_id
		 WHERE cp.page_id=$1 AND cp.section='product_grid' ORDER BY cp.sort_order ASC`, pageID)
	if err != nil {
		return nil, fmt.Errorf("get products: %w", err)
	}
	var products []model.ConferenceProductData
	for prodRows.Next() {
		var p model.ConferenceProductData
		if err := prodRows.Scan(&p.ID, &p.ProductID, &p.Name, &p.Category, &p.SubCategory, &p.ImageURL, &p.IsHidden); err != nil {
			prodRows.Close()
			return nil, err
		}
		products = append(products, p)
	}
	prodRows.Close()
	if products == nil {
		products = []model.ConferenceProductData{}
	}
	data.Products = products

	return &data, nil
}

func (r *ConferenceRepo) Save(ctx context.Context, data *model.ConferenceAdminData) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	var pageID string
	err = tx.QueryRow(ctx,
		`INSERT INTO conference_pages (slug, label, is_published)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (slug) DO UPDATE SET label=EXCLUDED.label, is_published=EXCLUDED.is_published
		 RETURNING id::text`,
		data.Slug, data.Label, data.IsPublished,
	).Scan(&pageID)
	if err != nil {
		return fmt.Errorf("upsert page: %w", err)
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO conference_hero (page_id, hero_image_url, badge_text, headline, sub_text)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (page_id) DO UPDATE SET
		   hero_image_url=EXCLUDED.hero_image_url, badge_text=EXCLUDED.badge_text,
		   headline=EXCLUDED.headline, sub_text=EXCLUDED.sub_text`,
		pageID, data.Hero.HeroImageURL, data.Hero.BadgeText, data.Hero.Headline, data.Hero.SubText,
	)
	if err != nil {
		return fmt.Errorf("upsert hero: %w", err)
	}

	type sectionRow struct{ key, title string }
	for _, s := range []sectionRow{
		{"product_grid", data.Titles.ProductGrid},
		{"workspace", data.Titles.Workspace},
		{"solutions", data.Titles.Solutions},
		{"contact", data.Titles.Contact},
	} {
		_, err = tx.Exec(ctx,
			`INSERT INTO conference_section_titles (page_id, section_key, title)
			 VALUES ($1, $2, $3)
			 ON CONFLICT (page_id, section_key) DO UPDATE SET title=EXCLUDED.title`,
			pageID, s.key, s.title,
		)
		if err != nil {
			return fmt.Errorf("upsert section title %s: %w", s.key, err)
		}
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO conference_workspace (page_id, description)
		 VALUES ($1, $2)
		 ON CONFLICT (page_id) DO UPDATE SET description=EXCLUDED.description`,
		pageID, data.WorkspaceDescription,
	)
	if err != nil {
		return fmt.Errorf("upsert workspace: %w", err)
	}

	// Load existing room solutions so we can delete removed ones
	existingRows, err := tx.Query(ctx,
		`SELECT id::text, room_size FROM conference_room_solutions WHERE page_id=$1`, pageID)
	if err != nil {
		return fmt.Errorf("query room solutions: %w", err)
	}
	existing := map[string]string{} // roomSize -> id
	for existingRows.Next() {
		var id, rs string
		if err := existingRows.Scan(&id, &rs); err != nil {
			existingRows.Close()
			return err
		}
		existing[rs] = id
	}
	existingRows.Close()

	incoming := map[string]bool{}
	for i, s := range data.Solutions {
		incoming[s.RoomSize] = true
		var solID string
		imgURL := sqlNullStr(s.ImageURL)
		imgURL2 := sqlNullStr(s.ImageURL2)
		err = tx.QueryRow(ctx,
			`INSERT INTO conference_room_solutions
			   (page_id, room_size, title, description, kit_label, image_url, image_url_2,
			    card1_name, card1_category, card1_sub_category,
			    card2_name, card2_category, card2_sub_category,
			    is_hidden, sort_order)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
			 ON CONFLICT (page_id, room_size) DO UPDATE SET
			   title=EXCLUDED.title, description=EXCLUDED.description,
			   kit_label=EXCLUDED.kit_label, image_url=EXCLUDED.image_url,
			   image_url_2=EXCLUDED.image_url_2,
			   card1_name=EXCLUDED.card1_name, card1_category=EXCLUDED.card1_category,
			   card1_sub_category=EXCLUDED.card1_sub_category,
			   card2_name=EXCLUDED.card2_name, card2_category=EXCLUDED.card2_category,
			   card2_sub_category=EXCLUDED.card2_sub_category,
			   is_hidden=EXCLUDED.is_hidden, sort_order=EXCLUDED.sort_order
			 RETURNING id::text`,
			pageID, s.RoomSize, s.Title, s.Description, s.KitLabel,
			imgURL, imgURL2,
			s.Card1Name, s.Card1Category, s.Card1SubCategory,
			s.Card2Name, s.Card2Category, s.Card2SubCategory,
			s.IsHidden, i+1,
		).Scan(&solID)
		if err != nil {
			return fmt.Errorf("upsert room solution %q: %w", s.RoomSize, err)
		}

		_, err = tx.Exec(ctx,
			`DELETE FROM conference_room_kit_items WHERE room_solution_id=$1`, solID)
		if err != nil {
			return fmt.Errorf("clear kit items: %w", err)
		}
		for j, item := range s.KitItems {
			if item == "" {
				continue
			}
			_, err = tx.Exec(ctx,
				`INSERT INTO conference_room_kit_items (room_solution_id, item, sort_order)
				 VALUES ($1, $2, $3)`,
				solID, item, j+1,
			)
			if err != nil {
				return fmt.Errorf("insert kit item: %w", err)
			}
		}
	}

	for roomSize, id := range existing {
		if !incoming[roomSize] {
			_, err = tx.Exec(ctx, `DELETE FROM conference_room_solutions WHERE id=$1`, id)
			if err != nil {
				return fmt.Errorf("delete room solution: %w", err)
			}
		}
	}

	_, err = tx.Exec(ctx,
		`DELETE FROM conference_products WHERE page_id=$1 AND section='product_grid'`, pageID)
	if err != nil {
		return fmt.Errorf("delete products: %w", err)
	}
	for i, p := range data.Products {
		_, err = tx.Exec(ctx,
			`INSERT INTO conference_products (page_id, section, product_id, is_hidden, sort_order)
			 VALUES ($1,'product_grid',$2,$3,$4)`,
			pageID, p.ProductID, p.IsHidden, i+1,
		)
		if err != nil {
			return fmt.Errorf("insert product: %w", err)
		}
	}

	return tx.Commit(ctx)
}

func sqlNullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}
