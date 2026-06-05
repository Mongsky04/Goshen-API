#!/bin/sh
set -e
psql "$DATABASE_URL" -f db/schema.sql
psql "$DATABASE_URL" -c "INSERT INTO admins (email, password_hash) VALUES ('admin@goshen.id', crypt('admin123', gen_salt('bf', 10))) ON CONFLICT (email) DO NOTHING"
exec ./server
