# Data model

## Entities

| Entity | Purpose |
|--------|---------|
| **User** | Authentication; email + bcrypt password hash. |
| **Tanque** | Master record for tanks; listed in Tanques screen. |
| **Balanza** | Master record for scales; listed in Balanzas screen. |
| **Asiento** | Home “Asientos”: date, tank, scale, description. FKs to `tanques` and `balanzas`. |
| **TanqueHistorial** | Per-tank history: assignment date, optional link to an `asiento`, description. |
| **BalanzaHistorial** | Per-scale history: same shape as tank history. |

## Relationships

- `asientos.tanque_id` → `tanques.id`
- `asientos.balanza_id` → `balanzas.id`
- `tanque_historiales.tanque_id` → `tanques.id`
- `tanque_historiales.asiento_id` → `asientos.id` (nullable, `ON DELETE SET NULL`)
- `balanza_historiales.balanza_id` → `balanzas.id`
- `balanza_historiales.asiento_id` → `asientos.id` (nullable)

## SQL source of truth

Executable DDL lives in `backend/migrations/001_init.sql`. The backend also runs GORM `AutoMigrate` on startup so schema stays aligned in dev and when init scripts are not mounted.
