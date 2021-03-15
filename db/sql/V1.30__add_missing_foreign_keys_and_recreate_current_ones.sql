-- Add foreign keys to asset_group --
ALTER TABLE asset_group ADD CONSTRAINT fk_assetGroupAssetID FOREIGN KEY (asset_id) REFERENCES assets(id) ON DELETE CASCADE;
ALTER TABLE asset_group ADD CONSTRAINT fk_assetGroupGroupID FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE;

-- Add foreign keys to assets --
ALTER TABLE assets ADD CONSTRAINT fk_assetsTeamID FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE;
ALTER TABLE assets ADD CONSTRAINT fk_assetsAssetTypeID FOREIGN KEY (asset_type_id) REFERENCES asset_types(id) ON DELETE CASCADE;

-- Add foreign keys to checktype_settings --
ALTER TABLE checktype_settings ADD CONSTRAINT fk_checktypeSettingsTeamID FOREIGN KEY (policy_id) REFERENCES policies(id) ON DELETE CASCADE;

-- Add foreign keys to groups --
ALTER TABLE groups ADD CONSTRAINT fk_groupsTeamID FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE;

-- Add foreign keys to policies --
ALTER TABLE policies ADD CONSTRAINT fk_policiesTeamID FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE;

-- Add foreign keys to recipients --
ALTER TABLE recipients ADD CONSTRAINT fk_recipientsTeamID FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE;

-- Add foreign key to user_team --
ALTER TABLE user_team ADD CONSTRAINT fk_userteamUserID FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
ALTER TABLE user_team ADD CONSTRAINT fk_userteamTeamID FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE;

-- Recreate current foreign keys with ON DELETE CASCADE
ALTER TABLE global_programs_metadata DROP CONSTRAINT global_programs_metadata_team_id_fkey;
ALTER TABLE global_programs_metadata ADD  CONSTRAINT global_programs_metadata_team_id_fkey FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE;

ALTER TABLE programs DROP CONSTRAINT programs_team_id_fkey;
ALTER TABLE programs ADD  CONSTRAINT programs_team_id_fkey FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE;

ALTER TABLE programs_groups_policies DROP CONSTRAINT programs_groups_policies_group_id_fkey;
ALTER TABLE programs_groups_policies ADD  CONSTRAINT programs_groups_policies_group_id_fkey FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE;

ALTER TABLE programs_groups_policies DROP CONSTRAINT programs_groups_policies_policy_id_fkey;
ALTER TABLE programs_groups_policies ADD  CONSTRAINT programs_groups_policies_policy_id_fkey FOREIGN KEY (policy_id) REFERENCES policies(id) ON DELETE CASCADE;

ALTER TABLE programs_groups_policies DROP CONSTRAINT programs_groups_policies_program_id_fkey;
ALTER TABLE programs_groups_policies ADD  CONSTRAINT programs_groups_policies_program_id_fkey FOREIGN KEY (program_id) REFERENCES programs(id) ON DELETE CASCADE;
