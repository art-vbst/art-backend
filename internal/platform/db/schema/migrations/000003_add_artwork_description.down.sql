ALTER TABLE artworks
ALTER COLUMN medium TYPE TEXT;

UPDATE artworks
SET medium = 'unknown'
WHERE medium IN (
        'clay_sculpture',
        'plaster_sculpture',
        'ink_on_paper',
        'mixed_media_on_paper'
    );

UPDATE artworks
SET medium = CASE
        medium
        WHEN 'oil_on_panel' THEN 'oil_panel'
        WHEN 'acrylic_on_panel' THEN 'acrylic_panel'
        WHEN 'oil_on_mdf' THEN 'oil_mdf'
        WHEN 'oil_on_oil_paper' THEN 'oil_paper'
        ELSE medium
    END;

DROP TYPE artwork_medium;

CREATE TYPE artwork_medium AS ENUM (
    'oil_panel',
    'acrylic_panel',
    'oil_mdf',
    'oil_paper',
    'unknown'
);

ALTER TABLE artworks
ALTER COLUMN medium TYPE artwork_medium USING medium::artwork_medium;

ALTER TABLE artworks DROP COLUMN description;