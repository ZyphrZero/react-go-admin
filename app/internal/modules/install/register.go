package install

import (
	"github.com/gin-gonic/gin"

	"react-go-admin/app/internal/framework"
)

const (
	moduleOrderInstall = 10
	installServiceKey  = "install.service"
	installHandlerKey  = "install.handler"
)

func init() {
	framework.RegisterModule(framework.ModuleSpec{
		Key:   "install",
		Order: moduleOrderInstall,
		Bind: func(container *framework.Container) error {
			service := NewService(container)
			container.Set(installServiceKey, service)
			container.Set(installHandlerKey, NewHandler(service))
			return nil
		},
		RouteGroups: func(container *framework.Container) []framework.RouteGroup {
			handler := framework.MustResolve[*Handler](container, installHandlerKey)
			return []framework.RouteGroup{
				{
					BasePath: "/api/v1/init",
					Routes: []framework.Route{
						{Method: "POST", Path: "/checkdb", Summary: "检测是否需要初始化数据库", Tags: "系统初始化", Handlers: []gin.HandlerFunc{gin.WrapF(handler.CheckDB)}},
						{Method: "POST", Path: "/initdb", Summary: "执行数据库初始化", Tags: "系统初始化", Handlers: []gin.HandlerFunc{gin.WrapF(handler.InitDB)}},
					},
				},
			}
		},
	})
}
