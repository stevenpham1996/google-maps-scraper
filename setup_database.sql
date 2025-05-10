-- setup_database.sql
-- Script to set up the PostgreSQL database for Google Maps scraper

-- Create database if it doesn't exist
-- Note: This needs to be run as a superuser
-- CREATE DATABASE gmapsdb;

-- Connect to the database
\c gmapsdb

-- Create user if it doesn't exist
-- Note: This needs to be run as a superuser
-- CREATE USER gmapsuser WITH PASSWORD 'justbeginagin';
-- GRANT ALL PRIVILEGES ON DATABASE gmapsdb TO gmapsuser;

-- Create tables
CREATE TABLE IF NOT EXISTS gmaps_jobs (
    id UUID PRIMARY KEY,
    priority INTEGER NOT NULL DEFAULT 1,
    payload_type VARCHAR(50) NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    status VARCHAR(20) NOT NULL DEFAULT 'new',
    error TEXT,
    worker_id VARCHAR(100)
);

CREATE TABLE IF NOT EXISTS gmaps_results (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id UUID NOT NULL REFERENCES gmaps_jobs(id) ON DELETE CASCADE,
    title TEXT,
    address TEXT,
    latitude DOUBLE PRECISION,
    longitude DOUBLE PRECISION,
    review_rating DOUBLE PRECISION,
    review_count INTEGER,
    website TEXT,
    link TEXT,
    phone TEXT,
    emails JSONB,
    open_hours JSONB,
    images JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_place_per_job UNIQUE (job_id, title, address)
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_gmaps_jobs_status ON gmaps_jobs(status);
CREATE INDEX IF NOT EXISTS idx_gmaps_jobs_priority ON gmaps_jobs(priority);
CREATE INDEX IF NOT EXISTS idx_gmaps_results_job_id ON gmaps_results(job_id);
CREATE INDEX IF NOT EXISTS idx_gmaps_results_coords ON gmaps_results(latitude, longitude);

-- Grant privileges to the user
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO gmapsuser;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO gmapsuser;

-- Sample data for testing (optional)
-- INSERT INTO gmaps_jobs (id, priority, payload_type, payload, status)
-- VALUES (
--     gen_random_uuid(),
--     1,
--     'search',
--     '{"name": "Test Job", "keywords": ["coworking space in Bangkok"], "lang": "en", "zoom": 15, "depth": 20, "max_time": 3600, "fields": "title,address,latitude,longitude,review_rating,review_count,website,link,phone,emails,open_hours,images"}',
--     'new'
-- );
