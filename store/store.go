// Copyright Turing Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package store store the world - state data
package store

import (
	"github.com/turingchain2020/turingchain/queue"
	"github.com/turingchain2020/turingchain/system/store"
	"github.com/turingchain2020/turingchain/types"
)

// New new store queue module
func New(cfg *types.TuringchainConfig) queue.Module {
	mcfg := cfg.GetModuleConfig().Store
	sub := cfg.GetSubConfig().Store
	s, err := store.Load(mcfg.Name)
	if err != nil {
		panic("Unsupported store type:" + mcfg.Name + " " + err.Error())
	}
	subcfg, ok := sub[mcfg.Name]
	if !ok {
		subcfg = nil
	}
	return s(mcfg, subcfg, cfg)
}
