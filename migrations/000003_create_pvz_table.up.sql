CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE
  pvz (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    registration_date TIMESTAMPTZ NOT NULL,
    city_id UUID NOT NULL,
    CONSTRAINT fk_pvz_city FOREIGN KEY (city_id) REFERENCES cities (id)
  );

CREATE INDEX IF NOT EXISTS idx_pvz_city_id ON pvz (city_id);