CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE
  reception_statuses (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    name VARCHAR(255) NOT NULL UNIQUE
  );

CREATE INDEX IF NOT EXISTS idx_reception_statuses_name ON reception_statuses (name);