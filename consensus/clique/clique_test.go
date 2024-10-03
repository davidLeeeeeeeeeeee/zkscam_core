package clique_test

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	single "github.com/ethereum/go-ethereum/singleton"
	"github.com/stretchr/testify/assert"
)

func decodeBase64(input string) []byte {
	data, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		panic(fmt.Sprintf("Failed to decode base64 string: %v", err))
	}
	return data
}

func mockBlockHeader() *types.Header {
	return &types.Header{
		MinerAddresses: []common.Address{
			common.HexToAddress("0xda56cc60cd2bdf454d164c92713422dc8be5190f"),
			common.HexToAddress("0xfe2a7e374320abe858c21310e533e169236e0f7e"),
			common.HexToAddress("0x6b65ab268a38107981eda2e1e64213c49a8540de"),
			common.HexToAddress("0xb9303cf61804f343d576e11074a7d8a93b882711"),
			common.HexToAddress("0x6559682d6f4e14fa650281d4f7407fad18533881"),
			common.HexToAddress("0xb96174af20d19f3954c8306606a035edc123cb5a"),
			common.HexToAddress("0x89e91694a2241eb6b8ae3546198d6f78b41ffc26"),
			common.HexToAddress("0xb1d5313d684998e960ded080679b536f65b6d8f2"),
			common.HexToAddress("0x6fea8531ab23300705573986b17331f2297246af"),
			common.HexToAddress("0x203ca20aed3b0718caa2c899bd3ea990322a106d"),
			common.HexToAddress("0xa9ef73c8badfd8526deb81973ff2ebb5f6f9f95c"),
			common.HexToAddress("0x239494ea3fd1494c742d28826adc848bfc41f90a"),
		},
		ZkscamHash: common.HexToHash("0xf0ed5140612b276bd7cb7b1f1536ab6dc3598b9e8449bfca4c414a82c6c1f576"),
		Signatures: [][]byte{
			decodeBase64("/X6SHy4Q7ysLf77Vw/j+9t9EQcPQ7WBPKX/PE5VNxqZpLo0T9SLeznRMWG7fdnW7e/y2l8NpZ35OKcKR/+NIvgA="),
			decodeBase64("5a2x0lTuVd5jvIdT1eG8IG5UQAzy7U7Jz2iFenerkJINl6ZeVjVYeSqp8fA/sQdwk+dMgd9xExHVw3A928rElAA="),
			decodeBase64("ir68RkUhPVK7XQgTqYvN7wLJSIbDHaC2EtvUvOxUFEIbak8KTQwBd/TRsKCsulVe3vLyScXgq9z+rv69l7wPxgA="),
			decodeBase64("N8fMSHkwlU4xTgneK3HZdqvklWtf4nNTLYY1TAOoF4EvhZSts4VcAAlbxic2FSASsdhm2K6z0pI8iScUkECr5wE="),
			decodeBase64("A7mNzmlN7fsGe/sfDs3yjA+GndBAHqzPt0vnp84AxdFFnbsF2ryFZytcKcX/XlMMwOQkZlgtBoRfGU2AKGsnsgE="),
			decodeBase64("ai3XjTQazczdiO7XumJe0z9nWtAEip9FIpI9ppAZJAE5kqlzc+4ZZFzBHB/2Kno0lX5JigGTfbSZ96p2LoAJrwA="),
			decodeBase64("38tY+IKlOLexE/juUfX9kTdgNLz5vzERDIIbAiFMDBA5VKY3DL/+Px7b5l7i5WvPqi5gn2x/3P9Z4ZBcv7TmcgA="),
			decodeBase64("drNlXA/i0G+tiHsjbAoZUADt4n6AgA/IMkRhAaFU6T4EWoID+rXI8QyJa6qhq51jfAOFjXZe1CGnUSHtqG8BDQE="),
			decodeBase64("BAcVCuEQa/lEq9O6PlmQc/FcHB5aVkxsmyQO7Cjj8wEGlANdq/3/TgTgt6D71XT5yLqYRHEjkntXPf4vwxaLVAA="),
			decodeBase64("V9NVM1tBsdZK6Mv/7QvGkSUadxeZxiV3TO+zvUhlDzxKyr2XTT3ZdWV8nIQjel8s9W42tJTIpoJHKGQJZAEULgE="),
			decodeBase64("DIF6e+8If6VucXPnMEf8VJL2IINHUYL0OScjF23ttg5SnQ3W76q1YsobypbptEmR3y0jpgEXVsCtOzNoS77y7AE="),
			decodeBase64("ODn1doEToJ2BI1YPuUOdTTSq6HsL5BzsyHezH16Z4xIY4kH4qx09EnDXF2QFql+P2DYgT6k7HQPqVGRGCXK+aAA="),
		},
		BLSPublicKeys: [][]byte{
			decodeBase64("WmWEwy96mMFV9FGqyx+fyjKlz45r7DA00EV6UIjT1XRkUGpweLw5zvnRydzoWQoNYGhk6GHCj9cnnOnt0xN17065Ha46ehyPvvWXbxVKOIhntIcFrC+UDIL8roI4fQs1i3yUn2LvxgjDAntouHN0ZymgoN5SMOHFSqgqF5x/5fg="),
			decodeBase64("WxvQxFCdCg3thlLvhs/1SNi7G4BlKFNgkQcLumvtQ5so0T24jNIxcx05XuX25oYc6zy53CbzG2pPuwRB0m3lzGXr07Q9pWw410BdrPlxTZUqzYWk6K33wV5ulT1sNU9EB/HzNcwWxersA+87337fApQhXs3FpZhA7ak2gfiYx6U="),
			decodeBase64("bn04GwMNAcECVEif3vPlTebJjAUwnXs/DhRV/vJ0w3MVJeW+fd7NWnIbRbQGdl+Z35wBu5Az9Y12LSBT0VkScoVs8Du00dt+aPEdcJCYS64j3dOHzsod9cxo7jnr2MLMP+hjIRV71Rm2SIYof13Bm5AZxPJ2YKauRqwOfY8TPCg="),
			decodeBase64("dS98JkQ1CPG8+cXLBUzkXGAU7jmfaghgo8FpdFhh7rkKBJz63ZmUj5K3wGFrwYh3VK4dobeHjConDwvEURVhR0SuVMiSTuXoQmDV9Ed0JMsoWf8gB0nyt1GE5+k670hPPPdPmxoeeKuUowk4tZKYMZe1zjRXYEE1FrMpHUNLyq4="),
			decodeBase64("JoutfJVfVbON/FOuRcwNbXsavskKlZCiOAwkzDM1NIVd0nnH7QkmO+c3UTIXTknKAXi7m6hT+LbZVqY6e2UmKxl0MAJ9r+bzIJ/K9ZvEGD42+bZLXn59OTze4Hh9UcU/ArMhuw3aAJGHel8wIsbR734xP4ewR/c+/7IVbiDPtXI="),
			decodeBase64("Oqsz2fF/eJmqbJITeCEfEsKO6/mlkxn/6uV5/nUBQQoKlakWJPlG0DpJj6KYfEeL7wbtVJdyNZUTFdTt4DE9/WKdnd8sR9UH+CcERiJqxMXG2VUt6eVnPvFBBMX4HvgkjnTV0fI18Znz/r8NR1p3l4QfwtiIxm0htaaqWkkm6L8="),
			decodeBase64("jKEVQKTz5aH7M5cjfhDJpWPtk+Tg2QOMKZ7tMME41V0dN38MwfoJ0EesshmYP3EOl3iQnnHSIPnLk8a2P5TFkhuzftdENV26UrGo9I5CVX8MeFaob8vpF4L2X0XrofUEGrq04RDb1KQaFYpO9qy4AbiCvMuQhhn6sakuBqid9T8="),
			decodeBase64("dMa7/rHZPmi+L+Hfvx5w7LbWkZwVmIdtRajblWXwjHgL4e3f8QKFzSPzbyOmhIxDv0RGc0tQ9rThBReHc/6VJo4Ajga2cMh+IcriqmFxKmrbZpAbww5i70dU3Pr6UlQpIZLuUPFzETqvOlrqZchcRA6C2eGRyfzGgKMvYdfT6bE="),
			decodeBase64("KMNYs9amqPOoivq+/fXxEo0T2cX/Qcb3PAwki9aYz9YzyLZYEFgGKO9A9NXD5d6De6G1TlY5PJlB1nThTLjT1S69yCYmOmValU1wCf2qdd5mrJ9evrYQsQhjMgZUmIunOsnlTNrXdXJ71AYM6iCCrRgA0S+IAfVhm+FNpj4v4k4="),
			decodeBase64("EhsvYvd7LlLKBEBSvluWOigLXpBbwhzLUVe+0gM/W6GEdCslj7lhbPJrDC0SohxaW3+n3ixpNxNmixcPGFYsFx9lyIQK7L80SbdY7JcZ4XTteMoqg7WAezvHAepwd49ZBIJonm5/sqORycvSiG3yw3Y04ir5XC/A0MSgM45LmTQ="),
			decodeBase64("WQQXL1N7g2sDYPEZYLSTJyZtvHtCJBbHAmddXnyjyzx8uckzf1kCgrjI5zb1R50DJ3oDlrc1U0HPniWEv0H1mV9WZ8qMMNTqIkbjH/kfY/UzKRzXGJkgRWevYtxjj8YDCrsO4A8NvflMq4weOE6nFFE3f6IXZr25rqASTQoQhsE="),
			decodeBase64("SPTSN/FsuZrVrjmZYosCGd7j/oj//vMB09A9QyeGno6BU6EDqrL4BI4sMIRg5L7YoW0X1WUCZyWtdlEmIfhRNl0KAjX21T4ITfvq2ta/KfONsicwXgqva5AVyWc0Y/2CP3s+TZwvYQY0LcloEgV+MT+K7cIyYdpj8ZDhq1j1a64="),
		},
		AuthBLSSignatures: [][]byte{
			decodeBase64("ZVCxjnIR5288B5kcdBl/xx14A602XwebmXpqVbrncEpx7MH/wu4ip5n0icIAFYoV4GxagFwJzeY6wA1I1HhIpgA="),
			decodeBase64("fo6nOZispIlF5NV03ZpgHD3eJroH1s7Th+Xxm/manbdKRIMR/8gjWNEY82m8ibzqbxXflzYiKAwg7dmd4/SRMgA="),
			decodeBase64("jsISlVtIu7zGn/DMPeQCY3OeARohhAeqxlMNHiblz3Nl4hKvqacmb8C44fS7O4uGwE4hDbHZcZsn3xKkpuSmUwE="),
			decodeBase64("UFGufQxTrjVlTZA5fYQa0P0QWBz9cIcc5Z0+kxzZU10D3AU5FWXixd1IpqX5hKpTkGdgeIA4OFcGZsUcweFkuAE="),
			decodeBase64("FVDm5p1BoZCPXflg0t5BreObxhJpERzm/L+b/5rcahQkAddpD4SYBIRaKy0JUn4V00mzy58HzwE6OpymreccqwE="),
			decodeBase64("OhvLMAgGAJrUZh0dHPI37Hj6vmx8LDFlQxyyBYzZ8e9P7ypcKCACjq8byJS6Wq8qFzbOnzHigx3A7XKQks/cCwE="),
			decodeBase64("JbxLWeJFF7r5ngBqYZHZZrEX0eFgXBXlTJCNf2IR97cjPCN4Wgu0v+VHf9eKWGtEiMbETtIgKbw9AFyDNT9WwgE="),
			decodeBase64("o0Ybn0F3lc9fv38dXF06LUUYIoZovu0ucd29fWBTkKlh2jgghhBrJ1E1E2+eSP/uM4Aj1JsI1oKfT5FRhGC/4wA="),
			decodeBase64("5eV0Qplwdslkom4Gv9fqnX28sdBtuBcVB01JsYXTHmU9NW0ACgSeApYQ/HnXCXQ+s36eq6SfDsDWoWAOVFtBfQE="),
			decodeBase64("iafwOxsSVOs4T1lHh+DDYQPeBqkqVClCxRKy/nGOsZ4tkGSynLMrKd80uyUBVj+lfRhRWubbyDRljacaWT0mmQE="),
			decodeBase64("MWYqZ+SE9dhDIpqv4HEI3SIE4V71aXRW364z2MPIXT1kj+Q1WT5Wx3g2gpc7/U4bOknnlEq4YC84cuDienC9ngA="),
			decodeBase64("gRJVwwebbLzlE1xHSBaNAHTn0uqO8innrUKUE+C8ztA6ClPifG5vowkI+Mx9GogVSJ7e8meL0AHADMfewANcUQE="),
		},
		AggregatedSignature: decodeBase64("jReglHCw9ymq5ley4zY0uUXXCYi5r4H6slbnlHCMc2Q8AyxT7YdNZXIj3wIZYHaXSZzfGKPWyNXoCiaCCJU7iw=="),
	}
}
func TestBLSVerification(t *testing.T) {
	// 模拟区块头
	header := mockBlockHeader()

	// 打印数组长度以确保它们匹配
	fmt.Printf("MinerAddresses length: %d\n", len(header.MinerAddresses))
	fmt.Printf("BLSPublicKeys length: %d\n", len(header.BLSPublicKeys))
	fmt.Printf("AuthBLSSignatures length: %d\n", len(header.AuthBLSSignatures))

	// 遍历矿工地址，逐个验证BLS签名
	for i, minerAddress := range header.MinerAddresses {
		if i >= len(header.BLSPublicKeys) || i >= len(header.AuthBLSSignatures) {
			t.Errorf("Index out of range: minerAddress[%d] exceeds length of BLSPublicKeys or AuthBLSSignatures", i)
			return
		}

		// 调试输出公钥、签名和地址
		fmt.Printf("\nMiner #%d:\n", i)
		fmt.Printf("Expected Miner Address: %s\n", minerAddress.Hex())
		fmt.Printf("BLS Public Key (Base64): %s\n", base64.StdEncoding.EncodeToString(header.BLSPublicKeys[i]))
		fmt.Printf("Auth BLS Signature (Base64): %s\n", base64.StdEncoding.EncodeToString(header.AuthBLSSignatures[i]))

		// 验证授权签名
		passSigBLSKey, err := single.VerifyAnyLengthMessageSignatureWithAddress(
			header.BLSPublicKeys[i],
			header.AuthBLSSignatures[i],
			minerAddress,
		)

		if err != nil {
			t.Errorf("Error verifying BLS key signature for miner %s: %v", minerAddress.Hex(), err)
		}

		assert.True(t, passSigBLSKey, fmt.Sprintf("Invalid BLS key signature for miner %s", minerAddress.Hex()))

		// **新增：验证每个矿工的单个签名**
		isValid, err := single.BLSVerify(header.ZkscamHash.Bytes(), header.Signatures[i], header.BLSPublicKeys[i])
		if err != nil || !isValid {
			t.Errorf("Individual BLS signature verification failed for miner %s: %v", minerAddress.Hex(), err)
		} else {
			fmt.Printf("Individual BLS signature verified for miner %s\n", minerAddress.Hex())
		}
	}

	// 聚合签名验证
	isValid, err := single.BLSAggregateVerify(header.ZkscamHash.Bytes(), header.AggregatedSignature, header.BLSPublicKeys)
	if err != nil {
		t.Errorf("Aggregated signature verification failed: %v", err)
	}
	assert.True(t, isValid, "Aggregated signature verification failed")
}
