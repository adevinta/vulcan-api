ALTER TABLE assets ADD classified_at TIMESTAMP WITH TIME ZONE;

/*
Get asset_id and datetime for the assets which rolfp value 
has been modified at a different time than the update performed 
by migration V1.39.

For each one, update assets table's classified_at field.
*/
WITH last_classifieds AS (
    SELECT (old_val ->> 'id')::uuid AS asset_id, MAX(date) as date
    FROM audit
    WHERE tablename = 'assets' AND operation = 'UPDATE'
        AND old_val ->> 'rolfp' <> new_val ->> 'rolfp'
        AND date <> (
            SELECT installed_on
            FROM flyway_schema_history
            WHERE version = '1.39'
        )
    GROUP BY asset_id
)
UPDATE assets
SET classified_at = last_classifieds.date
FROM last_classifieds
WHERE id = last_classifieds.asset_id;
