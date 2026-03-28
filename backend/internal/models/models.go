package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Email        string         `gorm:"uniqueIndex;not null;size:255" json:"email"`
	PasswordHash string         `gorm:"not null;size:255" json:"-"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

type Tanque struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Nombre    string         `gorm:"not null;size:255" json:"nombre"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type Balanza struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Nombre    string         `gorm:"not null;size:255" json:"nombre"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type Asiento struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Fecha       time.Time      `gorm:"type:date;not null" json:"fecha"`
	TanqueID    uint           `gorm:"not null;index" json:"tanque_id"`
	BalanzaID   uint           `gorm:"not null;index" json:"balanza_id"`
	Descripcion string         `gorm:"type:text;not null" json:"descripcion"`
	Tanque      Tanque         `gorm:"foreignKey:TanqueID" json:"tanque,omitempty"`
	Balanza     Balanza        `gorm:"foreignKey:BalanzaID" json:"balanza,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

type TanqueHistorial struct {
	ID               uint           `gorm:"primaryKey" json:"id"`
	TanqueID         uint           `gorm:"not null;index" json:"tanque_id"`
	FechaAsignacion  time.Time      `gorm:"type:date;not null" json:"fecha_asignacion"`
	AsientoID        *uint          `gorm:"index" json:"asiento_id,omitempty"`
	Descripcion      string         `gorm:"type:text;not null" json:"descripcion"`
	Tanque           Tanque         `gorm:"foreignKey:TanqueID" json:"tanque,omitempty"`
	Asiento          *Asiento       `gorm:"foreignKey:AsientoID" json:"asiento,omitempty"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

func (TanqueHistorial) TableName() string { return "tanque_historiales" }

type BalanzaHistorial struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	BalanzaID       uint           `gorm:"not null;index" json:"balanza_id"`
	FechaAsignacion time.Time      `gorm:"type:date;not null" json:"fecha_asignacion"`
	AsientoID       *uint          `gorm:"index" json:"asiento_id,omitempty"`
	Descripcion     string         `gorm:"type:text;not null" json:"descripcion"`
	Balanza         Balanza        `gorm:"foreignKey:BalanzaID" json:"balanza,omitempty"`
	Asiento         *Asiento       `gorm:"foreignKey:AsientoID" json:"asiento,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

func (BalanzaHistorial) TableName() string { return "balanza_historiales" }
