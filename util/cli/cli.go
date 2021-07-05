// Copyright Turing Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cli

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/turingchain2020/turingchain/common/log"
	"github.com/turingchain2020/turingchain/pluginmgr"
	"github.com/turingchain2020/turingchain/rpc/jsonclient"
	rpctypes "github.com/turingchain2020/turingchain/rpc/types"
	"github.com/turingchain2020/turingchain/system/dapp/commands"
	"github.com/turingchain2020/turingchain/types"
	"github.com/spf13/cobra"
)

//Run :
func Run(RPCAddr, ParaName, name string) {
	// cli 命令只打印错误级别到控制台
	log.SetLogLevel("error")
	configPath := ""
	for i, arg := range os.Args[:] {
		if arg == "--conf" && i+1 <= len(os.Args)-1 { // --conf turingchain.toml 可以配置读入cli配置文件路径
			configPath = os.Args[i+1]
			break
		}
		if strings.HasPrefix(arg, "--conf=") { // --conf="turingchain.toml"
			configPath = strings.TrimPrefix(arg, "--conf=")
			break
		}
	}
	if configPath == "" {
		if name == "" {
			configPath = "turingchain.toml"
		} else {
			configPath = name + ".toml"
		}
	}

	exist, _ := pathExists(configPath)
	var turingchainCfg *types.TuringchainConfig
	if exist {
		turingchainCfg = types.NewTuringchainConfig(types.ReadFile(configPath))
	} else {
		cfgstring := types.GetDefaultCfgstring()
		if ParaName != "" {
			cfgstring = strings.Replace(cfgstring, "Title=\"local\"", fmt.Sprintf("Title=\"%s\"", ParaName), 1)
			cfgstring = strings.Replace(cfgstring, "FixTime=false", "CoinSymbol=\"para\"", 1)
		}
		turingchainCfg = types.NewTuringchainConfig(cfgstring)
	}

	types.SetCliSysParam(turingchainCfg.GetTitle(), turingchainCfg)

	rootCmd := &cobra.Command{
		Use:   turingchainCfg.GetTitle() + "-cli",
		Short: turingchainCfg.GetTitle() + " client tools",
	}

	closeCmd := &cobra.Command{
		Use:   "close",
		Short: "Close " + turingchainCfg.GetTitle(),
		Run: func(cmd *cobra.Command, args []string) {
			rpcLaddr, err := cmd.Flags().GetString("rpc_laddr")
			if err != nil {
				panic(err)
			}
			//		rpc, _ := jsonrpc.NewJSONClient(rpcLaddr)
			//		rpc.Call("Turingchain.CloseQueue", nil, nil)
			var res rpctypes.Reply
			ctx := jsonclient.NewRPCCtx(rpcLaddr, "Turingchain.CloseQueue", nil, &res)
			ctx.Run()
		},
	}

	rootCmd.AddCommand(
		commands.CertCmd(),
		commands.AccountCmd(),
		commands.BlockCmd(),
		commands.CoinsCmd(),
		commands.ExecCmd(),
		commands.MempoolCmd(),
		commands.NetCmd(),
		commands.SeedCmd(),
		commands.StatCmd(),
		commands.TxCmd(),
		commands.WalletCmd(),
		commands.VersionCmd(),
		commands.OneStepSendCmd(),
		closeCmd,
		commands.AssetCmd(),
	)

	//test tls is enable
	RPCAddr = testTLS(RPCAddr)
	pluginmgr.AddCmd(rootCmd)
	log.SetLogLevel("error")
	turingchainCfg.S("RPCAddr", RPCAddr)
	turingchainCfg.S("ParaName", ParaName)
	rootCmd.PersistentFlags().String("rpc_laddr", turingchainCfg.GStr("RPCAddr"), "http url")
	rootCmd.PersistentFlags().String("paraName", turingchainCfg.GStr("ParaName"), "parachain")
	rootCmd.PersistentFlags().String("title", turingchainCfg.GetTitle(), "get title name")
	rootCmd.PersistentFlags().MarkHidden("title")
	rootCmd.PersistentFlags().String("conf", "", "cli config")
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func testTLS(RPCAddr string) string {
	rpcaddr := RPCAddr
	if !strings.HasPrefix(rpcaddr, "http://") {
		return RPCAddr
	}
	// if http://
	if rpcaddr[len(rpcaddr)-1] != '/' {
		rpcaddr += "/"
	}
	rpcaddr += "test"
	/* #nosec */
	resp, err := http.Get(rpcaddr)
	if err != nil {
		return "https://" + RPCAddr[7:]
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		return RPCAddr
	}
	return "https://" + RPCAddr[7:]
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
