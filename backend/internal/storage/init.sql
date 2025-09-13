-- Fix database schema issues
-- Run this SQL to create missing tables and fix column names

-- Create search_logs table if it doesn't exist
CREATE TABLE IF NOT EXISTS search_logs (
    id SERIAL PRIMARY KEY,
    query TEXT NOT NULL,
    normalized_title TEXT,
    category TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create products table with correct schema if it doesn't exist
CREATE TABLE IF NOT EXISTS products (
    id VARCHAR(50) PRIMARY KEY,
    title TEXT NOT NULL,
    url TEXT NOT NULL,
    image_url TEXT,
    image_local TEXT,
    price_gbp VARCHAR(20),
    price_eur DECIMAL(10,2),
    offer_count VARCHAR(10),
    features JSONB,
    category TEXT,
    additional_images JSONB,
    google_product_id TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- If you have existing products table with 'categories' column, rename it
-- ALTER TABLE products RENAME COLUMN categories TO category;

-- Add indexes for better performance
CREATE INDEX IF NOT EXISTS idx_products_category ON products(category);
CREATE INDEX IF NOT EXISTS idx_products_title ON products(title);
CREATE INDEX IF NOT EXISTS idx_search_logs_query ON search_logs(query);
CREATE INDEX IF NOT EXISTS idx_search_logs_category ON search_logs(category);