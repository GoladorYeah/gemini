CREATE TABLE IF NOT EXISTS search_logs (
    id SERIAL PRIMARY KEY,
    query TEXT NOT NULL,
    normalized_title TEXT,
    category TEXT,
    timestamp TIMESTAMPTZ DEFAULT NOW()
);