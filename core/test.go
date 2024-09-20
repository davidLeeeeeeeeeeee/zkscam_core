package core

//
//func (bc *BlockChain) insertChain(chain types.Blocks, setHead bool) (int, error) {
//	// If the chain is terminating, don't even bother starting up.
//	if bc.insertStopped() {
//		return 0, nil
//	}
//
//	// Start a parallel signature recovery (signer will fluke on fork transition, minimal perf loss)
//	SenderCacher.RecoverFromBlocks(types.MakeSigner(bc.chainConfig, chain[0].Number(), chain[0].Time()), chain)
//
//	var (
//		stats     = insertStats{startTime: mclock.Now()}
//		lastCanon *types.Block
//	)
//	// Fire a single chain head event if we've progressed the chain
//	defer func() {
//		if lastCanon != nil && bc.CurrentBlock().Hash() == lastCanon.Hash() {
//			bc.chainHeadFeed.Send(ChainHeadEvent{lastCanon})
//		}
//	}()
//
//	// Start the parallel header verifier
//	for _, block := range chain {
//		// 将每个 block 封装成一个单独的 types.Blocks 列表
//		singleBlockChain := types.Blocks{block}
//
//		// 提取该 block 的 header
//		headers := make([]*types.Header, 1)
//		headers[0] = block.Header()
//
//		// 调用 VerifyHeaders 验证该单个 header
//		abort, results := bc.engine.VerifyHeaders(bc, headers)
//		defer close(abort)
//
//		// Peek the error for the first block to decide the directing import logic
//		it := newInsertIterator(singleBlockChain, results, bc.validator)
//		block, err := it.next()
//
//		// Left-trim all the known blocks that don't need to build snapshot
//		if bc.skipBlock(err, it) {
//			// Reorganize logic (保持和之前一样，处理跳过的区块)
//			var (
//				reorg   bool
//				current = bc.CurrentBlock()
//			)
//			for block != nil && bc.skipBlock(err, it) {
//				reorg, err = bc.forker.ReorgNeeded(current, block.Header())
//				if err != nil {
//					return it.index, err
//				}
//				if reorg {
//					if block.NumberU64() > current.Number.Uint64() || bc.GetCanonicalHash(block.NumberU64()) != block.Hash() {
//						break
//					}
//				}
//				log.Debug("Ignoring already known block", "number", block.Number(), "hash", block.Hash())
//				stats.ignored++
//				block, err = it.next()
//			}
//
//			for block != nil && bc.skipBlock(err, it) {
//				log.Debug("Writing previously known block", "number", block.Number(), "hash", block.Hash())
//				if err := bc.writeKnownBlock(block); err != nil {
//					return it.index, err
//				}
//				lastCanon = block
//				block, err = it.next()
//			}
//		}
//
//		// 继续处理剩下的部分，验证后插入区块
//		switch {
//		case errors.Is(err, consensus.ErrPrunedAncestor):
//			if setHead {
//				log.Debug("Pruned ancestor, inserting as sidechain", "number", block.Number(), "hash", block.Hash())
//				return bc.insertSideChain(block, it)
//			} else {
//				log.Debug("Pruned ancestor", "number", block.Number(), "hash", block.Hash())
//				_, err := bc.recoverAncestors(block)
//				return it.index, err
//			}
//		case errors.Is(err, consensus.ErrFutureBlock) || (errors.Is(err, consensus.ErrUnknownAncestor) && bc.futureBlocks.Contains(it.first().ParentHash())):
//			for block != nil && (it.index == 0 || errors.Is(err, consensus.ErrUnknownAncestor)) {
//				log.Debug("Future block, postponing import", "number", block.Number(), "hash", block.Hash())
//				if err := bc.addFutureBlock(block); err != nil {
//					return it.index, err
//				}
//				block, err = it.next()
//			}
//			stats.queued += it.processed()
//			stats.ignored += it.remaining()
//			return it.index, err
//		case err != nil && !errors.Is(err, ErrKnownBlock):
//			bc.futureBlocks.Remove(block.Hash())
//			stats.ignored += len(it.chain)
//			bc.reportBlock(block, nil, err)
//			return it.index, err
//		}
//
//		// 验证并插入每一个区块
//		var activeState *state.StateDB
//		defer func() {
//			if activeState != nil {
//				activeState.StopPrefetcher()
//			}
//		}()
//
//		for ; block != nil && err == nil || errors.Is(err, ErrKnownBlock); block, err = it.next() {
//			if bc.insertStopped() {
//				log.Debug("Abort during block processing")
//				break
//			}
//
//			if BadHashes[block.Hash()] {
//				bc.reportBlock(block, nil, ErrBannedHash)
//				return it.index, ErrBannedHash
//			}
//
//			if bc.skipBlock(err, it) {
//				logger := log.Debug
//				if bc.chainConfig.Clique == nil {
//					logger = log.Warn
//				}
//				logger("Inserted known block", "number", block.Number(), "hash", block.Hash(),
//					"uncles", len(block.Uncles()), "txs", len(block.Transactions()), "gas", block.GasUsed(),
//					"root", block.Root())
//
//				if err := bc.writeKnownBlock(block); err != nil {
//					return it.index, err
//				}
//				stats.processed++
//				lastCanon = block
//				continue
//			}
//
//			parent := it.previous()
//			if parent == nil {
//				parent = bc.GetHeader(block.ParentHash(), block.NumberU64()-1)
//			}
//			statedb, err := state.New(parent.Root, bc.stateCache, bc.snaps)
//			if err != nil {
//				return it.index, err
//			}
//			statedb.StartPrefetcher("chain")
//			activeState = statedb
//
//			receipts, logs, usedGas, err := bc.processor.Process(block, statedb, bc.vmConfig)
//			if err != nil {
//				bc.reportBlock(block, receipts, err)
//				return it.index, err
//			}
//
//			if err := bc.validator.ValidateState(block, statedb, receipts, usedGas); err != nil {
//				bc.reportBlock(block, receipts, err)
//				return it.index, err
//			}
//
//			var status WriteStatus
//			if !setHead {
//				err = bc.writeBlockWithState(block, receipts, statedb)
//			} else {
//				status, err = bc.writeBlockAndSetHead(block, receipts, logs, statedb, false)
//			}
//			if err != nil {
//				return it.index, err
//			}
//
//			stats.processed++
//			stats.usedGas += usedGas
//		}
//	}
//	return it.index, err
//
//}
