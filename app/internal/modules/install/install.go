package install

import (
	"context"
	"fmt"
	"net/http"

	"react-go-admin/app/internal/framework"
	"react-go-admin/app/internal/migrate"
	"react-go-admin/app/internal/platform/database"
	"react-go-admin/app/internal/platform/response"
	"react-go-admin/app/internal/seed"
)

type Service struct {
	container *framework.Container
}

type Handler struct {
	service *Service
}

func NewService(container *framework.Container) *Service {
	return &Service{container: container}
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (s *Service) CheckDB(ctx context.Context) (map[string]interface{}, error) {
	if s.container == nil || s.container.DB == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	needSchema, err := migrate.NeedsMigration(ctx, s.container.DB)
	if err != nil {
		return nil, err
	}
	needSeed, err := s.needsSeed(ctx)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"need_init":        needSchema || needSeed,
		"need_schema":      needSchema,
		"need_seed":        needSeed,
		"init_api_enabled": s.container.Config.EnableInitAPI,
		"db_connected":     true,
	}, nil
}

func (s *Service) Bootstrap(ctx context.Context) error {
	if s.container == nil || s.container.DB == nil {
		return fmt.Errorf("数据库未初始化")
	}
	if !s.container.Config.EnableInitAPI {
		return fmt.Errorf("初始化接口已禁用")
	}
	if err := migrate.Up(ctx, s.container.DB); err != nil {
		return err
	}
	return seed.Run(ctx, s.container)
}

func (s *Service) needsSeed(ctx context.Context) (bool, error) {
	cfg := s.container.Config
	if !s.container.DB.Migrator().HasTable(&database.User{}) {
		return true, nil
	}
	var user database.User
	result := s.container.DB.WithContext(ctx).Where("username = ?", cfg.InitialAdminUsername).Limit(1).Find(&user)
	if result.Error != nil {
		return false, result.Error
	}
	if result.RowsAffected > 0 {
		return !user.IsSuperuser, nil
	}
	return true, nil
}

func (h *Handler) CheckDB(w http.ResponseWriter, r *http.Request) {
	payload, err := h.service.CheckDB(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	response.Success(w, payload, "成功", nil)
}

func (h *Handler) InitDB(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Bootstrap(r.Context()); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
	response.Success(w, map[string]interface{}{"initialized": true}, "初始化成功", nil)
}
