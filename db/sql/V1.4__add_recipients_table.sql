CREATE TABLE recipients (
    team_id UUID,
    email text,
    PRIMARY KEY(team_id, email),
    UNIQUE(team_id, email),
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);
