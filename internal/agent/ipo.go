package agent

import (
	"context"
	"investot/internal/model"
)

type IPOAgent struct{}

func NewIPOAgent() *IPOAgent {
	return &IPOAgent{}
}

func (a *IPOAgent) Name() string {
	return "IPOAgent"
}

func (a *IPOAgent) Process(ctx context.Context, msg *model.InternalMessage) (string, error) {
	// TODO: Fetch real data from DataService
	return "【本周新股】\n1. 某某科技 (688xxx): 预计周三申购\n2. 某某医疗 (300xxx): 预计周四上市\n\n(模拟数据)", nil
}
