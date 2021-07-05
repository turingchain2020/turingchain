// Copyright Turing Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package executor 管理插件执行器
package executor

import (
	log "github.com/turingchain2020/turingchain/common/log/log15"
	drivers "github.com/turingchain2020/turingchain/system/dapp"
	"github.com/turingchain2020/turingchain/types"
)

var (
	clog       = log.New("module", "execs.manage")
	driverName = "manage"
)

// Init resister a dirver
func Init(name string, cfg *types.TuringchainConfig, sub []byte) {
	// 需要先 RegisterDappFork才可以Register dapp
	drivers.Register(cfg, GetName(), newManage, cfg.GetDappFork(driverName, "Enable"))
	InitExecType()
}

// InitExecType initials coins functions.
func InitExecType() {
	ety := types.LoadExecutorType(driverName)
	ety.InitFuncList(types.ListMethod(&Manage{}))
}

// GetName return manage name
func GetName() string {
	return newManage().GetName()
}

// Manage defines Manage object
type Manage struct {
	drivers.DriverBase
}

func newManage() drivers.Driver {
	c := &Manage{}
	c.SetChild(c)
	c.SetExecutorType(types.LoadExecutorType(driverName))
	return c
}

// GetDriverName return a drivername
func (c *Manage) GetDriverName() string {
	return driverName
}

// CheckTx checkout transaction
func (c *Manage) CheckTx(tx *types.Transaction, index int) error {
	return nil
}

// IsSuperManager is supper manager or not
func IsSuperManager(cfg *types.TuringchainConfig, addr string) bool {
	conf := types.ConfSub(cfg, driverName)
	for _, m := range conf.GStrList("superManager") {
		if addr == m {
			return true
		}
	}
	return false
}

// CheckReceiptExecOk return true to check if receipt ty is ok
func (c *Manage) CheckReceiptExecOk() bool {
	return true
}
