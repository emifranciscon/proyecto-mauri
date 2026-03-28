package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/proyecto-mauri/backend/internal/models"
	"gorm.io/gorm"
)

type AsientoHandler struct{ DB *gorm.DB }

type asientoBody struct {
	Fecha       string `json:"fecha" binding:"required"` // YYYY-MM-DD
	TanqueID    uint   `json:"tanque_id" binding:"required"`
	BalanzaID   uint   `json:"balanza_id" binding:"required"`
	Descripcion string `json:"descripcion" binding:"required"`
}

func (h *AsientoHandler) List(c *gin.Context) {
	var rows []models.Asiento
	if err := h.DB.Preload("Tanque").Preload("Balanza").Order("fecha desc, id desc").Find(&rows).Error; err != nil {
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
	t, err := time.Parse("2006-01-02", body.Fecha)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "fecha must be YYYY-MM-DD"})
		return
	}
	a := models.Asiento{
		Fecha:       t,
		TanqueID:    body.TanqueID,
		BalanzaID:   body.BalanzaID,
		Descripcion: body.Descripcion,
	}
	if err := h.DB.Create(&a).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "could not create asiento (check tanque/balanza ids)"})
		return
	}
	_ = h.DB.Preload("Tanque").Preload("Balanza").First(&a, a.ID).Error
	c.JSON(http.StatusCreated, a)
}

func (h *AsientoHandler) Get(c *gin.Context) {
	id := c.Param("id")
	var a models.Asiento
	if err := h.DB.Preload("Tanque").Preload("Balanza").First(&a, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, a)
}

func (h *AsientoHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var a models.Asiento
	if err := h.DB.First(&a, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	var body asientoBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	t, err := time.Parse("2006-01-02", body.Fecha)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "fecha must be YYYY-MM-DD"})
		return
	}
	a.Fecha = t
	a.TanqueID = body.TanqueID
	a.BalanzaID = body.BalanzaID
	a.Descripcion = body.Descripcion
	if err := h.DB.Save(&a).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "could not update"})
		return
	}
	_ = h.DB.Preload("Tanque").Preload("Balanza").First(&a, a.ID).Error
	c.JSON(http.StatusOK, a)
}

func (h *AsientoHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.DB.Delete(&models.Asiento{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
