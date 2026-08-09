-- ============================================================================
-- Enterprise AlloyDB AI Extension Initialization Script
-- ============================================================================
-- Enables pgvector, ScaNN indexing, and Vertex AI Google ML integrations.
-- This script is executed per database (e.g. skills_dev, skills_qa, skills_prod).

-- 1. Enable standard PostgreSQL pgvector extension
CREATE EXTENSION IF NOT EXISTS vector;

-- 2. Enable Google AlloyDB ScaNN (Scalable Nearest Neighbors) vector indexing extension
CREATE EXTENSION IF NOT EXISTS alloydb_scann;

-- 3. Enable Google ML integration extension (for in-database embedding inference)
CREATE EXTENSION IF NOT EXISTS google_ml CASCADE;

-- 4. Verify Installed Extensions
SELECT extname, extversion FROM pg_extension WHERE extname IN ('vector', 'alloydb_scann', 'google_ml');
