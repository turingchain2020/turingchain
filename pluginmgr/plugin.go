// Copyright Turing Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pluginmgr

import (
	"github.com/turingchain2020/turingchain/rpc/types"
	typ "github.com/turingchain2020/turingchain/types"
	wcom "github.com/turingchain2020/turingchain/wallet/common"
	"github.com/spf13/cobra"
)

// Plugin plugin module struct
type Plugin interface {
	// 获取整个插件的包名，用以计算唯一值、做前缀等
	GetName() string
	// 获取插件中执行器名
	GetExecutorName() string
	// 初始化执行器时会调用该接口
	InitExec(cfg *typ.TuringchainConfig)
	InitWallet(wallet wcom.WalletOperate, sub map[string][]byte)
	AddCmd(rootCmd *cobra.Command)
	AddRPC(s types.RPCServer)
}
