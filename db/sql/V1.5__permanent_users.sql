-- Create Vulcan Service Team --
INSERT INTO teams (
    id,
    name,
    description,
    created_at,
    updated_at,
    deleted_at
) VALUES ('ba2f2a9b-1ea8-4a28-9519-eab4ed290866',
'Vulcan Team', 'Vulcan Service Team', current_date, current_date,null);

-- Crontinuous Service User --
INSERT INTO users(id,
    firstname,
    lastname,
    email,
    api_token,
    disabled,
    admin,
    created_at,
    updated_at,
    deleted_at) VALUES ('4434140a-05fc-4411-8fee-73acdb0c4c95',
'Crontinuous','The Scheduler',
'vulcan@vulcan.com',null,false,true,current_date,current_date,null);

INSERT INTO user_team(
    user_id,
    team_id,
    role,
    is_default
) VALUES (
    '4434140a-05fc-4411-8fee-73acdb0c4c95',
    'ba2f2a9b-1ea8-4a28-9519-eab4ed290866',
    'owner',
    true
    );
