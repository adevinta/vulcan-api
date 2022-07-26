ALTER TABLE finding_overwrites ADD COLUMN team_id UUID NOT NULL;
UPDATE finding_overwrites SET team_id = teams.id FROM teams WHERE teams.tag = finding_overwrites.tag;
ALTER TABLE finding_overwrites DROP COLUMN tag;
