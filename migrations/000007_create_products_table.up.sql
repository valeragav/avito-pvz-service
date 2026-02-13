CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TABLE products (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
  date_time TIMESTAMPTZ NOT NULL,
  type_id UUID NOT NULL,
  reception_id UUID NOT NULL,
  CONSTRAINT fk_products_type_id FOREIGN KEY (type_id) REFERENCES product_types (id),
  CONSTRAINT fk_products_reception_id FOREIGN KEY (reception_id) REFERENCES receptions (id)
);
CREATE INDEX IF NOT EXISTS idx_products_type_id ON products (type_id);
CREATE INDEX IF NOT EXISTS idx_products_reception_id ON products (reception_id);