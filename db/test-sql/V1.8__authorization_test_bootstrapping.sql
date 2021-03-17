/* Token is stored in _resources/config/testuser.token */
UPDATE users SET api_token='f276623c6f980dcbd73ab434b75ceebbf8295b75b3102b904969cbe87def7e9d' WHERE email='testuser@vulcan.example.com';

/* Token is stored in _resources/config/testuser1.token */
INSERT INTO users (firstname, lastname, email, disabled, admin, api_token) VALUES ('User1', 'AuthTest', 'testuser1@vulcan.example.com', false, false, '91d3d673f008ab6862b863e4f85cccf5e408c341c020b4850d5b723d5f06a9b6');

/* Token is stored in _resources/config/testuser2.token */
INSERT INTO users (firstname, lastname, email, disabled, admin, api_token) VALUES ('User2', 'AuthTest', 'testuser2@vulcan.example.com', false, false, '40624b6171829a1f6fca48db0e9b2b3ff12bea159df0fa125a3363a42f6eeb94');

INSERT INTO teams (name, description) VALUES ('Team1', 'Testing Team for User1');
INSERT INTO teams (name, description) VALUES ('Team2', 'Testing Team for User2');

/* Add User1 in Team1 as owner */
INSERT INTO user_team (user_id, team_id, role, is_default) SELECT u.id as user_id, t.id as team_id, 'owner' as role, false as is_default FROM users u, teams t WHERE u.email = 'testuser1@vulcan.example.com' AND t.name = 'Team1';
/* Add User2 in Team2 as owner */
INSERT INTO user_team (user_id, team_id, role, is_default) SELECT u.id as user_id, t.id as team_id, 'owner' as role, false as is_default FROM users u, teams t WHERE u.email = 'testuser2@vulcan.example.com' AND t.name = 'Team2';

INSERT INTO groups (team_id, name) SELECT t.id as team_id, 'Group1' as name FROM teams t WHERE t.name = 'Team1';
INSERT INTO groups (team_id, name) SELECT t.id as team_id, 'Group2' as name FROM teams t WHERE t.name = 'Team2';

INSERT INTO assets (team_id, asset_type_id, identifier) SELECT t.id as team_id, a.id as asset_type_id, 'Asset1' as identifier FROM teams t, asset_types a WHERE t.name = 'Team1' and a.name = 'IP';
INSERT INTO assets (team_id, asset_type_id, identifier) SELECT t.id as team_id, a.id as asset_type_id, 'Asset2' as identifier FROM teams t, asset_types a WHERE t.name = 'Team2' and a.name = 'IP';

INSERT INTO recipients (team_id, email) SELECT t.id as team_id, 'testuser2@vulcan.example.com' as email FROM teams t WHERE t.name = 'Team2';

INSERT INTO asset_group (asset_id, group_id) SELECT a.id as asset_id, g.id as group_id FROM assets a, groups g WHERE a.identifier = 'Asset2' and g.name='Group2';
