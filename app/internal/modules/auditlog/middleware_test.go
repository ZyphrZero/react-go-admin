package auditlog

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"react-go-admin/app/internal/config"
	"react-go-admin/app/internal/framework"
	"react-go-admin/app/internal/platform/database"
)

func TestGinMiddlewarePersistsLogAfterRequestContextCanceled(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:auditlog-context-cancel?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&database.AuditLog{}); err != nil {
		t.Fatalf("migrate audit_log: %v", err)
	}

	cfg := &config.Config{
		AppEnv:      "test",
		AppTitle:    "Test",
		ProjectName: "Test",
		BaseDir:     t.TempDir(),
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	container := framework.NewContainer(cfg, logger, db)
	if err := framework.BuildModules(container); err != nil {
		t.Fatalf("build modules: %v", err)
	}

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(GinMiddleware(container))

	reqCtx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	engine.GET("/api/v1/base/userapi", func(c *gin.Context) {
		cancel()
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": 401,
			"msg":  "用户不存在",
			"data": nil,
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/base/userapi", nil).WithContext(reqCtx)
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)

	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", resp.Code, http.StatusUnauthorized)
	}

	var record database.AuditLog
	if err := db.Where("path = ?", "/api/v1/base/userapi").Order("id DESC").First(&record).Error; err != nil {
		t.Fatalf("expected audit log to be written: %v", err)
	}
	if record.Status != http.StatusUnauthorized {
		t.Fatalf("audit status = %d, want %d", record.Status, http.StatusUnauthorized)
	}
}
