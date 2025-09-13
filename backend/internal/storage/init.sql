CREATE TABLE IF NOT EXISTS products (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    category TEXT,
    features TEXT[],
    google_product_id TEXT,
    image_url TEXT
);