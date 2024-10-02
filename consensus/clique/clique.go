// Copyright 2017 The go-ethereum Authors
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

// Package clique implements the proof-of-authority consensus engine.
package clique

import (
	"bytes"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/contracts"
	"github.com/ethereum/go-ethereum/eth/fetcher"
	eth2 "github.com/ethereum/go-ethereum/eth/protocols/eth"
	single "github.com/ethereum/go-ethereum/singleton"
	"github.com/holiman/uint256"
	"io"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	lru "github.com/ethereum/go-ethereum/common/lru"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/consensus/misc"
	"github.com/ethereum/go-ethereum/consensus/misc/eip1559"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum/trie"
	"golang.org/x/crypto/sha3"
)

const (
	checkpointInterval  = 1024 // Number of blocks after which to save the vote snapshot to the database
	inmemorySnapshots   = 128  // Number of recent vote snapshots to keep in memory
	inmemorySignatures  = 4096 // Number of recent block signatures to keep in memory
	miner_waiting_block = 10
	wiggleTime          = 500 * time.Millisecond // Random delay (per signer) to allow concurrent signers
)

// Clique proof-of-authority protocol constants.
var (
	epochLength = uint64(30000) // Default number of blocks after which to checkpoint and reset the pending votes

	extraVanity = 32                     // Fixed number of extra-data prefix bytes reserved for signer vanity
	extraSeal   = crypto.SignatureLength // Fixed number of extra-data suffix bytes reserved for signer seal

	nonceAuthVote = hexutil.MustDecode("0xffffffffffffffff") // Magic nonce number to vote on adding a new signer
	nonceDropVote = hexutil.MustDecode("0x0000000000000000") // Magic nonce number to vote on removing a signer.

	uncleHash = types.CalcUncleHash(nil) // Always Keccak256(RLP([])) as uncles are meaningless outside of PoW.

	diffInTurn = big.NewInt(2) // Block difficulty for in-turn signatures
	diffNoTurn = big.NewInt(1) // Block difficulty for out-of-turn signatures
)

// Various error messages to mark blocks invalid. These should be private to
// prevent engine specific errors from being referenced in the remainder of the
// codebase, inherently breaking if the engine is swapped out. Please put common
// error types into the consensus package.
var (
	// errUnknownBlock is returned when the list of signers is requested for a block
	// that is not part of the local blockchain.
	errUnknownBlock = errors.New("unknown block")

	errBalanceNotEnough = errors.New("stoke balance not enough")
	errMinerVotesIsNil  = errors.New("miner votes is nil")
	// errInvalidCheckpointBeneficiary is returned if a checkpoint/epoch transition
	// block has a beneficiary set to non-zeroes.
	errInvalidCheckpointBeneficiary = errors.New("beneficiary in checkpoint block non-zero")

	// errInvalidVote is returned if a nonce value is something else that the two
	// allowed constants of 0x00..0 or 0xff..f.
	errInvalidVote = errors.New("vote nonce not 0x00..0 or 0xff..f")

	// errInvalidCheckpointVote is returned if a checkpoint/epoch transition block
	// has a vote nonce set to non-zeroes.
	errInvalidCheckpointVote = errors.New("vote nonce in checkpoint block non-zero")

	// errMissingVanity is returned if a block's extra-data section is shorter than
	// 32 bytes, which is required to store the signer vanity.
	errMissingVanity = errors.New("extra-data 32 byte vanity prefix missing")

	// errMissingSignature is returned if a block's extra-data section doesn't seem
	// to contain a 65 byte secp256k1 signature.
	errMissingSignature = errors.New("extra-data 65 byte signature suffix missing")

	// errExtraSigners is returned if non-checkpoint block contain signer data in
	// their extra-data fields.
	errExtraSigners = errors.New("non-checkpoint block contains extra signer list")

	// errInvalidCheckpointSigners is returned if a checkpoint block contains an
	// invalid list of signers (i.e. non divisible by 20 bytes).
	errInvalidCheckpointSigners = errors.New("invalid signer list on checkpoint block")

	// errMismatchingCheckpointSigners is returned if a checkpoint block contains a
	// list of signers different than the one the local node calculated.
	errMismatchingCheckpointSigners = errors.New("mismatching signer list on checkpoint block")

	// errInvalidMixDigest is returned if a block's mix digest is non-zero.
	errInvalidMixDigest = errors.New("non-zero mix digest")

	// errInvalidUncleHash is returned if a block contains an non-empty uncle list.
	errInvalidUncleHash = errors.New("non empty uncle hash")

	// errInvalidDifficulty is returned if the difficulty of a block neither 1 or 2.
	errInvalidDifficulty = errors.New("invalid difficulty")

	// errWrongDifficulty is returned if the difficulty of a block doesn't match the
	// turn of the signer.
	errWrongDifficulty = errors.New("wrong difficulty")

	// errInvalidTimestamp is returned if the timestamp of a block is lower than
	// the previous block's timestamp + the minimum block period.
	errInvalidTimestamp = errors.New("invalid timestamp")

	// errInvalidVotingChain is returned if an authorization list is attempted to
	// be modified via out-of-range or non-contiguous headers.
	errInvalidVotingChain = errors.New("invalid voting chain")

	// errUnauthorizedSigner is returned if a header is signed by a non-authorized entity.
	errUnauthorizedSigner = errors.New("unauthorized signer")

	// errRecentlySigned is returned if a header is signed by an authorized entity
	// that already signed a header recently, thus is temporarily not allowed to.
	errRecentlySigned = errors.New("recently signed")
)

// SignerFn hashes and signs the data to be signed by a backing account.
type SignerFn func(signer accounts.Account, mimeType string, message []byte) ([]byte, error)

// ecrecover extracts the Ethereum account address from a signed header.
func ecrecover(header *types.Header, sigcache *sigLRU) (common.Address, error) {
	// If the signature's already cached, return that
	hash := header.Hash()
	if address, known := sigcache.Get(hash); known {
		return address, nil
	}
	// Retrieve the signature from the header extra-data
	if len(header.Extra) < extraSeal {
		return common.Address{}, errMissingSignature
	}
	signature := header.Extra[len(header.Extra)-extraSeal:]

	// Recover the public key and the Ethereum address
	pubkey, err := crypto.Ecrecover(SealHash(header).Bytes(), signature)
	if err != nil {
		return common.Address{}, err
	}
	var signer common.Address
	copy(signer[:], crypto.Keccak256(pubkey[1:])[12:])

	sigcache.Add(hash, signer)
	return signer, nil
}

// 用于缓存最近写入的区块头信息
type HeaderCache struct {
	cache map[common.Hash]*types.Header
	mu    sync.RWMutex
}

func (hc *HeaderCache) Get(hash common.Hash) *types.Header {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	return hc.cache[hash]
}

func (hc *HeaderCache) Set(hash common.Hash, header *types.Header) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	hc.cache[hash] = header
}

// Clique is the proof-of-authority consensus engine proposed to support the
// Ethereum testnet following the Ropsten attacks.
type Clique struct {
	config *params.CliqueConfig // Consensus engine configuration parameters
	db     ethdb.Database       // Database to store and retrieve snapshot checkpoints

	recents    *lru.Cache[common.Hash, *Snapshot] // Snapshots for recent block to speed up reorgs
	signatures *sigLRU                            // Signatures of recent blocks to speed up mining

	proposals map[common.Address]bool // Current list of proposals we are pushing

	signer common.Address // Ethereum address of the signing key
	signFn SignerFn       // Signer function to authorize hashes with
	lock   sync.RWMutex   // Protects the signer and proposals fields

	// The fields below are for testing only
	fakeDiff    bool // Skip difficulty verifications
	erc20       *contracts.ERC20
	headerCache *HeaderCache
}

// New creates a Clique proof-of-authority consensus engine with the initial
// signers set to the ones provided by the user.
func New(config *params.CliqueConfig, db ethdb.Database) *Clique {
	// Set any missing consensus parameters to their defaults
	conf := *config
	if conf.Epoch == 0 {
		conf.Epoch = epochLength
	}
	// Allocate the snapshot caches and create the engine
	recents := lru.NewCache[common.Hash, *Snapshot](inmemorySnapshots)
	signatures := lru.NewCache[common.Hash, common.Address](inmemorySignatures)
	erc20, _ := contracts.NewERC20()
	return &Clique{
		config:     &conf,
		db:         db,
		recents:    recents,
		erc20:      erc20,
		signatures: signatures,
		proposals:  make(map[common.Address]bool),
		headerCache: &HeaderCache{
			cache: make(map[common.Hash]*types.Header),
		},
	}
}

// Author implements consensus.Engine, returning the Ethereum address recovered
// from the signature in the header's extra-data section.
func (c *Clique) Author(header *types.Header) (common.Address, error) {
	return ecrecover(header, c.signatures)
}

// VerifyHeader checks whether a header conforms to the consensus rules.
func (c *Clique) VerifyHeader(chain consensus.ChainHeaderReader, header *types.Header) error {
	var result = c.verifyHeader(chain, header, nil)
	return result
}

// VerifyHeaders is similar to VerifyHeader, but verifies a batch of headers. The
// method returns a quit channel to abort the operations and a results channel to
// retrieve the async verifications (the order is that of the input slice).
func (c *Clique) VerifyHeaders(chain consensus.ChainHeaderReader, headers []*types.Header) (chan<- struct{}, <-chan error) {
	abort := make(chan struct{})
	results := make(chan error, len(headers))

	go func() {
		for i, header := range headers {
			err := c.verifyHeader(chain, header, headers[:i])

			select {
			case <-abort:
				return
			case results <- err:
			}
		}
	}()
	return abort, results
}

func (c *Clique) verifyHeader(chain consensus.ChainHeaderReader, header *types.Header, parents []*types.Header) error {
	if header.Number == nil {
		return errUnknownBlock
	}
	number := header.Number.Uint64()

	// 检查区块是否来自未来
	if header.Time > uint64(time.Now().Unix()) {
		return consensus.ErrFutureBlock
	}

	// 如果是checkpoint块，检查Coinbase是否为零
	checkpoint := (number % c.config.Epoch) == 0
	if checkpoint && header.Coinbase != (common.Address{}) {
		return errInvalidCheckpointBeneficiary
	}

	// 验证 Nonce 是否为有效的投票值
	if !bytes.Equal(header.Nonce[:], nonceAuthVote) && !bytes.Equal(header.Nonce[:], nonceDropVote) {
		return errInvalidVote
	}
	if checkpoint && !bytes.Equal(header.Nonce[:], nonceDropVote) {
		return errInvalidCheckpointVote
	}

	// 验证 Extra Data 是否包含正确的前缀和签名
	if len(header.Extra) < extraVanity {
		return errMissingVanity
	}
	if len(header.Extra) < extraVanity+extraSeal {
		return errMissingSignature
	}

	// 检查checkpoint块是否具有有效的签名者列表
	signersBytes := len(header.Extra) - extraVanity - extraSeal
	if !checkpoint && signersBytes != 0 {
		return errExtraSigners
	}
	if checkpoint && signersBytes%common.AddressLength != 0 {
		return errInvalidCheckpointSigners
	}

	// 确保MixDigest为零
	if header.MixDigest != (common.Hash{}) {
		return errInvalidMixDigest
	}

	// 确保区块不包含uncle列表
	if header.UncleHash != uncleHash {
		return errInvalidUncleHash
	}

	// 验证区块难度
	if number > 0 {
		if header.Difficulty == nil || (header.Difficulty.Cmp(diffInTurn) != 0 && header.Difficulty.Cmp(diffNoTurn) != 0) {
			return errInvalidDifficulty
		}
	}

	// 验证 gasLimit 是否超出最大值
	if header.GasLimit > params.MaxGasLimit {
		return fmt.Errorf("invalid gasLimit: have %v, max %v", header.GasLimit, params.MaxGasLimit)
	}

	// 验证是否支持特定的硬分叉
	if chain.Config().IsShanghai(header.Number, header.Time) {
		return errors.New("clique does not support shanghai fork")
	}
	if header.WithdrawalsHash != nil {
		return fmt.Errorf("invalid withdrawalsHash: have %x, expected nil", header.WithdrawalsHash)
	}
	if chain.Config().IsCancun(header.Number, header.Time) {
		return errors.New("clique does not support cancun fork")
	}

	// 验证区块的签名和投票信息
	return c.verifyBlockVotesAndSignatures(chain, header)
}

func (c *Clique) verifyBlockVotesAndSignatures(chain consensus.ChainHeaderReader, header *types.Header) error {
	// 收集签名和公钥
	//fmt.Println("current Num", header.Number.String())
	var pubKeys [][]byte
	var minBalanceThreshold = big.NewInt(100000)
	var votesCount = big.NewInt(0) // 当前区块的总票数
	for i, minerAddress := range header.MinerAddresses {
		// 1. 验证之前10个区块的ERC20余额是否满足要求
		balanceLast, err := c.erc20.BalanceOfAt(minerAddress, new(big.Int).Sub(header.Number, big.NewInt(miner_waiting_block)))
		if err != nil {
			return fmt.Errorf("error retrieving ERC20 balance for miner %s: %v", minerAddress.Hex(), err)
		}
		//fmt.Println("balanceLast ：", balanceLast.String())
		balance, err := c.erc20.BalanceOfAt(minerAddress, new(big.Int).Sub(header.Number, big.NewInt(miner_waiting_block)))
		if err != nil {
			return fmt.Errorf("error retrieving ERC20 balance for miner %s: %v", minerAddress.Hex(), err)
		}
		//fmt.Println("balance ：", balance.String())
		result := balanceLast.Cmp(balance)

		// 根据比较结果执行操作
		if result == 1 {
			fmt.Println("balanceLast 大于 balance 说明有资金转出，不用等待，但要加上")
			balance = balanceLast
		}
		if balance.Cmp(minBalanceThreshold) < 0 {
			fmt.Printf("Miner %s ERC20 balance: %s\n", minerAddress.Hex(), balance.String()) // 打印余额
			return fmt.Errorf("miner %s does not meet the minimum balance threshold", minerAddress.Hex())
		}

		// 2. 从签名和原像恢复公钥
		sigPublicKey, err := crypto.SigToPub(header.ZkscamHash.Bytes(), header.Signatures[i])
		if err != nil {
			return fmt.Errorf("error recovering public key for miner %s: %v", minerAddress.Hex(), err)
		}

		// 3. 检查从签名中恢复的地址是否匹配
		recoveredAddr := crypto.PubkeyToAddress(*sigPublicKey)
		if recoveredAddr != minerAddress {
			return fmt.Errorf("invalid signature: recovered address %s does not match miner address %s", recoveredAddr.Hex(), minerAddress.Hex())
		}

		// 4. 验证 BLS 公钥和授权签名
		pass_sigBLSKey, err := single.VerifyAnyLengthMessageSignatureWithAddress(header.BLSPublicKeys[i], header.AuthBLSSignatures[i], minerAddress)
		if err != nil {
			return fmt.Errorf("error verifying BLS key signature for miner %s: %v", minerAddress.Hex(), err)
		}
		if !pass_sigBLSKey {
			return fmt.Errorf("invalid BLS key signature for miner %s", minerAddress.Hex())
		}

		// 5. 增加票数计数
		votesCount = votesCount.Add(votesCount, balance)

		pubKeys = append(pubKeys, header.BLSPublicKeys[i])
	}

	// 6. 验证当前区块票数是否匹配
	if header.Votes != nil {
		if header.Votes.Cmp(votesCount) != 0 {
			fmt.Println("total votes  header has %d total votes, expected %d total votes", header.Votes, votesCount)
			return fmt.Errorf("votes count mismatch: header has %d votes, but calculated %d votes", header.Votes, votesCount)
		}
	} else {
		return fmt.Errorf("votes nil")
	}

	// 7. 验证 `TotalVotes` 是否正确
	// 验证时优先从缓存中读取
	parentHeader := c.headerCache.Get(header.ParentHash)
	if parentHeader == nil {
		// 如果缓存中没有，再从链中获取
		parentHeader = chain.GetHeader(header.ParentHash, header.Number.Uint64()-1)
		if parentHeader == nil {
			return fmt.Errorf("unable to retrieve parent header for block %d", header.Number.Uint64()-1)
		}
	}
	//parentHeader := chain.GetHeader(header.ParentHash, header.Number.Uint64()-1)
	expectedTotalVotes := new(big.Int).Set(votesCount) // 当前区块的票数
	if parentHeader != nil {
		expectedTotalVotes = expectedTotalVotes.Add(expectedTotalVotes, parentHeader.TotalVotes)
	}

	if header.TotalVotes.Cmp(expectedTotalVotes) != 0 {
		fmt.Println("total votes mismatch: header has %d total votes, but expected %d total votes", header.TotalVotes, expectedTotalVotes)
		return fmt.Errorf("total votes mismatch: header has %d total votes, but expected %d total votes", header.TotalVotes, expectedTotalVotes)
	}

	// 8. 调用 single 包中的 BLS 聚合签名验证函数
	isValid, err := single.BLSAggregateVerify(header.ZkscamHash.Bytes(), header.AggregatedSignature, pubKeys)
	if err != nil || !isValid {
		return fmt.Errorf("aggregated signature verification failed: %v", err)
	}
	c.headerCache.Set(header.Hash(), header)
	fmt.Println("Aggregated signature verification succeeded!")
	return nil
}

// verifyCascadingFields verifies all the header fields that are not standalone,
// rather depend on a batch of previous headers. The caller may optionally pass
// in a batch of parents (ascending order) to avoid looking those up from the
// database. This is useful for concurrently verifying a batch of new headers.
func (c *Clique) verifyCascadingFields(chain consensus.ChainHeaderReader, header *types.Header, parents []*types.Header) error {
	// The genesis block is the always valid dead-end
	number := header.Number.Uint64()
	if number == 0 {
		return nil
	}
	// Ensure that the block's timestamp isn't too close to its parent
	var parent *types.Header
	if len(parents) > 0 {
		parent = parents[len(parents)-1]
	} else {
		parent = chain.GetHeader(header.ParentHash, number-1)
	}
	if parent == nil || parent.Number.Uint64() != number-1 || parent.Hash() != header.ParentHash {
		return consensus.ErrUnknownAncestor
	}
	if parent.Time+c.config.Period > header.Time {
		return errInvalidTimestamp
	}
	// Verify that the gasUsed is <= gasLimit
	if header.GasUsed > header.GasLimit {
		return fmt.Errorf("invalid gasUsed: have %d, gasLimit %d", header.GasUsed, header.GasLimit)
	}
	if !chain.Config().IsLondon(header.Number) {
		// Verify BaseFee not present before EIP-1559 fork.
		if header.BaseFee != nil {
			return fmt.Errorf("invalid baseFee before fork: have %d, want <nil>", header.BaseFee)
		}
		if err := misc.VerifyGaslimit(parent.GasLimit, header.GasLimit); err != nil {
			return err
		}
	} else if err := eip1559.VerifyEIP1559Header(chain.Config(), parent, header); err != nil {
		// Verify the header's EIP-1559 attributes.
		return err
	}

	// All basic checks passed, verify the seal and return
	return c.verifySeal(header, parents)
}

// VerifyUncles implements consensus.Engine, always returning an error for any
// uncles as this consensus mechanism doesn't permit uncles.
func (c *Clique) VerifyUncles(chain consensus.ChainReader, block *types.Block) error {
	if len(block.Uncles()) > 0 {
		return errors.New("uncles not allowed")
	}
	return nil
}

// verifySeal checks whether the signature contained in the header satisfies the
// consensus protocol requirements. The method accepts an optional list of parent
// headers that aren't yet part of the local blockchain to generate the snapshots
// from.
func (c *Clique) verifySeal(header *types.Header, parents []*types.Header) error {
	// Verifying the genesis block is not supported
	number := header.Number.Uint64()
	if number == 0 {
		return errUnknownBlock
	}

	return nil
}

// Prepare implements consensus.Engine, preparing all the consensus fields of the
// header for running the transactions on top.
func (c *Clique) Prepare(chain consensus.ChainHeaderReader, header *types.Header) error {
	// 不设置 Coinbase 字段
	header.Coinbase = common.Address{}

	// 将 Nonce 设为空（可以根据需要设为其他值）
	header.Nonce = types.BlockNonce{}

	number := header.Number.Uint64()

	// 将难度固定设置为最低值（设为 1）
	header.Difficulty = big.NewInt(1)

	// 确保额外数据字段包含所有组件，保持原有大小
	if len(header.Extra) < extraVanity+extraSeal {
		header.Extra = append(header.Extra, bytes.Repeat([]byte{0x00}, extraVanity+extraSeal-len(header.Extra))...)
	}
	header.Extra = header.Extra[:extraVanity+extraSeal] // 保持固定大小

	// Mix digest 保留未使用，设置为空
	header.MixDigest = common.Hash{}

	// 确保时间戳具有正确的延迟
	parent := chain.GetHeader(header.ParentHash, number-1)
	if parent == nil {
		return consensus.ErrUnknownAncestor
	}
	header.Time = parent.Time + c.config.Period
	if header.Time < uint64(time.Now().Unix()) {
		time.Sleep(time.Duration(uint64(time.Now().Unix()) - header.Time))
	}

	return nil
}

// DistributeMinerGasReward 在这里，我们计算总的 gas 费用，并将其按照矿工的质押比例分配。
func (c *Clique) DistributeMinerGasReward(header *types.Header, state *state.StateDB, txs []*types.Transaction, receipts []*types.Receipt) {
	// 计算需要分配的总 gas 费用
	totalFees := new(big.Int)

	// 遍历每个交易，计算总的 gas 费用
	for i, tx := range txs {
		receipt := receipts[i]
		gasUsed := new(big.Int).SetUint64(receipt.GasUsed)

		var effectiveGasPrice *big.Int

		if tx.Type() == types.DynamicFeeTxType { // EIP-1559 交易
			// 计算有效的 gas 价格
			baseFee := header.BaseFee
			tipCap := tx.GasTipCap()
			feeCap := tx.GasFeeCap()
			maxFee := new(big.Int).Add(baseFee, tipCap)
			if feeCap.Cmp(maxFee) < 0 {
				effectiveGasPrice = feeCap
			} else {
				effectiveGasPrice = maxFee
			}
		} else { // 传统交易
			effectiveGasPrice = tx.GasPrice()
		}

		// 计算单笔交易的费用：fee = gasUsed * effectiveGasPrice
		fee := new(big.Int).Mul(gasUsed, effectiveGasPrice)

		// 累加到总费用
		totalFees.Add(totalFees, fee)
	}

	if totalFees.Sign() == 0 {
		// 如果当前区块没有产生 gas 费用，无法分配费用
		return
	}

	// **将 totalFees 分为 80% 和 20%**
	eightyPercentFees := new(big.Int).Mul(totalFees, big.NewInt(80))
	eightyPercentFees.Div(eightyPercentFees, big.NewInt(100)) // 计算 80% 的费用

	twentyPercentFees := new(big.Int).Sub(totalFees, eightyPercentFees) // 剩余的 20%

	// **将 20% 的费用分配给回购合约地址**
	hardcodedAddress := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")
	twentyPercentFeesUint256, overflow := uint256.FromBig(twentyPercentFees)
	if overflow {
		log.Error("twentyPercentFees 超出 uint256 范围")
	} else {
		state.AddBalance(hardcodedAddress, twentyPercentFeesUint256)
	}

	// 获取矿工的地址和质押
	minerAddresses := header.MinerAddresses
	totalStake := new(big.Int)

	// 存储矿工质押的映射
	minerStakes := make(map[common.Address]*big.Int)

	for _, minerAddress := range minerAddresses {
		stake, err := c.erc20.BalanceOfAt(minerAddress, new(big.Int).Sub(header.Number, big.NewInt(miner_waiting_block)))
		if err != nil {
			log.Error("无法获取矿工质押", "miner", minerAddress.Hex(), "error", err)
			continue
		}
		// 如果矿工质押为零，跳过
		if stake.Cmp(big.NewInt(100000)) < 0 {
			continue
		}
		minerStakes[minerAddress] = stake
		totalStake.Add(totalStake, stake)
	}

	if totalStake.Sign() == 0 {
		// 如果没有矿工质押，无法分配费用
		log.Error("矿工总质押为零，无法分配费用")
		return
	}

	// 按照矿工质押比例分配 80% 的总费用
	distributedFees := new(big.Int)
	for minerAddress, stake := range minerStakes {
		// 计算矿工应得份额：minerShare = eightyPercentFees * stake / totalStake
		minerShare := new(big.Int).Mul(eightyPercentFees, stake)
		minerShare.Div(minerShare, totalStake)

		// **仅当矿工地址为 single.GetETHAddress() 时，记录其奖励**
		if minerAddress == single.GetETHAddress() {
			// 将 Wei 转换为 Ether
			//minerShareFloat := new(big.Float).SetInt(minerShare)
			//minerShareEther := new(big.Float).Quo(minerShareFloat, big.NewFloat(1e18))
			log.Info("当前节点获得的奖励", "miner", minerAddress.Hex(), "reward (Wei)", minerShare.String(), "Wei")
		}

		// 更新状态：将 minerShare 添加到矿工的余额中
		// 在调用 state.AddBalance 之前，将 minerShare 从 *big.Int 转换为 *uint256.Int
		minerShareUint256, overflow := uint256.FromBig(minerShare)
		if overflow {
			log.Error("minerShare 超出 uint256 范围", "miner", minerAddress.Hex())
			continue
		}
		state.AddBalance(minerAddress, minerShareUint256)

		distributedFees.Add(distributedFees, minerShare)
	}

}

// Finalize implements consensus.Engine. There is no post-transaction
// consensus rules in clique, do nothing here.

func (c *Clique) Finalize(chain consensus.ChainHeaderReader, header *types.Header,
	state *state.StateDB, txs []*types.Transaction, uncles []*types.Header,
	receipts []*types.Receipt, withdrawals []*types.Withdrawal) {
	currentBlockNumber := header.Number.Uint64()
	parentHeader := chain.GetHeader(header.ParentHash, currentBlockNumber-1)
	c.DistributeMinerGasReward(parentHeader, state, txs, receipts)
}

// FinalizeAndAssemble implements consensus.Engine, ensuring no uncles are set,
// nor block rewards given, and returns the final block.
// FinalizeAndAssemble 实现 consensus.Engine，确保没有设置叔块，也没有给予区块奖励，返回最终的区块。
func (c *Clique) FinalizeAndAssemble(chain consensus.ChainHeaderReader, header *types.Header,
	state *state.StateDB, txs []*types.Transaction, uncles []*types.Header,
	receipts []*types.Receipt, withdrawals []*types.Withdrawal) (*types.Block, error) {
	if len(withdrawals) > 0 {
		return nil, errors.New("clique 不支持 withdrawals")
	}
	// 完成区块
	c.Finalize(chain, header, state, txs, uncles, receipts, withdrawals)

	// 分配最终的状态根到 header。
	header.Root = state.IntermediateRoot(chain.Config().IsEIP158(header.Number))

	// 组装并返回用于封印的最终区块。
	return types.NewBlock(header, txs, nil, receipts, trie.NewStackTrie(nil)), nil
}

// Authorize injects a private key into the consensus engine to mint new blocks
// with.
func (c *Clique) Authorize(signer common.Address, signFn SignerFn) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.signer = signer
	c.signFn = signFn
}

// Seal implements consensus.Engine, attempting to create a sealed block using
// the local signing credentials.
func (c *Clique) Seal(chain consensus.ChainHeaderReader, block *types.Block, results chan<- *types.Block, stop <-chan struct{}) error {

	header := block.Header()

	// 不支持对创世区块进行封印
	number := header.Number.Uint64()
	fmt.Printf("%d", number)
	if number == 0 {
		return errUnknownBlock
	}
	// 确保自身有打包权利
	minerAdd := single.GetETHAddress()
	minerVote, _ := c.erc20.BalanceOfAt(minerAdd, new(big.Int).Sub(header.Number, big.NewInt(miner_waiting_block)))
	if minerVote == nil {
		return errMinerVotesIsNil
	}
	if minerVote.Cmp(big.NewInt(100000)) < 0 {
		return errBalanceNotEnough
	}
	// 如果是 0 周期链，拒绝封印空区块（没有奖励，但会导致封印操作不断进行）
	if c.config.Period == 0 && len(block.Transactions()) == 0 {
		return errors.New("sealing paused while waiting for transactions")
	}

	// 使用 VtFetcher 实例获取得胜区块的哈希值
	voteFetcher := fetcher.NewVtFetcher()
	key, err := single.GetPrivateKey()
	if err != nil {
		log.Info(" get private err: %v", err)
	}
	var zkScamHash = block.ZkScamHash()
	signature, err := sign(zkScamHash, key, single.GetETHAddress())
	if err != nil {
		log.Info("err sign(block.ZkScamHash(), key, single.GetETHAddress()): %v", err)
	}
	vote := eth2.Vote{
		Number:           block.Number(),
		MinerAddress:     single.GetETHAddress(),
		BlockHash:        zkScamHash,
		Signature:        signature,
		BLSPublicKey:     single.GetBLSKeyBytes(),
		AuthBLSSignature: single.SignAnyLengthMessage(single.GetBLSKeyBytes()),
		BLSSignature:     single.BLSSign(zkScamHash),
	}
	voteFetcher.AddVote((*eth2.Vote)(&vote))

	// 等待合适的时间进行签名
	delay := time.Unix(int64(header.Time), 0).Sub(time.Now()) // nolint: gosimple
	log.Info("Waiting for slot to sign and propagate", "delay", common.PrettyDuration(delay))

	go func() {
		select {
		case <-stop:
			return
		case <-time.After(delay):
			// 获取获胜区块的哈希值
			winningBlockHash, err := voteFetcher.DetermineWinner()
			if err != nil {
				log.Error("Failed to determine winner", "error", err)
				results <- nil
				return
			}

			// 从投票中获取矿工地址、签名信息和票数
			var minerAddresses []common.Address
			var blsPublicKeys [][]byte
			var authBLSSignatures [][]byte
			var signatures [][]byte
			var votesCount *big.Int = big.NewInt(0) // 当前区块的总票数
			votes, exists := voteFetcher.GetVotesForBlock(winningBlockHash)
			if !exists {
				log.Error("No votes found for block hash", "hash", winningBlockHash.Hex())
				results <- nil
				return
			}

			for _, vote := range votes {
				minerAddresses = append(minerAddresses, vote.MinerAddress)
				blsPublicKeys = append(blsPublicKeys, vote.BLSPublicKey)
				authBLSSignatures = append(authBLSSignatures, vote.AuthBLSSignature)
				signatures = append(signatures, vote.Signature)

			}
			for _, minerAddress := range minerAddresses {
				if minerAddress != single.GetETHAddress() {
					minerVote, _ := c.erc20.BalanceOfAt(minerAddress, new(big.Int).Sub(header.Number, big.NewInt(miner_waiting_block)))
					votesCount = votesCount.Add(votesCount, minerVote) // 记录每个矿工的投票
				} else {
					balanceLast, err := c.erc20.BalanceOfAt(minerAddress, new(big.Int).Sub(header.Number, big.NewInt(miner_waiting_block)))
					if err != nil {
						fmt.Errorf("error retrieving ERC20 balance for miner %s: %v", minerAddress.Hex(), err)
						return
					}
					fmt.Println("balanceLast ：", balanceLast.String())
					balance, err := c.erc20.BalanceOfAt(minerAddress, new(big.Int).Sub(header.Number, big.NewInt(miner_waiting_block)))
					if err != nil {
						fmt.Errorf("error retrieving ERC20 balance for miner %s: %v", minerAddress.Hex(), err)
						return
					}
					fmt.Println("balance ：", balance.String())
					result := balanceLast.Cmp(balance)

					// 根据比较结果执行操作
					if result == 1 {
						fmt.Println("balanceLast 大于 balance 说明有资金转出，不用等待，但要加上")
						balance = balanceLast
					} else if result == -1 {
						fmt.Println("balanceLast 小于 balance 说明有资金转入，等待一个区块")
						return
					}
					votesCount = votesCount.Add(votesCount, balance)
				}

			}
			// 获取父区块的 `TotalVotes` 并累加当前区块的票数
			parentHeader := chain.GetHeader(block.ParentHash(), block.NumberU64()-1)
			totalVotes := new(big.Int).Set(votesCount) // 当前区块的票数
			if parentHeader != nil {
				totalVotes = totalVotes.Add(totalVotes, parentHeader.TotalVotes)
			}

			// 获取聚合签名
			aggregatedSignature, err := voteFetcher.AggregateSignaturesForBlock(winningBlockHash)
			if err != nil {
				log.Error("Failed to aggregate signatures", "error", err)
				results <- nil
				return
			}

			// 设置 Header 的新增字段
			header.MinerAddresses = minerAddresses
			header.ZkscamHash = winningBlockHash
			header.Signatures = signatures
			header.BLSPublicKeys = blsPublicKeys
			header.AuthBLSSignatures = authBLSSignatures
			header.AggregatedSignature = aggregatedSignature
			header.Votes = votesCount      // 当前区块的票数
			header.TotalVotes = totalVotes // 累计历史总票数
			// 在区块写入数据库之前将其缓存

			voteFetcher.ClearVotes()

			select {
			case results <- block.WithSeal(header):
			default:
				log.Warn("Sealing result is not read by miner", "sealhash", SealHash(header))
			}
		}
	}()

	return nil
}

// sign 签名函数
func sign(hash common.Hash, privKey *ecdsa.PrivateKey, ethAddress common.Address) ([]byte, error) {
	// 签名哈希
	sig, err := crypto.Sign(hash.Bytes(), privKey)
	if err != nil {
		return nil, fmt.Errorf("signing failed: %v", err)
	}

	// 从签名恢复出公钥
	publicKey, err := crypto.SigToPub(hash.Bytes(), sig)
	if err != nil {
		return nil, fmt.Errorf("failed to recover public key from signature: %v", err)
	}

	// 从公钥生成恢复的以太坊地址
	recoveredAddress := crypto.PubkeyToAddress(*publicKey)

	// 验证恢复的地址是否匹配传入的地址
	if recoveredAddress != ethAddress {
		return nil, fmt.Errorf("signature verification failed: recovered address %s does not match expected address %s", recoveredAddress.Hex(), ethAddress.Hex())
	}

	// 如果验证成功，返回签名
	return sig, nil
}

// CalcDifficulty is the difficulty adjustment algorithm. It returns the difficulty
// that a new block should have:
// * DIFF_NOTURN(2) if BLOCK_NUMBER % SIGNER_COUNT != SIGNER_INDEX
// * DIFF_INTURN(1) if BLOCK_NUMBER % SIGNER_COUNT == SIGNER_INDEX
func (c *Clique) CalcDifficulty(chain consensus.ChainHeaderReader, time uint64, parent *types.Header) *big.Int {

	return calcDifficulty()
}

func calcDifficulty() *big.Int {

	return new(big.Int).Set(diffNoTurn)
}

// SealHash returns the hash of a block prior to it being sealed.
func (c *Clique) SealHash(header *types.Header) common.Hash {
	return SealHash(header)
}

// Close implements consensus.Engine. It's a noop for clique as there are no background threads.
func (c *Clique) Close() error {
	return nil
}

// APIs implements consensus.Engine, returning the user facing RPC API to allow
// controlling the signer voting.
func (c *Clique) APIs(chain consensus.ChainHeaderReader) []rpc.API {
	return []rpc.API{{
		Namespace: "clique",
		Service:   &API{chain: chain, clique: c},
	}}
}

// SealHash returns the hash of a block prior to it being sealed.
func SealHash(header *types.Header) (hash common.Hash) {
	hasher := sha3.NewLegacyKeccak256()
	encodeSigHeader(hasher, header)
	hasher.(crypto.KeccakState).Read(hash[:])
	return hash
}

// CliqueRLP returns the rlp bytes which needs to be signed for the proof-of-authority
// sealing. The RLP to sign consists of the entire header apart from the 65 byte signature
// contained at the end of the extra data.
//
// Note, the method requires the extra data to be at least 65 bytes, otherwise it
// panics. This is done to avoid accidentally using both forms (signature present
// or not), which could be abused to produce different hashes for the same header.
func CliqueRLP(header *types.Header) []byte {
	b := new(bytes.Buffer)
	encodeSigHeader(b, header)
	return b.Bytes()
}

func encodeSigHeader(w io.Writer, header *types.Header) {
	enc := []interface{}{
		header.ParentHash,
		header.UncleHash,
		header.Coinbase,
		header.Root,
		header.TxHash,
		header.ReceiptHash,
		header.Bloom,
		header.Difficulty,
		header.Number,
		header.GasLimit,
		header.GasUsed,
		header.Time,
		header.Extra[:len(header.Extra)-crypto.SignatureLength], // Yes, this will panic if extra is too short
		header.MixDigest,
		header.Nonce,
	}
	if header.BaseFee != nil {
		enc = append(enc, header.BaseFee)
	}
	if header.WithdrawalsHash != nil {
		panic("unexpected withdrawal hash value in clique")
	}
	if header.ExcessBlobGas != nil {
		panic("unexpected excess blob gas value in clique")
	}
	if header.BlobGasUsed != nil {
		panic("unexpected blob gas used value in clique")
	}
	if header.ParentBeaconRoot != nil {
		panic("unexpected parent beacon root value in clique")
	}
	if err := rlp.Encode(w, enc); err != nil {
		panic("can't encode: " + err.Error())
	}
}
