CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE
  statuses (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    name VARCHAR(255) NOT NULL UNIQUE
  );

CREATE INDEX IF NOT EXISTS idx_statuses_name ON statuses (name);