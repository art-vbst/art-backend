DROP INDEX IF EXISTS idx_images_artwork_id;

DROP INDEX IF EXISTS idx_artworks_status_sort_order;

DROP INDEX IF EXISTS idx_artworks_created_at;

DROP INDEX IF EXISTS idx_artworks_sort_order;

DROP INDEX IF EXISTS idx_artworks_status;

DROP TABLE IF EXISTS images CASCADE;

DROP TABLE IF EXISTS artworks CASCADE;

DROP TYPE IF EXISTS artwork_category;

DROP TYPE IF EXISTS artwork_medium;

DROP TYPE IF EXISTS artwork_status;