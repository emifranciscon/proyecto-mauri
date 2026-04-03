-- Extends tanque_historiales for asiento-driven traceability. Run once on existing DBs.
-- Skip if GORM AutoMigrate already applied these columns / index.
USE asientos;

ALTER TABLE tanque_historiales
    ADD COLUMN tipo_operacion VARCHAR(32) NOT NULL DEFAULT '' AFTER descripcion,
    ADD COLUMN cantidad_movimiento DECIMAL(18,6) NOT NULL DEFAULT 0 AFTER tipo_operacion,
    ADD COLUMN saldo_resultante DECIMAL(18,6) NOT NULL DEFAULT 0 AFTER cantidad_movimiento;

-- Index on asiento_id: created by docker-init/001_init on new DBs, or by GORM when schema
-- is created without SQL init. Do not duplicate here (Error 1061).
