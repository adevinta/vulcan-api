-- create partial index on recipients table
-- this will enable soft delete
-- based on http://shuber.io/porting-activerecord-soft-delete-behavior-to-postgres/
alter table groups drop constraint groups_team_id_name_key;
create unique index unique_group ON groups (team_id, name) WHERE deleted_at IS NULL;
