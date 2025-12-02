ALTER TABLE artworks
ALTER COLUMN medium TYPE TEXT;

UPDATE artworks
SET medium = CASE
        medium
        WHEN 'oil_panel' THEN 'oil_on_panel'
        WHEN 'acrylic_panel' THEN 'acrylic_on_panel'
        WHEN 'oil_mdf' THEN 'oil_on_mdf'
        WHEN 'oil_paper' THEN 'oil_on_oil_paper'
        ELSE medium
    END;

DROP TYPE artwork_medium;

CREATE TYPE artwork_medium AS ENUM (
    'oil_on_panel',
    'acrylic_on_panel',
    'oil_on_mdf',
    'oil_on_oil_paper',
    'clay_sculpture',
    'plaster_sculpture',
    'ink_on_paper',
    'mixed_media_on_paper',
    'unknown'
);

ALTER TABLE artworks
ALTER COLUMN medium TYPE artwork_medium USING medium::artwork_medium;

ALTER TABLE artworks
ADD COLUMN description TEXT;