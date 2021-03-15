-- Add column risk to reports --
ALTER TABLE users rename column disabled to active;
UPDATE users set active = NOT active;
