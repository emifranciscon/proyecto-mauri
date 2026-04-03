package database

import (
	"fmt"
	"log"
	"time"

	"github.com/proyecto-mauri/backend/internal/config"
	"github.com/proyecto-mauri/backend/internal/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	maxRetries     = 30
	initialBackoff = 2 * time.Second
	maxBackoff     = 10 * time.Second
)

func Connect(cfg *config.Config) (*gorm.DB, error) {
	var db *gorm.DB
	var err error
	backoff := initialBackoff
	for attempt := 1; attempt <= maxRetries; attempt++ {
		db, err = gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Warn),
		})
		if err == nil {
			sqlDB, serr := db.DB()
			if serr != nil {
				return nil, serr
			}
			sqlDB.SetMaxOpenConns(cfg.DBMaxOpen())
			sqlDB.SetMaxIdleConns(cfg.DBMaxIdle())
			sqlDB.SetConnMaxLifetime(time.Hour)
			if err = sqlDB.Ping(); err == nil {
				log.Printf("database: connected after %d attempt(s)", attempt)
				log.Printf("database: running AutoMigrate…")
				if merr := db.AutoMigrate(
					&models.User{},
					&models.Tanque{},
					&models.Balanza{},
					&models.Asiento{},
					&models.AsientoTanque{},
					&models.AsientoBalanza{},
					&models.TanqueHistorial{},
					&models.BalanzaHistorial{},
				); merr != nil {
					return nil, fmt.Errorf("automigrate: %w", merr)
				}
				log.Printf("database: AutoMigrate finished OK")
				return db, nil
			}
			_ = sqlDB.Close()
		}
		log.Printf("database: connect failed (attempt %d/%d): %v; retry in %s", attempt, maxRetries, err, backoff)
		time.Sleep(backoff)
		if backoff < maxBackoff {
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
		}
	}
	return nil, fmt.Errorf("database: gave up after %d attempts: %w", maxRetries, err)
}
