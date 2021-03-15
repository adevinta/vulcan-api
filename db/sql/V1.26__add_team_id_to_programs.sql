-- Add team_id to programs --
ALTER TABLE programs
    ADD team_id UUID;

-- Fill team_id for existing programs --
UPDATE programs p
SET team_id = (select team_id from policies po where po.id=p.policy_id);

-- Add current policies groups to the programs_groups_policies table --
INSERT INTO programs_groups_policies SELECT id,policy_id,group_id FROM programs;

-- Remove deprecated columns the programs table --
ALTER TABLE programs
DROP COLUMN  policy_id;

ALTER TABLE programs
DROP COLUMN  group_id;

-- Enable constrainsts in the new table column --

ALTER TABLE programs
     ADD FOREIGN KEY (team_id) REFERENCES teams(id);
