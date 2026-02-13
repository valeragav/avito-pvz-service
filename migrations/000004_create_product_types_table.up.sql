CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TABLE product_types (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
  name VARCHAR(255) NOT NULL UNIQUE
);
CREATE INDEX IF NOT EXISTS idx_product_types_name ON product_types (name);
