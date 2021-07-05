package blockchain

import (
	"testing"

	"github.com/turingchain2020/turingchain/queue"
	"github.com/turingchain2020/turingchain/types"
)

func TestCheckClockDrift(t *testing.T) {
	cfg := types.NewTuringchainConfig(types.GetDefaultCfgstring())
	q := queue.New("channel")
	q.SetConfig(cfg)

	blockchain := &BlockChain{}
	blockchain.client = q.Client()
	blockchain.checkClockDrift()

	cfg.GetModuleConfig().NtpHosts = append(cfg.GetModuleConfig().NtpHosts, types.NtpHosts...)
	blockchain.checkClockDrift()
}
