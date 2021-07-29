CREATE TABLE asset_annotations (
    asset_id UUID NOT NULL,
    key TEXT NOT NULL,
    value TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,

    PRIMARY KEY(asset_id, key),
    CONSTRAINT fk_asset_id FOREIGN KEY (asset_id) REFERENCES assets(id) ON DELETE CASCADE
);
