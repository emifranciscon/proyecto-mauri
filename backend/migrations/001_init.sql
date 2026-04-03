-- Schema for asientos app (GORM AutoMigrate adds secondary indexes; see database.go)
CREATE DATABASE IF NOT EXISTS asientos CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE asientos;

CREATE TABLE IF NOT EXISTS users (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at DATETIME(3) NULL,
    updated_at DATETIME(3) NULL,
    deleted_at DATETIME(3) NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS tanques (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    nombre VARCHAR(255) NOT NULL,
    saldo DECIMAL(18,6) NOT NULL DEFAULT 0,
    created_at DATETIME(3) NULL,
    updated_at DATETIME(3) NULL,
    deleted_at DATETIME(3) NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS balanzas (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    nombre VARCHAR(255) NOT NULL,
    saldo DECIMAL(18,6) NOT NULL DEFAULT 0,
    created_at DATETIME(3) NULL,
    updated_at DATETIME(3) NULL,
    deleted_at DATETIME(3) NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS asientos (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    fecha DATETIME(3) NOT NULL,
    descripcion TEXT NOT NULL,
    created_at DATETIME(3) NULL,
    updated_at DATETIME(3) NULL,
    deleted_at DATETIME(3) NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS asiento_tanques (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    asiento_id BIGINT UNSIGNED NOT NULL,
    tanque_id BIGINT UNSIGNED NOT NULL,
    cantidad DECIMAL(18,6) NOT NULL,
    tipo_operacion VARCHAR(16) NOT NULL DEFAULT 'ingreso',
    created_at DATETIME(3) NULL,
    updated_at DATETIME(3) NULL,
    UNIQUE KEY ux_asiento_tanque_line (asiento_id, tanque_id, tipo_operacion),
    CONSTRAINT fk_asiento_tanques_asiento FOREIGN KEY (asiento_id) REFERENCES asientos(id) ON DELETE CASCADE,
    CONSTRAINT fk_asiento_tanques_tanque FOREIGN KEY (tanque_id) REFERENCES tanques(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS asiento_balanzas (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    asiento_id BIGINT UNSIGNED NOT NULL,
    balanza_id BIGINT UNSIGNED NOT NULL,
    cantidad DECIMAL(18,6) NOT NULL,
    tipo_operacion VARCHAR(16) NOT NULL DEFAULT 'ingreso',
    created_at DATETIME(3) NULL,
    updated_at DATETIME(3) NULL,
    UNIQUE KEY ux_asiento_balanza_line (asiento_id, balanza_id, tipo_operacion),
    CONSTRAINT fk_asiento_balanzas_asiento FOREIGN KEY (asiento_id) REFERENCES asientos(id) ON DELETE CASCADE,
    CONSTRAINT fk_asiento_balanzas_balanza FOREIGN KEY (balanza_id) REFERENCES balanzas(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS tanque_historiales (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    tanque_id BIGINT UNSIGNED NOT NULL,
    fecha_asignacion DATE NOT NULL,
    asiento_id BIGINT UNSIGNED NULL,
    descripcion TEXT NOT NULL,
    balanzas_resumen VARCHAR(4096) NOT NULL DEFAULT '',
    tipo_operacion VARCHAR(32) NOT NULL DEFAULT '',
    cantidad_movimiento DECIMAL(18,6) NOT NULL DEFAULT 0,
    created_at DATETIME(3) NULL,
    updated_at DATETIME(3) NULL,
    deleted_at DATETIME(3) NULL,
    CONSTRAINT fk_tanque_historial_tanque FOREIGN KEY (tanque_id) REFERENCES tanques(id),
    CONSTRAINT fk_tanque_historial_asiento FOREIGN KEY (asiento_id) REFERENCES asientos(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS balanza_historiales (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    balanza_id BIGINT UNSIGNED NOT NULL,
    fecha_asignacion DATE NOT NULL,
    asiento_id BIGINT UNSIGNED NULL,
    descripcion TEXT NOT NULL,
    tanques_resumen VARCHAR(4096) NOT NULL DEFAULT '',
    tipo_operacion VARCHAR(32) NOT NULL DEFAULT '',
    cantidad_movimiento DECIMAL(18,6) NOT NULL DEFAULT 0,
    created_at DATETIME(3) NULL,
    updated_at DATETIME(3) NULL,
    deleted_at DATETIME(3) NULL,
    CONSTRAINT fk_balanza_historial_balanza FOREIGN KEY (balanza_id) REFERENCES balanzas(id),
    CONSTRAINT fk_balanza_historial_asiento FOREIGN KEY (asiento_id) REFERENCES asientos(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
