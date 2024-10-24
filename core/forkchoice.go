// Copyright 2021 The go-ethereum Authors
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

package core

import (
	crand "crypto/rand"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
	single "github.com/ethereum/go-ethereum/singleton"
	"math/big"
	mrand "math/rand"
)

// ChainReader defines a small collection of methods needed to access the local
// blockchain during header verification. It's implemented by both blockchain
// and lightchain.
type ChainReader interface {
	// Config retrieves the header chain's chain configuration.
	Config() *params.ChainConfig

	// GetTd returns the total difficulty of a local block.
	GetTd(common.Hash, uint64) *big.Int
}

// ForkChoice is the fork chooser based on the highest total difficulty of the
// chain(the fork choice used in the eth1) and the external fork choice (the fork
// choice used in the eth2). This main goal of this ForkChoice is not only for
// offering fork choice during the eth1/2 merge phase, but also keep the compatibility
// for all other proof-of-work networks.
type ForkChoice struct {
	chain ChainReader
	rand  *mrand.Rand

	// preserve is a helper function used in td fork choice.
	// Miners will prefer to choose the local mined block if the
	// local td is equal to the extern one. It can be nil for light
	// client
	preserve func(header *types.Header) bool
}

func NewForkChoice(chainReader ChainReader, preserve func(header *types.Header) bool) *ForkChoice {
	// Seed a fast but crypto originating random generator
	seed, err := crand.Int(crand.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		log.Crit("Failed to initialize random seed", "err", err)
	}
	return &ForkChoice{
		chain:    chainReader,
		rand:     mrand.New(mrand.NewSource(seed.Int64())),
		preserve: preserve,
	}
}

// ReorgNeeded returns whether the reorg should be applied
// based on the given external header and local canonical chain.
// In the new POS mode, the new head is chosen if the corresponding
// TotalVotes is higher. The trusted header is not selected based on
// external trust but by direct vote comparison.
func (f *ForkChoice) ReorgNeeded(current *types.Header, extern *types.Header) (bool, error) {
	log.Info("func (f *ForkChoice) ReorgNeeded ", "current", current.Number.String(), "extern", extern.Number.String())
	var (
		localVotes  = current.TotalVotes
		externVotes = extern.TotalVotes
	)
	if localVotes == nil || externVotes == nil {
		log.Info("return false, errors.New(\"missing votes\")")
		single.IsReorging = false
		return false, errors.New("missing votes")
	}
	// If the total votes are higher in the external header, choose it as the new head
	if diff := externVotes.Cmp(localVotes); diff > 0 {
		//log.Info("return true, nil", externVotes.String(), localVotes.String())

		single.IsReorging = true
		return true, nil
	} else if diff < 0 {
		log.Info("return false, nil")
		single.IsReorging = false
		return false, nil
	}
	// 投票相同，比较MinerAddresses的数量
	localMinersCount := len(current.MinerAddresses)
	externMinersCount := len(extern.MinerAddresses)

	if externMinersCount > localMinersCount {
		single.IsReorging = true
		return true, nil
	} else if externMinersCount < localMinersCount {
		return false, nil
	}
	// Local and external votes are identical.
	// Second clause reduces the vulnerability to selfish mining attacks.
	// Please refer to http://www.cs.cornell.edu/~ie53/publications/btcProcFC.pdf
	single.IsReorging = false
	reorg := false
	externNum, localNum := extern.Number.Uint64(), current.Number.Uint64()
	//log.Info("externNum, localNum :")
	if externNum < localNum {
		single.IsReorging = true
		reorg = true
	} else if externNum == localNum {
		if current.ZkscamHash == extern.ZkscamHash {
			single.IsReorging = false
			reorg = false
		}
	}
	//log.Info("return reorg, nil", reorg)
	return reorg, nil
}
