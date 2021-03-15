CREATE TABLE reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    scan_id text,
    report text,
    report_json text,
    email_body text,
    delivered_to text,
    update_status_at timestamp with time zone,
    status text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);
