-- AlloyDB / PostgreSQL Schema for Multi-Content Assets
CREATE TABLE IF NOT EXISTS asset_encoding_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    asset_name VARCHAR(255) NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    payload_hash CHAR(64) NOT NULL,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_asset_records_mime ON asset_encoding_records(mime_type);
