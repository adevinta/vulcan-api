-- Drop deleted_at column from asset_group --
DELETE FROM asset_group WHERE deleted_at IS NOT NULL;
ALTER TABLE asset_group drop column deleted_at;

-- Drop deleted_at column from assets --
DELETE FROM assets WHERE deleted_at IS NOT NULL;
ALTER TABLE assets drop column deleted_at;

-- Drop deleted_at column from checktype_settings --
DELETE FROM checktype_settings WHERE deleted_at IS NOT NULL;
ALTER TABLE checktype_settings drop column deleted_at;

-- Drop deleted_at column from global_programs_metadata --
DELETE FROM global_programs_metadata WHERE deleted_at IS NOT NULL;
ALTER TABLE global_programs_metadata drop column deleted_at;

DELETE FROM programs_groups_policies WHERE program_id in (SELECT id FROM programs WHERE deleted_at IS NOT NULL);
DELETE FROM programs_groups_policies WHERE group_id in (SELECT id FROM groups WHERE deleted_at IS NOT NULL);
DELETE FROM programs_groups_policies WHERE policy_id in (SELECT id FROM policies WHERE deleted_at IS NOT NULL);

-- Drop deleted_at column from groups --
DELETE FROM groups WHERE deleted_at IS NOT NULL;
ALTER TABLE groups drop column deleted_at;

-- Drop deleted_at column from policies --
DELETE FROM policies WHERE deleted_at IS NOT NULL;
ALTER TABLE policies drop column deleted_at;

-- Drop deleted_at column from programs --
DELETE FROM programs WHERE deleted_at IS NOT NULL;
ALTER TABLE programs drop column deleted_at;

-- Drop deleted_at column from recipients --
DELETE FROM recipients WHERE deleted_at IS NOT NULL;
ALTER TABLE recipients drop column deleted_at;

-- Drop deleted_at column from reports --
DELETE FROM reports WHERE deleted_at IS NOT NULL;
ALTER TABLE reports drop column deleted_at;

-- Drop deleted_at column from teams --
DELETE FROM teams WHERE deleted_at IS NOT NULL;
ALTER TABLE teams drop column deleted_at;

-- Drop deleted_at column from user_team --
DELETE FROM user_team WHERE deleted_at IS NOT NULL;
ALTER TABLE user_team drop column deleted_at;

-- Drop deleted_at column from users --
DELETE FROM users WHERE deleted_at IS NOT NULL;
ALTER TABLE users drop column deleted_at;

-- Recreate unique constraint to assets: team_id, identifier, asset_type_id --
CREATE UNIQUE index unique_asset_team ON assets (team_id, identifier, asset_type_id);

-- Recreate unique constraint to recipients: team_id, email --
CREATE UNIQUE index unique_recipient ON recipients (team_id, email);

-- Recreate unique constraint to user_team: user_id, team_id --
CREATE UNIQUE index unique_user_team ON user_team (user_id, team_id);

-- Recreate unique constraint to asset_group: asset_id, group_id --
CREATE UNIQUE index unique_asset_group ON asset_group (asset_id, group_id);

-- Recreate unique constraint to groups: team_id, name --
CREATE UNIQUE index unique_group ON groups (team_id, name);
