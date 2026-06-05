-- ============================================================
-- reset.sql — Wipe all data and schema
-- WARNING: Destructive. Local development only.
-- Usage: psql "$DATABASE_URL" -f db/reset.sql
-- ============================================================

DROP TABLE IF EXISTS conference_room_kit_items   CASCADE;
DROP TABLE IF EXISTS conference_room_solutions    CASCADE;
DROP TABLE IF EXISTS conference_workspace         CASCADE;
DROP TABLE IF EXISTS conference_products          CASCADE;
DROP TABLE IF EXISTS conference_section_titles    CASCADE;
DROP TABLE IF EXISTS conference_hero              CASCADE;
DROP TABLE IF EXISTS conference_pages             CASCADE;

DROP TABLE IF EXISTS performer_videos    CASCADE;
DROP TABLE IF EXISTS performer_products  CASCADE;
DROP TABLE IF EXISTS performer_pages     CASCADE;

DROP TABLE IF EXISTS homepage_support_cards  CASCADE;
DROP TABLE IF EXISTS customers               CASCADE;
DROP TABLE IF EXISTS brands                  CASCADE;
DROP TABLE IF EXISTS slider                  CASCADE;
DROP TABLE IF EXISTS articles                CASCADE;
DROP TABLE IF EXISTS featured                CASCADE;
DROP TABLE IF EXISTS products                CASCADE;
DROP TABLE IF EXISTS admins                  CASCADE;

DROP FUNCTION IF EXISTS set_updated_at() CASCADE;
