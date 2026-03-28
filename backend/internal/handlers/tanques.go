package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/proyecto-mauri/backend/internal/models"
	"gorm.io/gorm"
)

type TanqueHandler struct{ DB *gorm.DB }

type tanqueBody struct {
	Nombre string `json:"nombre" binding:"required"`
}

type tanqueHistorialBody struct {
	FechaAsignacion string `json:"fecha_asignacion" binding:"required"`
	AsientoID       *uint  `json:"asiento_id"`
	Descripcion     string `json:"descripcion" binding:"required"`
}

func (h *TanqueHandler) List(c *gin.Context) {
	var rows []models.Tanque
	if err := h.DB.Order("nombre asc").Find(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rows)
}

func (h *TanqueHandler) Create(c *gin.Context) {
	var body tanqueBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	t := models.Tanque{Nombre: body.Nombre}
	if err := h.DB.Create(&t).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, t)
}

func (h *TanqueHandler) Get(c *gin.Context) {
	var t models.Tanque
	if err := h.DB.First(&t, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, t)
}

func (h *TanqueHandler) Update(c *gin.Context) {
	var t models.Tanque
	if err := h.DB.First(&t, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	var body tanqueBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	t.Nombre = body.Nombre
	if err := h.DB.Save(&t).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, t)
}

func (h *TanqueHandler) Delete(c *gin.Context) {
	if err := h.DB.Delete(&models.Tanque{}, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *TanqueHandler) ListHistorial(c *gin.Context) {
	tanqueID := c.Param("id")
	var rows []models.TanqueHistorial
	if err := h.DB.Where("tanque_id = ?", tanqueID).Preload("Asiento").Order("fecha_asignacion desc, id desc").Find(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rows)
}

func (h *TanqueHandler) CreateHistorial(c *gin.Context) {
	tanqueID := c.Param("id")
	var t models.Tanque
	if err := h.DB.First(&t, tanqueID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "tanque not found"})
		return
	}
	var body tanqueHistorialBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fa, err := time.Parse("2006-01-02", body.FechaAsignacion)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "fecha_asignacion must be YYYY-MM-DD"})
		return
	}
	row := models.TanqueHistorial{
		TanqueID:        t.ID,
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
