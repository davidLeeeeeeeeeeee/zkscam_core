// Copyright 2015 The go-ethereum Authors
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

package eth

import (
	"errors"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/txpool"
	"github.com/ethereum/go-ethereum/eth/downloader"
	"github.com/ethereum/go-ethereum/eth/protocols/eth"
	"github.com/ethereum/go-ethereum/log"
)

const (
	forceSyncCycle      = 1 * time.Second // Time interval to force syncs, even if few peers are available
	defaultMinSyncPeers = 1               // Amount of peers desired to start syncing
)

// syncTransactions starts sending all currently pending transactions to the given peer.
func (h *handler) syncTransactions(p *eth.Peer) {
	var hashes []common.Hash
	for _, batch := range h.txpool.Pending(txpool.PendingFilter{OnlyPlainTxs: true}) {
		for _, tx := range batch {
			hashes = append(hashes, tx.Hash)
		}
	}
	if len(hashes) == 0 {
		return
	}
	p.AsyncSendPooledTransactionHashes(hashes)
}

// chainSyncer coordinates blockchain sync components.
type chainSyncer struct {
	handler     *handler
	force       *time.Timer
	forced      bool // true when force timer fired
	warned      time.Time
	peerEventCh chan struct{}
	doneCh      chan error // non-nil when sync is running
}

// chainSyncOp is a scheduled sync operation.
type chainSyncOp struct {
	mode       downloader.SyncMode
	peer       *eth.Peer
	td         *big.Int
	totalVotes *big.Int
	head       common.Hash
}

// newChainSyncer creates a chainSyncer.
func newChainSyncer(handler *handler) *chainSyncer {
	return &chainSyncer{
		handler:     handler,
		peerEventCh: make(chan struct{}),
	}
}

// handlePeerEvent notifies the syncer about a change in the peer set.
// This is called for new peers and every time a peer announces a new
// chain head.
func (cs *chainSyncer) handlePeerEvent() bool {
	select {
	case cs.peerEventCh <- struct{}{}:
		//log.Info("case cs.peerEventCh <- struct{}{}:")
		return true
	case <-cs.handler.quitSync:
		//log.Info("case <-cs.handler.quitSync:")
		return false
	}
}

// loop runs in its own goroutine and launches the sync when necessary.
func (cs *chainSyncer) loop() {
	defer cs.handler.wg.Done()

	cs.handler.blockFetcher.Start()
	cs.handler.txFetcher.Start()
	defer cs.handler.blockFetcher.Stop()
	defer cs.handler.txFetcher.Stop()
	defer cs.handler.downloader.Terminate()

	// The force timer lowers the peer count threshold down to one when it fires.
	// This ensures we'll always start sync even if there aren't enough peers.
	cs.force = time.NewTimer(forceSyncCycle)
	defer cs.force.Stop()

	for {
		if op := cs.nextSyncOp(); op != nil {
			cs.startSync(op)
		}
		select {
		case <-cs.peerEventCh:
			// Peer information changed, recheck.
		case err := <-cs.doneCh:
			cs.doneCh = nil
			cs.force.Reset(forceSyncCycle)
			cs.forced = false

			// If we've reached the merge transition but no beacon client is available, or
			// it has not yet switched us over, keep warning the user that their infra is
			// potentially flaky.
			if errors.Is(err, downloader.ErrMergeTransition) && time.Since(cs.warned) > 10*time.Second {
				log.Warn("Local chain is post-merge, waiting for beacon client sync switch-over...")
				cs.warned = time.Now()
			}
		case <-cs.force.C:
			cs.forced = true

		case <-cs.handler.quitSync:
			// Disable all insertion on the blockchain. This needs to happen before
			// terminating the downloader because the downloader waits for blockchain
			// inserts, and these can take a long time to finish.
			cs.handler.chain.StopInsert()
			cs.handler.downloader.Terminate()
			if cs.doneCh != nil {
				<-cs.doneCh
			}
			return
		}
	}
}
func (cs *chainSyncer) modeAndLocalVotes() (downloader.SyncMode, *big.Int) {
	head := cs.handler.chain.CurrentBlock()
	totalVotes := head.TotalVotes // 从区块头获取 TotalVotes
	return downloader.FullSync, totalVotes
}

// nextSyncOp determines whether sync is required at this time.
func (cs *chainSyncer) nextSyncOp() *chainSyncOp {
	//log.Info("进入 nextSyncOp 函数 / Entering nextSyncOp function", "时间 / Time", time.Now())      // 中英文描述: 函数开始的日志
	//defer log.Info("退出 nextSyncOp 函数 / Exiting nextSyncOp function", "时间 / Time", time.Now()) // 中英文描述: 函数退出的日志

	if cs.doneCh != nil {
		//log.Info("同步操作正在进行，跳过新的同步请求 / Sync already running, skipping new sync operation") // 中英文描述: 当前同步正在进行
		return nil // Sync already running
	}

	// 确保达到最小对等节点数 / Ensure we're at minimum peer count.
	minPeers := defaultMinSyncPeers
	if cs.forced {
		minPeers = 1
		//log.Info("强制同步已启动，将最小对等节点数设置为 1 / Forced sync started, setting minimum peer count to 1") // 中英文描述: 强制同步时将最小对等节点数设为 1
	} else if minPeers > cs.handler.maxPeers {
		minPeers = cs.handler.maxPeers
		//log.Info("减少最小对等节点数以匹配最大对等节点数 / Reducing minimum peer count to match maxPeers", "最大对等节点数 / Max Peers", cs.handler.maxPeers) // 中英文描述: 匹配最大对等节点数
	}

	// 记录当前对等节点数量 / Log current peer count
	//log.Info("检查对等节点数量 / Checking peer count", "当前对等节点数 / Current Peers", cs.handler.peers.len(), "最小对等节点数 / Minimum Peers", minPeers) // 中英文描述: 检查对等节点数量
	if cs.handler.peers.len() < minPeers {
		//log.Info("对等节点数量不足，无法开始同步 / Not enough peers to start sync", "当前对等节点数 / Current Peers", cs.handler.peers.len(), "最小对等节点数 / Minimum Peers", minPeers) // 中英文描述: 对等节点数量不足
		return nil
	}

	// 选择 TotalVotes 更高的 peer 来进行同步 / Select peer with higher TotalVotes for sync
	peer := cs.handler.peers.peerWithHighestVotes()
	if peer == nil {
		//log.Info("没有找到拥有更高 TotalVotes 的合适对等节点 / No suitable peer with higher TotalVotes found") // 中英文描述: 没有找到合适的对等节点
		return nil
	}

	// 记录选择的对等节点及其链头信息 / Log the selected peer and its head info

	mode, ourTotalVotes := cs.modeAndLocalVotes() // 基于 Votes 的本地链头 / Local head based on Votes
	//log.Info("本地节点链头信息 / Local node head info", "同步模式 / Sync Mode", mode, "本地 TotalVotes / Local TotalVotes", ourTotalVotes) // 中英文描述: 本地节点链头信息

	op := peerToSyncOp(mode, peer)
	//log.Info("选择了对等节点进行同步 / Selected peer for sync", "对等节点ID / Peer ID", peer.ID(), "对等节点 TotalVotes / Peer TotalVotes", op.totalVotes.String()) // 中英文描述: 记录选择的对等节点及其链头信息
	// 使用 TotalVotes 来决定是否已经同步 / Use TotalVotes to decide if already synced
	if op.totalVotes.Cmp(ourTotalVotes) <= 0 {
		//log.Info("已经同步到最新的链，跳过同步 / Already synced to the latest chain, skipping sync", "对等节点 TotalVotes / Peer TotalVotes", op.totalVotes, "本地 TotalVotes / Local TotalVotes", ourTotalVotes) // 中英文描述: 已经同步到最新的链
		return nil
	}

	//log.Info("准备开始同步 / Ready to start sync", "对等节点 TotalVotes / Peer TotalVotes", op.totalVotes, "本地 TotalVotes / Local TotalVotes", ourTotalVotes) // 中英文描述: 准备开始同步
	return op
}

func peerToSyncOp(mode downloader.SyncMode, p *eth.Peer) *chainSyncOp {
	_, peerTotalVotes := p.Head() // 假设 Head 返回 Votes 和 TotalVotes
	return &chainSyncOp{mode: mode, peer: p, totalVotes: peerTotalVotes}
}

func (cs *chainSyncer) modeAndLocalHead() (downloader.SyncMode, *big.Int) {
	// If we're in snap sync mode, return that directly
	if cs.handler.snapSync.Load() {
		block := cs.handler.chain.CurrentSnapBlock()
		td := cs.handler.chain.GetTd(block.Hash(), block.Number.Uint64())
		return downloader.SnapSync, td
	}
	// We are probably in full sync, but we might have rewound to before the
	// snap sync pivot, check if we should re-enable snap sync.
	head := cs.handler.chain.CurrentBlock()
	if pivot := rawdb.ReadLastPivotNumber(cs.handler.database); pivot != nil {
		if head.Number.Uint64() < *pivot {
			block := cs.handler.chain.CurrentSnapBlock()
			td := cs.handler.chain.GetTd(block.Hash(), block.Number.Uint64())
			return downloader.SnapSync, td
		}
	}
	// We are in a full sync, but the associated head state is missing. To complete
	// the head state, forcefully rerun the snap sync. Note it doesn't mean the
	// persistent state is corrupted, just mismatch with the head block.
	if !cs.handler.chain.HasState(head.Root) {
		block := cs.handler.chain.CurrentSnapBlock()
		td := cs.handler.chain.GetTd(block.Hash(), block.Number.Uint64())
		log.Info("Reenabled snap sync as chain is stateless")
		return downloader.SnapSync, td
	}
	// Nope, we're really full syncing
	td := cs.handler.chain.GetTd(head.Hash(), head.Number.Uint64())
	return downloader.FullSync, td
}

// startSync launches doSync in a new goroutine.
func (cs *chainSyncer) startSync(op *chainSyncOp) {
	cs.doneCh = make(chan error, 1)
	go func() { cs.doneCh <- cs.handler.doSync(op) }()
}

func (h *handler) doSync(op *chainSyncOp) error {
	log.Info("Sending sync request", "peer", op.peer.ID(), "head", op.head, "totalVotes", op.totalVotes, "time", time.Now())

	// 修改为使用 TotalVotes 进行同步
	err := h.downloader.LegacySync(op.peer.ID(), op.head, op.totalVotes, op.mode)
	if err != nil {
		log.Error("Sync failed", "peer", op.peer.ID(), "error", err)
		return err
	}
	h.enableSyncedFeatures()

	// 同步完成后的逻辑，广播最新的块头
	head := h.chain.CurrentBlock()
	if head.Number.Uint64() > 0 {
		if block := h.chain.GetBlock(head.Hash(), head.Number.Uint64()); block != nil {
			h.BroadcastBlock(block, false)
		}
	}
	return nil
}
