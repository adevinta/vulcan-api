-- Add the table that contains the tuples (asset_group,policy) for the programs --
CREATE TABLE programs_groups_policies (
    program_id UUID NOT NULL,
    policy_id UUID  NOT NULL,
    group_id UUID NOT NULL,
    PRIMARY KEY (program_id,policy_id,group_id),
    FOREIGN KEY (program_id) REFERENCES programs(id),
    FOREIGN KEY (policy_id) REFERENCES policies(id),
    FOREIGN KEY (group_id) REFERENCES groups(id)
)
