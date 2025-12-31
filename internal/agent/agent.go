package agent

import (
	"context"
	"investor/internal/model"
)

type Agent interface {
	Name() string
	Process(ctx context.Context, msg *model.InternalMessage) (string, error)
}
