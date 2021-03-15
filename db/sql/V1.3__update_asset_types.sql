UPDATE asset_types SET name='IP' WHERE name='ip';
UPDATE asset_types SET name='DomainName' WHERE name='dns';
UPDATE asset_types SET name='Hostname' WHERE name='hostname';
DELETE FROM asset_types WHERE name='cidr';
DELETE FROM asset_types WHERE name='s3';
DELETE FROM asset_types WHERE name='git';
