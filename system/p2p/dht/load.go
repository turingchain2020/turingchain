package dht

import (
	_ "github.com/turingchain2020/turingchain/system/p2p/dht/protocol/broadcast" //register init package
	_ "github.com/turingchain2020/turingchain/system/p2p/dht/protocol/download"  //register init package
	_ "github.com/turingchain2020/turingchain/system/p2p/dht/protocol/p2pstore"  //register init package
	_ "github.com/turingchain2020/turingchain/system/p2p/dht/protocol/peer"      //register init package
)
