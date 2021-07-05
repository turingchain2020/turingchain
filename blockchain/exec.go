// Copyright Turing Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package blockchain

import (
	"github.com/turingchain2020/turingchain/queue"
	"github.com/turingchain2020/turingchain/types"
	"github.com/turingchain2020/turingchain/util"
)

//执行区块将变成一个私有的函数
func execBlock(client queue.Client, prevStateRoot []byte, block *types.Block, errReturn bool, sync bool) (*types.BlockDetail, []*types.Transaction, error) {
	return util.ExecBlock(client, prevStateRoot, block, errReturn, sync, true)
}

//从本地执行区块
func execBlockUpgrade(client queue.Client, prevStateRoot []byte, block *types.Block, sync bool) error {
	return util.ExecBlockUpgrade(client, prevStateRoot, block, sync)
}
