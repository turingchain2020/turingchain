package mempool

import (
	"github.com/turingchain2020/turingchain/queue"
	"github.com/turingchain2020/turingchain/system/mempool"
	"github.com/turingchain2020/turingchain/types"
)

// New new mempool queue module
func New(cfg *types.TuringchainConfig) queue.Module {
	mcfg := cfg.GetModuleConfig().Mempool
	sub := cfg.GetSubConfig().Mempool
	con, err := mempool.Load(mcfg.Name)
	if err != nil {
		panic("Unsupported mempool type:" + mcfg.Name + " " + err.Error())
	}
	subcfg, ok := sub[mcfg.Name]
	if !ok {
		subcfg = nil
	}
	obj := con(mcfg, subcfg)
	return obj
}
