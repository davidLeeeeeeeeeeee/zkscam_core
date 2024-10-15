package fetcher

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/contracts"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	eth2 "github.com/ethereum/go-ethereum/eth/protocols/eth"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	single "github.com/ethereum/go-ethereum/singleton"
	"go.dedis.ch/kyber/v3/pairing/bn256"
	"go.dedis.ch/kyber/v3/sign/bls"
	"math/big"
	"sync"
	"time"
)

// VtFetcher manages the fetching process=
type VtFetcher struct {
	mu             sync.Mutex
	votes          map[common.Hash][]*eth2.Vote
	notifyData     map[common.Hash]notifyEntry
	erc20          *contracts.ERC20
	rpcClient      *rpc.Client
	winningBlk     common.Hash
	voteTracker    map[string]struct{}
	broadcastVotes func(votes eth2.Votes)
	blockFetcher   *BlockFetcher // 新增的字段
}

// notifyEntry defines the structure for storing notification data
type notifyEntry struct {
	PeerID        string
	Number        uint64
	Timestamp     time.Time
	RequestHeader headerRequesterFn
	RequestBodies bodyRequesterFn
}

// Singleton instance
var (
	instance *VtFetcher
	once     sync.Once
)

// NewVtFetcher creates or returns the singleton instance of VtFetcher
func NewVtFetcher(optionalArgs ...interface{}) *VtFetcher {
	once.Do(func() {
		erc20, _ := contracts.NewERC20()
		var (
			callback     func(votes eth2.Votes)
			blockFetcher *BlockFetcher
		)

		// 解析可选参数
		for _, arg := range optionalArgs {
			switch v := arg.(type) {
			case func(votes eth2.Votes):
				callback = v
			case *BlockFetcher:
				blockFetcher = v
			}
		}

		// 如果没有传入回调函数，则使用空函数作为默认值
		if callback == nil {
			callback = func(votes eth2.Votes) {}
		}
		// 调用 New 方法获取私钥和地址
		_, _, err := single.New()
		if err != nil {
			fmt.Println("Failed to initialize: %v", err)
		}
		// 初始化 VtFetcher 实例
		instance = &VtFetcher{
			votes:          make(map[common.Hash][]*eth2.Vote),
			notifyData:     make(map[common.Hash]notifyEntry),
			voteTracker:    make(map[string]struct{}),
			erc20:          erc20,
			rpcClient:      erc20.Client,
			broadcastVotes: callback,
			blockFetcher:   blockFetcher, // 使用传入的 blockFetcher
		}
	})
	return instance
}

// AddNotifyData adds a new entry to the notifyData map
func (f *VtFetcher) AddNotifyData(hash common.Hash, peerID string, number uint64, timestamp time.Time, headerFetcher headerRequesterFn, bodyFetcher bodyRequesterFn) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.notifyData[hash] = notifyEntry{
		PeerID:        peerID,
		Number:        number,
		Timestamp:     timestamp,
		RequestHeader: headerFetcher,
		RequestBodies: bodyFetcher,
	}
}

// GetNotifyData retrieves an entry from the notifyData map
func (f *VtFetcher) GetNotifyData(hash common.Hash) (notifyEntry, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()

	entry, exists := f.notifyData[hash]
	return entry, exists
}

// ClearNotifyData clears the notifyData map
func (f *VtFetcher) ClearNotifyData() {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.notifyData = make(map[common.Hash]notifyEntry)
}

// ReceiveVotes processes multiple votes
func (f *VtFetcher) ReceiveVotes(votesData eth2.Votes) error {

	for _, vote := range votesData.Votes {
		sigPublicKey, err := crypto.SigToPub(vote.BlockHash.Bytes(), vote.Signature)
		if err != nil {
			fmt.Println("Error recovering public key:", err)
			continue
		}
		// 说明收到了自己发出去的vote了
		if vote.MinerAddress == single.GetETHAddress() {
			continue
		}

		recoveredAddr := crypto.PubkeyToAddress(*sigPublicKey)
		if recoveredAddr != vote.MinerAddress {
			fmt.Println("Invalid signature: recovered address does not match miner address")
			continue
		}
		pass_sigBLSKey, err := single.VerifyAnyLengthMessageSignatureWithAddress(vote.BLSPublicKey, vote.AuthBLSSignature, vote.MinerAddress)
		if err != nil {
			fmt.Println("Error recovering BLSKey:", err)
			continue
		}
		if !pass_sigBLSKey {
			continue
		}
		pass_bls, err := single.BLSVerify(vote.BlockHash.Bytes(), vote.BLSSignature, vote.BLSPublicKey)
		if err != nil {
			fmt.Println("Error recovering BLSVerify:", err)
			continue
		}
		if !pass_bls {
			continue
		}

		if err != nil {
			fmt.Println("Error adding vote:", err)
			continue
		}
		var minBalanceThreshold = big.NewInt(100000)
		// 验证余额是否满足要求
		balance, err := f.erc20.BalanceOfMinus10(vote.MinerAddress)
		if err != nil {
			fmt.Println("Error retrieving ERC20 balance:", err)
			continue
		}
		if balance.Cmp(minBalanceThreshold) < 0 {
			fmt.Println("Miner does not meet the minimum balance threshold:", vote.MinerAddress.Hex())
			continue
		}
		err = f.AddVote((*eth2.Vote)(&vote))
	}
	return nil
}

// AddVote adds a new vote to the fetcher, ensuring no duplicates
func (f *VtFetcher) AddVote(vote *eth2.Vote) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	voteKey := fmt.Sprintf("%s-%s", vote.MinerAddress.Hex(), vote.BlockHash.Hex())

	if _, exists := f.voteTracker[voteKey]; exists {
		// 如果已经存在相同的vote，不再添加
		return nil
	}
	// 不存在时添加到字典
	f.voteTracker[voteKey] = struct{}{}
	f.votes[vote.BlockHash] = append(f.votes[vote.BlockHash], vote)
	//并广播
	votes := eth2.Votes{Votes: []eth2.Vote{*vote}} // 解引用 vote
	f.broadcastVotes(votes)
	return nil
}

// DetermineWinner determines the block with the highest total votes from qualified voters
func (f *VtFetcher) DetermineWinner() (common.Hash, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	// 定义一个固定的最小余额门槛，例如 10000
	minBalance := big.NewInt(100000)

	var maxVotes *big.Int = big.NewInt(0)
	var winningBlock common.Hash

	for blockHash, votes := range f.votes {
		totalVotes := big.NewInt(0)
		for _, vote := range votes {
			balance, err := f.erc20.BalanceOfMinus10(vote.MinerAddress)
			if err != nil {
				return common.Hash{}, err
			}
			fmt.Println("balance", balance, "MinerAddress", vote.MinerAddress)
			// 过滤掉余额小于 minBalance 的投票者
			if balance.Cmp(minBalance) >= 0 {
				totalVotes.Add(totalVotes, balance) // 将投票者的余额累加到总票数中
			}
		}
		fmt.Println("totalVotes", totalVotes)
		// 找出拥有最多有效投票的区块
		if totalVotes.Cmp(maxVotes) > 0 {
			maxVotes = totalVotes
			winningBlock = blockHash
		}
	}

	f.winningBlk = winningBlock
	return f.winningBlk, nil
}

// fetchBlockByHash fetches a block by its hash
func (f *VtFetcher) fetchBlockByHash(blockHash common.Hash) (*types.Block, error) {
	client := ethclient.NewClient(f.rpcClient)
	block, err := client.BlockByHash(context.Background(), blockHash)
	if err != nil {
		return nil, err
	}
	return block, nil
}

// ClearVotes clears the votes map and the vote tracker
func (f *VtFetcher) ClearVotes() {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.votes = make(map[common.Hash][]*eth2.Vote)
	f.voteTracker = make(map[string]struct{})
}

// AggregateSignaturesForBlock aggregates all BLS signatures for a specified block hash
func (f *VtFetcher) AggregateSignaturesForBlock(blockHash common.Hash) ([]byte, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	// 获取指定 blockHash 对应的所有 votes
	votes, exists := f.votes[blockHash]
	if !exists {
		return nil, fmt.Errorf("no votes found for block hash: %s", blockHash.Hex())
	}

	// 收集所有的 BLS 签名
	var signaturesOrder1 [][]byte
	for _, vote := range votes {
		signaturesOrder1 = append(signaturesOrder1, vote.BLSSignature)
	}
	suite := bn256.NewSuite()
	// 使用 bls.AggregateSignatures 方法聚合所有签名
	aggregatedSignature1, err := bls.AggregateSignatures(suite, signaturesOrder1...)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate signatures: %v", err)
	}

	return aggregatedSignature1, nil
}
func (f *VtFetcher) GetVotesForBlock(blockHash common.Hash) ([]*eth2.Vote, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()

	votes := f.votes[blockHash]
	if len(votes) == 0 {
		return nil, false
	}

	return votes, true
}
