package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/proyecto-mauri/backend/internal/middleware"
	"github.com/proyecto-mauri/backend/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthHandler struct {
	DB        *gorm.DB
	JWTSecret []byte
	JWTTTL    time.Duration
}

type loginReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type registerReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var body registerReq
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not hash password"})
		return
	}
	u := models.User{Email: body.Email, PasswordHash: string(hash)}
	if err := h.DB.Create(&u).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
		return
	}
	token, err := middleware.SignToken(h.JWTSecret, u.ID, u.Email, h.JWTTTL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not sign token"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"token": token, "user": gin.H{"id": u.ID, "email": u.Email}})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var body loginReq
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var u models.User
	if err := h.DB.Where("email = ?", body.Email).First(&u).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(body.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	token, err := middleware.SignToken(h.JWTSecret, u.ID, u.Email, h.JWTTTL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not sign token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token, "user": gin.H{"id": u.ID, "email": u.Email}})
}

func (h *AuthHandler) Me(c *gin.Context) {
	uid, _ := c.Get("user_id")
	email, _ := c.Get("email")
	c.JSON(http.StatusOK, gin.H{"id": uid, "email": email})
}
