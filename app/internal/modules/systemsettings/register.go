package systemsettings

import (
	"github.com/gin-gonic/gin"

	"react-go-admin/app/internal/framework"
	"react-go-admin/app/internal/modules/base"
	"react-go-admin/app/internal/platform/database"
)

const (
	moduleOrderSystemSettings = 600
	settingsServiceKey        = "systemsettings.service"
	settingsHandlerKey        = "systemsettings.handler"
)

func init() {
	framework.RegisterModule(framework.ModuleSpec{
		Key:   "systemsettings",
		Order: moduleOrderSystemSettings,
		Bind: func(container *framework.Container) error {
			service := NewService(container.Config, container.DB)
			container.Set(settingsServiceKey, service)
			container.Set(settingsHandlerKey, NewHandler(service))
			return nil
		},
		Models: func() []interface{} {
			return []interface{}{
				&database.SystemSetting{},
			}
		},
		RouteGroups: func(container *framework.Container) []framework.RouteGroup {
			handler := framework.MustResolve[*Handler](container, settingsHandlerKey)
			baseService := framework.MustResolve[*base.Service](container, "base.service")

			return []framework.RouteGroup{
				{
					BasePath:    "/api/v1/system_settings",
					Middlewares: []gin.HandlerFunc{base.GinAuthMiddleware(baseService)},
					Routes: []framework.Route{
						{Method: "GET", Path: "/application", Summary: "获取应用设置", Tags: "系统设置", Handlers: []gin.HandlerFunc{gin.WrapF(handler.GetApplication)}},
						{Method: "POST", Path: "/application", Summary: "更新应用设置", Tags: "系统设置", Handlers: []gin.HandlerFunc{gin.WrapF(handler.UpdateApplication)}},
						{Method: "GET", Path: "/logging", Summary: "获取日志设置", Tags: "系统设置", Handlers: []gin.HandlerFunc{gin.WrapF(handler.GetLogging)}},
						{Method: "POST", Path: "/logging", Summary: "更新日志设置", Tags: "系统设置", Handlers: []gin.HandlerFunc{gin.WrapF(handler.UpdateLogging)}},
						{Method: "GET", Path: "/security", Summary: "获取安全设置", Tags: "系统设置", Handlers: []gin.HandlerFunc{gin.WrapF(handler.GetSecurity)}},
						{Method: "POST", Path: "/security", Summary: "更新安全设置", Tags: "系统设置", Handlers: []gin.HandlerFunc{gin.WrapF(handler.UpdateSecurity)}},
						{Method: "GET", Path: "/storage", Summary: "获取存储设置", Tags: "系统设置", Handlers: []gin.HandlerFunc{gin.WrapF(handler.GetStorage)}},
						{Method: "POST", Path: "/storage", Summary: "更新存储设置", Tags: "系统设置", Handlers: []gin.HandlerFunc{gin.WrapF(handler.UpdateStorage)}},
					},
				},
			}
		},
	})
}
