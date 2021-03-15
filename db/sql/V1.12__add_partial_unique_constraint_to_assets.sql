-- create partial index on assets table
-- this will enable soft delete
-- based on http://shuber.io/porting-activerecord-soft-delete-behavior-to-postgres/
CREATE UNIQUE index unique_asset_team ON assets (team_id, identifier, asset_type_id) WHERE deleted_at IS NULL;
