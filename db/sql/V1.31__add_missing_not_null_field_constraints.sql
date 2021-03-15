-- Add not null constraints on assets --
ALTER TABLE assets ALTER COLUMN team_id SET NOT NULL;
ALTER TABLE assets ALTER COLUMN asset_type_id SET NOT NULL;

-- Add not null constraints on groups --
ALTER TABLE groups ALTER COLUMN team_id SET NOT NULL;

-- Add not null constraints on checktype_settings --
ALTER TABLE checktype_settings ALTER COLUMN policy_id SET NOT NULL;

-- Add not null constraints on policies --
ALTER TABLE policies ALTER COLUMN team_id SET NOT NULL; 
