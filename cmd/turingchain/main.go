// Copyright Turing Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build go1.8

// Package main turingchain程序入口
package main

import (
	_ "github.com/turingchain2020/turingchain/system"
	"github.com/turingchain2020/turingchain/util/cli"
)

func main() {
	cli.RunTuringchain("", "")
}
