CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TABLE receptions (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
  date_time TIMESTAMPTZ NOT NULL,
  pvz_id UUID NOT NULL,
  status_id UUID NOT NULL,
  CONSTRAINT fk_receptions_pvz_id FOREIGN KEY (pvz_id) REFERENCES pvz (id),
  CONSTRAINT fk_receptions_status_id FOREIGN KEY (status_id) REFERENCES reception_statuses (id)
);
CREATE INDEX IF NOT EXISTS idx_receptions_pvz_id ON receptions (pvz_id);
CREATE INDEX IF NOT EXISTS idx_receptions_status_id ON receptions (status_id);