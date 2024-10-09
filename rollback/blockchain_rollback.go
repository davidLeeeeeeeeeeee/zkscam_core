package rollback

//import (
//	"fmt"
//	"log"
//	"math/big"
//	"path/filepath"
//
//	"github.com/ethereum/go-ethereum/common"
//	"github.com/ethereum/go-ethereum/consensus"
//	"github.com/ethereum/go-ethereum/consensus/ethash"
//	"github.com/ethereum/go-ethereum/core"
//	"github.com/ethereum/go-ethereum/core/rawdb"
//	"github.com/ethereum/go-ethereum/core/state"
//	"github.com/ethereum/go-ethereum/core/vm"
//	"github.com/ethereum/go-ethereum/ethdb"
//	"github.com/ethereum/go-ethereum/ethdb/leveldb"
//	"github.com/ethereum/go-ethereum/params"
//	"github.com/ethereum/go-ethereum/triedb"
//)
//
//func main() {
//	// 替换为您的以太坊数据目录
//	datadir := "/path/to/ethereum/data/geth"
//
//	// 打开数据库
//	chaindataPath := filepath.Join(datadir, "chaindata")
//	freezerPath := filepath.Join(datadir, "geth", "chaindata")
//	db, err := rawdb.NewLevelDBDatabaseWithFreezer(chaindataPath, 1024, 1024, freezerPath, "", false)
//	if err != nil {
//		log.Fatalf("无法打开数据库: %v", err)
//	}
//	defer db.Close()
//
//	// 加载创世区块配置
//	genesis := &core.Genesis{}
//	genesisHash := rawdb.ReadCanonicalHash(db, 0)
//	if genesisHash == (common.Hash{}) {
//		log.Fatalf("数据库中没有找到创世区块")
//	} else {
//		genesisBlock := rawdb.ReadBlock(db, genesisHash, 0)
//		if genesisBlock == nil {
//			log.Fatalf("无法读取创世区块")
//		}
//		chainConfig := rawdb.ReadChainConfig(db, genesisHash)
//		if chainConfig == nil {
//			log.Fatalf("无法读取链配置")
//		}
//		genesis = &core.Genesis{
//			Config:     chainConfig,
//			Difficulty: genesisBlock.Difficulty(),
//			GasLimit:   genesisBlock.GasLimit(),
//			ExtraData:  genesisBlock.Extra(),
//			Timestamp:  genesisBlock.Time(),
//			Mixhash:    genesisBlock.MixDigest(),
//			Coinbase:   genesisBlock.Coinbase(),
//			Number:     genesisBlock.Number(),
//			GasUsed:    genesisBlock.GasUsed(),
//			ParentHash: genesisBlock.ParentHash(),
//			BaseFee:    genesisBlock.BaseFee(),
//		}
//	}
//
//	// 初始化 Trie 数据库
//	trieDB := triedb.New(db)
//
//	// 初始化区块链配置
//	cacheConfig := &core.CacheConfig{
//		TrieCleanLimit:      256,
//		TrieDirtyLimit:      256,
//		TrieTimeLimit:       5 * 60 * 1e9, // 5 分钟
//		SnapshotLimit:       256,
//		SnapshotWait:        true,
//		TrieDirtyDisabled:   false,
//		TrieCleanNoPrefetch: false,
//		StateScheme:         rawdb.HashScheme, // 或者使用 rawdb.PathScheme，取决于您的节点配置
//	}
//
//	// 初始化共识引擎（如果您使用的是其他共识算法，请替换为相应的引擎）
//	engine := ethash.New(ethash.Config{
//		PowMode: ethash.ModeNormal,
//	})
//
//	// 初始化区块链
//	blockchain, err := core.NewBlockChain(db, cacheConfig, genesis, nil, engine, vm.Config{}, nil, nil)
//	if err != nil {
//		log.Fatalf("无法创建区块链: %v", err)
//	}
//	defer blockchain.Stop()
//
//	// 设置要回滚到的目标区块号（替换为您的目标区块号）
//	rollbackTo := big.NewInt(1000000)
//
//	// 调用 Rollback 函数
//	err = blockchain.Rollback(rollbackTo)
//	if err != nil {
//		log.Fatalf("回滚区块链失败: %v", err)
//	}
//
//	fmt.Println("回滚成功")
//}
