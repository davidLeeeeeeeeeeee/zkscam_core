package singleton

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing/bn256"
	"go.dedis.ch/kyber/v3/sign/bls"
)

var (
	instance          *ecdsa.PrivateKey
	address           common.Address
	mu                sync.Mutex
	initialized       bool
	errNotInitialized = errors.New("private key is not initialized")
)

// New initializes the singleton instance by reading the private key and address from a file.
// The file should contain the private key on the first line and the address on the second line.
func New() (*ecdsa.PrivateKey, common.Address, error) {
	mu.Lock() // 使用互斥锁来确保并发安全
	defer mu.Unlock()

	// 如果已经成功初始化，直接返回实例
	if initialized {
		return instance, address, nil
	}

	// 执行初始化逻辑
	data, err := os.ReadFile("./miner_private_key.txt")
	if err != nil {
		return nil, common.Address{}, fmt.Errorf("failed to read file: %v", err)
	}

	// Split the file content into lines
	lines := splitLines(string(data))
	if len(lines) < 2 {
		return nil, common.Address{}, fmt.Errorf("file format is incorrect: expected private key and address")
	}

	// Parse the private key
	privateKeyBytes, err := hex.DecodeString(lines[0])
	if err != nil {
		return nil, common.Address{}, fmt.Errorf("failed to decode private key: %v", err)
	}
	instance, err = crypto.ToECDSA(privateKeyBytes)
	if err != nil {
		return nil, common.Address{}, fmt.Errorf("failed to parse private key: %v", err)
	}

	// Parse the address
	address = common.HexToAddress(lines[1])

	// 如果成功，标记为已初始化
	initialized = true

	return instance, address, nil
}

// Helper function to split string by newlines and return non-empty lines.
func splitLines(data string) []string {
	lines := []string{}
	for _, line := range strings.Split(data, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

// SetPrivateKey 设置单例的私钥值
func SetPrivateKey(key *ecdsa.PrivateKey) {
	mu.Lock()
	defer mu.Unlock()
	instance = key
}

// GetPrivateKey 获取单例的私钥值
func GetPrivateKey() (*ecdsa.PrivateKey, error) {
	mu.Lock()
	defer mu.Unlock()
	if instance == nil {
		return nil, errNotInitialized
	}
	return instance, nil
}

// GetBLSPrivateKey 通过ECDSA私钥生成BLS私钥
func GetBLSPrivateKey() (kyber.Scalar, error) {
	mu.Lock()
	defer mu.Unlock()
	if instance == nil {
		return nil, errNotInitialized
	}

	// 对ECDSA私钥的D值进行哈希，生成BLS私钥
	hash := sha256.Sum256(instance.D.Bytes())
	suite := bn256.NewSuite()
	blsPrivateKey := suite.G2().Scalar().SetBytes(hash[:])

	return blsPrivateKey, nil
}

// GetBLSPublicKey 获取与BLS私钥对应的BLS公钥
func GetBLSPublicKey() (kyber.Point, error) {
	blsPrivateKey, err := GetBLSPrivateKey()
	if err != nil {
		return nil, err
	}

	suite := bn256.NewSuite()
	blsPublicKey := suite.G2().Point().Mul(blsPrivateKey, nil)

	return blsPublicKey, nil
}

// BLSSign 对消息进行BLS签名
func BLSSign(message common.Hash) []byte {
	blsPrivateKey, err := GetBLSPrivateKey()
	if err != nil {
		return nil
	}

	suite := bn256.NewSuite()
	signature, err := bls.Sign(suite, blsPrivateKey, message.Bytes())
	if err != nil {
		return nil
	}

	return signature
}

// BLSVerify 验证BLS签名
func BLSVerify(message []byte, signature []byte, pubKey []byte) (bool, error) {
	blsPublicKey, err := UnmarshalBLSKeyBytes(pubKey)
	if err != nil {
		return false, err
	}

	suite := bn256.NewSuite()
	err = bls.Verify(suite, blsPublicKey, message, signature)
	if err != nil {
		return false, err
	}

	return true, nil
}

// BLSAggregateVerify 验证BLS聚合签名
func BLSAggregateVerify(message []byte, aggregatedSignature []byte, pubKeys [][]byte) (bool, error) {
	// 初始化 BLS 套件
	suite := bn256.NewSuite()

	// 解析每个 BLS 公钥
	var publicKeys []kyber.Point
	for _, pubKeyBytes := range pubKeys {
		blsPublicKey, err := UnmarshalBLSKeyBytes(pubKeyBytes)
		if err != nil {
			return false, fmt.Errorf("failed to unmarshal BLS public key: %v", err)
		}
		publicKeys = append(publicKeys, blsPublicKey)
	}

	// 聚合公钥
	aggregatedPublicKey := bls.AggregatePublicKeys(suite, publicKeys...)

	// 验证聚合签名
	err := bls.Verify(suite, aggregatedPublicKey, message, aggregatedSignature)
	if err != nil {
		return false, fmt.Errorf("aggregated signature verification failed: %v", err)
	}

	// 返回验证成功
	return true, nil
}

// GetBLSKeyBytes 序列化BLS公钥
func GetBLSKeyBytes() []byte {
	blsPublicKey, err := GetBLSPublicKey()
	if err != nil {
		return nil
	}
	blsKey, err := blsPublicKey.MarshalBinary()
	if err != nil {
		return nil
	}
	return blsKey
}

// UnmarshalBLSKeyBytes 反序列化BLS公钥
func UnmarshalBLSKeyBytes(blsKeyBytes []byte) (kyber.Point, error) {
	suite := bn256.NewSuite()
	blsPublicKey := suite.G2().Point()
	err := blsPublicKey.UnmarshalBinary(blsKeyBytes)
	if err != nil {
		return nil, err
	}
	return blsPublicKey, nil
}

// GetETHAddress 返回与当前ECDSA私钥对应的以太坊地址
func GetETHAddress() common.Address {
	mu.Lock()
	defer mu.Unlock()
	if instance == nil {
		return common.Address{}
	}

	// 使用Geth的crypto库来获取ECDSA公钥并生成以太坊地址
	pubKey := instance.Public()
	pubKeyECDSA, ok := pubKey.(*ecdsa.PublicKey)
	if !ok {
		return common.Address{}
	}

	// 使用Geth的crypto库生成以太坊地址
	address := crypto.PubkeyToAddress(*pubKeyECDSA)

	return address
}

// SignAnyLengthMessage 使用ETH私钥对任意长度的数据进行签名
func SignAnyLengthMessage(message []byte) []byte {
	privateKey, err := GetPrivateKey()
	if err != nil {
		return nil
	}

	// 对消息进行哈希处理 (使用 SHA-256)
	hash := sha256.Sum256(message)

	// 使用以太坊的 crypto 库签名哈希值，生成 65 字节的签名
	signature, err := crypto.Sign(hash[:], privateKey)
	if err != nil {
		return nil
	}

	return signature
}

// VerifyAnyLengthMessageSignatureWithAddress 验证任意长度消息的签名
func VerifyAnyLengthMessageSignatureWithAddress(message []byte, signature []byte, address common.Address) (bool, error) {
	// 对消息进行哈希处理 (使用 SHA-256)
	hash := sha256.Sum256(message)

	// ECDSA签名的前32字节是r值，后32字节是s值
	if len(signature) != crypto.SignatureLength {
		return false, errors.New("signature length is incorrect")
	}

	// 使用加密库恢复公钥
	pubKey, err := crypto.SigToPub(hash[:], signature)
	if err != nil {
		return false, err
	}

	// 从恢复的公钥生成以太坊地址
	recoveredAddr := crypto.PubkeyToAddress(*pubKey)

	// 比较恢复的地址和传入的地址是否相同
	if recoveredAddr == address {
		return true, nil
	}

	return false, errors.New("signature verification failed: address mismatch")
}
