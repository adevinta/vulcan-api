CREATE TABLE finding_overwrites (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL,
    finding_id      UUID NOT NULL,
    status          TEXT NOT NULL,
    status_previous TEXT NOT NULL,
    notes           TEXT NOT NULL,
    tag             TEXT NOT NULL,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT fk_findings_overwrites_user
        FOREIGN KEY(user_id) 
	    REFERENCES users(id)
);
