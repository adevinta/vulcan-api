UPDATE assets SET scannable = TRUE;
ALTER TABLE assets ALTER COLUMN scannable SET NOT NULL;
ALTER TABLE assets ALTER COLUMN scannable SET DEFAULT TRUE;
