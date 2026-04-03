package models

import (
	"time"

	"github.com/shopspring/decimal"
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
	ID        uint            `gorm:"primaryKey" json:"id"`
	Nombre    string          `gorm:"not null;size:255" json:"nombre"`
	Saldo     decimal.Decimal `gorm:"type:decimal(18,6);not null;default:0" json:"saldo"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	DeletedAt gorm.DeletedAt  `gorm:"index" json:"-"`
}

type Balanza struct {
	ID        uint            `gorm:"primaryKey" json:"id"`
	Nombre    string          `gorm:"not null;size:255" json:"nombre"`
	Saldo     decimal.Decimal `gorm:"type:decimal(18,6);not null;default:0" json:"saldo"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	DeletedAt gorm.DeletedAt  `gorm:"index" json:"-"`
}

// Asiento is the header; tanques and balanzas with cantidades live in AsientoTanque / AsientoBalanza.
type Asiento struct {
	ID          uint            `gorm:"primaryKey" json:"id"`
	Fecha       time.Time       `gorm:"type:datetime(3);not null" json:"fecha"`
	Descripcion string          `gorm:"type:text;not null" json:"descripcion"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	DeletedAt   gorm.DeletedAt  `gorm:"index" json:"-"`

	AsientoTanques  []AsientoTanque  `gorm:"foreignKey:AsientoID" json:"asiento_tanques,omitempty"`
	AsientoBalanzas []AsientoBalanza `gorm:"foreignKey:AsientoID" json:"asiento_balanzas,omitempty"`
}

// AsientoTanque links one asiento to a tanque with cantidad and explicit ingreso/egreso.
type AsientoTanque struct {
	ID             uint            `gorm:"primaryKey" json:"id"`
	AsientoID      uint            `gorm:"not null;uniqueIndex:ux_asiento_tanque_line" json:"asiento_id"`
	TanqueID       uint            `gorm:"not null;uniqueIndex:ux_asiento_tanque_line" json:"tanque_id"`
	Cantidad       decimal.Decimal `gorm:"type:decimal(18,6);not null" json:"cantidad"`
	TipoOperacion  string          `gorm:"size:16;not null;default:ingreso;uniqueIndex:ux_asiento_tanque_line" json:"tipo_operacion"`
	Asiento        Asiento         `gorm:"foreignKey:AsientoID" json:"-"`
	Tanque         Tanque          `gorm:"foreignKey:TanqueID" json:"tanque,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

func (AsientoTanque) TableName() string { return "asiento_tanques" }

// AsientoBalanza links one asiento to a balanza with cantidad and explicit ingreso/egreso.
type AsientoBalanza struct {
	ID             uint            `gorm:"primaryKey" json:"id"`
	AsientoID      uint            `gorm:"not null;uniqueIndex:ux_asiento_balanza_line" json:"asiento_id"`
	BalanzaID      uint            `gorm:"not null;uniqueIndex:ux_asiento_balanza_line" json:"balanza_id"`
	Cantidad       decimal.Decimal `gorm:"type:decimal(18,6);not null" json:"cantidad"`
	TipoOperacion  string          `gorm:"size:16;not null;default:ingreso;uniqueIndex:ux_asiento_balanza_line" json:"tipo_operacion"`
	Asiento        Asiento         `gorm:"foreignKey:AsientoID" json:"-"`
	Balanza        Balanza         `gorm:"foreignKey:BalanzaID" json:"balanza,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

func (AsientoBalanza) TableName() string { return "asiento_balanzas" }

// TanqueHistorial: one row per (asiento, tanque) for automatic asiento-driven traceability.
type TanqueHistorial struct {
	ID                 uint            `gorm:"primaryKey" json:"id"`
	TanqueID           uint            `gorm:"not null;index" json:"tanque_id"`
	FechaAsignacion    time.Time       `gorm:"type:date;not null" json:"fecha_asignacion"`
	AsientoID          *uint           `gorm:"index" json:"asiento_id,omitempty"`
	Descripcion        string          `gorm:"type:text;not null" json:"descripcion"`
	BalanzasResumen    string          `gorm:"size:4096;not null;default:''" json:"balanzas_resumen"`
	TipoOperacion      string          `gorm:"size:32;not null;default:''" json:"tipo_operacion"`
	CantidadMovimiento decimal.Decimal `gorm:"type:decimal(18,6);not null;default:0" json:"cantidad_movimiento"`
	Tanque             Tanque          `gorm:"foreignKey:TanqueID" json:"tanque,omitempty"`
	Asiento            *Asiento        `gorm:"foreignKey:AsientoID" json:"asiento,omitempty"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
	DeletedAt          gorm.DeletedAt  `gorm:"index" json:"-"`
}

func (TanqueHistorial) TableName() string { return "tanque_historiales" }

// BalanzaHistorial: one row per (asiento, balanza) when generated from an asiento.
type BalanzaHistorial struct {
	ID                 uint            `gorm:"primaryKey" json:"id"`
	BalanzaID          uint            `gorm:"not null;index" json:"balanza_id"`
	FechaAsignacion    time.Time       `gorm:"type:date;not null" json:"fecha_asignacion"`
	AsientoID          *uint           `gorm:"index" json:"asiento_id,omitempty"`
	Descripcion        string          `gorm:"type:text;not null" json:"descripcion"`
	TanquesResumen     string          `gorm:"size:4096;not null;default:''" json:"tanques_resumen"`
	TipoOperacion      string          `gorm:"size:32;not null;default:''" json:"tipo_operacion"`
	CantidadMovimiento decimal.Decimal `gorm:"type:decimal(18,6);not null;default:0" json:"cantidad_movimiento"`
	Balanza            Balanza         `gorm:"foreignKey:BalanzaID" json:"balanza,omitempty"`
	Asiento            *Asiento        `gorm:"foreignKey:AsientoID" json:"asiento,omitempty"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
	DeletedAt          gorm.DeletedAt  `gorm:"index" json:"-"`
}

func (BalanzaHistorial) TableName() string { return "balanza_historiales" }
