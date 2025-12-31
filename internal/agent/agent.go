package agent

import (
	"context"
	"investot/internal/model"
)

type Agent interface {
	Name() string
	Process(ctx context.Context, msg *model.InternalMessage) (string, error)
}
