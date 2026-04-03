-- Drop saldo_resultante from historial; add tipo_operacion on asiento junction lines.
-- Run on existing DBs after backup. Fresh installs: see docker-init/01_schema.sql
USE asientos;

ALTER TABLE tanque_historiales DROP COLUMN saldo_resultante;
ALTER TABLE balanza_historiales DROP COLUMN saldo_resultante;

ALTER TABLE asiento_tanques DROP INDEX ux_asiento_tanque;
ALTER TABLE asiento_tanques ADD COLUMN tipo_operacion VARCHAR(16) NOT NULL DEFAULT 'ingreso' AFTER cantidad;
ALTER TABLE asiento_tanques ADD UNIQUE KEY ux_asiento_tanque_line (asiento_id, tanque_id, tipo_operacion);

ALTER TABLE asiento_balanzas DROP INDEX ux_asiento_balanza;
ALTER TABLE asiento_balanzas ADD COLUMN tipo_operacion VARCHAR(16) NOT NULL DEFAULT 'ingreso' AFTER cantidad;
ALTER TABLE asiento_balanzas ADD UNIQUE KEY ux_asiento_balanza_line (asiento_id, balanza_id, tipo_operacion);
