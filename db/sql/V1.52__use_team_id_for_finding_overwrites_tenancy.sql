ALTER TABLE finding_overwrites ADD COLUMN team_id UUID REFERENCES teams (id) ON DELETE CASCADE;
UPDATE finding_overwrites SET team_id = teams.id FROM teams WHERE teams.tag = finding_overwrites.tag;
ALTER TABLE finding_overwrites ALTER COLUMN tag DROP NOT NULL;
