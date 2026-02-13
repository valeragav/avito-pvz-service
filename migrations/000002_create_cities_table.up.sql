CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE
  cities (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    name VARCHAR(255) NOT NULL UNIQUE
  );

CREATE INDEX IF NOT EXISTS idx_cities_name ON cities (name);