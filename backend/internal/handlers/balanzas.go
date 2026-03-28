package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/proyecto-mauri/backend/internal/models"
	"gorm.io/gorm"
)

type BalanzaHandler struct{ DB *gorm.DB }

type balanzaBody struct {
	Nombre string `json:"nombre" binding:"required"`
}

type balanzaHistorialBody struct {
	FechaAsignacion string `json:"fecha_asignacion" binding:"required"`
	AsientoID       *uint  `json:"asiento_id"`
	Descripcion     string `json:"descripcion" binding:"required"`
}

func (h *BalanzaHandler) List(c *gin.Context) {
	var rows []models.Balanza
	if err := h.DB.Order("nombre asc").Find(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rows)
}

func (h *BalanzaHandler) Create(c *gin.Context) {
	var body balanzaBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	b := models.Balanza{Nombre: body.Nombre}
	if err := h.DB.Create(&b).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, b)
}

func (h *BalanzaHandler) Get(c *gin.Context) {
	var b models.Balanza
	if err := h.DB.First(&b, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, b)
}

func (h *BalanzaHandler) Update(c *gin.Context) {
	var b models.Balanza
	if err := h.DB.First(&b, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	var body balanzaBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	b.Nombre = body.Nombre
	if err := h.DB.Save(&b).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, b)
}

func (h *BalanzaHandler) Delete(c *gin.Context) {
	if err := h.DB.Delete(&models.Balanza{}, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *BalanzaHandler) ListHistorial(c *gin.Context) {
	balanzaID := c.Param("id")
	var rows []models.BalanzaHistorial
	if err := h.DB.Where("balanza_id = ?", balanzaID).Preload("Asiento").Order("fecha_asignacion desc, id desc").Find(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rows)
}

func (h *BalanzaHandler) CreateHistorial(c *gin.Context) {
	balanzaID := c.Param("id")
	var b models.Balanza
	if err := h.DB.First(&b, balanzaID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "balanza not found"})
		return
	}
	var body balanzaHistorialBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fa, err := time.Parse("2006-01-02", body.FechaAsignacion)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "fecha_asignacion must be YYYY-MM-DD"})
		return
	}
	row := models.BalanzaHistorial{
		BalanzaID:       b.ID,
		FechaAsignacion: fa,
		AsientoID:       body.AsientoID,
		Descripcion:     body.Descripcion,
	}
	if err := h.DB.Create(&row).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "could not create historial"})
		return
	}
	_ = h.DB.Preload("Asiento").First(&row, row.ID).Error
	c.JSON(http.StatusCreated, row)
}
