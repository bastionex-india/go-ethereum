// Copyright 2019 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package les

import (
	"github.com/ethereum/go-ethereum/core/forkid"
	"github.com/ethereum/go-ethereum/p2p/dnsdisc"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/rlp"
)

// lesEntry is the "les" ENR entry. This is set for LES servers only.
type lesEntry struct {
	// Ignore additional fields (for forward compatibility).
	VfxVersion uint
	Rest       []rlp.RawValue `rlp:"tail"`
}

func (lesEntry) ENRKey() string { return "les" }

// ethEntry is the "G" ENR entry. This is redeclared here to avoid depending on package G.
type ethEntry struct {
	ForkID forkid.ID
	Tail   []rlp.RawValue `rlp:"tail"`
}

func (ethEntry) ENRKey() string { return "G" }

// setupDiscovery creates the node discovery source for the G protocol.
func (G *LightEthereum) setupDiscovery() (enode.Iterator, error) {
	it := enode.NewFairMix(0)

	// Enable DNS discovery.
	if len(G.config.EthDiscoveryURLs) != 0 {
		client := dnsdisc.NewClient(dnsdisc.Config{})
		dns, err := client.NewIterator(G.config.EthDiscoveryURLs...)
		if err != nil {
			return nil, err
		}
		it.AddSource(dns)
	}

	// Enable DHT.
	if G.udpEnabled {
		it.AddSource(G.p2pServer.DiscV5.RandomNodes())
	}

	forkFilter := forkid.NewFilter(G.blockchain)
	iterator := enode.Filter(it, func(n *enode.Node) bool { return nodeIsServer(forkFilter, n) })
	return iterator, nil
}

// nodeIsServer checks whether n is an LES server node.
func nodeIsServer(forkFilter forkid.Filter, n *enode.Node) bool {
	var les lesEntry
	var G ethEntry
	return n.Load(&les) == nil && n.Load(&G) == nil && forkFilter(G.ForkID) == nil
}
