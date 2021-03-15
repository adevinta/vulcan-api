-- Add historical fields to groups table
alter table groups add column deleted_at timestamp with time zone;

-- Add historical fields to recipients table
alter table recipients drop constraint recipients_pkey;
alter table recipients add column deleted_at text;

-- create partial index on recipients table
-- this will enable soft delete
-- based on http://shuber.io/porting-activerecord-soft-delete-behavior-to-postgres/
create unique index unique_recipient ON recipients (team_id, email) WHERE deleted_at IS NULL;

-- Add historical fields to user_team table
alter table user_team add column created_at timestamp with time zone;
alter table user_team add column updated_at timestamp with time zone;
alter table user_team add column deleted_at timestamp with time zone;

-- drop current primary key
alter table user_team drop constraint user_team_pkey;

-- create partial index on user_team table
-- this will enable soft delete
-- based on http://shuber.io/porting-activerecord-soft-delete-behavior-to-postgres/
create unique index unique_user_team ON user_team (user_id, team_id) WHERE deleted_at IS NULL;

-- Add historical fields to asset_group table
alter table asset_group add column deleted_at timestamp with time zone;

-- drop current primary key
alter table asset_group drop constraint asset_group_pkey;

-- create partial index on user_team table
-- this will enable soft delete
-- based on http://shuber.io/porting-activerecord-soft-delete-behavior-to-postgres/
create unique index unique_asset_group ON asset_group (asset_id, group_id) WHERE deleted_at IS NULL;
