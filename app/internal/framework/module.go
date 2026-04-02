package framework

import (
	"context"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
)

type Route struct {
	Method   string
	Path     string
	Summary  string
	Tags     string
	Handlers []gin.HandlerFunc
}

type RouteGroup struct {
	BasePath    string
	Middlewares []gin.HandlerFunc
	Routes      []Route
}

type RouteMeta struct {
	Method  string
	Path    string
	Summary string
	Tags    string
}

type InitTask struct {
	Name  string
	Order int
	Run   func(context.Context, *Container) error
}

type ModuleSpec struct {
	Key         string
	Order       int
	Bind        func(*Container) error
	Models      func() []interface{}
	RouteGroups func(*Container) []RouteGroup
	InitTasks   func(*Container) []InitTask
}

var (
	moduleRegistry = map[string]ModuleSpec{}
	moduleKeys     []string
)

func RegisterModule(spec ModuleSpec) {
	key := strings.TrimSpace(spec.Key)
	if key == "" {
		panic("framework: module key is empty")
	}
	if _, exists := moduleRegistry[key]; exists {
		panic("framework: module already registered: " + key)
	}
	spec.Key = key
	moduleRegistry[key] = spec
	moduleKeys = append(moduleKeys, key)
}

func BuildModules(container *Container) error {
	for _, spec := range RegisteredModules() {
		if spec.Bind == nil {
			continue
		}
		if err := spec.Bind(container); err != nil {
			return err
		}
	}
	return nil
}

func RegisteredModules() []ModuleSpec {
	modules := make([]ModuleSpec, 0, len(moduleKeys))
	for _, key := range moduleKeys {
		modules = append(modules, moduleRegistry[key])
	}
	sort.Slice(modules, func(i, j int) bool {
		if modules[i].Order == modules[j].Order {
			return modules[i].Key < modules[j].Key
		}
		return modules[i].Order < modules[j].Order
	})
	return modules
}

func RegisteredModels() []interface{} {
	models := make([]interface{}, 0)
	for _, spec := range RegisteredModules() {
		if spec.Models == nil {
			continue
		}
		models = append(models, spec.Models()...)
	}
	return models
}

func RegisteredRouteGroups(container *Container) []RouteGroup {
	groups := make([]RouteGroup, 0)
	for _, spec := range RegisteredModules() {
		if spec.RouteGroups == nil {
			continue
		}
		groups = append(groups, spec.RouteGroups(container)...)
	}
	return groups
}

func RegisteredInitTasks(container *Container) []InitTask {
	tasks := make([]InitTask, 0)
	for _, spec := range RegisteredModules() {
		if spec.InitTasks == nil {
			continue
		}
		tasks = append(tasks, spec.InitTasks(container)...)
	}
	sort.Slice(tasks, func(i, j int) bool {
		if tasks[i].Order == tasks[j].Order {
			return tasks[i].Name < tasks[j].Name
		}
		return tasks[i].Order < tasks[j].Order
	})
	return tasks
}

func RegisteredRouteMetadata(container *Container) []RouteMeta {
	groups := RegisteredRouteGroups(container)
	seen := map[string]struct{}{}
	metas := make([]RouteMeta, 0)
	for _, group := range groups {
		for _, route := range group.Routes {
			fullPath := normalizeCatalogPath(joinRoutePath(group.BasePath, route.Path))
			key := strings.ToUpper(strings.TrimSpace(route.Method)) + " " + fullPath
			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = struct{}{}
			metas = append(metas, RouteMeta{
				Method:  strings.ToUpper(strings.TrimSpace(route.Method)),
				Path:    fullPath,
				Summary: strings.TrimSpace(route.Summary),
				Tags:    strings.TrimSpace(route.Tags),
			})
		}
	}
	sort.Slice(metas, func(i, j int) bool {
		if metas[i].Tags == metas[j].Tags {
			if metas[i].Path == metas[j].Path {
				return metas[i].Method < metas[j].Method
			}
			return metas[i].Path < metas[j].Path
		}
		return metas[i].Tags < metas[j].Tags
	})
	return metas
}

func joinRoutePath(basePath string, routePath string) string {
	base := strings.TrimRight(strings.TrimSpace(basePath), "/")
	path := strings.TrimSpace(routePath)
	if path == "" || path == "/" {
		if base == "" {
			return "/"
		}
		return base
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if base == "" {
		return path
	}
	return base + path
}

func normalizeCatalogPath(path string) string {
	if path == "" {
		return "/"
	}
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if strings.HasPrefix(part, ":") && len(part) > 1 {
			parts[i] = "{" + strings.TrimPrefix(part, ":") + "}"
		}
	}
	normalized := strings.Join(parts, "/")
	if normalized == "" {
		return "/"
	}
	return normalized
}
