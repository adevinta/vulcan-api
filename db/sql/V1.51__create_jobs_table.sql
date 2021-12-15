CREATE TABLE jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	team_id UUID,
    operation TEXT NOT NULL,
    status TEXT NOT NULL,
    result TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE,

    CONSTRAINT fk_team_id FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE
);
