package auditlog

import (
	"github.com/gin-gonic/gin"

	"react-go-admin/app/internal/framework"
	"react-go-admin/app/internal/modules/base"
	"react-go-admin/app/internal/platform/database"
)

const (
	moduleOrderAuditLog = 500
	auditServiceKey     = "auditlog.service"
	auditHandlerKey     = "auditlog.handler"
)

func init() {
	framework.RegisterModule(framework.ModuleSpec{
		Key:   "auditlog",
		Order: moduleOrderAuditLog,
		Bind: func(container *framework.Container) error {
			service := NewService(container.Config, container.DB)
			container.Set(auditServiceKey, service)
			container.Set(auditHandlerKey, NewHandler(service))
			return nil
		},
		Models: func() []interface{} {
			return []interface{}{
				&database.AuditLog{},
			}
		},
		RouteGroups: func(container *framework.Container) []framework.RouteGroup {
			handler := framework.MustResolve[*Handler](container, auditHandlerKey)
			baseService := framework.MustResolve[*base.Service](container, "base.service")

			return []framework.RouteGroup{
				{
					BasePath:    "/api/v1/auditlog",
					Middlewares: []gin.HandlerFunc{base.GinAuthMiddleware(baseService), base.GinPermissionMiddleware(baseService)},
					Routes: []framework.Route{
						{Method: "GET", Path: "/list", Summary: "查看操作日志", Tags: "审计日志", Handlers: []gin.HandlerFunc{gin.WrapF(handler.List)}},
						{Method: "GET", Path: "/detail/:log_id", Summary: "查看操作日志详情", Tags: "审计日志", Handlers: []gin.HandlerFunc{gin.WrapF(handler.Detail)}},
						{Method: "DELETE", Path: "/delete/:log_id", Summary: "删除操作日志", Tags: "审计日志", Handlers: []gin.HandlerFunc{gin.WrapF(handler.Delete)}},
						{Method: "DELETE", Path: "/batch_delete", Summary: "批量删除操作日志", Tags: "审计日志", Handlers: []gin.HandlerFunc{gin.WrapF(handler.BatchDelete)}},
						{Method: "DELETE", Path: "/clear", Summary: "清空操作日志", Tags: "审计日志", Handlers: []gin.HandlerFunc{gin.WrapF(handler.Clear)}},
						{Method: "POST", Path: "/export", Summary: "导出操作日志", Tags: "审计日志", Handlers: []gin.HandlerFunc{gin.WrapF(handler.Export)}},
						{Method: "GET", Path: "/download/:filename", Summary: "下载导出的日志文件", Tags: "审计日志", Handlers: []gin.HandlerFunc{gin.WrapF(handler.Download)}},
						{Method: "GET", Path: "/statistics", Summary: "获取操作日志统计信息", Tags: "审计日志", Handlers: []gin.HandlerFunc{gin.WrapF(handler.Statistics)}},
					},
				},
			}
		},
	})
}
