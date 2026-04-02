package initialize

import (
	"react-go-admin/app/internal/config"
	"react-go-admin/app/internal/framework"
	moduleregistry "react-go-admin/app/internal/modules"
	"react-go-admin/app/internal/platform/database"
	applogger "react-go-admin/app/internal/platform/logger"
)

// Runtime contains the initialized Go backend dependencies.
type Runtime struct {
	*framework.Container
}

// InitRuntime initializes config, logger, database, and module handlers.
func InitRuntime() (*Runtime, error) {
	moduleregistry.Register()

	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	logger, err := applogger.New(cfg)
	if err != nil {
		return nil, err
	}

	db, err := database.Open(cfg, logger)
	if err != nil {
		return nil, err
	}

	container := framework.NewContainer(cfg, logger, db)
	if err := framework.BuildModules(container); err != nil {
		return nil, err
	}

	return &Runtime{
		Container: container,
	}, nil
}

// Close releases the underlying SQL connection.
func (rt *Runtime) Close() error {
	if rt == nil || rt.DB == nil {
		return nil
	}
	sqlDB, err := rt.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
