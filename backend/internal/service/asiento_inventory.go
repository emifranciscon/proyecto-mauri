package service

import (
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/proyecto-mauri/backend/internal/models"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrTanqueNotFound           = errors.New("tanque not found")
	ErrBalanzaNotFound          = errors.New("balanza not found")
	ErrCantidadInvalid          = errors.New("cantidad must be positive")
	ErrTanqueSaldoInsuficiente  = errors.New("tank balance would become negative for this egreso")
	ErrBalanzaSaldoInsuficiente = errors.New("scale balance would become negative for this egreso")
	ErrAsientoSinTanques        = errors.New("asiento must include at least one tanque with cantidad")
	ErrAsientoSinBalanzas       = errors.New("asiento must include at least one balanza with cantidad")
)

// TanqueCantidad / BalanzaCantidad are normalized lines (same entity + tipo merged by summing cantidad).
type TanqueCantidad struct {
	TanqueID      uint
	Cantidad      decimal.Decimal
	TipoOperacion string // "ingreso" | "egreso"
}

type BalanzaCantidad struct {
	BalanzaID     uint
	Cantidad      decimal.Decimal
	TipoOperacion string // "ingreso" | "egreso"
}

type tanqueKey struct {
	TanqueID uint
	Tipo     string
}

type balanzaKey struct {
	BalanzaID uint
	Tipo      string
}

func signedMovimiento(cantidad decimal.Decimal, tipo string) decimal.Decimal {
	if strings.EqualFold(strings.TrimSpace(tipo), "egreso") {
		return cantidad.Neg()
	}
	return cantidad
}

// AsientoInput is the domain payload for asiento writes.
type AsientoInput struct {
	Fecha       time.Time
	Descripcion string
	Tanques     []TanqueCantidad
	Balanzas    []BalanzaCantidad
}

type AsientoService struct {
	DB *gorm.DB
}

func NewAsientoService(db *gorm.DB) *AsientoService {
	return &AsientoService{DB: db}
}

func sortedUniqueUint(xs []uint) []uint {
	seen := make(map[uint]struct{}, len(xs))
	for _, x := range xs {
		seen[x] = struct{}{}
	}
	out := make([]uint, 0, len(seen))
	for x := range seen {
		out = append(out, x)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func mergeTanqueLines(lines []TanqueCantidad) ([]TanqueCantidad, error) {
	if len(lines) == 0 {
		return nil, ErrAsientoSinTanques
	}
	m := make(map[tanqueKey]decimal.Decimal)
	for _, l := range lines {
		if l.TanqueID == 0 {
			return nil, errors.New("invalid tanque_id")
		}
		tipo := strings.TrimSpace(strings.ToLower(l.TipoOperacion))
		if tipo != "ingreso" && tipo != "egreso" {
			return nil, errors.New("tanque line tipo_operacion must be ingreso or egreso")
		}
		if !l.Cantidad.GreaterThan(decimal.Zero) {
			return nil, ErrCantidadInvalid
		}
		k := tanqueKey{TanqueID: l.TanqueID, Tipo: tipo}
		m[k] = m[k].Add(l.Cantidad)
	}
	out := make([]TanqueCantidad, 0, len(m))
	for k, q := range m {
		out = append(out, TanqueCantidad{TanqueID: k.TanqueID, Cantidad: q, TipoOperacion: k.Tipo})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].TanqueID != out[j].TanqueID {
			return out[i].TanqueID < out[j].TanqueID
		}
		return out[i].TipoOperacion < out[j].TipoOperacion
	})
	return out, nil
}

func mergeBalanzaLines(lines []BalanzaCantidad) ([]BalanzaCantidad, error) {
	if len(lines) == 0 {
		return nil, ErrAsientoSinBalanzas
	}
	m := make(map[balanzaKey]decimal.Decimal)
	for _, l := range lines {
		if l.BalanzaID == 0 {
			return nil, errors.New("invalid balanza_id")
		}
		tipo := strings.TrimSpace(strings.ToLower(l.TipoOperacion))
		if tipo != "ingreso" && tipo != "egreso" {
			return nil, errors.New("balanza line tipo_operacion must be ingreso or egreso")
		}
		if !l.Cantidad.GreaterThan(decimal.Zero) {
			return nil, ErrCantidadInvalid
		}
		k := balanzaKey{BalanzaID: l.BalanzaID, Tipo: tipo}
		m[k] = m[k].Add(l.Cantidad)
	}
	out := make([]BalanzaCantidad, 0, len(m))
	for k, q := range m {
		out = append(out, BalanzaCantidad{BalanzaID: k.BalanzaID, Cantidad: q, TipoOperacion: k.Tipo})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].BalanzaID != out[j].BalanzaID {
			return out[i].BalanzaID < out[j].BalanzaID
		}
		return out[i].TipoOperacion < out[j].TipoOperacion
	})
	return out, nil
}

func lockTanques(tx *gorm.DB, ids []uint) error {
	for _, id := range sortedUniqueUint(ids) {
		var t models.Tanque
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&t, id).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrTanqueNotFound
			}
			return err
		}
	}
	return nil
}

func lockBalanzas(tx *gorm.DB, ids []uint) error {
	for _, id := range sortedUniqueUint(ids) {
		var b models.Balanza
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&b, id).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrBalanzaNotFound
			}
			return err
		}
	}
	return nil
}

func applyTanqueSaldo(tx *gorm.DB, tanqueID uint, qty decimal.Decimal) (decimal.Decimal, error) {
	var t models.Tanque
	if err := tx.First(&t, tanqueID).Error; err != nil {
		return decimal.Zero, err
	}
	newT := t.Saldo.Add(qty)
	if newT.LessThan(decimal.Zero) {
		return decimal.Zero, ErrTanqueSaldoInsuficiente
	}
	if err := tx.Model(&models.Tanque{}).Where("id = ?", tanqueID).Update("saldo", newT).Error; err != nil {
		return decimal.Zero, err
	}
	return newT, nil
}

func revertTanqueSaldo(tx *gorm.DB, tanqueID uint, appliedQty decimal.Decimal) error {
	var t models.Tanque
	if err := tx.First(&t, tanqueID).Error; err != nil {
		return err
	}
	// Undo the original movement exactly (inverse of apply). Do not block when current
	// saldo is lower than the reverted ingreso (stock may have moved out since).
	newT := t.Saldo.Sub(appliedQty)
	return tx.Model(&models.Tanque{}).Where("id = ?", tanqueID).Update("saldo", newT).Error
}

func applyBalanzaSaldo(tx *gorm.DB, balanzaID uint, qty decimal.Decimal) (decimal.Decimal, error) {
	var b models.Balanza
	if err := tx.First(&b, balanzaID).Error; err != nil {
		return decimal.Zero, err
	}
	newB := b.Saldo.Add(qty)
	if newB.LessThan(decimal.Zero) {
		return decimal.Zero, ErrBalanzaSaldoInsuficiente
	}
	if err := tx.Model(&models.Balanza{}).Where("id = ?", balanzaID).Update("saldo", newB).Error; err != nil {
		return decimal.Zero, err
	}
	return newB, nil
}

func revertBalanzaSaldo(tx *gorm.DB, balanzaID uint, appliedQty decimal.Decimal) error {
	var b models.Balanza
	if err := tx.First(&b, balanzaID).Error; err != nil {
		return err
	}
	newB := b.Saldo.Sub(appliedQty)
	return tx.Model(&models.Balanza{}).Where("id = ?", balanzaID).Update("saldo", newB).Error
}

func balanzasResumen(tx *gorm.DB, lines []BalanzaCantidad) (string, error) {
	parts := make([]string, 0, len(lines))
	for _, l := range lines {
		var b models.Balanza
		if err := tx.First(&b, l.BalanzaID).Error; err != nil {
			return "", err
		}
		parts = append(parts, fmt.Sprintf("%s %s (%s)", b.Nombre, l.TipoOperacion, l.Cantidad.String()))
	}
	return strings.Join(parts, ", "), nil
}

func tanquesResumen(tx *gorm.DB, lines []TanqueCantidad) (string, error) {
	parts := make([]string, 0, len(lines))
	for _, l := range lines {
		var t models.Tanque
		if err := tx.First(&t, l.TanqueID).Error; err != nil {
			return "", err
		}
		parts = append(parts, fmt.Sprintf("%s %s (%s)", t.Nombre, l.TipoOperacion, l.Cantidad.String()))
	}
	return strings.Join(parts, ", "), nil
}

func (s *AsientoService) Create(input AsientoInput) (*models.Asiento, error) {
	tanques, err := mergeTanqueLines(input.Tanques)
	if err != nil {
		return nil, err
	}
	balanzas, err := mergeBalanzaLines(input.Balanzas)
	if err != nil {
		return nil, err
	}

	var out models.Asiento
	err = s.DB.Transaction(func(tx *gorm.DB) error {
		tIDs := make([]uint, len(tanques))
		for i := range tanques {
			tIDs[i] = tanques[i].TanqueID
		}
		bIDs := make([]uint, len(balanzas))
		for i := range balanzas {
			bIDs[i] = balanzas[i].BalanzaID
		}
		if err := lockTanques(tx, tIDs); err != nil {
			return err
		}
		if err := lockBalanzas(tx, bIDs); err != nil {
			return err
		}

		br, err := balanzasResumen(tx, balanzas)
		if err != nil {
			return err
		}
		tr, err := tanquesResumen(tx, tanques)
		if err != nil {
			return err
		}

		out = models.Asiento{
			Fecha:       input.Fecha,
			Descripcion: input.Descripcion,
		}
		if err := tx.Create(&out).Error; err != nil {
			return err
		}
		aid := out.ID

		for _, line := range tanques {
			delta := signedMovimiento(line.Cantidad, line.TipoOperacion)
			if _, err := applyTanqueSaldo(tx, line.TanqueID, delta); err != nil {
				return err
			}
			row := models.AsientoTanque{
				AsientoID:     aid,
				TanqueID:      line.TanqueID,
				Cantidad:      line.Cantidad,
				TipoOperacion: line.TipoOperacion,
			}
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
			th := models.TanqueHistorial{
				TanqueID:           line.TanqueID,
				FechaAsignacion:    input.Fecha,
				AsientoID:          &aid,
				Descripcion:        input.Descripcion,
				BalanzasResumen:    br,
				TipoOperacion:      line.TipoOperacion,
				CantidadMovimiento: line.Cantidad,
			}
			if err := tx.Create(&th).Error; err != nil {
				return err
			}
		}

		for _, line := range balanzas {
			delta := signedMovimiento(line.Cantidad, line.TipoOperacion)
			if _, err := applyBalanzaSaldo(tx, line.BalanzaID, delta); err != nil {
				return err
			}
			row := models.AsientoBalanza{
				AsientoID:     aid,
				BalanzaID:     line.BalanzaID,
				Cantidad:      line.Cantidad,
				TipoOperacion: line.TipoOperacion,
			}
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
			bh := models.BalanzaHistorial{
				BalanzaID:          line.BalanzaID,
				FechaAsignacion:    input.Fecha,
				AsientoID:          &aid,
				Descripcion:        input.Descripcion,
				TanquesResumen:     tr,
				TipoOperacion:      line.TipoOperacion,
				CantidadMovimiento: line.Cantidad,
			}
			if err := tx.Create(&bh).Error; err != nil {
				return err
			}
		}

		log.Printf("asiento service: Create id=%d tanques=%d balanzas=%d", aid, len(tanques), len(balanzas))
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *AsientoService) Update(id uint, input AsientoInput) (*models.Asiento, error) {
	tanques, err := mergeTanqueLines(input.Tanques)
	if err != nil {
		return nil, err
	}
	balanzas, err := mergeBalanzaLines(input.Balanzas)
	if err != nil {
		return nil, err
	}

	var out models.Asiento
	err = s.DB.Transaction(func(tx *gorm.DB) error {
		var cur models.Asiento
		if err := tx.Preload("AsientoTanques").Preload("AsientoBalanzas").First(&cur, id).Error; err != nil {
			return err
		}

		allT := make([]uint, 0)
		allB := make([]uint, 0)
		for _, x := range cur.AsientoTanques {
			allT = append(allT, x.TanqueID)
		}
		for _, x := range tanques {
			allT = append(allT, x.TanqueID)
		}
		for _, x := range cur.AsientoBalanzas {
			allB = append(allB, x.BalanzaID)
		}
		for _, x := range balanzas {
			allB = append(allB, x.BalanzaID)
		}
		if err := lockTanques(tx, allT); err != nil {
			return err
		}
		if err := lockBalanzas(tx, allB); err != nil {
			return err
		}

		for _, row := range cur.AsientoTanques {
			applied := signedMovimiento(row.Cantidad, row.TipoOperacion)
			if err := revertTanqueSaldo(tx, row.TanqueID, applied); err != nil {
				return err
			}
		}
		for _, row := range cur.AsientoBalanzas {
			applied := signedMovimiento(row.Cantidad, row.TipoOperacion)
			if err := revertBalanzaSaldo(tx, row.BalanzaID, applied); err != nil {
				return err
			}
		}
		if err := tx.Where("asiento_id = ?", id).Delete(&models.AsientoTanque{}).Error; err != nil {
			return err
		}
		if err := tx.Where("asiento_id = ?", id).Delete(&models.AsientoBalanza{}).Error; err != nil {
			return err
		}
		if err := tx.Where("asiento_id = ?", id).Delete(&models.TanqueHistorial{}).Error; err != nil {
			return err
		}
		if err := tx.Where("asiento_id = ?", id).Delete(&models.BalanzaHistorial{}).Error; err != nil {
			return err
		}

		br, err := balanzasResumen(tx, balanzas)
		if err != nil {
			return err
		}
		tr, err := tanquesResumen(tx, tanques)
		if err != nil {
			return err
		}

		cur.Fecha = input.Fecha
		cur.Descripcion = input.Descripcion
		if err := tx.Save(&cur).Error; err != nil {
			return err
		}

		for _, line := range tanques {
			delta := signedMovimiento(line.Cantidad, line.TipoOperacion)
			if _, err := applyTanqueSaldo(tx, line.TanqueID, delta); err != nil {
				return err
			}
			if err := tx.Create(&models.AsientoTanque{
				AsientoID:     id,
				TanqueID:      line.TanqueID,
				Cantidad:      line.Cantidad,
				TipoOperacion: line.TipoOperacion,
			}).Error; err != nil {
				return err
			}
			aid := id
			if err := tx.Create(&models.TanqueHistorial{
				TanqueID:           line.TanqueID,
				FechaAsignacion:    input.Fecha,
				AsientoID:          &aid,
				Descripcion:        input.Descripcion,
				BalanzasResumen:    br,
				TipoOperacion:      line.TipoOperacion,
				CantidadMovimiento: line.Cantidad,
			}).Error; err != nil {
				return err
			}
		}
		for _, line := range balanzas {
			delta := signedMovimiento(line.Cantidad, line.TipoOperacion)
			if _, err := applyBalanzaSaldo(tx, line.BalanzaID, delta); err != nil {
				return err
			}
			if err := tx.Create(&models.AsientoBalanza{
				AsientoID:     id,
				BalanzaID:     line.BalanzaID,
				Cantidad:      line.Cantidad,
				TipoOperacion: line.TipoOperacion,
			}).Error; err != nil {
				return err
			}
			aid := id
			if err := tx.Create(&models.BalanzaHistorial{
				BalanzaID:          line.BalanzaID,
				FechaAsignacion:    input.Fecha,
				AsientoID:          &aid,
				Descripcion:        input.Descripcion,
				TanquesResumen:     tr,
				TipoOperacion:      line.TipoOperacion,
				CantidadMovimiento: line.Cantidad,
			}).Error; err != nil {
				return err
			}
		}

		out = cur
		log.Printf("asiento service: Update id=%d", id)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *AsientoService) Delete(id uint) error {
	return s.DB.Transaction(func(tx *gorm.DB) error {
		var cur models.Asiento
		if err := tx.Preload("AsientoTanques").Preload("AsientoBalanzas").First(&cur, id).Error; err != nil {
			return err
		}
		tIDs := make([]uint, len(cur.AsientoTanques))
		for i, row := range cur.AsientoTanques {
			tIDs[i] = row.TanqueID
		}
		bIDs := make([]uint, len(cur.AsientoBalanzas))
		for i, row := range cur.AsientoBalanzas {
			bIDs[i] = row.BalanzaID
		}
		if err := lockTanques(tx, tIDs); err != nil {
			return err
		}
		if err := lockBalanzas(tx, bIDs); err != nil {
			return err
		}
		for _, row := range cur.AsientoTanques {
			applied := signedMovimiento(row.Cantidad, row.TipoOperacion)
			if err := revertTanqueSaldo(tx, row.TanqueID, applied); err != nil {
				return err
			}
		}
		for _, row := range cur.AsientoBalanzas {
			applied := signedMovimiento(row.Cantidad, row.TipoOperacion)
			if err := revertBalanzaSaldo(tx, row.BalanzaID, applied); err != nil {
				return err
			}
		}
		if err := tx.Where("asiento_id = ?", id).Delete(&models.TanqueHistorial{}).Error; err != nil {
			return err
		}
		if err := tx.Where("asiento_id = ?", id).Delete(&models.BalanzaHistorial{}).Error; err != nil {
			return err
		}
		if err := tx.Where("asiento_id = ?", id).Delete(&models.AsientoTanque{}).Error; err != nil {
			return err
		}
		if err := tx.Where("asiento_id = ?", id).Delete(&models.AsientoBalanza{}).Error; err != nil {
			return err
		}
		log.Printf("asiento service: Delete id=%d", id)
		return tx.Delete(&cur).Error
	})
}
