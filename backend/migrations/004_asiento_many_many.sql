-- Upgrade path: single tanque_id/balanza_id/cantidad on asientos → junction tables.
-- Run manually on existing DBs if AutoMigrate cannot reconcile; backup first.
-- Adjust FK constraint names if your schema differs (SHOW CREATE TABLE asientos;).
USE asientos;

CREATE TABLE IF NOT EXISTS asiento_tanques (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    asiento_id BIGINT UNSIGNED NOT NULL,
    tanque_id BIGINT UNSIGNED NOT NULL,
    cantidad DECIMAL(18,6) NOT NULL,
    created_at DATETIME(3) NULL,
    updated_at DATETIME(3) NULL,
    UNIQUE KEY ux_asiento_tanque (asiento_id, tanque_id),
    CONSTRAINT fk_asiento_tanques_asiento FOREIGN KEY (asiento_id) REFERENCES asientos(id) ON DELETE CASCADE,
    CONSTRAINT fk_asiento_tanques_tanque FOREIGN KEY (tanque_id) REFERENCES tanques(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS asiento_balanzas (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    asiento_id BIGINT UNSIGNED NOT NULL,
    balanza_id BIGINT UNSIGNED NOT NULL,
    cantidad DECIMAL(18,6) NOT NULL,
    created_at DATETIME(3) NULL,
    updated_at DATETIME(3) NULL,
    UNIQUE KEY ux_asiento_balanza (asiento_id, balanza_id),
    CONSTRAINT fk_asiento_balanzas_asiento FOREIGN KEY (asiento_id) REFERENCES asientos(id) ON DELETE CASCADE,
    CONSTRAINT fk_asiento_balanzas_balanza FOREIGN KEY (balanza_id) REFERENCES balanzas(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- If legacy columns exist, backfill then drop (uncomment after verifying column names):
-- INSERT IGNORE INTO asiento_tanques (asiento_id, tanque_id, cantidad, created_at, updated_at)
--   SELECT id, tanque_id, cantidad, NOW(), NOW() FROM asientos WHERE deleted_at IS NULL AND tanque_id IS NOT NULL;
-- INSERT IGNORE INTO asiento_balanzas (asiento_id, balanza_id, cantidad, created_at, updated_at)
--   SELECT id, balanza_id, cantidad, NOW(), NOW() FROM asientos WHERE deleted_at IS NULL AND balanza_id IS NOT NULL;
-- ALTER TABLE asientos DROP FOREIGN KEY fk_asientos_tanque;
-- ALTER TABLE asientos DROP FOREIGN KEY fk_asientos_balanza;
-- ALTER TABLE asientos DROP COLUMN tanque_id, DROP COLUMN balanza_id, DROP COLUMN cantidad;

-- VARCHAR allows DEFAULT; TEXT/BLOB cannot have DEFAULT in MySQL (Error 1101).
ALTER TABLE tanque_historiales ADD COLUMN IF NOT EXISTS balanzas_resumen VARCHAR(4096) NOT NULL DEFAULT '' AFTER descripcion;
-- MySQL <8 lacks IF NOT EXISTS for columns; use separate migration tools or run ALTER manually.
