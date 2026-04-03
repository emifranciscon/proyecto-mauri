-- Upgrade: asientos.fecha DATE → DATETIME(3) for time-of-day support.
-- Safe: existing DATE values become midnight. Backup before running on production.
USE asientos;

ALTER TABLE asientos MODIFY COLUMN fecha DATETIME(3) NOT NULL;
