-- Upgrade path for DBs created before inventory columns (saldo / cantidad).
-- If GORM AutoMigrate already added these columns, skip the ALTERs and only run the UPDATEs if saldos need reconciling.
-- Running ALTER twice will fail (duplicate column).
USE asientos;

ALTER TABLE tanques
    ADD COLUMN saldo DECIMAL(18,6) NOT NULL DEFAULT 0 AFTER nombre;
ALTER TABLE balanzas
    ADD COLUMN saldo DECIMAL(18,6) NOT NULL DEFAULT 0 AFTER nombre;
ALTER TABLE asientos
    ADD COLUMN cantidad DECIMAL(18,6) NOT NULL DEFAULT 1 AFTER descripcion;

-- Recompute cumulative totals (same rules as application: +cantidad per referenced entity).
UPDATE tanques t
SET saldo = COALESCE(
    (SELECT SUM(a.cantidad)
     FROM asientos a
     WHERE a.tanque_id = t.id AND a.deleted_at IS NULL),
    0
);

UPDATE balanzas b
SET saldo = COALESCE(
    (SELECT SUM(a.cantidad)
     FROM asientos a
     WHERE a.balanza_id = b.id AND a.deleted_at IS NULL),
    0
);
