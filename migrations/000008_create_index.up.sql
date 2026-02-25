CREATE INDEX idx_receptions_pvz_date ON receptions(pvz_id, date_time);
CREATE INDEX idx_pvz_registration_date ON pvz(registration_date DESC);