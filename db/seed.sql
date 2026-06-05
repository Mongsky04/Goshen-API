-- ============================================================
-- seed.sql — Initial data for local development
-- Usage: psql "$DATABASE_URL" -f db/seed.sql
--
-- Images are intentionally left empty — upload them via the
-- admin panel at http://localhost:3000/admin
-- ============================================================

-- ── Admin user ───────────────────────────────────────────────
-- Login: admin@goshen.id / admin123

INSERT INTO admins (email, password_hash)
VALUES ('admin@goshen.id', crypt('admin123', gen_salt('bf', 10)))
ON CONFLICT (email) DO UPDATE SET password_hash = EXCLUDED.password_hash;

-- ── Brands ───────────────────────────────────────────────────

INSERT INTO brands (name, sort_order) VALUES
  ('Shure',       1),
  ('Sennheiser',  2),
  ('Sony',        3),
  ('Yamaha',      4),
  ('Bose',        5),
  ('JBL',         6),
  ('Audio-Technica', 7),
  ('Rode',        8)
ON CONFLICT DO NOTHING;

-- ── Products (grid) ──────────────────────────────────────────

INSERT INTO products (name, sort_order) VALUES
  ('Wireless Microphone System',  1),
  ('Conference Speakerphone',     2),
  ('Digital Mixing Console',      3),
  ('PTZ Camera 4K',               4),
  ('AV Receiver',                 5),
  ('Ceiling Speaker Set',         6)
ON CONFLICT DO NOTHING;

-- ── Featured products ────────────────────────────────────────

INSERT INTO featured (name, category, sub_category, featured_categories, sort_order) VALUES
  ('MXA920 Ceiling Array Mic', 'Microphone', 'Ceiling Array',  ARRAY['New Products', 'Best Sellers'], 1),
  ('ULXD4 Receiver',           'Wireless',   'UHF Receiver',   ARRAY['Best Sellers'],                 2),
  ('IntelliMix P300',          'DSP',        'Audio Processor', ARRAY['New Products'],                3)
ON CONFLICT DO NOTHING;

-- ── Articles ─────────────────────────────────────────────────

INSERT INTO articles (title, description, sort_order) VALUES
  (
    'Solusi Audio untuk Ruang Konferensi Modern',
    'Pelajari bagaimana sistem audio terpadu dapat meningkatkan kualitas meeting hybrid dan remote collaboration di era kerja modern.',
    1
  ),
  (
    'Panduan Memilih Mikrofon untuk Live Event',
    'Dari wireless handheld hingga clip-on lavalier — temukan jenis mikrofon yang paling tepat untuk kebutuhan panggung dan presentasi Anda.',
    2
  ),
  (
    'Teknologi AV Terbaru untuk Auditorium',
    'Review lengkap sistem tata suara dan visual untuk auditorium berkapasitas besar: dari speaker line array hingga sistem distribusi sinyal digital.',
    3
  )
ON CONFLICT DO NOTHING;

-- ── Customers ────────────────────────────────────────────────

INSERT INTO customers (alt_text, sort_order) VALUES
  ('Bank BRI',            1),
  ('Telkom Indonesia',    2),
  ('Garuda Indonesia',    3),
  ('Pertamina',           4),
  ('Bank Mandiri',        5),
  ('PLN',                 6)
ON CONFLICT DO NOTHING;

-- ── Slider (homepage hero) ───────────────────────────────────
-- image_url left empty — upload via admin panel

INSERT INTO slider (image_url, order_num) VALUES
  ('', 1),
  ('', 2),
  ('', 3)
ON CONFLICT DO NOTHING;

-- ── Homepage support cards ───────────────────────────────────

INSERT INTO homepage_support_cards (title, description, cta_label, cta_href, sort_order) VALUES
  (
    'Technical Support',
    'Tim teknisi berpengalaman kami siap membantu instalasi dan troubleshooting sistem AV Anda.',
    'Hubungi Kami',
    '/contact',
    1
  ),
  (
    'Product Demo',
    'Jadwalkan demo produk langsung di showroom kami atau via video call dengan spesialis produk.',
    'Buat Janji',
    '/demo',
    2
  ),
  (
    'Rental & Event',
    'Layanan rental peralatan AV profesional untuk event, seminar, dan conference Anda.',
    'Lihat Paket',
    '/rental',
    3
  )
ON CONFLICT DO NOTHING;
