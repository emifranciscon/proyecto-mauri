package router

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/proyecto-mauri/backend/internal/config"
	"github.com/proyecto-mauri/backend/internal/handlers"
	"github.com/proyecto-mauri/backend/internal/middleware"
	"github.com/proyecto-mauri/backend/internal/service"
	"gorm.io/gorm"
)

func New(cfg *config.Config, db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(corsMiddleware(mergeOrigins(cfg.CORSOrigins)), gin.Logger(), gin.Recovery())

	r.GET("/healthz", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	jwtSecret := []byte(cfg.JWTSecret)
	ttl := 24 * time.Hour
	auth := &handlers.AuthHandler{DB: db, JWTSecret: jwtSecret, JWTTTL: ttl}
	asientoSvc := service.NewAsientoService(db)
	asientos := &handlers.AsientoHandler{DB: db, Svc: asientoSvc}
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

// baselineOrigins are always merged with CORS_ORIGINS so production and common dev URLs stay allowed
// even if the deployment environment variable is missing or misconfigured.
var baselineOrigins = []string{
	"http://localhost:3000",
	"http://localhost:5173",
	"http://127.0.0.1:3000",
	"http://127.0.0.1:5173",
	"https://proyecto-mauri.vercel.app",
}

func normalizeOrigin(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, "/")
	return s
}

func mergeOrigins(fromConfig []string) []string {
	seen := make(map[string]struct{})
	out := make([]string, 0, len(baselineOrigins)+len(fromConfig))
	for _, list := range [][]string{baselineOrigins, fromConfig} {
		for _, o := range list {
			o = normalizeOrigin(o)
			if o == "" {
				continue
			}
			if _, dup := seen[o]; dup {
				continue
			}
			seen[o] = struct{}{}
			out = append(out, o)
		}
	}
	return out
}

// corsResponseWriter re-applies CORS on the first write so Allow-Origin survives error responses and recovery.
type corsResponseWriter struct {
	gin.ResponseWriter
	allowOrigin string
}

func (w *corsResponseWriter) WriteHeader(code int) {
	w.ensureCORS()
	w.ResponseWriter.WriteHeader(code)
}

func (w *corsResponseWriter) Write(b []byte) (int, error) {
	w.ensureCORS()
	return w.ResponseWriter.Write(b)
}

func (w *corsResponseWriter) ensureCORS() {
	if w.allowOrigin == "" {
		return
	}
	h := w.Header()
	if h.Get("Access-Control-Allow-Origin") == "" {
		h.Set("Access-Control-Allow-Origin", w.allowOrigin)
	}
	if h.Get("Access-Control-Allow-Credentials") == "" {
		h.Set("Access-Control-Allow-Credentials", "true")
	}
}

func corsMiddleware(origins []string) gin.HandlerFunc {
	allow := make(map[string]struct{}, len(origins))
	for _, o := range origins {
		allow[o] = struct{}{}
	}
	return func(c *gin.Context) {
		reqOrigin := normalizeOrigin(c.GetHeader("Origin"))
		_, originOK := allow[reqOrigin]
		allowed := originOK && reqOrigin != ""

		if allowed {
			c.Writer = &corsResponseWriter{ResponseWriter: c.Writer, allowOrigin: reqOrigin}
			c.Header("Access-Control-Allow-Origin", reqOrigin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Vary", "Origin")
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
