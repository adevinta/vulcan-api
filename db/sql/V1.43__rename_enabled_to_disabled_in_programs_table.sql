ALTER TABLE programs ADD COLUMN disabled boolean NOT NULL DEFAULT(false);
UPDATE programs set disabled=true WHERE enabled=false;
ALTER TABLE programs DROP COLUMN enabled;
