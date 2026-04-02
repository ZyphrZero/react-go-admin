package seed

import (
	"context"

	"fmt"

	"react-go-admin/app/internal/framework"
)

func Run(ctx context.Context, container *framework.Container) error {
	for _, task := range framework.RegisteredInitTasks(container) {
		if container.Logger != nil {
			container.Logger.Info("running init task", "task", task.Name)
		}
		if err := task.Run(ctx, container); err != nil {
			return fmt.Errorf("init task %s: %w", task.Name, err)
		}
	}
	return nil
}
