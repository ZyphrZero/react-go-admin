package router

import (
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	gincors "github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"react-go-admin/app/internal/config"
	"react-go-admin/app/internal/framework"
	"react-go-admin/app/internal/modules/auditlog"
	"react-go-admin/app/internal/platform/response"
)

func New(container *framework.Container) *gin.Engine {
	cfg := container.Config
	logger := container.Logger

	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gincors.New(buildCORSConfig(cfg)))
	r.Use(requestLogger(logger))
	r.Use(auditlog.GinMiddleware(container))
	r.StaticFS("/static", gin.Dir(filepath.Join(cfg.BaseDir, "storage"), false))
	registerFrontendFallback(r, cfg, logger)

	r.GET("/health", func(c *gin.Context) {
		response.Success(c.Writer, map[string]interface{}{
			"status":      "ok",
			"app":         cfg.ProjectName,
			"version":     cfg.Version,
			"environment": cfg.AppEnv,
		}, "成功", nil)
	})

	for _, groupSpec := range framework.RegisteredRouteGroups(container) {
		group := r.Group(groupSpec.BasePath, groupSpec.Middlewares...)
		for _, route := range groupSpec.Routes {
			handlers := append([]gin.HandlerFunc(nil), route.Handlers...)
			group.Handle(strings.ToUpper(strings.TrimSpace(route.Method)), route.Path, handlers...)
		}
	}

	return r
}

func requestLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		logger.Info("http request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration_ms", time.Since(start).Milliseconds(),
		)
	}
}

func registerFrontendFallback(r *gin.Engine, cfg *config.Config, logger *slog.Logger) {
	distDir := filepath.Join(cfg.BaseDir, "web", "dist")
	if stat, err := os.Stat(distDir); err != nil || !stat.IsDir() {
		if logger != nil {
			logger.Warn("frontend dist not found, skipping static fallback", "path", distDir)
		}
		return
	}

	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api/") || strings.HasPrefix(path, "/static/") || path == "/health" {
			response.Error(c.Writer, http.StatusNotFound, "资源不存在", nil)
			return
		}

		requested := strings.TrimPrefix(path, "/")
		if requested == "" {
			requested = "index.html"
		}

		requested = filepath.Clean(requested)
		if strings.HasPrefix(requested, "..") {
			response.Error(c.Writer, http.StatusNotFound, "资源不存在", nil)
			return
		}

		target := filepath.Join(distDir, requested)
		if stat, err := os.Stat(target); err == nil && !stat.IsDir() {
			c.File(target)
			return
		}

		c.File(filepath.Join(distDir, "index.html"))
	})
}

func buildCORSConfig(cfg *config.Config) gincors.Config {
	corsConfig := gincors.Config{
		AllowMethods:     cfg.CORSAllowMethods,
		AllowHeaders:     cfg.CORSAllowHeaders,
		AllowCredentials: cfg.CORSAllowCredentials,
		MaxAge:           300 * time.Second,
	}
	if len(cfg.CORSOrigins) == 1 && cfg.CORSOrigins[0] == "*" {
		corsConfig.AllowOriginFunc = func(origin string) bool {
			return true
		}
		return corsConfig
	}
	corsConfig.AllowOrigins = cfg.CORSOrigins
	return corsConfig
}
