CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    firstname text,
    lastname text,
    email text,
    api_token text,
    disabled boolean,
    admin boolean,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);

CREATE TABLE teams (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name text,
    description text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);

CREATE TABLE user_team (
    user_id UUID,
    team_id UUID,
    role text,
    is_default boolean,
    PRIMARY KEY(user_id, team_id),
    UNIQUE(user_id, team_id)
);

CREATE TABLE assets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id UUID,
    asset_type_id UUID,
    identifier text,
    options text,
    environmental_cvss text,
    scannable boolean default true,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);

CREATE TABLE asset_types (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name text
);

INSERT INTO asset_types (id, name) VALUES ('d53a9a5a-70ca-4c71-9b0d-808b64dadc40', 'ip');
INSERT INTO asset_types (id, name) VALUES ('d944c4b4-dfd9-4a3f-98cf-279d99d1297b','cidr');
INSERT INTO asset_types (id, name) VALUES ('e2e4b23e-b72c-40a6-9f72-e6ade33a7b00','dns');
INSERT INTO asset_types (id, name) VALUES ('1937b564-bbc4-47f6-9722-b4a8c8ac0595','hostname');
INSERT INTO asset_types (id, name) VALUES ('8ddd51a0-aa7a-46b8-bc28-b9a06c05051d','s3');
INSERT INTO asset_types (id, name) VALUES ('7b25c079-01bb-4df3-bfb1-2f65cef23fc0','git');


CREATE TABLE groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id UUID not null,
    name text not null,
    options text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    UNIQUE (team_id, name)
);

CREATE TABLE asset_group (
    asset_id UUID,
    group_id UUID,
    PRIMARY KEY(asset_id, group_id),
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

CREATE TABLE programs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID,
    policy_id UUID,
    name text,
    cron text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);

CREATE TABLE policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id UUID,
    name text,
    global boolean,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);

CREATE TABLE checktype_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    policy_id UUID,
    check_type_name text,
    options text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);

CREATE TABLE scans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    program_id UUID,
    scheduled_time timestamp with time zone,
    end_time timestamp with time zone,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);