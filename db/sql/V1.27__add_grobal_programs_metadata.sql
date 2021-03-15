-- Add  global program metedata --

CREATE TABLE global_programs_metadata (
    team_id UUID,
    program text,
    autosend boolean NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    PRIMARY KEY (program,team_id),
    FOREIGN KEY (team_id) REFERENCES teams(id)
)
