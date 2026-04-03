package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/proyecto-mauri/backend/internal/models"
	"github.com/proyecto-mauri/backend/internal/service"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type AsientoHandler struct {
	DB  *gorm.DB
	Svc *service.AsientoService
}

type tanqueLineBody struct {
	TanqueID      uint     `json:"tanque_id" binding:"required"`
	Cantidad      *float64 `json:"cantidad"`
	TipoOperacion string   `json:"tipo_operacion"`
}

type balanzaLineBody struct {
	BalanzaID     uint     `json:"balanza_id" binding:"required"`
	Cantidad      *float64 `json:"cantidad"`
	TipoOperacion string   `json:"tipo_operacion"`
}

type asientoBody struct {
	Fecha       string           `json:"fecha" binding:"required"`
	Descripcion string           `json:"descripcion" binding:"required"`
	Tanques     []tanqueLineBody `json:"tanques" binding:"required"`
	Balanzas    []balanzaLineBody `json:"balanzas" binding:"required"`
}

// parseAsientoFecha accepts RFC3339 (e.g. from JSON) or local "YYYY-MM-DDTHH:mm" / date-only.
func parseAsientoFecha(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, fmt.Errorf("empty fecha")
	}
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return t.Local(), nil
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t.Local(), nil
	}
	layouts := []string{
		"2006-01-02T15:04:05",
		"2006-01-02T15:04",
		"2006-01-02",
	}
	var lastErr error
	for _, layout := range layouts {
		t, err := time.ParseInLocation(layout, s, time.Local)
		if err == nil {
			return t, nil
		}
		lastErr = err
	}
	return time.Time{}, fmt.Errorf("invalid fecha: %w", lastErr)
}

func lineCantidad(v *float64) (decimal.Decimal, error) {
	if v == nil {
		return decimal.NewFromInt(1), nil
	}
	if *v <= 0 {
		return decimal.Zero, service.ErrCantidadInvalid
	}
	return decimal.NewFromFloat(*v), nil
}

var errInvalidTipoOperacion = errors.New("tipo_operacion must be ingreso, egreso, or mantenimiento")

func parseTipoOperacion(s string) (string, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return "ingreso", nil
	}
	if s == "ingreso" || s == "egreso" || s == "mantenimiento" {
		return s, nil
	}
	return "", errInvalidTipoOperacion
}

func bodyToAsientoInput(body *asientoBody) (service.AsientoInput, error) {
	if len(body.Tanques) == 0 {
		return service.AsientoInput{}, service.ErrAsientoSinTanques
	}
	if len(body.Balanzas) == 0 {
		return service.AsientoInput{}, service.ErrAsientoSinBalanzas
	}
	t, err := parseAsientoFecha(body.Fecha)
	if err != nil {
		return service.AsientoInput{}, err
	}
	tanques := make([]service.TanqueCantidad, 0, len(body.Tanques))
	for _, row := range body.Tanques {
		q, err := lineCantidad(row.Cantidad)
		if err != nil {
			return service.AsientoInput{}, err
		}
		tipo, err := parseTipoOperacion(row.TipoOperacion)
		if err != nil {
			return service.AsientoInput{}, err
		}
		tanques = append(tanques, service.TanqueCantidad{TanqueID: row.TanqueID, Cantidad: q, TipoOperacion: tipo})
	}
	balanzas := make([]service.BalanzaCantidad, 0, len(body.Balanzas))
	for _, row := range body.Balanzas {
		q, err := lineCantidad(row.Cantidad)
		if err != nil {
			return service.AsientoInput{}, err
		}
		tipo, err := parseTipoOperacion(row.TipoOperacion)
		if err != nil {
			return service.AsientoInput{}, err
		}
		balanzas = append(balanzas, service.BalanzaCantidad{BalanzaID: row.BalanzaID, Cantidad: q, TipoOperacion: tipo})
	}
	return service.AsientoInput{
		Fecha:       t,
		Descripcion: body.Descripcion,
		Tanques:     tanques,
		Balanzas:    balanzas,
	}, nil
}

func mapAsientoWriteErr(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return true
	case errors.Is(err, service.ErrTanqueNotFound), errors.Is(err, service.ErrBalanzaNotFound):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return true
	case errors.Is(err, service.ErrCantidadInvalid),
		errors.Is(err, service.ErrTanqueSaldoInsuficiente),
		errors.Is(err, service.ErrBalanzaSaldoInsuficiente),
		errors.Is(err, service.ErrAsientoSinTanques),
		errors.Is(err, service.ErrAsientoSinBalanzas):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return true
	default:
		return false
	}
}

func (h *AsientoHandler) List(c *gin.Context) {
	var rows []models.Asiento
	if err := h.DB.Preload("AsientoTanques.Tanque").Preload("AsientoBalanzas.Balanza").
		Order("fecha desc, id desc").Find(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rows)
}

func (h *AsientoHandler) Create(c *gin.Context) {
	var body asientoBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	inp, err := bodyToAsientoInput(&body)
	if err != nil {
		if errors.Is(err, errInvalidTipoOperacion) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, service.ErrCantidadInvalid) || errors.Is(err, service.ErrAsientoSinTanques) || errors.Is(err, service.ErrAsientoSinBalanzas) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "fecha inválida: use fecha ISO o fecha y hora (ej. 2026-04-02T15:30)"})
		return
	}
	a, err := h.Svc.Create(inp)
	if mapAsientoWriteErr(c, err) {
		return
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "could not create asiento"})
		return
	}
	_ = h.DB.Preload("AsientoTanques.Tanque").Preload("AsientoBalanzas.Balanza").First(a, a.ID).Error
	c.JSON(http.StatusCreated, a)
}

func (h *AsientoHandler) Get(c *gin.Context) {
	id := c.Param("id")
	var a models.Asiento
	if err := h.DB.Preload("AsientoTanques.Tanque").Preload("AsientoBalanzas.Balanza").First(&a, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, a)
}

func (h *AsientoHandler) Update(c *gin.Context) {
	id := c.Param("id")
	idU64, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	idU := uint(idU64)
	var body asientoBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	inp, err := bodyToAsientoInput(&body)
	if err != nil {
		if errors.Is(err, errInvalidTipoOperacion) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, service.ErrCantidadInvalid) || errors.Is(err, service.ErrAsientoSinTanques) || errors.Is(err, service.ErrAsientoSinBalanzas) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "fecha inválida: use fecha ISO o fecha y hora (ej. 2026-04-02T15:30)"})
		return
	}
	a, err := h.Svc.Update(idU, inp)
	if mapAsientoWriteErr(c, err) {
		return
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "could not update"})
		return
	}
	_ = h.DB.Preload("AsientoTanques.Tanque").Preload("AsientoBalanzas.Balanza").First(a, a.ID).Error
	c.JSON(http.StatusOK, a)
}

func (h *AsientoHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	idU64, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	err = h.Svc.Delete(uint(idU64))
	if mapAsientoWriteErr(c, err) {
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
