package framework

import (
	"fmt"
	"log/slog"
	"sync"

	"gorm.io/gorm"

	"react-go-admin/app/internal/config"
)

// Container holds the shared runtime dependencies for scaffold modules.
type Container struct {
	Config *config.Config
	Logger *slog.Logger
	DB     *gorm.DB

	mu     sync.RWMutex
	values map[string]interface{}
}

func NewContainer(cfg *config.Config, logger *slog.Logger, db *gorm.DB) *Container {
	return &Container{
		Config: cfg,
		Logger: logger,
		DB:     db,
		values: map[string]interface{}{},
	}
}

func (c *Container) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.values[key] = value
}

func (c *Container) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, ok := c.values[key]
	return value, ok
}

func Resolve[T any](c *Container, key string) (T, bool) {
	var zero T
	if c == nil {
		return zero, false
	}
	value, ok := c.Get(key)
	if !ok {
		return zero, false
	}
	typed, ok := value.(T)
	if !ok {
		return zero, false
	}
	return typed, true
}

func MustResolve[T any](c *Container, key string) T {
	value, ok := Resolve[T](c, key)
	if !ok {
		panic(fmt.Sprintf("framework: missing dependency %q", key))
	}
	return value
}
