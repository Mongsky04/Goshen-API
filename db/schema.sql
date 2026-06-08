-- ============================================================
-- schema.sql — Full database schema for Goshen
-- Usage: psql "$DATABASE_URL" -f db/schema.sql
-- ============================================================

-- Extensions
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ── Homepage content ─────────────────────────────────────────

CREATE TABLE IF NOT EXISTS admins (
  id            BIGSERIAL    PRIMARY KEY,
  email         TEXT         NOT NULL UNIQUE,
  password_hash TEXT         NOT NULL,
  created_at    TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS slider (
  id         BIGSERIAL    PRIMARY KEY,
  title      TEXT         NOT NULL DEFAULT '',
  image_url  TEXT         NOT NULL DEFAULT '',
  order_num  INT          NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ  NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ  NOT NULL DEFAULT now()
);
ALTER TABLE slider ADD COLUMN IF NOT EXISTS title TEXT NOT NULL DEFAULT '';

CREATE TABLE IF NOT EXISTS products (
  id           BIGSERIAL    PRIMARY KEY,
  name         TEXT         NOT NULL,
  image_url    TEXT         NOT NULL DEFAULT '',
  category     TEXT         NOT NULL DEFAULT '',
  sub_category TEXT         NOT NULL DEFAULT '',
  sort_order   INT          NOT NULL DEFAULT 0,
  created_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
  updated_at   TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS featured (
  id                  BIGSERIAL    PRIMARY KEY,
  product_id          BIGINT       NOT NULL REFERENCES products(id) ON DELETE CASCADE,
  name                TEXT         NOT NULL DEFAULT '',
  image_url           TEXT         NOT NULL DEFAULT '',
  category            TEXT         NOT NULL DEFAULT '',
  sub_category        TEXT         NOT NULL DEFAULT '',
  featured_categories TEXT[]       NOT NULL DEFAULT '{}',
  sort_order          INT          NOT NULL DEFAULT 0,
  created_at          TIMESTAMPTZ  NOT NULL DEFAULT now(),
  updated_at          TIMESTAMPTZ  NOT NULL DEFAULT now()
);
ALTER TABLE featured ADD COLUMN IF NOT EXISTS name TEXT NOT NULL DEFAULT '';
ALTER TABLE featured ADD COLUMN IF NOT EXISTS image_url TEXT NOT NULL DEFAULT '';
ALTER TABLE featured ADD COLUMN IF NOT EXISTS category TEXT NOT NULL DEFAULT '';
ALTER TABLE featured ADD COLUMN IF NOT EXISTS sub_category TEXT NOT NULL DEFAULT '';

CREATE TABLE IF NOT EXISTS homepage_grid_products (
  id         BIGSERIAL    PRIMARY KEY,
  product_id BIGINT       NOT NULL REFERENCES products(id) ON DELETE CASCADE,
  sort_order INT          NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS banners (
  id         BIGSERIAL    PRIMARY KEY,
  name       TEXT         NOT NULL DEFAULT '',
  image_url  TEXT         NOT NULL DEFAULT '',
  sort_order INT          NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ  NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS articles (
  id           BIGSERIAL    PRIMARY KEY,
  title        TEXT         NOT NULL,
  description  TEXT         NOT NULL DEFAULT '',
  image_url    TEXT         NOT NULL DEFAULT '',
  sort_order   INT          NOT NULL DEFAULT 0,
  published_at TIMESTAMPTZ  NOT NULL DEFAULT now(),
  created_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
  updated_at   TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS brands (
  id         BIGSERIAL    PRIMARY KEY,
  name       TEXT         NOT NULL,
  image_url  TEXT         NOT NULL DEFAULT '',
  sort_order INT          NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ  NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS customers (
  id         BIGSERIAL    PRIMARY KEY,
  image_url  TEXT         NOT NULL DEFAULT '',
  alt_text   TEXT         NOT NULL DEFAULT '',
  sort_order INT          NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ  NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS homepage_support_cards (
  id          UUID  PRIMARY KEY DEFAULT gen_random_uuid(),
  title       TEXT  NOT NULL DEFAULT '',
  description TEXT  NOT NULL DEFAULT '',
  cta_label   TEXT  NOT NULL DEFAULT '',
  cta_href    TEXT  NOT NULL DEFAULT '#',
  sort_order  INT   NOT NULL DEFAULT 0
);

-- ── Conference pages ─────────────────────────────────────────

CREATE TABLE IF NOT EXISTS conference_pages (
  id           UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  slug         TEXT         NOT NULL UNIQUE,
  label        TEXT         NOT NULL,
  is_published BOOLEAN      NOT NULL DEFAULT false,
  created_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
  updated_at   TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS conference_hero (
  id             UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  page_id        UUID         NOT NULL REFERENCES conference_pages(id) ON DELETE CASCADE,
  hero_image_url TEXT         NOT NULL DEFAULT '',
  badge_text     TEXT         NOT NULL DEFAULT '',
  headline       TEXT         NOT NULL DEFAULT '',
  sub_text       TEXT         NOT NULL DEFAULT '',
  created_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
  updated_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
  UNIQUE (page_id)
);

CREATE TABLE IF NOT EXISTS conference_section_titles (
  id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  page_id     UUID         NOT NULL REFERENCES conference_pages(id) ON DELETE CASCADE,
  section_key TEXT         NOT NULL,
  title       TEXT         NOT NULL DEFAULT '',
  created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
  updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
  UNIQUE (page_id, section_key)
);

CREATE TABLE IF NOT EXISTS conference_products (
  id         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  page_id    UUID         NOT NULL REFERENCES conference_pages(id) ON DELETE CASCADE,
  section    TEXT         NOT NULL CHECK (section IN ('product_grid', 'workspace', 'solutions')),
  product_id BIGINT       NOT NULL REFERENCES products(id) ON DELETE CASCADE,
  is_hidden  BOOLEAN      NOT NULL DEFAULT false,
  sort_order INT          NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ  NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS conference_workspace (
  id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  page_id     UUID         NOT NULL REFERENCES conference_pages(id) ON DELETE CASCADE,
  description TEXT         NOT NULL DEFAULT '',
  created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
  updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
  UNIQUE (page_id)
);

CREATE TABLE IF NOT EXISTS conference_room_solutions (
  id                UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  page_id           UUID         NOT NULL REFERENCES conference_pages(id) ON DELETE CASCADE,
  room_size         TEXT         NOT NULL,
  title             TEXT         NOT NULL DEFAULT '',
  description       TEXT         NOT NULL DEFAULT '',
  kit_label         TEXT         NOT NULL DEFAULT 'IMX ROOM KIT 30:',
  image_url         TEXT         NOT NULL DEFAULT '',
  image_url_2       TEXT         NOT NULL DEFAULT '',
  card1_name        TEXT         NOT NULL DEFAULT '',
  card1_category    TEXT         NOT NULL DEFAULT '',
  card1_sub_category TEXT        NOT NULL DEFAULT '',
  card2_name        TEXT         NOT NULL DEFAULT '',
  card2_category    TEXT         NOT NULL DEFAULT '',
  card2_sub_category TEXT        NOT NULL DEFAULT '',
  is_hidden         BOOLEAN      NOT NULL DEFAULT false,
  sort_order        INT          NOT NULL DEFAULT 0,
  created_at        TIMESTAMPTZ  NOT NULL DEFAULT now(),
  updated_at        TIMESTAMPTZ  NOT NULL DEFAULT now(),
  UNIQUE (page_id, room_size)
);

CREATE TABLE IF NOT EXISTS conference_room_kit_items (
  id               UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  room_solution_id UUID         NOT NULL REFERENCES conference_room_solutions(id) ON DELETE CASCADE,
  item             TEXT         NOT NULL DEFAULT '',
  sort_order       INT          NOT NULL DEFAULT 0,
  created_at       TIMESTAMPTZ  NOT NULL DEFAULT now(),
  updated_at       TIMESTAMPTZ  NOT NULL DEFAULT now()
);

-- ── Performer pages ──────────────────────────────────────────

CREATE TABLE IF NOT EXISTS performer_pages (
  id                   UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  slug                 TEXT         NOT NULL UNIQUE,
  label                TEXT         NOT NULL,
  is_published         BOOLEAN      NOT NULL DEFAULT false,
  hero_image_url       TEXT         NOT NULL DEFAULT '',
  product_grid_title   TEXT         NOT NULL DEFAULT '',
  videos_section_title TEXT         NOT NULL DEFAULT '',
  created_at           TIMESTAMPTZ  NOT NULL DEFAULT now(),
  updated_at           TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS performer_products (
  id         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  page_id    UUID         NOT NULL REFERENCES performer_pages(id) ON DELETE CASCADE,
  product_id BIGINT       NOT NULL REFERENCES products(id) ON DELETE CASCADE,
  is_hidden  BOOLEAN      NOT NULL DEFAULT false,
  sort_order INT          NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ  NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS performer_videos (
  id            UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  page_id       UUID         NOT NULL REFERENCES performer_pages(id) ON DELETE CASCADE,
  is_main       BOOLEAN      NOT NULL DEFAULT false,
  title         TEXT         NOT NULL DEFAULT '',
  subtitle      TEXT         NOT NULL DEFAULT '',
  thumbnail_url TEXT         NOT NULL DEFAULT '',
  video_url     TEXT         NOT NULL DEFAULT '',
  sort_order    INT          NOT NULL DEFAULT 0,
  created_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
  updated_at    TIMESTAMPTZ  NOT NULL DEFAULT now()
);

-- ── Media library ───────────────────────────────────────────

CREATE TABLE IF NOT EXISTS media_assets (
  id         BIGSERIAL    PRIMARY KEY,
  filename   TEXT         NOT NULL DEFAULT '',
  url        TEXT         NOT NULL,
  size       BIGINT       NOT NULL DEFAULT 0,
  mime_type  TEXT         NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ  NOT NULL DEFAULT now()
);

-- ── updated_at trigger ───────────────────────────────────────

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$;

CREATE TABLE IF NOT EXISTS page_banners (
  id         BIGSERIAL   PRIMARY KEY,
  page_slug  TEXT        NOT NULL,
  banner_id  BIGINT      NOT NULL REFERENCES slider(id) ON DELETE CASCADE,
  order_num  INT         NOT NULL DEFAULT 0,
  UNIQUE(page_slug, banner_id)
);

DO $$
DECLARE t TEXT;
BEGIN
  FOREACH t IN ARRAY ARRAY[
    'slider', 'products', 'featured', 'banners', 'articles', 'brands', 'customers',
    'conference_pages', 'conference_hero', 'conference_section_titles',
    'conference_products', 'conference_workspace',
    'conference_room_solutions', 'conference_room_kit_items',
    'performer_pages', 'performer_products', 'performer_videos'
  ] LOOP
    EXECUTE format(
      'DROP TRIGGER IF EXISTS trg_%I_updated_at ON %I;
       CREATE TRIGGER trg_%I_updated_at
       BEFORE UPDATE ON %I
       FOR EACH ROW EXECUTE FUNCTION set_updated_at()',
      t, t, t, t
    );
  END LOOP;
END;
$$;
