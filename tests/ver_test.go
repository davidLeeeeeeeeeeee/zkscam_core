package tests

import (
	"fmt"
	"log"
	"testing"

	"encoding/base64"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func TestVer(t *testing.T) {
	// 示例数据
	zkscamHash := common.HexToHash("0x13c78fafa62ddd9a8f5271e4ca8770dccee79d74be04d28ad4a1e9f68e4d4c5f")          // 假设这个是你的ZkscamHash
	signatureBase64 := "ctP2d3xLY3B5+NBqaezLH69Op1hUc9FUHcpwx3pmOPZPfbncR5ymG4gOcMKmA/8KG9CUceZ9FvQTszRfEWY2CAE=" // 签名
	minerAddress := common.HexToAddress("0xfe2a7e374320abe858c21310e533e169236e0f7e")                             // 矿工地址

	// 将Base64签名转换为字节
	signature, err := base64.StdEncoding.DecodeString(signatureBase64)
	if err != nil {
		log.Fatalf("error decoding base64 signature: %v", err)
	}

	// 2. 验证签名是否匹配 (从签名中恢复公钥)
	sigPublicKey, err := crypto.SigToPub(zkscamHash.Bytes(), signature)
	if err != nil {
		log.Fatalf("error recovering public key: %v", err)
	}

	// 3. 检查从签名中恢复的地址是否匹配
	recoveredAddr := crypto.PubkeyToAddress(*sigPublicKey)
	if recoveredAddr != minerAddress {
		log.Fatalf("invalid signature: recovered address %s does not match miner address %s", recoveredAddr.Hex(), minerAddress.Hex())
	} else {
		fmt.Printf("Signature is valid! Recovered address: %s matches miner address: %s\n", recoveredAddr.Hex(), minerAddress.Hex())
	}
}
