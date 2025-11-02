-- psql "$DATABASE_URL" < internal/platform/db/seed/artworks.sql
-- Insert test artworks
WITH inserted_artworks AS (
  INSERT INTO artworks (
      title,
      painting_number,
      painting_year,
      width_inches,
      height_inches,
      price_cents,
      paper,
      sort_order,
      status,
      medium,
      category
    )
  VALUES (
      'Sunset Over the Mountains',
      101,
      2024,
      24.0,
      18.0,
      125000,
      false,
      1,
      'available',
      'oil_panel',
      'landscape'
    ),
    (
      'Portrait of a Woman',
      102,
      2024,
      16.0,
      20.0,
      95000,
      false,
      2,
      'available',
      'acrylic_panel',
      'figure'
    ),
    (
      'Urban Street Scene',
      103,
      2023,
      12.0,
      16.0,
      75000,
      false,
      3,
      'available',
      'oil_mdf',
      'other'
    ),
    (
      'Beach at Dawn',
      104,
      2023,
      18.0,
      24.0,
      150000,
      false,
      4,
      'pending',
      'oil_panel',
      'landscape'
    ),
    (
      'Family Gathering',
      105,
      2022,
      30.0,
      24.0,
      200000,
      false,
      5,
      'sold',
      'oil_panel',
      'multi_figure'
    ),
    (
      'Abstract Composition',
      106,
      2024,
      9.0,
      12.0,
      45000,
      true,
      6,
      'available',
      'oil_paper',
      'other'
    ),
    (
      'Forest Path',
      107,
      2024,
      20.0,
      16.0,
      110000,
      false,
      7,
      'available',
      'acrylic_panel',
      'landscape'
    ),
    (
      'Reclining Figure',
      108,
      2023,
      14.0,
      11.0,
      65000,
      false,
      8,
      'not_for_sale',
      'oil_panel',
      'figure'
    )
  RETURNING id,
    title
) -- Insert images for each artwork
INSERT INTO images (
    artwork_id,
    is_main_image,
    object_name,
    image_url,
    image_width,
    image_height
  )
SELECT id,
  true,
  CASE
    WHEN title = 'Sunset Over the Mountains' THEN 'seed/artworks/101-sunset-over-the-mountains/main.jpg'
    WHEN title = 'Portrait of a Woman' THEN 'seed/artworks/102-portrait-of-a-woman/main.jpg'
    WHEN title = 'Urban Street Scene' THEN 'seed/artworks/103-urban-street-scene/main.jpg'
    WHEN title = 'Beach at Dawn' THEN 'seed/artworks/104-beach-at-dawn/main.jpg'
    WHEN title = 'Family Gathering' THEN 'seed/artworks/105-family-gathering/main.jpg'
    WHEN title = 'Abstract Composition' THEN 'seed/artworks/106-abstract-composition/main.jpg'
    WHEN title = 'Forest Path' THEN 'seed/artworks/107-forest-path/main.jpg'
    WHEN title = 'Reclining Figure' THEN 'seed/artworks/108-reclining-figure/main.jpg'
  END,
  CASE
    WHEN title = 'Sunset Over the Mountains' THEN 'https://picsum.photos/id/1015/1200/900'
    WHEN title = 'Portrait of a Woman' THEN 'https://picsum.photos/id/64/800/1000'
    WHEN title = 'Urban Street Scene' THEN 'https://picsum.photos/id/146/900/1200'
    WHEN title = 'Beach at Dawn' THEN 'https://picsum.photos/id/1018/1200/1600'
    WHEN title = 'Family Gathering' THEN 'https://picsum.photos/id/91/1500/1200'
    WHEN title = 'Abstract Composition' THEN 'https://picsum.photos/id/136/900/1200'
    WHEN title = 'Forest Path' THEN 'https://picsum.photos/id/1019/1600/1280'
    WHEN title = 'Reclining Figure' THEN 'https://picsum.photos/id/837/1120/880'
  END,
  CASE
    WHEN title = 'Sunset Over the Mountains' THEN 1200
    WHEN title = 'Portrait of a Woman' THEN 800
    WHEN title = 'Urban Street Scene' THEN 900
    WHEN title = 'Beach at Dawn' THEN 1200
    WHEN title = 'Family Gathering' THEN 1500
    WHEN title = 'Abstract Composition' THEN 900
    WHEN title = 'Forest Path' THEN 1600
    WHEN title = 'Reclining Figure' THEN 1120
  END,
  CASE
    WHEN title = 'Sunset Over the Mountains' THEN 900
    WHEN title = 'Portrait of a Woman' THEN 1000
    WHEN title = 'Urban Street Scene' THEN 1200
    WHEN title = 'Beach at Dawn' THEN 1600
    WHEN title = 'Family Gathering' THEN 1200
    WHEN title = 'Abstract Composition' THEN 1200
    WHEN title = 'Forest Path' THEN 1280
    WHEN title = 'Reclining Figure' THEN 880
  END
FROM inserted_artworks
UNION ALL
-- Add secondary images for some artworks
SELECT id,
  false,
  CASE
    WHEN title = 'Sunset Over the Mountains' THEN 'seed/artworks/101-sunset-over-the-mountains/secondary-1.jpg'
    WHEN title = 'Portrait of a Woman' THEN 'seed/artworks/102-portrait-of-a-woman/secondary-1.jpg'
    WHEN title = 'Beach at Dawn' THEN 'seed/artworks/104-beach-at-dawn/secondary-1.jpg'
    WHEN title = 'Forest Path' THEN 'seed/artworks/107-forest-path/secondary-1.jpg'
  END,
  CASE
    WHEN title = 'Sunset Over the Mountains' THEN 'https://picsum.photos/id/1016/1200/900'
    WHEN title = 'Portrait of a Woman' THEN 'https://picsum.photos/id/65/800/1000'
    WHEN title = 'Beach at Dawn' THEN 'https://picsum.photos/id/1024/1200/1600'
    WHEN title = 'Forest Path' THEN 'https://picsum.photos/id/1020/1600/1280'
  END,
  1200,
  900
FROM inserted_artworks
WHERE title IN (
    'Sunset Over the Mountains',
    'Portrait of a Woman',
    'Beach at Dawn',
    'Forest Path'
  );

-- Display summary
SELECT 'Artworks created: ' || COUNT(DISTINCT a.id) as summary
FROM artworks a
WHERE a.created_at > NOW() - INTERVAL '1 minute';

SELECT 'Images created: ' || COUNT(i.id) as summary
FROM images i
WHERE i.created_at > NOW() - INTERVAL '1 minute';

-- Display the created artworks
SELECT a.id,
  a.title,
  a.status,
  a.medium,
  a.category,
  '$' || (a.price_cents / 100.0) as price,
  COUNT(i.id) as image_count
FROM artworks a
  LEFT JOIN images i ON i.artwork_id = a.id
WHERE a.created_at > NOW() - INTERVAL '1 minute'
GROUP BY a.id,
  a.title,
  a.status,
  a.medium,
  a.category,
  a.price_cents
ORDER BY a.sort_order;