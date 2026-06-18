-- ============================================================
-- seed-dev-homepage.sql — DEV-ONLY visual verification data
-- NEVER run on production. Rows are marked sort_order 900+ so
-- this file is idempotent and trivially removable:
--   DELETE FROM featured WHERE sort_order BETWEEN 900 AND 999;
--   DELETE FROM articles WHERE sort_order BETWEEN 900 AND 999;
--   DELETE FROM products WHERE sort_order BETWEEN 900 AND 999;  -- cascades grid links
-- Images point at goshen-web /public assets (served by :3000).
-- ============================================================

-- Clean any previous dev seed (idempotent) ───────────────────
DELETE FROM featured WHERE sort_order BETWEEN 900 AND 999;
DELETE FROM articles WHERE sort_order BETWEEN 900 AND 999;
DELETE FROM products WHERE sort_order BETWEEN 900 AND 999; -- cascades homepage_grid_products

-- Products (grid) — 11 items, cycling the 5 local product images
INSERT INTO products (name, image_url, category, sub_category, sort_order) VALUES
  ('Shure NXN8 Vocal Microphone Cardiod',   '/assets/grid/Shure NXN8.png',   'Vocal', 'Microphone', 900),
  ('Shure ANX4 Scalable Wireless Receiver', '/assets/grid/Shure ANX4.png',   'Wireless', 'Receiver',  901),
  ('Shure MXA710 Linear Array Microphone',  '/assets/grid/Shure MXA710.png', 'Linear Array', 'Microphone', 902),
  ('Shure MXA920 Ceiling Array Microphone', '/assets/grid/Shure MXA920.png', 'Ceiling Array', 'Microphone', 903),
  ('Shure NXN8 Vocal Microphone Cardiod',   '/assets/grid/Shure NXN8.png',   'Vocal', 'Microphone', 904),
  ('Shure ANX4 Scalable Wireless Receiver', '/assets/grid/Shure ANX4.png',   'Wireless', 'Receiver',  905),
  ('Shure MXA710 Linear Array Microphone',  '/assets/grid/Shure MXA710.png', 'Linear Array', 'Microphone', 906),
  ('Shure MXA920 Ceiling Array Microphone', '/assets/grid/Shure MXA920.png', 'Ceiling Array', 'Microphone', 907),
  ('Shure NXN8 Vocal Microphone Cardiod',   '/assets/grid/Shure NXN8.png',   'Vocal', 'Microphone', 908),
  ('Shure ANX4 Scalable Wireless Receiver', '/assets/grid/Shure ANX4.png',   'Wireless', 'Receiver',  909),
  ('Shure MXA710 Linear Array Microphone',  '/assets/grid/Shure MXA710.png', 'Linear Array', 'Microphone', 910);

-- Link all dev products into the homepage grid
INSERT INTO homepage_grid_products (product_id, sort_order)
SELECT id, sort_order - 900 FROM products WHERE sort_order BETWEEN 900 AND 910;

-- Featured — 6 items across the three tabs (product_id = any dev product)
INSERT INTO featured (product_id, name, image_url, category, sub_category, featured_categories, sort_order)
SELECT (SELECT id FROM products WHERE sort_order BETWEEN 900 AND 910 ORDER BY id LIMIT 1),
       v.name, v.image_url, v.category, v.sub_category, v.cats, v.so
FROM (VALUES
  ('Shure MXA920-B-R', '/assets/grid/Shure MXA920.png', 'Conferencing Ceiling', 'Array Microphone', ARRAY['New Products','Best Sellers'], 900),
  ('Shure ANX4',       '/assets/grid/Shure ANX4.png',   'Scalable Wireless',    'Receiver',          ARRAY['New Products'],                901),
  ('Shure MXA710',     '/assets/grid/Shure MXA710.png', 'Linear Array',         'Microphone',        ARRAY['New Products','Special Offers'], 902),
  ('Shure NXN8',       '/assets/grid/Shure NXN8.png',   'Vocal',                'Microphone',        ARRAY['New Products','Best Sellers'], 903),
  ('Shure MXA920-B-R', '/assets/grid/Shure MXA920.png', 'Conferencing Ceiling', 'Array Microphone',  ARRAY['Best Sellers'],                904),
  ('Shure ANX4',       '/assets/grid/Shure ANX4.png',   'Scalable Wireless',    'Receiver',          ARRAY['Special Offers'],              905)
) AS v(name, image_url, category, sub_category, cats, so);

-- Articles — 4 items with local header images
INSERT INTO articles (title, description, image_url, sort_order) VALUES
  ('Menyatukan Golf, Hiburan, dan Teknologi AV dengan Shure, Q-SYS, dan LEA Professional', 'Topgolf Jakarta telah menetapkan standar baru dalam dunia golf dengan memadukan teknologi, hiburan, dan kenyamanan kelas dunia.', '/assets/header/Shure-Ceiling-Array.jpg', 900),
  ('Sistem Line Array Martin Audio untuk Venue Besar', 'Bagaimana Martin Audio TORUS memberikan cakupan suara konsisten untuk auditorium dan ruang pertunjukan berkapasitas besar.', '/assets/header/Martin-TORUS.jpg', 901),
  ('Optimal Audio Cuboid: Loudspeaker Serbaguna', 'Review lengkap loudspeaker Optimal Audio Cuboid untuk instalasi komersial — dari kafe hingga ruang ritel.', '/assets/header/Optimal-Audio-Cuboid.jpg', 902),
  ('Inovasi Wireless Shure ADPSM untuk Live Event', 'Teknologi transmitter Shure ADPSM menghadirkan keandalan sinyal wireless untuk panggung dan presentasi profesional.', '/assets/header/Shure-ADPSM.jpg', 903);
