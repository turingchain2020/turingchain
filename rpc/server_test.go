// Copyright Turing Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/turingchain2020/turingchain/client/mocks"
	"github.com/turingchain2020/turingchain/common"
	qmocks "github.com/turingchain2020/turingchain/queue/mocks"
	"github.com/turingchain2020/turingchain/rpc/jsonclient"
	rpctypes "github.com/turingchain2020/turingchain/rpc/types"
	"github.com/turingchain2020/turingchain/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func TestCheckIpWhitelist(t *testing.T) {
	address := "127.0.0.1"
	assert.True(t, checkIPWhitelist(address))

	address = "::1"
	assert.True(t, checkIPWhitelist(address))

	address = "192.168.3.1"
	remoteIPWhitelist[address] = true
	assert.False(t, checkIPWhitelist("192.168.3.2"))

	remoteIPWhitelist["0.0.0.0"] = true
	assert.True(t, checkIPWhitelist(address))
	assert.True(t, checkIPWhitelist("192.168.3.2"))

}

func TestCheckBasicAuth(t *testing.T) {
	rpcCfg = new(types.RPC)
	var r = &http.Request{Header: make(http.Header)}
	assert.True(t, checkBasicAuth(r))
	r.SetBasicAuth("1212121", "turingchain-mypasswd")
	assert.True(t, checkBasicAuth(r))
	rpcCfg.JrpcUserName = "turingchain-user"
	rpcCfg.JrpcUserPasswd = "turingchain-mypasswd"
	r.SetBasicAuth("", "turingchain-mypasswd")
	assert.False(t, checkBasicAuth(r))
	r.SetBasicAuth("", "")
	assert.False(t, checkBasicAuth(r))
	r.SetBasicAuth("turingchain-user", "")
	assert.False(t, checkBasicAuth(r))
	r.SetBasicAuth("turingchain", "1234")
	assert.False(t, checkBasicAuth(r))
	r.SetBasicAuth("turingchain-user", "turingchain-mypasswd")
	assert.True(t, checkBasicAuth(r))

}

func TestJSONClient_Call(t *testing.T) {
	rpcCfg = new(types.RPC)
	rpcCfg.GrpcBindAddr = "127.0.0.1:8101"
	rpcCfg.JrpcBindAddr = "127.0.0.1:8200"
	rpcCfg.Whitelist = []string{"127.0.0.1", "0.0.0.0"}
	rpcCfg.JrpcFuncWhitelist = []string{"*"}
	rpcCfg.GrpcFuncWhitelist = []string{"*"}
	InitCfg(rpcCfg)
	api := new(mocks.QueueProtocolAPI)
	cfg := types.NewTuringchainConfig(types.GetDefaultCfgstring())
	api.On("GetConfig", mock.Anything).Return(cfg)
	qm := &qmocks.Client{}
	qm.On("GetConfig", mock.Anything).Return(cfg)
	server := NewJSONRPCServer(qm, api)
	assert.NotNil(t, server)
	done := make(chan struct{}, 1)
	go func() {
		done <- struct{}{}
		server.Listen()
	}()
	<-done
	time.Sleep(time.Millisecond)
	ret := &types.Reply{
		IsOk: true,
		Msg:  []byte("123"),
	}
	api.On("IsSync").Return(ret, nil)
	api.On("Close").Return()

	jsonClient, err := jsonclient.NewJSONClient("http://" + rpcCfg.JrpcBindAddr + "/root")
	assert.Nil(t, err)
	assert.NotNil(t, jsonClient)

	var result = ""
	err = jsonClient.Call("Turingchain.Version", nil, &result)
	assert.NotNil(t, err)
	assert.Empty(t, result)

	jsonClient, err = jsonclient.NewJSONClient("http://" + rpcCfg.JrpcBindAddr)
	assert.Nil(t, err)
	assert.NotNil(t, jsonClient)

	ver := &types.VersionInfo{Turingchain: "6.0.2"}
	api.On("Version").Return(ver, nil)
	var nodeVersion types.VersionInfo
	err = jsonClient.Call("Turingchain.Version", nil, &nodeVersion)
	assert.Nil(t, err)
	assert.Equal(t, "6.0.2", nodeVersion.Turingchain)

	var isSnyc bool
	err = jsonClient.Call("Turingchain.IsSync", &types.ReqNil{}, &isSnyc)
	assert.Nil(t, err)
	assert.Equal(t, ret.GetIsOk(), isSnyc)
	var nodeInfo rpctypes.NodeNetinfo
	api.On("GetNetInfo", mock.Anything).Return(&types.NodeNetInfo{Externaladdr: "123"}, nil)
	err = jsonClient.Call("Turingchain.GetNetInfo", &types.ReqNil{}, &nodeInfo)
	assert.Nil(t, err)
	assert.Equal(t, "123", nodeInfo.Externaladdr)

	var singRet = ""
	api.On("ExecWalletFunc", "wallet", "SignRawTx", mock.Anything).Return(&types.ReplySignRawTx{TxHex: "123"}, nil)
	err = jsonClient.Call("Turingchain.SignRawTx", &types.ReqSignRawTx{}, &singRet)
	assert.Nil(t, err)
	assert.Equal(t, "123", singRet)

	var fee types.TotalFee
	api.On("LocalGet", mock.Anything).Return(nil, errors.New("error value"))
	err = jsonClient.Call("Turingchain.QueryTotalFee", &types.LocalDBGet{Keys: [][]byte{[]byte("test")}}, &fee)
	assert.NotNil(t, err)

	var retNtp bool
	api.On("IsNtpClockSync", mock.Anything).Return(&types.Reply{IsOk: true, Msg: []byte("yes")}, nil)
	err = jsonClient.Call("Turingchain.IsNtpClockSync", &types.ReqNil{}, &retNtp)
	assert.Nil(t, err)
	assert.True(t, retNtp)
	api.On("GetProperFee", mock.Anything).Return(&types.ReplyProperFee{ProperFee: 2}, nil)
	testCreateTxCoins(t, cfg, jsonClient)
	server.Close()
	mock.AssertExpectationsForObjects(t, api)
}

func testDecodeTxHex(t *testing.T, txHex string) *types.Transaction {
	txbytes, err := common.FromHex(txHex)
	assert.Nil(t, err)
	var tx types.Transaction
	err = types.Decode(txbytes, &tx)
	assert.Nil(t, err)
	return &tx
}

func testCreateTxCoins(t *testing.T, cfg *types.TuringchainConfig, jsonClient *jsonclient.JSONClient) {
	req := &rpctypes.CreateTx{
		To:          "184wj4nsgVxKyz2NhM3Yb5RK5Ap6AFRFq2",
		Amount:      10,
		Fee:         1,
		Note:        "12312",
		IsWithdraw:  false,
		IsToken:     false,
		TokenSymbol: "",
		ExecName:    cfg.ExecName("coins"),
	}
	var res string
	err := jsonClient.Call("Turingchain.CreateRawTransaction", req, &res)
	assert.Nil(t, err)
	tx := testDecodeTxHex(t, res)
	assert.Equal(t, "184wj4nsgVxKyz2NhM3Yb5RK5Ap6AFRFq2", tx.To)
	assert.Equal(t, int64(1), tx.Fee)
	req.Fee = 0
	err = jsonClient.Call("Turingchain.CreateRawTransaction", req, &res)
	assert.Nil(t, err)
	tx = testDecodeTxHex(t, res)
	fee, _ := tx.GetRealFee(2)
	assert.Equal(t, fee, tx.Fee)
}

func TestGrpc_Call(t *testing.T) {
	rpcCfg := new(types.RPC)
	rpcCfg.GrpcBindAddr = "127.0.0.1:8101"
	rpcCfg.JrpcBindAddr = "127.0.0.1:8200"
	rpcCfg.Whitelist = []string{"127.0.0.1", "0.0.0.0"}
	rpcCfg.JrpcFuncWhitelist = []string{"*"}
	rpcCfg.GrpcFuncWhitelist = []string{"*"}
	InitCfg(rpcCfg)
	cfg := types.NewTuringchainConfig(types.GetDefaultCfgstring())
	api := new(mocks.QueueProtocolAPI)
	api.On("GetConfig", mock.Anything).Return(cfg)
	_ = NewGrpcServer()
	qm := &qmocks.Client{}
	qm.On("GetConfig", mock.Anything).Return(cfg)
	server := NewGRpcServer(qm, api)
	assert.NotNil(t, server)
	go server.Listen()
	time.Sleep(time.Second)
	ret := &types.Reply{
		IsOk: true,
		Msg:  []byte("123"),
	}
	api.On("IsSync").Return(ret, nil)
	api.On("Close").Return()

	ctx := context.Background()
	c, err := grpc.DialContext(ctx, rpcCfg.GrpcBindAddr, grpc.WithInsecure())
	assert.Nil(t, err)
	assert.NotNil(t, c)

	client := types.NewTuringchainClient(c)
	result, err := client.IsSync(ctx, &types.ReqNil{})

	assert.Nil(t, err)
	assert.Equal(t, ret.IsOk, result.IsOk)
	assert.Equal(t, ret.Msg, result.Msg)

	rst, err := client.GetFork(ctx, &types.ReqKey{Key: []byte("ForkBlockHash")})
	assert.Nil(t, err)
	assert.Equal(t, int64(1), rst.Data)

	api.On("GetBlockBySeq", mock.Anything).Return(&types.BlockSeq{}, nil)
	blockSeq, err := client.GetBlockBySeq(ctx, &types.Int64{Data: 1})
	assert.Nil(t, err)
	assert.Equal(t, &types.BlockSeq{}, blockSeq)

	server.Close()
	mock.AssertExpectationsForObjects(t, api)
}

func TestRPC(t *testing.T) {
	cfg := types.NewTuringchainConfig(types.GetDefaultCfgstring())
	rpcCfg := cfg.GetModuleConfig().RPC
	rpcCfg.JrpcBindAddr = "9671"
	rpcCfg.GrpcBindAddr = "9672"
	rpcCfg.Whitlist = []string{"127.0.0.1"}
	rpcCfg.JrpcFuncBlacklist = []string{"CloseQueue"}
	rpcCfg.GrpcFuncBlacklist = []string{"CloseQueue"}
	rpcCfg.EnableTrace = true
	InitCfg(rpcCfg)
	rpc := New(cfg)
	client := &qmocks.Client{}
	client.On("GetConfig", mock.Anything).Return(cfg)
	rpc.SetQueueClient(client)

	assert.Equal(t, client, rpc.GetQueueClient())
	assert.NotNil(t, rpc.GRPC())
	assert.NotNil(t, rpc.JRPC())
}

func TestCheckFuncList(t *testing.T) {
	funcName := "abc"
	jrpcFuncWhitelist = make(map[string]bool)
	assert.False(t, checkJrpcFuncWhitelist(funcName))
	jrpcFuncWhitelist["*"] = true
	assert.True(t, checkJrpcFuncWhitelist(funcName))

	delete(jrpcFuncWhitelist, "*")
	jrpcFuncWhitelist[funcName] = true
	assert.True(t, checkJrpcFuncWhitelist(funcName))

	grpcFuncWhitelist = make(map[string]bool)
	assert.False(t, checkGrpcFuncWhitelist(funcName))
	grpcFuncWhitelist["*"] = true
	assert.True(t, checkGrpcFuncWhitelist(funcName))

	delete(grpcFuncWhitelist, "*")
	grpcFuncWhitelist[funcName] = true
	assert.True(t, checkGrpcFuncWhitelist(funcName))

	jrpcFuncBlacklist = make(map[string]bool)
	assert.False(t, checkJrpcFuncBlacklist(funcName))
	jrpcFuncBlacklist[funcName] = true
	assert.True(t, checkJrpcFuncBlacklist(funcName))

	grpcFuncBlacklist = make(map[string]bool)
	assert.False(t, checkGrpcFuncBlacklist(funcName))
	grpcFuncBlacklist[funcName] = true
	assert.True(t, checkGrpcFuncBlacklist(funcName))

}
