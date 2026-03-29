package router

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/proyecto-mauri/backend/internal/config"
	"github.com/proyecto-mauri/backend/internal/handlers"
	"github.com/proyecto-mauri/backend/internal/middleware"
	"gorm.io/gorm"
)

func New(cfg *config.Config, db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.Use(corsMiddleware(cfg.CORSOrigins))

	r.GET("/healthz", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	jwtSecret := []byte(cfg.JWTSecret)
	ttl := 24 * time.Hour
	auth := &handlers.AuthHandler{DB: db, JWTSecret: jwtSecret, JWTTTL: ttl}
	asientos := &handlers.AsientoHandler{DB: db}
	tanques := &handlers.TanqueHandler{DB: db}
	balanzas := &handlers.BalanzaHandler{DB: db}

	api := r.Group("/api")
	{
		api.POST("/auth/register", auth.Register)
		api.POST("/auth/login", auth.Login)
	}

	protected := r.Group("/api")
	protected.Use(middleware.JWTAuth(jwtSecret))
	{
		protected.GET("/auth/me", auth.Me)

		protected.GET("/asientos", asientos.List)
		protected.POST("/asientos", asientos.Create)
		protected.GET("/asientos/:id", asientos.Get)
		protected.PUT("/asientos/:id", asientos.Update)
		protected.DELETE("/asientos/:id", asientos.Delete)

		protected.GET("/tanques", tanques.List)
		protected.POST("/tanques", tanques.Create)
		protected.GET("/tanques/:id/historial", tanques.ListHistorial)
		protected.POST("/tanques/:id/historial", tanques.CreateHistorial)
		protected.GET("/tanques/:id", tanques.Get)
		protected.PUT("/tanques/:id", tanques.Update)
		protected.DELETE("/tanques/:id", tanques.Delete)

		protected.GET("/balanzas", balanzas.List)
		protected.POST("/balanzas", balanzas.Create)
		protected.GET("/balanzas/:id/historial", balanzas.ListHistorial)
		protected.POST("/balanzas/:id/historial", balanzas.CreateHistorial)
		protected.GET("/balanzas/:id", balanzas.Get)
		protected.PUT("/balanzas/:id", balanzas.Update)
		protected.DELETE("/balanzas/:id", balanzas.Delete)
	}

	return r
}

func corsMiddleware(origins []string) gin.HandlerFunc {
	allow := make(map[string]struct{}, len(origins))
	for _, o := range origins {
		allow[o] = struct{}{}
	}
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if _, ok := allow[origin]; ok {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		c.Header("Access-Control-Allow-Headers", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
