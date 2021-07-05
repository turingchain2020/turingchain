// Copyright Turing Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package consensus 共识相关的模块
package consensus

import (
	"reflect"

	"github.com/turingchain2020/turingchain/queue"
	"github.com/turingchain2020/turingchain/system/consensus"
	"github.com/turingchain2020/turingchain/types"
)

// New new consensus queue module
func New(cfg *types.TuringchainConfig) queue.Module {
	mcfg := cfg.GetModuleConfig().Consensus
	sub := cfg.GetSubConfig().Consensus
	con, err := consensus.Load(mcfg.Name)
	if err != nil {
		panic("Unsupported consensus type:" + mcfg.Name + " " + err.Error())
	}
	subcfg, ok := sub[mcfg.Name]
	if !ok {
		subcfg = nil
	}
	obj := con(mcfg, subcfg)
	consensus.QueryData.SetThis(mcfg.Name, reflect.ValueOf(obj))
	return obj
}
