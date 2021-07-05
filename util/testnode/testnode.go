// Copyright Turing Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// package testnode 提供一个通用的测试节点，用于单元测试和集成测试。

package testnode

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/turingchain2020/turingchain/p2p"

	"github.com/turingchain2020/turingchain/account"
	"github.com/turingchain2020/turingchain/blockchain"
	"github.com/turingchain2020/turingchain/client"
	"github.com/turingchain2020/turingchain/common"
	"github.com/turingchain2020/turingchain/common/address"
	"github.com/turingchain2020/turingchain/common/crypto"
	"github.com/turingchain2020/turingchain/common/limits"
	"github.com/turingchain2020/turingchain/common/log"
	"github.com/turingchain2020/turingchain/common/log/log15"
	"github.com/turingchain2020/turingchain/consensus"
	"github.com/turingchain2020/turingchain/executor"
	"github.com/turingchain2020/turingchain/mempool"
	"github.com/turingchain2020/turingchain/pluginmgr"
	"github.com/turingchain2020/turingchain/queue"
	"github.com/turingchain2020/turingchain/rpc"
	"github.com/turingchain2020/turingchain/rpc/jsonclient"
	rpctypes "github.com/turingchain2020/turingchain/rpc/types"
	"github.com/turingchain2020/turingchain/store"
	"github.com/turingchain2020/turingchain/types"
	"github.com/turingchain2020/turingchain/util"
	"github.com/turingchain2020/turingchain/wallet"
)

func init() {
	err := limits.SetLimits()
	if err != nil {
		panic(err)
	}
	log.SetLogLevel("info")
}

//保证只有一个turingchain 会运行
var lognode = log15.New("module", "lognode")

//TuringchainMock :
type TuringchainMock struct {
	random   *rand.Rand
	q        queue.Queue
	client   queue.Client
	api      client.QueueProtocolAPI
	chain    *blockchain.BlockChain
	mem      queue.Module
	cs       queue.Module
	exec     *executor.Executor
	wallet   queue.Module
	network  queue.Module
	store    queue.Module
	rpc      *rpc.RPC
	cfg      *types.Config
	sub      *types.ConfigSubModule
	datadir  string
	lastsend []byte
	mu       sync.Mutex
}

//GetDefaultConfig :
func GetDefaultConfig() *types.TuringchainConfig {
	return types.NewTuringchainConfig(types.GetDefaultCfgstring())
}

//NewWithConfig :
func NewWithConfig(cfg *types.TuringchainConfig, mockapi client.QueueProtocolAPI) *TuringchainMock {
	return newWithConfig(cfg, mockapi)
}

// NewWithRPC 创建测试节点 并开放rpc服务
func NewWithRPC(cfg *types.TuringchainConfig, mockapi client.QueueProtocolAPI) *TuringchainMock {
	mock := newWithConfigNoLock(cfg, mockapi)
	mock.rpc.Close()
	server := rpc.New(cfg)
	server.SetAPI(mock.api)
	server.SetQueueClient(mock.q.Client())
	mock.rpc = server
	return mock
}

func newWithConfig(cfg *types.TuringchainConfig, mockapi client.QueueProtocolAPI) *TuringchainMock {
	return newWithConfigNoLock(cfg, mockapi)
}

func newWithConfigNoLock(cfg *types.TuringchainConfig, mockapi client.QueueProtocolAPI) *TuringchainMock {
	mfg := cfg.GetModuleConfig()
	sub := cfg.GetSubConfig()
	crypto.Init(mfg.Crypto, sub.Crypto)
	q := queue.New("channel")
	q.SetConfig(cfg)
	types.Debug = false
	datadir := util.ResetDatadir(mfg, "$TEMP/")
	mock := &TuringchainMock{cfg: mfg, sub: sub, q: q, datadir: datadir}
	mock.random = rand.New(rand.NewSource(types.Now().UnixNano()))

	mock.exec = executor.New(cfg)
	mock.exec.SetQueueClient(q.Client())
	lognode.Info("init exec")

	mock.store = store.New(cfg)
	mock.store.SetQueueClient(q.Client())
	lognode.Info("init store")

	mock.chain = blockchain.New(cfg)
	mock.chain.SetQueueClient(q.Client())
	lognode.Info("init blockchain")

	mock.cs = consensus.New(cfg)
	mock.cs.SetQueueClient(q.Client())
	lognode.Info("init consensus " + mfg.Consensus.Name)

	mock.mem = mempool.New(cfg)
	mock.mem.SetQueueClient(q.Client())
	mock.mem.Wait()
	lognode.Info("init mempool")
	if mfg.P2P.Enable {
		mock.network = p2p.NewP2PMgr(cfg)
		mock.network.SetQueueClient(q.Client())
	} else {
		mock.network = &mockP2P{}
		mock.network.SetQueueClient(q.Client())
	}
	lognode.Info("init P2P")
	cli := q.Client()
	w := wallet.New(cfg)
	mock.client = q.Client()
	mock.wallet = w
	mock.wallet.SetQueueClient(cli)
	lognode.Info("init wallet")
	if mockapi == nil {
		var err error
		mockapi, err = client.New(q.Client(), nil)
		if err != nil {
			return nil
		}
		newWalletRealize(mockapi)
	}
	mock.api = mockapi
	server := rpc.New(cfg)
	server.SetAPI(mock.api)
	server.SetQueueClientNoListen(q.Client())
	mock.rpc = server
	return mock
}

//New :
func New(cfgpath string, mockapi client.QueueProtocolAPI) *TuringchainMock {
	var cfg *types.TuringchainConfig
	if cfgpath == "" || cfgpath == "--notset--" || cfgpath == "--free--" {
		cfg = types.NewTuringchainConfig(types.GetDefaultCfgstring())
		if cfgpath == "--free--" {
			setFee(cfg.GetModuleConfig(), 0)
			cfg.SetMinFee(0)
		}
	} else {
		cfg = types.NewTuringchainConfig(types.ReadFile(cfgpath))
	}
	return newWithConfig(cfg, mockapi)
}

//Listen :
func (mock *TuringchainMock) Listen() {
	pluginmgr.AddRPC(mock.rpc)
	var portgrpc, portjsonrpc int
	for {
		portgrpc, portjsonrpc = mock.rpc.Listen()
		if portgrpc != 0 && portjsonrpc != 0 {
			break
		}
	}

	if strings.HasSuffix(mock.cfg.RPC.JrpcBindAddr, ":0") {
		l := len(mock.cfg.RPC.JrpcBindAddr)
		mock.cfg.RPC.JrpcBindAddr = mock.cfg.RPC.JrpcBindAddr[0:l-2] + ":" + fmt.Sprint(portjsonrpc)
	}
	if strings.HasSuffix(mock.cfg.RPC.GrpcBindAddr, ":0") {
		l := len(mock.cfg.RPC.GrpcBindAddr)
		mock.cfg.RPC.GrpcBindAddr = mock.cfg.RPC.GrpcBindAddr[0:l-2] + ":" + fmt.Sprint(portgrpc)
	}
}

//ModifyParaClient modify para config
func ModifyParaClient(cfg *types.TuringchainConfig, gaddr string) {
	sub := cfg.GetSubConfig()
	if sub.Consensus["para"] != nil {
		data, err := types.ModifySubConfig(sub.Consensus["para"], "ParaRemoteGrpcClient", gaddr)
		if err != nil {
			panic(err)
		}
		sub.Consensus["para"] = data
		cfg.S("config.consensus.sub.para.ParaRemoteGrpcClient", gaddr)
	}
}

//GetBlockChain :
func (mock *TuringchainMock) GetBlockChain() *blockchain.BlockChain {
	return mock.chain
}

func setFee(cfg *types.Config, fee int64) {
	cfg.Mempool.MinTxFeeRate = fee
	cfg.Wallet.MinFee = fee
}

//GetJSONC :
func (mock *TuringchainMock) GetJSONC() *jsonclient.JSONClient {
	jsonc, err := jsonclient.NewJSONClient("http://" + mock.cfg.RPC.JrpcBindAddr + "/")
	if err != nil {
		return nil
	}
	return jsonc
}

//SendAndSign :
func (mock *TuringchainMock) SendAndSign(priv crypto.PrivKey, hextx string) ([]byte, error) {
	txbytes, err := common.FromHex(hextx)
	if err != nil {
		return nil, err
	}
	tx := &types.Transaction{}
	err = types.Decode(txbytes, tx)
	if err != nil {
		return nil, err
	}
	tx.Fee = 1e6
	tx.Sign(types.SECP256K1, priv)
	reply, err := mock.api.SendTx(tx)
	if err != nil {
		return nil, err
	}
	return reply.GetMsg(), nil
}

//SendAndSignNonce 用外部传入的nonce 重写nonce
func (mock *TuringchainMock) SendAndSignNonce(priv crypto.PrivKey, hextx string, nonce int64) ([]byte, error) {
	txbytes, err := common.FromHex(hextx)
	if err != nil {
		return nil, err
	}
	tx := &types.Transaction{}
	err = types.Decode(txbytes, tx)
	if err != nil {
		return nil, err
	}
	tx.Nonce = nonce
	tx.ChainID = mock.q.GetConfig().GetChainID()

	tx.Fee = 1e6
	tx.Sign(types.SECP256K1, priv)
	reply, err := mock.api.SendTx(tx)
	if err != nil {
		return nil, err
	}
	return reply.GetMsg(), nil
}

func newWalletRealize(qAPI client.QueueProtocolAPI) {
	seed := &types.SaveSeedByPw{
		Seed:   "subject hamster apple parent vital can adult chapter fork business humor pen tiger void elephant",
		Passwd: "123456fuzamei",
	}
	reply, err := qAPI.ExecWalletFunc("wallet", "SaveSeed", seed)
	if !reply.(*types.Reply).IsOk && err != nil {
		panic(err)
	}
	reply, err = qAPI.ExecWalletFunc("wallet", "WalletUnLock", &types.WalletUnLock{Passwd: "123456fuzamei"})
	if !reply.(*types.Reply).IsOk && err != nil {
		panic(err)
	}
	for i, priv := range util.TestPrivkeyHex {
		privkey := &types.ReqWalletImportPrivkey{Privkey: priv, Label: fmt.Sprintf("label%d", i)}
		acc, err := qAPI.ExecWalletFunc("wallet", "WalletImportPrivkey", privkey)
		if err != nil {
			panic(err)
		}
		lognode.Info("import", "index", i, "addr", acc.(*types.WalletAccount).Acc.Addr)
	}
	req := &types.ReqAccountList{WithoutBalance: true}
	_, err = qAPI.ExecWalletFunc("wallet", "WalletGetAccountList", req)
	if err != nil {
		panic(err)
	}
}

//GetAPI :
func (mock *TuringchainMock) GetAPI() client.QueueProtocolAPI {
	return mock.api
}

//GetRPC :
func (mock *TuringchainMock) GetRPC() *rpc.RPC {
	return mock.rpc
}

//GetCfg :
func (mock *TuringchainMock) GetCfg() *types.Config {
	return mock.cfg
}

//GetLastSendTx :
func (mock *TuringchainMock) GetLastSendTx() []byte {
	return mock.lastsend
}

//Close :
func (mock *TuringchainMock) Close() {
	mock.closeNoLock()
}

func (mock *TuringchainMock) closeNoLock() {
	lognode.Info("network close")
	mock.network.Close()
	lognode.Info("network close")
	mock.rpc.Close()
	lognode.Info("rpc close")
	mock.mem.Close()
	lognode.Info("mem close")
	mock.exec.Close()
	lognode.Info("exec close")
	mock.cs.Close()
	lognode.Info("cs close")
	mock.wallet.Close()
	lognode.Info("wallet close")
	mock.chain.Close()
	lognode.Info("chain close")
	mock.store.Close()
	lognode.Info("store close")
	mock.client.Close()
	err := os.RemoveAll(mock.datadir)
	if err != nil {
		return
	}
}

//WaitHeight :
func (mock *TuringchainMock) WaitHeight(height int64) error {
	for {
		header, err := mock.api.GetLastHeader()
		if err != nil {
			return err
		}
		if header.Height >= height {
			break
		}
		time.Sleep(time.Second / 10)
	}
	return nil
}

//WaitTx :
func (mock *TuringchainMock) WaitTx(hash []byte) (*rpctypes.TransactionDetail, error) {
	if hash == nil {
		return nil, nil
	}
	for {
		param := &types.ReqHash{Hash: hash}
		_, err := mock.api.QueryTx(param)
		if err != nil {
			time.Sleep(time.Second / 10)
			continue
		}
		var testResult rpctypes.TransactionDetail
		data := rpctypes.QueryParm{
			Hash: common.ToHex(hash),
		}
		err = mock.GetJSONC().Call("Turingchain.QueryTransaction", data, &testResult)
		return &testResult, err
	}
}

//SendHot :
func (mock *TuringchainMock) SendHot() error {
	types.AssertConfig(mock.client)
	tx := util.CreateCoinsTx(mock.client.GetConfig(), mock.GetGenesisKey(), mock.GetHotAddress(), 10000*types.Coin)
	mock.SendTx(tx)
	return mock.Wait()
}

//SendTx :
func (mock *TuringchainMock) SendTx(tx *types.Transaction) []byte {
	reply, err := mock.GetAPI().SendTx(tx)
	if err != nil {
		panic(err)
	}
	mock.SetLastSend(reply.GetMsg())
	return reply.GetMsg()
}

//SetLastSend :
func (mock *TuringchainMock) SetLastSend(hash []byte) {
	mock.mu.Lock()
	mock.lastsend = hash
	mock.mu.Unlock()
}

//SendTxRPC :
func (mock *TuringchainMock) SendTxRPC(tx *types.Transaction) []byte {
	var txhash string
	hextx := common.ToHex(types.Encode(tx))
	err := mock.GetJSONC().Call("Turingchain.SendTransaction", &rpctypes.RawParm{Data: hextx}, &txhash)
	if err != nil {
		panic(err)
	}
	hash, err := common.FromHex(txhash)
	if err != nil {
		panic(err)
	}
	mock.lastsend = hash
	return hash
}

//Wait :
func (mock *TuringchainMock) Wait() error {
	if mock.lastsend == nil {
		return nil
	}
	_, err := mock.WaitTx(mock.lastsend)
	return err
}

//GetAccount :
func (mock *TuringchainMock) GetAccount(stateHash []byte, addr string) *types.Account {
	statedb := executor.NewStateDB(mock.client, stateHash, nil, nil)
	types.AssertConfig(mock.client)
	acc := account.NewCoinsAccount(mock.client.GetConfig())
	acc.SetDB(statedb)
	return acc.LoadAccount(addr)
}

//GetExecAccount :get execer account info
func (mock *TuringchainMock) GetExecAccount(stateHash []byte, execer, addr string) *types.Account {
	statedb := executor.NewStateDB(mock.client, stateHash, nil, nil)
	types.AssertConfig(mock.client)
	acc := account.NewCoinsAccount(mock.client.GetConfig())
	acc.SetDB(statedb)
	return acc.LoadExecAccount(addr, address.ExecAddress(execer))
}

//GetBlock :
func (mock *TuringchainMock) GetBlock(height int64) *types.Block {
	blocks, err := mock.api.GetBlocks(&types.ReqBlocks{Start: height, End: height})
	if err != nil {
		panic(err)
	}
	return blocks.Items[0].Block
}

//GetLastBlock :
func (mock *TuringchainMock) GetLastBlock() *types.Block {
	header, err := mock.api.GetLastHeader()
	if err != nil {
		panic(err)
	}
	return mock.GetBlock(header.Height)
}

//GetClient :
func (mock *TuringchainMock) GetClient() queue.Client {
	return mock.client
}

//GetHotKey :
func (mock *TuringchainMock) GetHotKey() crypto.PrivKey {
	return util.TestPrivkeyList[0]
}

//GetHotAddress :
func (mock *TuringchainMock) GetHotAddress() string {
	return address.PubKeyToAddress(mock.GetHotKey().PubKey().Bytes()).String()
}

//GetGenesisKey :
func (mock *TuringchainMock) GetGenesisKey() crypto.PrivKey {
	return util.TestPrivkeyList[1]
}

//GetGenesisAddress :
func (mock *TuringchainMock) GetGenesisAddress() string {
	return address.PubKeyToAddress(mock.GetGenesisKey().PubKey().Bytes()).String()
}

type mockP2P struct {
}

//SetQueueClient :
func (m *mockP2P) SetQueueClient(client queue.Client) {
	go func() {
		p2pKey := "p2p"
		client.Sub(p2pKey)
		for msg := range client.Recv() {
			switch msg.Ty {
			case types.EventPeerInfo:
				msg.Reply(client.NewMessage(p2pKey, types.EventPeerList, &types.PeerList{}))
			case types.EventGetNetInfo:
				msg.Reply(client.NewMessage(p2pKey, types.EventPeerList, &types.NodeNetInfo{}))
			case types.EventTxBroadcast, types.EventBlockBroadcast:
				client.FreeMessage(msg)
			default:
				msg.ReplyErr("p2p->Do not support "+types.GetEventName(int(msg.Ty)), types.ErrNotSupport)
			}
		}
	}()
}

//Wait for ready
func (m *mockP2P) Wait() {}

//Close :
func (m *mockP2P) Close() {
}
