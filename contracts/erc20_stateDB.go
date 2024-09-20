package contracts

//
//import (
//	"github.com/ethereum/go-ethereum/common"
//	"github.com/ethereum/go-ethereum/core/state"
//	"github.com/ethereum/go-ethereum/crypto"
//	"log"
//	"math/big"
//	"sync"
//)
//
//// ERC20 represents a module for retrieving ERC20 balances
//type ERC20 struct {
//	StateDB *state.StateDB // StateDB 由外部赋值
//}
//
//var (
//	instance *ERC20
//	once     sync.Once
//)
//var tokenAddress = common.HexToAddress("0x4b75210419009994c7f856f0b5c5b79750dbed22")
//
//// NewERC20 creates or returns the single instance of ERC20 without any parameters
//func NewERC20() *ERC20 {
//	once.Do(func() {
//		instance = &ERC20{}
//	})
//	return instance
//}
//
//// BalanceOf retrieves the balance of the ERC20 token for a specific address from the local stateDB
//func (erc20 *ERC20) BalanceOf(accountAddress common.Address) (*big.Int, error) {
//	if erc20.StateDB == nil {
//		log.Printf("StateDB is not initialized")
//		return nil, nil
//	}
//
//	// ERC20 balanceOf 函数实际上是从合约存储中读取账户余额，因此我们需要从合约存储槽位读取余额
//	// 计算存储槽位（ERC20 合约中的存储结构根据账户地址计算余额槽位）
//	balanceSlot := calculateERC20BalanceSlot(accountAddress)
//
//	// 从 stateDB 中读取该地址的余额
//	balanceBytes := erc20.StateDB.GetState(tokenAddress, balanceSlot)
//	balance := new(big.Int).SetBytes(balanceBytes.Bytes())
//
//	return balance, nil
//}
//
//// BalanceOfAt retrieves the balance of the ERC20 token for a specific address at a specific block number
//// 在重放的状态中，通过指定的区块高度获取余额，这里的 stateDB 通常应该和具体的区块高度关联
//func (erc20 *ERC20) BalanceOfAt(accountAddress common.Address, blockNumber *big.Int) (*big.Int, error) {
//	// 在重放模式下，stateDB 代表的是某个特定区块的状态
//	return erc20.BalanceOf(accountAddress) // 基于重放时的 stateDB
//}
//
//// calculateERC20BalanceSlot calculates the storage slot for an account's ERC20 balance based on its address.
//// ERC20 合约中的余额存储位置是通过 Keccak256(账户地址) 计算得到的。
//func calculateERC20BalanceSlot(accountAddress common.Address) common.Hash {
//	return common.BytesToHash(crypto.Keccak256(common.LeftPadBytes(accountAddress.Bytes(), 32)))
//}
